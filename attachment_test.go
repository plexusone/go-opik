package opik

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestAttachmentTypes(t *testing.T) {
	tests := []struct {
		attachType AttachmentType
		want       string
	}{
		{AttachmentTypeImage, "image"},
		{AttachmentTypeAudio, "audio"},
		{AttachmentTypeVideo, "video"},
		{AttachmentTypeDocument, "document"},
		{AttachmentTypeText, "text"},
		{AttachmentTypeOther, "other"},
	}

	for _, tt := range tests {
		if string(tt.attachType) != tt.want {
			t.Errorf("AttachmentType = %q, want %q", tt.attachType, tt.want)
		}
	}
}

func TestNewAttachmentFromBytes(t *testing.T) {
	data := []byte("Hello, World!")

	t.Run("with mime type", func(t *testing.T) {
		a := NewAttachmentFromBytes("test.txt", data, "text/plain")

		if a.Name != "test.txt" {
			t.Errorf("Name = %q, want %q", a.Name, "test.txt")
		}
		if a.MimeType != "text/plain" {
			t.Errorf("MimeType = %q, want %q", a.MimeType, "text/plain")
		}
		if a.Type != AttachmentTypeText {
			t.Errorf("Type = %v, want %v", a.Type, AttachmentTypeText)
		}
		if !bytes.Equal(a.Data, data) {
			t.Error("Data doesn't match")
		}
	})

	t.Run("auto detect mime type", func(t *testing.T) {
		a := NewAttachmentFromBytes("test.bin", data, "")

		// http.DetectContentType should detect text/plain for this data
		if a.MimeType == "" {
			t.Error("MimeType should be auto-detected")
		}
	})
}

func TestNewAttachmentFromReader(t *testing.T) {
	data := []byte("Test content")
	reader := bytes.NewReader(data)

	a, err := NewAttachmentFromReader("test.txt", reader, "text/plain")
	if err != nil {
		t.Fatalf("NewAttachmentFromReader error = %v", err)
	}

	if a.Name != "test.txt" {
		t.Errorf("Name = %q, want %q", a.Name, "test.txt")
	}
	if !bytes.Equal(a.Data, data) {
		t.Error("Data doesn't match")
	}
}

func TestNewAttachmentFromURL(t *testing.T) {
	a := NewAttachmentFromURL("image.png", "https://example.com/image.png", AttachmentTypeImage)

	if a.Name != "image.png" {
		t.Errorf("Name = %q, want %q", a.Name, "image.png")
	}
	if a.URL != "https://example.com/image.png" {
		t.Errorf("URL = %q, want %q", a.URL, "https://example.com/image.png")
	}
	if a.Type != AttachmentTypeImage {
		t.Errorf("Type = %v, want %v", a.Type, AttachmentTypeImage)
	}
	if len(a.Data) != 0 {
		t.Error("Data should be empty for URL attachment")
	}
}

func TestNewTextAttachment(t *testing.T) {
	content := "This is some text content"
	a := NewTextAttachment("notes.txt", content)

	if a.Name != "notes.txt" {
		t.Errorf("Name = %q, want %q", a.Name, "notes.txt")
	}
	if a.Type != AttachmentTypeText {
		t.Errorf("Type = %v, want %v", a.Type, AttachmentTypeText)
	}
	if a.MimeType != "text/plain" {
		t.Errorf("MimeType = %q, want %q", a.MimeType, "text/plain")
	}
	if string(a.Data) != content {
		t.Errorf("Data = %q, want %q", string(a.Data), content)
	}
}

func TestNewImageAttachment(t *testing.T) {
	// PNG header bytes
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	t.Run("with mime type", func(t *testing.T) {
		a := NewImageAttachment("test.png", data, "image/png")

		if a.Type != AttachmentTypeImage {
			t.Errorf("Type = %v, want %v", a.Type, AttachmentTypeImage)
		}
		if a.MimeType != "image/png" {
			t.Errorf("MimeType = %q, want %q", a.MimeType, "image/png")
		}
	})

	t.Run("auto detect mime type", func(t *testing.T) {
		a := NewImageAttachment("test.png", data, "")

		// Should detect image/png from the header bytes
		if a.MimeType != "image/png" {
			t.Errorf("MimeType = %q, want %q", a.MimeType, "image/png")
		}
	})
}

