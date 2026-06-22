package agentobs

import (
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"
)

// MediaRef represents a media reference extracted from content.
type MediaRef struct {
	// Type is the media type (image, audio, video, document, etc.)
	Type string `json:"type"`

	// MimeType is the MIME type of the media.
	MimeType string `json:"mime_type,omitempty"`

	// Source indicates where the media came from (url, file, data, markdown).
	Source string `json:"source"`

	// URL is the media URL if applicable.
	URL string `json:"url,omitempty"`

	// Data contains inline data (e.g., base64-decoded bytes).
	Data []byte `json:"data,omitempty"`

	// Alt is the alt text or description if available.
	Alt string `json:"alt,omitempty"`

	// Original is the original reference string as found in content.
	Original string `json:"original,omitempty"`
}

// Pre-compiled patterns for media extraction.
var (
	// Markdown image pattern: ![alt](url)
	markdownImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	// Data URL pattern: data:mime/type;base64,...
	dataURLPattern = regexp.MustCompile(`data:([a-zA-Z0-9/+.-]+);base64,([A-Za-z0-9+/=]+)`)

	// media: protocol pattern (custom protocol for inline media)
	mediaProtocolPattern = regexp.MustCompile(`media:([a-zA-Z0-9/+.-]+):([A-Za-z0-9+/=]+)`)

	// file:// URL pattern
	fileURLPattern = regexp.MustCompile(`file://([^\s"'<>]+)`)

	// HTTP(S) URL pattern for images
	httpImagePattern = regexp.MustCompile(`https?://[^\s"'<>]+\.(png|jpg|jpeg|gif|webp|svg|bmp|ico)(?:\?[^\s"'<>]*)?`)

	// HTTP(S) URL pattern for audio
	httpAudioPattern = regexp.MustCompile(`https?://[^\s"'<>]+\.(mp3|wav|ogg|flac|aac|m4a)(?:\?[^\s"'<>]*)?`)

	// HTTP(S) URL pattern for video
	httpVideoPattern = regexp.MustCompile(`https?://[^\s"'<>]+\.(mp4|webm|avi|mov|mkv)(?:\?[^\s"'<>]*)?`)

	// HTTP(S) URL pattern for documents
	httpDocumentPattern = regexp.MustCompile(`https?://[^\s"'<>]+\.(pdf|doc|docx|xls|xlsx|ppt|pptx|txt)(?:\?[^\s"'<>]*)?`)
)

// ExtractMedia extracts all media references from content.
// It recognizes:
//   - Markdown images: ![alt](url)
//   - Data URLs: data:mime/type;base64,...
//   - media: protocol: media:mime/type:base64data
//   - file:// URLs: file:///path/to/file
//   - HTTP URLs with media extensions
func ExtractMedia(content string) []MediaRef {
	if content == "" {
		return nil
	}

	var refs []MediaRef

	// Extract markdown images
	for _, match := range markdownImagePattern.FindAllStringSubmatch(content, -1) {
		alt := match[1]
		urlStr := match[2]

		ref := MediaRef{
			Source:   "markdown",
			Alt:      alt,
			Original: match[0],
		}

		// Check if URL is a data URL
		if strings.HasPrefix(urlStr, "data:") {
			if dataMatch := dataURLPattern.FindStringSubmatch(urlStr); len(dataMatch) == 3 {
				ref.MimeType = dataMatch[1]
				ref.Type = mimeToType(dataMatch[1])
				if data, err := base64.StdEncoding.DecodeString(dataMatch[2]); err == nil {
					ref.Data = data
				}
				ref.Source = "data"
			}
		} else {
			ref.URL = urlStr
			ref.Type = urlToType(urlStr)
		}

		refs = append(refs, ref)
	}

	// Extract data URLs (not in markdown)
	for _, match := range dataURLPattern.FindAllStringSubmatch(content, -1) {
		// Skip if already captured in markdown
		if containsOriginal(refs, match[0]) {
			continue
		}

		ref := MediaRef{
			MimeType: match[1],
			Type:     mimeToType(match[1]),
			Source:   "data",
			Original: match[0],
		}
		if data, err := base64.StdEncoding.DecodeString(match[2]); err == nil {
			ref.Data = data
		}
		refs = append(refs, ref)
	}

	// Extract media: protocol references
	for _, match := range mediaProtocolPattern.FindAllStringSubmatch(content, -1) {
		ref := MediaRef{
			MimeType: match[1],
			Type:     mimeToType(match[1]),
			Source:   "media",
			Original: match[0],
		}
		if data, err := base64.StdEncoding.DecodeString(match[2]); err == nil {
			ref.Data = data
		}
		refs = append(refs, ref)
	}

	// Extract file:// URLs
	for _, match := range fileURLPattern.FindAllStringSubmatch(content, -1) {
		path := match[1]
		if decodedPath, err := url.PathUnescape(path); err == nil {
			path = decodedPath
		}
		refs = append(refs, MediaRef{
			Type:     urlToType(path),
			Source:   "file",
			URL:      "file://" + path,
			Original: match[0],
		})
	}

	// Extract HTTP image URLs (not already captured)
	for _, match := range httpImagePattern.FindAllString(content, -1) {
		if containsURL(refs, match) {
			continue
		}
		refs = append(refs, MediaRef{
			Type:     "image",
			Source:   "url",
			URL:      match,
			Original: match,
		})
	}

	// Extract HTTP audio URLs
	for _, match := range httpAudioPattern.FindAllString(content, -1) {
		if containsURL(refs, match) {
			continue
		}
		refs = append(refs, MediaRef{
			Type:     "audio",
			Source:   "url",
			URL:      match,
			Original: match,
		})
	}

	// Extract HTTP video URLs
	for _, match := range httpVideoPattern.FindAllString(content, -1) {
		if containsURL(refs, match) {
			continue
		}
		refs = append(refs, MediaRef{
			Type:     "video",
			Source:   "url",
			URL:      match,
			Original: match,
		})
	}

	// Extract HTTP document URLs
	for _, match := range httpDocumentPattern.FindAllString(content, -1) {
		if containsURL(refs, match) {
			continue
		}
		refs = append(refs, MediaRef{
			Type:     "document",
			Source:   "url",
			URL:      match,
			Original: match,
		})
	}

	return refs
}

