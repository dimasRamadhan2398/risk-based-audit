package media

import (
	"audit-service/models"
	"audit-service/pkg/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GDriveProvider struct {
	service         *drive.Service
	defaultFolderID string
}

func NewGDriveProvider(cfg *config.GDriveConfig) (MediaProvider, error) {
	ctx := context.Background()
	var opts []option.ClientOption

	if cfg.AuthMode == "service_account" {
		if cfg.CredentialsJSON != "" {
			var creds map[string]interface{}
			if err := json.Unmarshal([]byte(cfg.CredentialsJSON), &creds); err == nil {
				if pk, ok := creds["private_key"].(string); ok {
					// Clean literal \n sequences into real newlines inside the parsed private key string
					creds["private_key"] = strings.ReplaceAll(pk, `\n`, "\n")
				}
				if cleanedJSON, err := json.Marshal(creds); err == nil {
					opts = append(opts, option.WithCredentialsJSON(cleanedJSON))
				} else {
					opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
				}
			} else {
				opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
			}
		} else if cfg.CredentialsJSONPath != "" {
			opts = append(opts, option.WithCredentialsFile(cfg.CredentialsJSONPath))
		}
	} else if cfg.AuthMode == "oauth2" {
		oauthCfg := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     google.Endpoint,
		}
		token := &oauth2.Token{
			RefreshToken: cfg.RefreshToken,
			Expiry:       time.Now(),
		}
		tokenSource := oauthCfg.TokenSource(ctx, token)
		opts = append(opts, option.WithTokenSource(tokenSource))
	}

	// Always ensure drive scopes are added
	opts = append(opts, option.WithScopes(drive.DriveScope, drive.DriveFileScope))

	srv, err := drive.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create drive client: %w", err)
	}

	return &GDriveProvider{
		service:         srv,
		defaultFolderID: cfg.DefaultFolderID,
	}, nil
}

func (p *GDriveProvider) Upload(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error) {
	folderID := p.defaultFolderID
	if folder != "" && folder != "audit" && folder != "default" {
		folderID = folder
	}

	f := &drive.File{
		Name: fileName,
	}
	if folderID != "" {
		f.Parents = []string{folderID}
	}

	res, err := p.service.Files.Create(f).Media(file).Context(ctx).Fields("id", "name", "size", "mimeType", "webViewLink").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to drive: %w", err)
	}

	return &models.MediaAttachment{
		FileID:     res.Id,
		FileName:   res.Name,
		FilePath:   res.WebViewLink,
		FileSize:   res.Size,
		FileType:   res.MimeType,
		UploadedAt: time.Now(),
	}, nil
}

func (p *GDriveProvider) Delete(ctx context.Context, fileID string) error {
	err := p.service.Files.Delete(fileID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete file from drive: %w", err)
	}
	return nil
}