func TestAttachmentToBase64(t *testing.T) {
	data := []byte("Hello")
	a := NewAttachmentFromBytes("test.txt", data, "text/plain")

	b64 := a.ToBase64()
	expected := base64.StdEncoding.EncodeToString(data)

	if b64 != expected {
		t.Errorf("ToBase64() = %q, want %q", b64, expected)
	}
}

func TestAttachmentToBase64Cached(t *testing.T) {
	a := &Attachment{
		Name:   "test",
		Base64: "Y2FjaGVk", // "cached" in base64
		Data:   []byte("different"),
	}

	// Should return cached Base64, not re-encode Data
	if a.ToBase64() != "Y2FjaGVk" {
		t.Error("ToBase64() should return cached value")
	}
}

func TestAttachmentToBase64Empty(t *testing.T) {
	a := &Attachment{Name: "empty"}

	if a.ToBase64() != "" {
		t.Error("ToBase64() should return empty for attachment without data")
	}
}

func TestAttachmentToDataURL(t *testing.T) {
	data := []byte("Hello")
	a := NewAttachmentFromBytes("test.txt", data, "text/plain")

	dataURL := a.ToDataURL()

	if !strings.HasPrefix(dataURL, "data:text/plain;base64,") {
		t.Errorf("ToDataURL() = %q, should start with data:text/plain;base64,", dataURL)
	}
}

func TestAttachmentToDataURLWithURL(t *testing.T) {
	a := NewAttachmentFromURL("image.png", "https://example.com/image.png", AttachmentTypeImage)

	// When URL is set, ToDataURL should return it directly
	if a.ToDataURL() != "https://example.com/image.png" {
		t.Errorf("ToDataURL() = %q, want URL", a.ToDataURL())
	}
}

func TestAttachmentToDataURLDefaultMime(t *testing.T) {
	a := &Attachment{
		Name: "test",
		Data: []byte("data"),
	}

	dataURL := a.ToDataURL()
	if !strings.HasPrefix(dataURL, "data:application/octet-stream;base64,") {
		t.Error("ToDataURL() should use application/octet-stream for unknown mime type")
	}
}

func TestAttachmentSize(t *testing.T) {
	data := []byte("Hello, World!")
	a := NewAttachmentFromBytes("test.txt", data, "text/plain")

	if a.Size() != len(data) {
		t.Errorf("Size() = %d, want %d", a.Size(), len(data))
	}
}

func TestAttachmentIsEmbeddable(t *testing.T) {
	smallData := make([]byte, 100)
	largeData := make([]byte, 2000)

	small := NewAttachmentFromBytes("small.bin", smallData, "")
	large := NewAttachmentFromBytes("large.bin", largeData, "")

	if !small.IsEmbeddable(1000) {
		t.Error("small attachment should be embeddable with 1000 byte limit")
	}
	if large.IsEmbeddable(1000) {
		t.Error("large attachment should not be embeddable with 1000 byte limit")
	}
}

