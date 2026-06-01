package infra

import "testing"

func TestAllowedImageType(t *testing.T) {
	for _, contentType := range []string{"image/png", "image/jpeg", "image/webp"} {
		if !allowedImageType(contentType) {
			t.Fatalf("expected %s to be allowed", contentType)
		}
	}
	if allowedImageType("text/plain") {
		t.Fatalf("expected text/plain to be rejected")
	}
}

func TestObjectKeyUsesSafeImagePrefix(t *testing.T) {
	key := objectKey("cover.png", "image/png")
	if len(key) < len("images/2026/01/") || key[:7] != "images/" || key[len(key)-4:] != ".png" {
		t.Fatalf("unexpected object key: %s", key)
	}
}
