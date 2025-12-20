package opik

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// AttachmentType represents the type of attachment.
type AttachmentType string

const (
	AttachmentTypeImage    AttachmentType = "image"
	AttachmentTypeAudio    AttachmentType = "audio"
	AttachmentTypeVideo    AttachmentType = "video"
	AttachmentTypeDocument AttachmentType = "document"
	AttachmentTypeText     AttachmentType = "text"
	AttachmentTypeOther    AttachmentType = "other"
)

// Attachment represents a file attachment for a trace or span.
type Attachment struct {
	// Name is the display name of the attachment.
	Name string
	// Type is the attachment type (image, audio, etc.).
	Type AttachmentType
	// MimeType is the MIME type of the attachment.
	MimeType string
	// Data is the raw attachment data.
	Data []byte
	// URL is an optional URL for externally hosted attachments.
	URL string
	// Base64 is the base64-encoded data (for embedding).
	Base64 string
}

// NewAttachmentFromFile creates an attachment from a file path.
func NewAttachmentFromFile(path string) (*Attachment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	name := filepath.Base(path)
	mimeType := detectMimeType(path, data)
	attachType := mimeTypeToAttachmentType(mimeType)

	return &Attachment{
		Name:     name,
		Type:     attachType,
		MimeType: mimeType,
		Data:     data,
	}, nil
}

// NewAttachmentFromBytes creates an attachment from raw bytes.
func NewAttachmentFromBytes(name string, data []byte, mimeType string) *Attachment {
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	attachType := mimeTypeToAttachmentType(mimeType)

	return &Attachment{
		Name:     name,
		Type:     attachType,
		MimeType: mimeType,
		Data:     data,
	}
}

// NewAttachmentFromReader creates an attachment from an io.Reader.
func NewAttachmentFromReader(name string, r io.Reader, mimeType string) (*Attachment, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return NewAttachmentFromBytes(name, data, mimeType), nil
}

// NewAttachmentFromURL creates an attachment from a URL.
func NewAttachmentFromURL(name string, url string, attachType AttachmentType) *Attachment {
	return &Attachment{
		Name: name,
		Type: attachType,
		URL:  url,
	}
}

// NewTextAttachment creates a text attachment.
func NewTextAttachment(name, content string) *Attachment {
	return &Attachment{
		Name:     name,
		Type:     AttachmentTypeText,
		MimeType: "text/plain",
		Data:     []byte(content),
	}
}

// NewImageAttachment creates an image attachment from bytes.
func NewImageAttachment(name string, data []byte, mimeType string) *Attachment {
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	return &Attachment{
		Name:     name,
		Type:     AttachmentTypeImage,
		MimeType: mimeType,
		Data:     data,
	}
}

// ToBase64 returns the base64-encoded attachment data.
func (a *Attachment) ToBase64() string {
	if a.Base64 != "" {
		return a.Base64
	}
	if len(a.Data) > 0 {
		return base64.StdEncoding.EncodeToString(a.Data)
	}
	return ""
}

// ToDataURL returns a data URL for the attachment.
func (a *Attachment) ToDataURL() string {
	if a.URL != "" {
		return a.URL
	}
	mimeType := a.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, a.ToBase64())
}

// Size returns the size of the attachment in bytes.
func (a *Attachment) Size() int {
	return len(a.Data)
}

// IsEmbeddable returns true if the attachment is small enough to embed inline.
func (a *Attachment) IsEmbeddable(maxSize int) bool {
	return len(a.Data) <= maxSize
}

// detectMimeType detects the MIME type from file extension and content.
func detectMimeType(path string, data []byte) string {
	// Try extension first
	ext := filepath.Ext(path)
	if ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType
		}
	}

	// Fall back to content detection
	return http.DetectContentType(data)
}

// mimeTypeToAttachmentType converts a MIME type to an attachment type.
func mimeTypeToAttachmentType(mimeType string) AttachmentType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return AttachmentTypeImage
	case strings.HasPrefix(mimeType, "audio/"):
		return AttachmentTypeAudio
	case strings.HasPrefix(mimeType, "video/"):
		return AttachmentTypeVideo
	case strings.HasPrefix(mimeType, "text/"):
		return AttachmentTypeText
	case strings.Contains(mimeType, "pdf"),
		strings.Contains(mimeType, "document"),
		strings.Contains(mimeType, "word"),
		strings.Contains(mimeType, "excel"),
		strings.Contains(mimeType, "powerpoint"):
		return AttachmentTypeDocument
	default:
		return AttachmentTypeOther
	}
}

// AttachmentUploader handles uploading attachments to storage.
type AttachmentUploader struct {
	client       *Client
	maxEmbedSize int
}

// NewAttachmentUploader creates a new attachment uploader.
func NewAttachmentUploader(client *Client) *AttachmentUploader {
	return &AttachmentUploader{
		client:       client,
		maxEmbedSize: 1024 * 1024, // 1MB default
	}
}

// SetMaxEmbedSize sets the maximum size for inline embedding.
func (u *AttachmentUploader) SetMaxEmbedSize(size int) {
	u.maxEmbedSize = size
}

