package media

import (
	"audit-service/models"
	"audit-service/pkg/config"
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "crypto/sha256" // ensure SHA256 is registered

	"github.com/google/uuid"
)

// GDriveProvider implements MediaProvider using the Drive REST API directly (no SDK).
type GDriveProvider struct {
	cfg             *config.GDriveConfig
	privateKey      *rsa.PrivateKey
	clientEmail     string
	defaultFolderID string
	httpClient      *http.Client
}

// NewGDriveProvider creates a new GDrive provider that uses raw REST calls.
func NewGDriveProvider(cfg *config.GDriveConfig) (MediaProvider, error) {
	p := &GDriveProvider{
		cfg:             cfg,
		defaultFolderID: cfg.DefaultFolderID,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
	}

	if cfg.AuthMode == "service_account" {
		if err := p.loadServiceAccountKey(cfg); err != nil {
			return nil, fmt.Errorf("failed to load service account key: %w", err)
		}
	}

	return p, nil
}

// loadServiceAccountKey parses the service account credentials JSON and extracts the private key.
func (p *GDriveProvider) loadServiceAccountKey(cfg *config.GDriveConfig) error {
	var credsJSON string
	if cfg.CredentialsJSON != "" {
		credsJSON = cfg.CredentialsJSON
	} else if cfg.CredentialsJSONPath != "" {
		data, err := os.ReadFile(cfg.CredentialsJSONPath)
		if err != nil {
			return fmt.Errorf("failed to read credentials file from %s: %w", cfg.CredentialsJSONPath, err)
		}
		credsJSON = string(data)
	} else {
		return fmt.Errorf("credentials_json and credentials_json_path are empty; set credentials in config")
	}

	var creds struct {
		Type        string `json:"type"`
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}
	if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
		return fmt.Errorf("failed to parse credentials_json: %w", err)
	}

	block, _ := pem.Decode([]byte(creds.PrivateKey))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from private_key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not RSA")
	}

	p.privateKey = rsaKey
	p.clientEmail = creds.ClientEmail
	return nil
}

// getAccessToken mints a JWT and exchanges it for a Google OAuth2 access token.
func (p *GDriveProvider) getAccessToken(ctx context.Context) (string, error) {
	now := time.Now().Unix()

	// Build JWT header
	headerJSON, _ := json.Marshal(map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	})
	header := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Build JWT payload
	payloadJSON, _ := json.Marshal(map[string]interface{}{
		"iss":   p.clientEmail,
		"aud":   "https://oauth2.googleapis.com/token",
		"scope": "https://www.googleapis.com/auth/drive",
		"iat":   now,
		"exp":   now + 3600,
	})
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Sign header.payload with RSA-SHA256
	signingInput := header + "." + payload
	h := sha256.New()
	h.Write([]byte(signingInput))
	digest := h.Sum(nil)

	sig, err := rsa.SignPKCS1v15(rand.Reader, p.privateKey, crypto.SHA256, digest)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	sigEncoded := base64.RawURLEncoding.EncodeToString(sig)

	jwtToken := signingInput + "." + sigEncoded

	// Exchange JWT for access token
	body := "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Ajwt-bearer&assertion=" + jwtToken
	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange JWT for token: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("token error %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	return tokenResp.AccessToken, nil
}

// Upload attempts to upload to GDrive first, but falls back to local storage if GDrive is unconfigured or fails (e.g. quota limits).
func (p *GDriveProvider) Upload(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error) {
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Try GDrive upload
	attachment, err := p.uploadToGDrive(ctx, fileContent, fileName, folder)
	if err == nil {
		return attachment, nil
	}

	// Fallback to local upload
	fmt.Printf("[GDrive Fallback] Uploading to Google Drive failed (%v). Falling back to local storage.\n", err)
	return p.uploadLocal(fileContent, fileName, folder)
}

