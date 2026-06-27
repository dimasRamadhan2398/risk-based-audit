package media

import (
	"audit-service/models"
	"audit-service/pkg/media"
	"audit-service/services/base"
	"context"
	"io"
)

type MediaServiceInterface interface {
	UploadFile(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error)
	DeleteFile(ctx context.Context, fileID string) error
}

type MediaService struct {
	*base.BaseService
	provider *media.ImageKitProvider
}

func NewMediaService(provider *media.ImageKitProvider) MediaServiceInterface {
	return &MediaService{
		BaseService: base.NewBaseService(),
		provider:    provider,
	}
}

func (s *MediaService) UploadFile(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error) {
	return s.provider.Upload(ctx, file, fileName, folder)
}

func (s *MediaService) DeleteFile(ctx context.Context, fileID string) error {
	return s.provider.Delete(ctx, fileID)
}