// ExtractMediaFromMap extracts media from string values in a map.
func ExtractMediaFromMap(data map[string]any) []MediaRef {
	var refs []MediaRef
	extractFromValue(data, &refs)
	return refs
}

func extractFromValue(v any, refs *[]MediaRef) {
	switch val := v.(type) {
	case string:
		*refs = append(*refs, ExtractMedia(val)...)
	case map[string]any:
		for _, item := range val {
			extractFromValue(item, refs)
		}
	case []any:
		for _, item := range val {
			extractFromValue(item, refs)
		}
	case []string:
		for _, item := range val {
			*refs = append(*refs, ExtractMedia(item)...)
		}
	}
}

// mimeToType converts a MIME type to a media type category.
func mimeToType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "text/"):
		return "text"
	case mimeType == "application/pdf":
		return "document"
	case strings.Contains(mimeType, "document") || strings.Contains(mimeType, "spreadsheet") || strings.Contains(mimeType, "presentation"):
		return "document"
	default:
		return "other"
	}
}

// urlToType guesses media type from a URL based on file extension.
func urlToType(urlStr string) string {
	lower := strings.ToLower(urlStr)

	// Remove query string for extension detection
	if idx := strings.Index(lower, "?"); idx != -1 {
		lower = lower[:idx]
	}

	switch {
	case strings.HasSuffix(lower, ".png"), strings.HasSuffix(lower, ".jpg"),
		strings.HasSuffix(lower, ".jpeg"), strings.HasSuffix(lower, ".gif"),
		strings.HasSuffix(lower, ".webp"), strings.HasSuffix(lower, ".svg"),
		strings.HasSuffix(lower, ".bmp"), strings.HasSuffix(lower, ".ico"):
		return "image"
	case strings.HasSuffix(lower, ".mp3"), strings.HasSuffix(lower, ".wav"),
		strings.HasSuffix(lower, ".ogg"), strings.HasSuffix(lower, ".flac"),
		strings.HasSuffix(lower, ".aac"), strings.HasSuffix(lower, ".m4a"):
		return "audio"
	case strings.HasSuffix(lower, ".mp4"), strings.HasSuffix(lower, ".webm"),
		strings.HasSuffix(lower, ".avi"), strings.HasSuffix(lower, ".mov"),
		strings.HasSuffix(lower, ".mkv"):
		return "video"
	case strings.HasSuffix(lower, ".pdf"), strings.HasSuffix(lower, ".doc"),
		strings.HasSuffix(lower, ".docx"), strings.HasSuffix(lower, ".xls"),
		strings.HasSuffix(lower, ".xlsx"), strings.HasSuffix(lower, ".ppt"),
		strings.HasSuffix(lower, ".pptx"):
		return "document"
	case strings.HasSuffix(lower, ".txt"), strings.HasSuffix(lower, ".md"),
		strings.HasSuffix(lower, ".csv"), strings.HasSuffix(lower, ".json"):
		return "text"
	default:
		return "other"
	}
}

func containsOriginal(refs []MediaRef, original string) bool {
	for _, ref := range refs {
		if strings.Contains(ref.Original, original) {
			return true
		}
	}
	return false
}

func containsURL(refs []MediaRef, url string) bool {
	for _, ref := range refs {
		if ref.URL == url {
			return true
		}
	}
	return false
}

// StripMedia removes media references from content, replacing them with placeholders.
func StripMedia(content string) string {
	result := content

	// Replace markdown images with placeholder
	result = markdownImagePattern.ReplaceAllString(result, "[image: $1]")

	// Replace data URLs with placeholder
	result = dataURLPattern.ReplaceAllString(result, "[data: $1]")

	// Replace media: protocol with placeholder
	result = mediaProtocolPattern.ReplaceAllString(result, "[media: $1]")

	return result
}