// getOrCreateFolderHierarchy traverses or creates the folder hierarchy in Google Drive
// pathParts: e.g. ["Auditsphere", "fieldwork", "ST-001"]
func (p *GDriveProvider) getOrCreateFolderHierarchy(ctx context.Context, accessToken string, pathParts []string) (string, error) {
	parentID := p.defaultFolderID

	for _, rawPart := range pathParts {
		part := strings.TrimSpace(rawPart)
		if part == "" {
			continue
		}

		// Search for an existing folder with this name under parentID
		query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", strings.ReplaceAll(part, "'", "\\'"))
		if parentID != "" {
			query += fmt.Sprintf(" and '%s' in parents", parentID)
		} else {
			query += " and 'root' in parents"
		}

		searchURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files?q=%s&fields=files(id,name)&spaces=drive", url.QueryEscape(query))
		req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := p.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to search folder '%s': %w", part, err)
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var searchResp struct {
			Files []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"files"`
		}
		_ = json.Unmarshal(respBody, &searchResp)

		if len(searchResp.Files) > 0 {
			parentID = searchResp.Files[0].ID
			continue
		}

		// Create folder if not found
		type createFolderMeta struct {
			Name     string   `json:"name"`
			MimeType string   `json:"mimeType"`
			Parents  []string `json:"parents,omitempty"`
		}
		folderMeta := createFolderMeta{
			Name:     part,
			MimeType: "application/vnd.google-apps.folder",
		}
		if parentID != "" {
			folderMeta.Parents = []string{parentID}
		}

		folderJSON, _ := json.Marshal(folderMeta)
		createReq, err := http.NewRequestWithContext(ctx, "POST", "https://www.googleapis.com/drive/v3/files?fields=id,name", bytes.NewReader(folderJSON))
		if err != nil {
			return "", err
		}
		createReq.Header.Set("Authorization", "Bearer "+accessToken)
		createReq.Header.Set("Content-Type", "application/json")

		createResp, err := p.httpClient.Do(createReq)
		if err != nil {
			return "", fmt.Errorf("failed to create folder '%s': %w", part, err)
		}
		createBody, _ := io.ReadAll(createResp.Body)
		createResp.Body.Close()

		var createdFolder struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(createBody, &createdFolder); err != nil || createdFolder.ID == "" {
			return "", fmt.Errorf("failed to parse created folder '%s' response: %s", part, string(createBody))
		}

		parentID = createdFolder.ID
	}

	return parentID, nil
}

// uploadToGDrive uploads file content to Google Drive.
func (p *GDriveProvider) uploadToGDrive(ctx context.Context, fileContent []byte, fileName string, folder string) (*models.MediaAttachment, error) {
	accessToken, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Resolve folder hierarchy if folder path is provided
	folderID := p.defaultFolderID
	cleanFolder := strings.Trim(folder, "/")
	if cleanFolder != "" && cleanFolder != "default" {
		parts := strings.Split(cleanFolder, "/")
		if len(parts) > 0 {
			resolvedID, err := p.getOrCreateFolderHierarchy(ctx, accessToken, parts)
			if err == nil && resolvedID != "" {
				folderID = resolvedID
			} else {
				fmt.Printf("[GDrive] Folder hierarchy resolution warning (%v), using defaultFolderID\n", err)
			}
		}
	}

	type fileMetadata struct {
		Name    string   `json:"name"`
		Parents []string `json:"parents,omitempty"`
	}

	meta := fileMetadata{Name: fileName}
	if folderID != "" {
		meta.Parents = []string{folderID}
	}

	metaJSON, _ := json.Marshal(meta)

	// Build multipart body
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Part 1: metadata (application/json)
	metaHeader := textproto.MIMEHeader{}
	metaHeader.Set("Content-Type", "application/json; charset=UTF-8")
	metaPart, err := w.CreatePart(metaHeader)
	if err != nil {
		return nil, err
	}
	metaPart.Write(metaJSON)

	// Part 2: file content
	fileHeader := textproto.MIMEHeader{}
	fileHeader.Set("Content-Type", "application/octet-stream")
	filePart, err := w.CreatePart(fileHeader)
	if err != nil {
		return nil, err
	}
	filePart.Write(fileContent)
	w.Close()

	// POST to Drive multipart upload endpoint
	uploadURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&fields=id,name,size,mimeType,webViewLink"
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "multipart/related; boundary="+w.Boundary())

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to drive: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("drive upload failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var driveFile struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Size        int64  `json:"size,string"`
		MimeType    string `json:"mimeType"`
		WebViewLink string `json:"webViewLink"`
	}
	if err := json.Unmarshal(respBody, &driveFile); err != nil {
		return nil, fmt.Errorf("failed to parse drive response: %w", err)
	}

	// Make the file publicly readable so the link works without login
	_ = p.makePublic(ctx, accessToken, driveFile.ID)

	return &models.MediaAttachment{
		FileID:     driveFile.ID,
		FileName:   driveFile.Name,
		FilePath:   driveFile.WebViewLink,
		FileSize:   driveFile.Size,
		FileType:   driveFile.MimeType,
		UploadedAt: time.Now(),
	}, nil
}

// uploadLocal saves the file content in the local filesystem preserving folder structure.
func (p *GDriveProvider) uploadLocal(fileContent []byte, fileName string, folder string) (*models.MediaAttachment, error) {
	cleanFolder := strings.Trim(folder, "/")
	if cleanFolder == "" {
		cleanFolder = "Auditsphere/general"
	}
	uploadsDir := filepath.Join("uploads", cleanFolder)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local uploads directory: %w", err)
	}

	fileID := uuid.New().String()
	filePath := filepath.Join(uploadsDir, fileName)

	// Write to file
	if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write local file: %w", err)
	}

	relPath := "/" + filepath.ToSlash(filepath.Join("uploads", cleanFolder, fileName))

	return &models.MediaAttachment{
		FileID:     fileID,
		FileName:   fileName,
		FilePath:   relPath,
		FileSize:   int64(len(fileContent)),
		FileType:   "application/octet-stream",
		UploadedAt: time.Now(),
	}, nil
}

// makePublic grants "anyone with the link can view" permission on a Drive file.
func (p *GDriveProvider) makePublic(ctx context.Context, accessToken, fileID string) error {
	permBody, _ := json.Marshal(map[string]string{
		"role": "reader",
		"type": "anyone",
	})
	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s/permissions", fileID),
		bytes.NewReader(permBody),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Delete removes a file from Google Drive by its file ID, or locally if it's stored locally.
func (p *GDriveProvider) Delete(ctx context.Context, fileID string) error {
	// Check if this matches a local file
	uploadsDir := "uploads"
	files, err := os.ReadDir(uploadsDir)
	if err == nil {
		for _, f := range files {
			if strings.HasPrefix(f.Name(), fileID) {
				localPath := filepath.Join(uploadsDir, f.Name())
				return os.Remove(localPath)
			}
		}
	}

	// Otherwise, call GDrive Delete API
	accessToken, err := p.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE",
		fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s", fileID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file from drive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("drive delete failed (HTTP %d): %s", resp.StatusCode, string(body))
	}
	return nil
}
