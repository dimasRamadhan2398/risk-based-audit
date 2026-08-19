package media

import (
	"audit-service/models"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type LocalProvider struct {
}

func NewLocalProvider() MediaProvider {
	return &LocalProvider{}
}

func (p *LocalProvider) Upload(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error) {
	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	// Generate safe file name
	safeFileName := fmt.Sprintf("%d-%s", time.Now().Unix(), fileName)
	filePath := filepath.Join(uploadDir, safeFileName)

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	// Copy data to the file
	written, err := io.Copy(out, file)
	if err != nil {
		return nil, err
	}

	return &models.MediaAttachment{
		FileID:     "local-" + safeFileName,
		FileName:   fileName,
		FilePath:   "/uploads/" + safeFileName,
		FileSize:   written,
		FileType:   "application/octet-stream",
		UploadedAt: time.Now(),
	}, nil
}

func (p *LocalProvider) Delete(ctx context.Context, fileID string) error {
	return nil
}
