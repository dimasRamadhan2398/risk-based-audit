package media

import (
	"audit-service/models"
	"audit-service/pkg/upload"
	"context"
	"io"
	"path/filepath"
	"time"
)

type LocalProvider struct {
}

func NewLocalProvider() MediaProvider {
	return &LocalProvider{}
}

// storeLocal writes an upload under uploads/<folder>/ with a server-generated
// name (O_EXCL, never overwriting). folder is confined to the uploads tree
// again here, whatever the caller did.
func storeLocal(data []byte, fileName, folder string) (*models.MediaAttachment, error) {
	cleanFolder, e := upload.ConfineFolder(folder)
	if e != nil {
		return nil, e
	}
	dir := filepath.Join(upload.UploadsRoot, filepath.FromSlash(cleanFolder))
	stored, _, err := upload.Store(dir, upload.DisplayName(fileName), data)
	if err != nil {
		return nil, err
	}
	relPath := "/" + filepath.ToSlash(filepath.Join("uploads", filepath.FromSlash(cleanFolder), stored))
	return &models.MediaAttachment{
		FileID: "local-" + stored,
		// The stored name, not the client name: the frontend downloads media by
		// file name (/media/download/<fileName>), and stored names are unique.
		FileName:   stored,
		FilePath:   relPath,
		FileSize:   int64(len(data)),
		FileType:   "application/octet-stream",
		UploadedAt: time.Now(),
	}, nil
}

func (p *LocalProvider) Upload(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error) {
	data, e := upload.ReadAll(file, upload.DefaultMaxBytes)
	if e != nil {
		return nil, e
	}
	return storeLocal(data, fileName, folder)
}

func (p *LocalProvider) Delete(ctx context.Context, fileID string) error {
	return nil
}
