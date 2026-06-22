package agentobs

import (
	"encoding/base64"
	"testing"
)

func TestExtractMediaMarkdownImages(t *testing.T) {
	content := "Check out this image: ![My Image](https://example.com/image.png) and this one ![](https://example.com/other.jpg)"

	refs := ExtractMedia(content)

	if len(refs) != 2 {
		t.Fatalf("expected 2 media refs, got %d", len(refs))
	}

	// First image with alt text
	if refs[0].Source != "markdown" {
		t.Errorf("expected source markdown, got %s", refs[0].Source)
	}
	if refs[0].Alt != "My Image" {
		t.Errorf("expected alt 'My Image', got %s", refs[0].Alt)
	}
	if refs[0].URL != "https://example.com/image.png" {
		t.Errorf("expected URL https://example.com/image.png, got %s", refs[0].URL)
	}
	if refs[0].Type != "image" {
		t.Errorf("expected type image, got %s", refs[0].Type)
	}

	// Second image without alt text
	if refs[1].Alt != "" {
		t.Errorf("expected empty alt, got %s", refs[1].Alt)
	}
	if refs[1].Type != "image" {
		t.Errorf("expected type image, got %s", refs[1].Type)
	}
}

func TestExtractMediaDataURLs(t *testing.T) {
	// Small PNG data URL
	pngData := base64.StdEncoding.EncodeToString([]byte("fake png data"))
	content := "Here is an image: data:image/png;base64," + pngData

	refs := ExtractMedia(content)

	if len(refs) != 1 {
		t.Fatalf("expected 1 media ref, got %d", len(refs))
	}

	if refs[0].Source != "data" {
		t.Errorf("expected source data, got %s", refs[0].Source)
	}
	if refs[0].MimeType != "image/png" {
		t.Errorf("expected mime type image/png, got %s", refs[0].MimeType)
	}
	if refs[0].Type != "image" {
		t.Errorf("expected type image, got %s", refs[0].Type)
	}
	if string(refs[0].Data) != "fake png data" {
		t.Errorf("expected decoded data 'fake png data', got %s", string(refs[0].Data))
	}
}

func TestExtractMediaProtocol(t *testing.T) {
	audioData := base64.StdEncoding.EncodeToString([]byte("audio bytes"))
	content := "Play this: media:audio/mp3:" + audioData

	refs := ExtractMedia(content)

	if len(refs) != 1 {
		t.Fatalf("expected 1 media ref, got %d", len(refs))
	}

	if refs[0].Source != "media" {
		t.Errorf("expected source media, got %s", refs[0].Source)
	}
	if refs[0].MimeType != "audio/mp3" {
		t.Errorf("expected mime type audio/mp3, got %s", refs[0].MimeType)
	}
	if refs[0].Type != "audio" {
		t.Errorf("expected type audio, got %s", refs[0].Type)
	}
}

func TestExtractMediaFileURLs(t *testing.T) {
	content := "Open file://Users/test/document.pdf for more info"

	refs := ExtractMedia(content)

	if len(refs) != 1 {
		t.Fatalf("expected 1 media ref, got %d", len(refs))
	}

	if refs[0].Source != "file" {
		t.Errorf("expected source file, got %s", refs[0].Source)
	}
	if refs[0].Type != "document" {
		t.Errorf("expected type document, got %s", refs[0].Type)
	}
	if refs[0].URL != "file://Users/test/document.pdf" {
		t.Errorf("expected URL file://Users/test/document.pdf, got %s", refs[0].URL)
	}
}

