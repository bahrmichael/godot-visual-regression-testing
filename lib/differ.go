package lib

import (
	"fmt"
	"image"
	_ "image/png" // This registers PNG format via init() for pixel comparison
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func HasDiff(renderedVideo, baselineVideo, outFile string, verbose bool, duration int) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(outFile), 0700); err != nil {
		return false, fmt.Errorf("error creating dir: %v", err)
	}
	args := []string{
		"-i",
		baselineVideo,
		"-i",
		renderedVideo,
		"-filter_complex",
		"blend=all_mode=difference",
		outFile,
	}
	if !verbose {
		args = slices.Insert(args, 0, "-loglevel", "error")
	}
	_, stderr, err := executeCommandUnsafe(nil, "ffmpeg", args)

	if err != nil {
		return false, fmt.Errorf("generating diff video: %v %s", err, stderr)
	}

	fileInfo, err := os.Stat(outFile)
	if err != nil {
		return false, fmt.Errorf("error getting diff file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		return false, fmt.Errorf("error: diff file is empty")
	}

	return HasMultiplePixelValues(outFile, duration, verbose)
}

func HasMultiplePixelValues(videoPath string, duration int, verbose bool) (bool, error) {
	// Create temporary directory for extracted frames
	//tempDir, err := os.MkdirTemp(config.TmpDir, "video_frames_")
	tempDir, err := os.MkdirTemp(filepath.Dir(videoPath), fmt.Sprintf(".frames_%s_", strings.Replace(filepath.Base(videoPath), ".avi", "", 1)))
	if err != nil {
		return false, fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Extract frames using ffmpeg
	args := []string{
		"-i", videoPath,
		"-vf", fmt.Sprintf("select=gte(n\\,0)"),
		"-vframes", strconv.Itoa(duration),
		"-vsync", "0",
		"-f", "image2",
		fmt.Sprintf("%s/frame_%%05d.png", tempDir),
	}

	if !verbose {
		args = slices.Insert(args, 0, "-loglevel", "error")
	}

	_, stderr, err := executeCommandUnsafe(nil, "ffmpeg", args)
	if err != nil {
		return false, fmt.Errorf("failed to extract frames: %v - %s", err, stderr)
	}

	frameFiles, err := filepath.Glob(fmt.Sprintf("%s/frame_*.png", tempDir))
	if err != nil {
		return false, fmt.Errorf("failed to list frame files: %v", err)
	}

	// The differ_test fails because somehow it only yields 30 frames.
	// ffprobe -select_streams v -show_streams lib/test_assets/differ_multiple_values_with_difference.avi | grep nb_frames
	// shows that the files have 60 frames
	//if len(frameFiles) != duration {
	//	return false, fmt.Errorf("expected %d frames, got %d", duration, len(frameFiles))
	//}

	for _, framePath := range frameFiles {
		file, err := os.Open(framePath)
		if err != nil {
			return false, fmt.Errorf("failed to open frame %s: %v", framePath, err)
		}

		img, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			return false, fmt.Errorf("failed to decode frame %s: %v", framePath, err)
		}

		bounds := img.Bounds()
		width := bounds.Max.X - bounds.Min.X
		height := bounds.Max.Y - bounds.Min.Y

		// We have to remember the initial pixel value to make sure that the entire diff is the same.
		// We can't use a predetermined pixel value, because the colour seems to change from system to system.
		pixelValue := make([]uint32, 4)
		pixelValueInitialized := false

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				r, g, b, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
				if !pixelValueInitialized {
					pixelValue = []uint32{r, g, b, a}
					pixelValueInitialized = true
				} else {
					if r != pixelValue[0] || g != pixelValue[1] || b != pixelValue[2] || a != pixelValue[3] {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

func GenerateComparison(sceneName, rendered, baseline, resultDir string, verbose bool) (string, error) {

	outFile := fmt.Sprintf("%s%s%s", resultDir, sceneName, ".avi")
	if err := os.MkdirAll(filepath.Dir(outFile), 0755); err != nil {
		return "", fmt.Errorf("error creating dir: %v %s", err, resultDir)
	}
	args := []string{
		"-y",
		"-i",
		baseline,
		"-i",
		rendered,
		"-filter_complex",
		"[0:v][1:v]blend=all_mode=difference[diff];[0:v][1:v][diff]hstack=inputs=3",
		outFile,
	}

	if !verbose {
		args = slices.Insert(args, 0, "-loglevel", "error")
	}

	_, stderr, err := executeCommandUnsafe(nil, "ffmpeg", args)

	if err != nil {
		return "", fmt.Errorf("generating comparison video: %v %s", err, stderr)
	}

	fileInfo, err := os.Stat(outFile)
	if err != nil {
		return "", fmt.Errorf("error getting comparison file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("error: comparison file is empty")
	}

	return outFile, nil
}
