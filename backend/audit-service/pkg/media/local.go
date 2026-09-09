package media

import (
	"audit-service/models"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalProvider struct {
}

func NewLocalProvider() MediaProvider {
	return &LocalProvider{}
}

func (p *LocalProvider) Upload(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error) {
	cleanFolder := strings.Trim(folder, "/")
	if cleanFolder == "" {
		cleanFolder = "Auditsphere/general"
	}
	// Create uploads directory with subfolder structure if it doesn't exist
	uploadDir := filepath.Join("./uploads", cleanFolder)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(uploadDir, fileName)

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

	relPath := "/" + filepath.ToSlash(filepath.Join("uploads", cleanFolder, fileName))

	return &models.MediaAttachment{
		FileID:     "local-" + fileName,
		FileName:   fileName,
		FilePath:   relPath,
		FileSize:   written,
		FileType:   "application/octet-stream",
		UploadedAt: time.Now(),
	}, nil
}

func (p *LocalProvider) Delete(ctx context.Context, fileID string) error {
	return nil
}
