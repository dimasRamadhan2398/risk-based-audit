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
	} else {
		return fmt.Errorf("credentials_json is empty; set credentials_json in config")
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
	return p.uploadLocal(fileContent, fileName)
}

// uploadToGDrive uploads file content to Google Drive.
func (p *GDriveProvider) uploadToGDrive(ctx context.Context, fileContent []byte, fileName string, folder string) (*models.MediaAttachment, error) {
	accessToken, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Build file metadata
	folderID := p.defaultFolderID
	if folder != "" && folder != "audit" && folder != "default" {
		folderID = folder
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

// uploadLocal saves the file content in the local filesystem.
func (p *GDriveProvider) uploadLocal(fileContent []byte, fileName string) (*models.MediaAttachment, error) {
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local uploads directory: %w", err)
	}

	// Generate a unique file name/ID to prevent overwrites
	fileID := uuid.New().String()
	safeFileName := fmt.Sprintf("%s_%s", fileID, fileName)
	filePath := filepath.Join(uploadsDir, safeFileName)

	// Write to file
	if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write local file: %w", err)
	}

	// Return local attachment metadata
	// Note: We use http://localhost:8080/api/v1/media/download/ as the public URL via Gateway proxy
	downloadURL := fmt.Sprintf("http://localhost:8080/api/v1/media/download/%s", safeFileName)

	return &models.MediaAttachment{
		FileID:     fileID,
		FileName:   fileName,
		FilePath:   downloadURL,
		FileSize:   int64(len(fileContent)),
		FileType:   "application/octet-stream", // Fallback mime type
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
