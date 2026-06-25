package media

import (
	"audit-service/pkg/config"
	"context"
	"github.com/rb-audit/shared"
	"io"
	"time"

	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
)

type ImageKitProvider struct {
	client imagekit.Client
}

func NewImageKitProvider(cfg *config.ImageKitConfig) *ImageKitProvider {
	client := imagekit.NewClient(
		option.WithPrivateKey(cfg.PrivateKey),
		option.WithBaseURL(cfg.UrlEndpoint),
	)

	return &ImageKitProvider{
		client: client,
	}
}

func (p *ImageKitProvider) Upload(ctx context.Context, file io.Reader, fileName string, folder string) (*shared.MediaAttachment, error) {
	resp, err := p.client.Files.Upload(ctx, imagekit.FileUploadParams{
		File:     file,
		FileName: fileName,
		Folder:   imagekit.String(folder),
	})
	if err != nil {
		return nil, err
	}

	return &shared.MediaAttachment{
		FileID:     resp.FileID,
		FileName:   resp.Name,
		FilePath:   resp.URL,
		FileSize:   int64(resp.Size),
		FileType:   resp.FileType,
		UploadedAt: time.Now(),
	}, nil
}

func (p *ImageKitProvider) Delete(ctx context.Context, fileID string) error {
	return p.client.Files.Delete(ctx, fileID)
}
