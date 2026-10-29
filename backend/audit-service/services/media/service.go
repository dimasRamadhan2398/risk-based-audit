package media

import (
	"audit-service/models"
	"audit-service/pkg/media"
	"audit-service/pkg/upload"
	"audit-service/services/base"
	"bytes"
	"context"
	"io"
)

type MediaServiceInterface interface {
	UploadFile(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error)
	DeleteFile(ctx context.Context, fileID string) error
}

type MediaService struct {
	*base.BaseService
	provider media.MediaProvider
}

func NewMediaService(provider media.MediaProvider) MediaServiceInterface {
	return &MediaService{
		BaseService: base.NewBaseService(),
		provider:    provider,
	}
}

// UploadFile validates the upload (file name, extension allowlist, magic
// bytes, 10 MB limit, malware scan) and confines folder to the uploads tree
// before handing it to the provider. Rejections are *upload.Error values that
// carry the HTTP status to send.
func (s *MediaService) UploadFile(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error) {
	data, e := upload.ReadAll(file, upload.DefaultMaxBytes)
	if e != nil {
		return nil, e
	}
	name, e := upload.Validate(ctx, fileName, data, upload.DefaultMaxBytes, upload.DocumentExtensions)
	if e != nil {
		return nil, e
	}
	cleanFolder, e := upload.ConfineFolder(folder)
	if e != nil {
		return nil, e
	}
	return s.provider.Upload(ctx, bytes.NewReader(data), name, cleanFolder)
}

func (s *MediaService) DeleteFile(ctx context.Context, fileID string) error {
	return s.provider.Delete(ctx, fileID)
}