func TestMimeTypeToAttachmentType(t *testing.T) {
	tests := []struct {
		mimeType string
		want     AttachmentType
	}{
		{"image/png", AttachmentTypeImage},
		{"image/jpeg", AttachmentTypeImage},
		{"audio/mpeg", AttachmentTypeAudio},
		{"audio/wav", AttachmentTypeAudio},
		{"video/mp4", AttachmentTypeVideo},
		{"text/plain", AttachmentTypeText},
		{"text/html", AttachmentTypeText},
		{"application/pdf", AttachmentTypeDocument},
		{"application/msword", AttachmentTypeDocument},
		{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", AttachmentTypeDocument},
		{"application/octet-stream", AttachmentTypeOther},
		{"application/json", AttachmentTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := mimeTypeToAttachmentType(tt.mimeType)
			if got != tt.want {
				t.Errorf("mimeTypeToAttachmentType(%q) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestNewAttachmentUploader(t *testing.T) {
	uploader := NewAttachmentUploader(nil)

	if uploader == nil {
		t.Fatal("NewAttachmentUploader returned nil")
	}
	if uploader.maxEmbedSize != 1024*1024 {
		t.Errorf("default maxEmbedSize = %d, want %d", uploader.maxEmbedSize, 1024*1024)
	}
}

func TestAttachmentUploaderSetMaxEmbedSize(t *testing.T) {
	uploader := NewAttachmentUploader(nil)
	uploader.SetMaxEmbedSize(500000)

	if uploader.maxEmbedSize != 500000 {
		t.Errorf("maxEmbedSize = %d, want 500000", uploader.maxEmbedSize)
	}
}

func TestAttachmentExtractorEnabled(t *testing.T) {
	extractor := NewAttachmentExtractor()

	if !extractor.enabled {
		t.Error("extractor should be enabled by default")
	}

	extractor.SetEnabled(false)
	if extractor.enabled {
		t.Error("extractor should be disabled")
	}
}

func TestAttachmentExtractorExtractDisabled(t *testing.T) {
	extractor := NewAttachmentExtractor()
	extractor.SetEnabled(false)

	result := extractor.Extract("data:image/png;base64,abc123")
	if result != nil {
		t.Error("disabled extractor should return nil")
	}
}

func TestAttachmentExtractorExtractDataURL(t *testing.T) {
	extractor := NewAttachmentExtractor()

	// Create a valid data URL
	data := []byte("Hello")
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURL := "data:text/plain;base64," + b64

	attachments := extractor.Extract(dataURL)

	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	if attachments[0].MimeType != "text/plain" {
		t.Errorf("MimeType = %q, want %q", attachments[0].MimeType, "text/plain")
	}
	if !bytes.Equal(attachments[0].Data, data) {
		t.Error("Data doesn't match")
	}
}

func TestAttachmentExtractorExtractFromMap(t *testing.T) {
	extractor := NewAttachmentExtractor()

	m := map[string]any{
		"image": "https://example.com/image.png",
	}

	attachments := extractor.Extract(m)

	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	if attachments[0].URL != "https://example.com/image.png" {
		t.Errorf("URL = %q, want %q", attachments[0].URL, "https://example.com/image.png")
	}
}

func TestAttachmentExtractorExtractFromArray(t *testing.T) {
	extractor := NewAttachmentExtractor()

	arr := []any{
		"data:text/plain;base64,SGVsbG8=",
		map[string]any{"url": "https://example.com/file"},
	}

	attachments := extractor.Extract(arr)

	if len(attachments) < 1 {
		t.Error("should extract attachments from array")
	}
}

func TestImageFromBase64(t *testing.T) {
	data := []byte("test image data")
	b64 := base64.StdEncoding.EncodeToString(data)

	a, err := ImageFromBase64("test.png", b64, "image/png")
	if err != nil {
		t.Fatalf("ImageFromBase64 error = %v", err)
	}

	if a.Name != "test.png" {
		t.Errorf("Name = %q, want %q", a.Name, "test.png")
	}
	if a.Type != AttachmentTypeImage {
		t.Errorf("Type = %v, want %v", a.Type, AttachmentTypeImage)
	}
	if !bytes.Equal(a.Data, data) {
		t.Error("Data doesn't match")
	}
	if a.Base64 != b64 {
		t.Error("Base64 should be preserved")
	}
}

func TestImageFromBase64Invalid(t *testing.T) {
	_, err := ImageFromBase64("test.png", "not-valid-base64!!!", "image/png")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestWithAttachmentOption(t *testing.T) {
	a := NewTextAttachment("test.txt", "content")
	opts := &attachmentOptions{}

	WithAttachment(a)(opts)

	if len(opts.attachments) != 1 {
		t.Errorf("attachments length = %d, want 1", len(opts.attachments))
	}
	if opts.attachments[0] != a {
		t.Error("attachment not added correctly")
	}
}

func TestWithTextAttachmentOption(t *testing.T) {
	opts := &attachmentOptions{}

	WithTextAttachment("notes.txt", "content")(opts)

	if len(opts.attachments) != 1 {
		t.Errorf("attachments length = %d, want 1", len(opts.attachments))
	}
	if opts.attachments[0].Name != "notes.txt" {
		t.Error("attachment name incorrect")
	}
}

func TestWithImageAttachmentOption(t *testing.T) {
	opts := &attachmentOptions{}
	data := []byte{0x89, 0x50, 0x4E, 0x47}

	WithImageAttachment("image.png", data)(opts)

	if len(opts.attachments) != 1 {
		t.Errorf("attachments length = %d, want 1", len(opts.attachments))
	}
	if opts.attachments[0].Type != AttachmentTypeImage {
		t.Error("attachment type should be image")
	}
}
