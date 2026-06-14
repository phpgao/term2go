package term2go

import (
	"context"
	"testing"
)

func TestScreenshotRegionRejectsNonPNG(t *testing.T) {
	err := screenshot(context.Background(), "/tmp/test.jpg", func(ctx context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-.png path")
	}
}

func TestScreenshotRegionRejectsNoExtension(t *testing.T) {
	err := screenshot(context.Background(), "/tmp/test", func(ctx context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected error for path without .png extension")
	}
}
