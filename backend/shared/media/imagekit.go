package media

import (
	"context"
	"io"
	"time"

	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/rb-audit/shared"
)

type ImageKitConfig struct {
	PublicKey   string
	PrivateKey  string
	UrlEndpoint string
}

type ImageKitProvider struct {
	client *imagekit.ImageKit
}

func NewImageKitProvider(cfg *ImageKitConfig) *ImageKitProvider {
	client := imagekit.NewClient(
		option.WithPublicKey(cfg.PublicKey),
		option.WithPrivateKey(cfg.PrivateKey),
		option.WithUrlEndpoint(cfg.UrlEndpoint),
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
