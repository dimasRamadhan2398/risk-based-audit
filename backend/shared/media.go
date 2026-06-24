package shared

import (
	"time"
)

// MediaAttachment represents a file attachment metadata
type MediaAttachment struct {
	FileName   string    `json:"fileName"`
	FilePath   string    `json:"filePath"`
	FileSize   int64     `json:"fileSize"`
	FileType   string    `json:"fileType"`
	UploadedAt time.Time `json:"uploadedAt"`
}