func TestExtractMediaHTTPUrls(t *testing.T) {
	content := `
		Image: https://example.com/photo.jpg
		Audio: https://example.com/song.mp3
		Video: https://example.com/video.mp4
		Doc: https://example.com/report.pdf
	`

	refs := ExtractMedia(content)

	if len(refs) != 4 {
		t.Fatalf("expected 4 media refs, got %d", len(refs))
	}

	expectedTypes := map[string]string{
		"https://example.com/photo.jpg":  "image",
		"https://example.com/song.mp3":   "audio",
		"https://example.com/video.mp4":  "video",
		"https://example.com/report.pdf": "document",
	}

	for _, ref := range refs {
		expected, ok := expectedTypes[ref.URL]
		if !ok {
			t.Errorf("unexpected URL: %s", ref.URL)
			continue
		}
		if ref.Type != expected {
			t.Errorf("URL %s: expected type %s, got %s", ref.URL, expected, ref.Type)
		}
		if ref.Source != "url" {
			t.Errorf("URL %s: expected source url, got %s", ref.URL, ref.Source)
		}
	}
}

func TestExtractMediaEmpty(t *testing.T) {
	refs := ExtractMedia("")
	if refs != nil {
		t.Error("expected nil for empty content")
	}

	refs = ExtractMedia("No media here, just text.")
	if len(refs) != 0 {
		t.Errorf("expected 0 media refs, got %d", len(refs))
	}
}

func TestExtractMediaNoDuplicates(t *testing.T) {
	// Markdown image with HTTP URL should not create duplicate
	content := "![Test](https://example.com/test.png)"

	refs := ExtractMedia(content)

	if len(refs) != 1 {
		t.Fatalf("expected 1 media ref (no duplicate), got %d", len(refs))
	}

	if refs[0].Source != "markdown" {
		t.Errorf("expected source markdown, got %s", refs[0].Source)
	}
}

func TestExtractMediaFromMap(t *testing.T) {
	data := map[string]any{
		"message": "See ![photo](https://example.com/photo.png)",
		"nested": map[string]any{
			"content": "Audio: https://example.com/sound.mp3",
		},
		"list": []any{
			"https://example.com/video.mp4",
		},
		"number": 42, // Should be ignored
	}

	refs := ExtractMediaFromMap(data)

	if len(refs) != 3 {
		t.Fatalf("expected 3 media refs, got %d", len(refs))
	}
}

func TestStripMedia(t *testing.T) {
	pngData := base64.StdEncoding.EncodeToString([]byte("png"))
	content := "Image: ![alt](https://example.com/img.png) and data:image/png;base64," + pngData + " and media:audio/mp3:" + pngData

	result := StripMedia(content)

	if result != "Image: [image: alt] and [data: image/png] and [media: audio/mp3]" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestMimeToType(t *testing.T) {
	tests := []struct {
		mime     string
		expected string
	}{
		{"image/png", "image"},
		{"image/jpeg", "image"},
		{"audio/mp3", "audio"},
		{"audio/wav", "audio"},
		{"video/mp4", "video"},
		{"video/webm", "video"},
		{"text/plain", "text"},
		{"text/html", "text"},
		{"application/pdf", "document"},
		{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "document"},
		{"application/octet-stream", "other"},
	}

	for _, tc := range tests {
		got := mimeToType(tc.mime)
		if got != tc.expected {
			t.Errorf("mimeToType(%s): expected %s, got %s", tc.mime, tc.expected, got)
		}
	}
}

func TestUrlToType(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/image.png", "image"},
		{"https://example.com/image.PNG", "image"},
		{"https://example.com/photo.jpg", "image"},
		{"https://example.com/photo.jpeg", "image"},
		{"https://example.com/audio.mp3", "audio"},
		{"https://example.com/sound.wav", "audio"},
		{"https://example.com/video.mp4", "video"},
		{"https://example.com/clip.webm", "video"},
		{"https://example.com/doc.pdf", "document"},
		{"https://example.com/data.json", "text"},
		{"https://example.com/file.bin", "other"},
		{"https://example.com/image.png?size=large", "image"},
	}

	for _, tc := range tests {
		got := urlToType(tc.url)
		if got != tc.expected {
			t.Errorf("urlToType(%s): expected %s, got %s", tc.url, tc.expected, got)
		}
	}
}