// Upload uploads an attachment and returns the URL or embedded data.
func (u *AttachmentUploader) Upload(ctx context.Context, attachment *Attachment) (string, error) {
	// For small files, embed inline
	if attachment.IsEmbeddable(u.maxEmbedSize) {
		return attachment.ToDataURL(), nil
	}

	// For larger files, would upload to S3 or similar
	// For now, just embed anyway (server will handle storage)
	return attachment.ToDataURL(), nil
}

// UploadMultiple uploads multiple attachments.
func (u *AttachmentUploader) UploadMultiple(ctx context.Context, attachments []*Attachment) ([]string, error) {
	urls := make([]string, len(attachments))
	for i, a := range attachments {
		url, err := u.Upload(ctx, a)
		if err != nil {
			return nil, fmt.Errorf("failed to upload attachment %s: %w", a.Name, err)
		}
		urls[i] = url
	}
	return urls, nil
}

// AttachmentOption is a functional option for attachments.
type AttachmentOption func(*attachmentOptions)

type attachmentOptions struct {
	attachments []*Attachment
}

// WithAttachment adds an attachment.
func WithAttachment(attachment *Attachment) AttachmentOption {
	return func(o *attachmentOptions) {
		o.attachments = append(o.attachments, attachment)
	}
}

// WithFileAttachment adds a file attachment.
func WithFileAttachment(path string) AttachmentOption {
	return func(o *attachmentOptions) {
		if a, err := NewAttachmentFromFile(path); err == nil {
			o.attachments = append(o.attachments, a)
		}
	}
}

// WithTextAttachment adds a text attachment.
func WithTextAttachment(name, content string) AttachmentOption {
	return func(o *attachmentOptions) {
		o.attachments = append(o.attachments, NewTextAttachment(name, content))
	}
}

// WithImageAttachment adds an image attachment.
func WithImageAttachment(name string, data []byte) AttachmentOption {
	return func(o *attachmentOptions) {
		o.attachments = append(o.attachments, NewImageAttachment(name, data, ""))
	}
}

// AttachmentExtractor extracts attachments from LLM responses.
type AttachmentExtractor struct {
	enabled bool
}

// NewAttachmentExtractor creates a new attachment extractor.
func NewAttachmentExtractor() *AttachmentExtractor {
	return &AttachmentExtractor{enabled: true}
}

// SetEnabled enables or disables attachment extraction.
func (e *AttachmentExtractor) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// Extract extracts attachments from content.
func (e *AttachmentExtractor) Extract(content any) []*Attachment {
	if !e.enabled {
		return nil
	}

	var attachments []*Attachment

	switch v := content.(type) {
	case string:
		// Look for data URLs
		attachments = append(attachments, e.extractDataURLs(v)...)
	case map[string]any:
		// Look for image URLs or base64 data in response
		attachments = append(attachments, e.extractFromMap(v)...)
	case []any:
		for _, item := range v {
			attachments = append(attachments, e.Extract(item)...)
		}
	}

	return attachments
}

func (e *AttachmentExtractor) extractDataURLs(s string) []*Attachment {
	var attachments []*Attachment

	// Simple data URL extraction
	if strings.HasPrefix(s, "data:") {
		parts := strings.SplitN(s, ",", 2)
		if len(parts) == 2 {
			mimeType := strings.TrimPrefix(parts[0], "data:")
			mimeType = strings.TrimSuffix(mimeType, ";base64")

			data, err := base64.StdEncoding.DecodeString(parts[1])
			if err == nil {
				attachments = append(attachments, &Attachment{
					Name:     "extracted",
					Type:     mimeTypeToAttachmentType(mimeType),
					MimeType: mimeType,
					Data:     data,
				})
			}
		}
	}

	return attachments
}

func (e *AttachmentExtractor) extractFromMap(m map[string]any) []*Attachment {
	var attachments []*Attachment

	// Look for common image/file keys
	keys := []string{"image", "images", "file", "files", "attachment", "attachments", "url", "data"}

	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case string:
				if strings.HasPrefix(v, "data:") || strings.HasPrefix(v, "http") {
					attachments = append(attachments, &Attachment{
						Name: key,
						URL:  v,
						Type: AttachmentTypeOther,
					})
				}
			case []byte:
				attachments = append(attachments, NewAttachmentFromBytes(key, v, ""))
			case map[string]any:
				attachments = append(attachments, e.extractFromMap(v)...)
			}
		}
	}

	return attachments
}

// ImageFromBase64 creates an Attachment from a base64 string.
func ImageFromBase64(name, base64Data, mimeType string) (*Attachment, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return &Attachment{
		Name:     name,
		Type:     AttachmentTypeImage,
		MimeType: mimeType,
		Data:     data,
		Base64:   base64Data,
	}, nil
}

// ImageFromURL fetches and creates an Attachment from a URL.
func ImageFromURL(ctx context.Context, name, url string) (*Attachment, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}

	mimeType := resp.Header.Get("Content-Type")

	return &Attachment{
		Name:     name,
		Type:     AttachmentTypeImage,
		MimeType: mimeType,
		Data:     buf.Bytes(),
		URL:      url,
	}, nil
}
