package lib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasMultiplePixelValues(t *testing.T) {
	tests := []struct {
		name     string
		video    string
		duration int
		want     bool
	}{
		{
			name:     "has only one pixel value",
			video:    "test_assets/differ_multiple_values_no_difference.avi",
			duration: 60,
			want:     false,
		},
		{
			name:     "has multiple pixel values",
			video:    "test_assets/differ_multiple_values_with_difference.avi",
			duration: 60,
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				dirs, err := filepath.Glob(filepath.Dir(tt.video) + "/.frames_*")
				if err != nil {
					t.Errorf("failed to resolve .frames dir for cleanup: %v", err)
				}
				for _, dir := range dirs {
					if err = os.RemoveAll(dir); err != nil {
						t.Errorf("error removing temp dirs after test: %v", err)
					}
				}
			})

			got, err := HasMultiplePixelValues(tt.video, tt.duration, false)
			if err != nil {
				t.Errorf("HasMultiplePixelValues() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("HasMultiplePixelValues() got = %v, want %v", got, tt.want)
			}
		})
	}
}
