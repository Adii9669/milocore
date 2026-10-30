package files

import "testing"

func TestIsPreviewableImage(t *testing.T) {
	for _, mimeType := range []string{"image/jpeg", "image/png", "image/gif", "image/webp", "image/avif", "image/bmp"} {
		if !isPreviewableImage(mimeType) {
			t.Errorf("expected %s to be previewable", mimeType)
		}
	}
	for _, mimeType := range []string{"image/svg+xml", "application/pdf", "text/html", "application/octet-stream"} {
		if isPreviewableImage(mimeType) {
			t.Errorf("expected %s to be download-only", mimeType)
		}
	}
}
