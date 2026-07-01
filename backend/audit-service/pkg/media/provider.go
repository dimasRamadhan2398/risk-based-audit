package media

import (
	"audit-service/models"
	"context"
	"io"
)

type MediaProvider interface {
	Upload(ctx context.Context, file io.Reader, fileName string, folder string) (*models.MediaAttachment, error)
	Delete(ctx context.Context, fileID string) error
}
