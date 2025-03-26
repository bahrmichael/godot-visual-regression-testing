package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/png" // This registers PNG format via init() for pixel comparison
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Config struct {
	Godot           string
	Scenes          string
	Duration        int
	Baseline        string
	Verbose         bool
	TmpDir          string
	ResultDir       string
	RetainRenderDir bool
}

var defaultConfig = Config{
	Duration:  60,
	TmpDir:    ".",
	ResultDir: ".",
}

func main() {
	command, config := parseFlags()

	err := verifyGodotInstallation(config.Godot)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println("Error: ffmpeg not found. Please install ffmpeg.")
		os.Exit(1)
	}

	if !strings.HasSuffix(config.TmpDir, "/") {
		config.TmpDir = config.TmpDir + "/"
	}
	if !strings.HasSuffix(config.ResultDir, "/") {
		config.ResultDir = config.ResultDir + "/"
	}

	if command == "baseline" {
		_, err := renderScenes(config)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else if command == "test" {
		failedDiffs, err := testScenes(config)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if len(failedDiffs) > 0 {
			fmt.Println("Failed diffs:")
			for _, diff := range failedDiffs {
				fmt.Println(diff)
			}
			os.Exit(1)
		}
	}
}

func testScenes(config Config) ([]string, error) {
	// list all sceneFiles at config.Scenes (that's a glob)
	sceneFiles, err := filepath.Glob(config.Scenes)
	if err != nil {
		return nil, fmt.Errorf("error listing sceneFiles: %v", err)
	}
	baselineFiles, err := filepath.Glob(config.Baseline)
	if err != nil {
		return nil, fmt.Errorf("error listing sceneFiles: %v", err)
	}

	// there must be a baseline for each scene if we're in test mode
	var missingBaselines []string
	for _, file := range sceneFiles {
		target := strings.Replace(file, ".tscn", ".avi", 1)
		if !slices.Contains(baselineFiles, target) {
			missingBaselines = append(missingBaselines, file)
		}
	}
	if len(missingBaselines) > 0 {
		return nil, fmt.Errorf("missing baselines for scenes: %v", missingBaselines)
	}

	renderDir, err := os.MkdirTemp(config.TmpDir, "renders_")
	if err != nil {
		return nil, fmt.Errorf("error creating temp dir: %v", err)
	}
	if !strings.HasSuffix(renderDir, "/") {
		renderDir = renderDir + "/"
	}
	if !config.RetainRenderDir {
		defer os.RemoveAll(renderDir)
	}

	// for each file, render it
	defaultArgs := []string{
		"--quit-after",
		strconv.Itoa(config.Duration),
	}
	var failedDiffs []string

	for _, file := range sceneFiles {
		sceneName := strings.Replace(file, ".tscn", "", 1)

		actualPathFile := fmt.Sprintf("%s%s%s", renderDir, sceneName, "_actual.avi")
		err := os.MkdirAll(filepath.Dir(actualPathFile), 0755)
		if err != nil {
			return nil, fmt.Errorf("error creating dir: %v", err)
		}
		renderedScene, err := renderScene(file, actualPathFile, defaultArgs, config)
		if err != nil {
			return nil, fmt.Errorf("error rendering file: %v", err)
		}

		baseline := fmt.Sprintf("%s%s", sceneName, ".avi")
		diffOutFile := fmt.Sprintf("%s%s%s", renderDir, sceneName, "_diff.avi")
		err = os.MkdirAll(filepath.Dir(diffOutFile), 0755)
		if err != nil {
			return nil, fmt.Errorf("error creating dir: %v", err)
		}
		h, err := hasDiff(renderedScene, baseline, diffOutFile, config)
		if h {
			d, err := generateComparison(sceneName, renderedScene, baseline, config)
			if err != nil {
				return nil, fmt.Errorf("error generating comparison: %v", err)
			}
			failedDiffs = append(failedDiffs, d)
		}
	}

	return failedDiffs, nil
}

func parseFlags() (string, Config) {
	config := defaultConfig

	flag.StringVar(&config.TmpDir, "workdir", config.TmpDir, "Path to workdir")
	flag.StringVar(&config.Godot, "godot", config.Godot, "Path to Godot executable")
	flag.StringVar(&config.Scenes, "scenes", config.Scenes, "Glob path to the .tscn files")
	flag.StringVar(&config.Baseline, "baseline", config.Baseline, "Glob path to the baseline .avi files")
	flag.IntVar(&config.Duration, "duration", config.Duration, "Duration of video capture in frames")
	flag.BoolVar(&config.Verbose, "verbose", config.Verbose, "Running verbose")
	flag.BoolVar(&config.RetainRenderDir, "retain-render-dir", config.RetainRenderDir, "Retain render dir")

	helpFlag := flag.Bool("help", false, "Show help information")
	command := flag.String("command", "", "Command to run (baseline or test)")

	flag.Parse()

	if command == nil || *command == "" {
		showHelp("")
		os.Exit(0)
	} else if *helpFlag {
		showHelp(*command)
		os.Exit(0)
	}

	var missingFlags []string

	if config.Godot == "" {
		missingFlags = append(missingFlags, "godot")
	}
	if config.Scenes == "" {
		missingFlags = append(missingFlags, "scenes")
	}

	if *command == "test" {
		if config.Baseline == "" {
			missingFlags = append(missingFlags, "baseline")
		}
	}

	if len(missingFlags) > 0 {
		fmt.Println("Error: The following required flags were not provided:")
		for _, flagName := range missingFlags {
			fmt.Printf("  --%s\n", flagName)
		}
		fmt.Println("\nUse --help for more information.")
		os.Exit(1)
	}

	return *command, config
}

func showHelp(command string) {
	fmt.Println("Godot Visual Regression Tester (VRT)")
	fmt.Println("Renders two Godot scenes and generates a visual diff video")
	fmt.Println()
	if command == "" {
		fmt.Println("Available commands: ")
		fmt.Println("  baseline")
		fmt.Println("  test")
		fmt.Println()
	} else {
		fmt.Println("Required flags:")
		fmt.Println("  --godot    Godot executable (can be a path or the name of the executable)")
		fmt.Println("  --scenes   Glob path to the .tscn files")
		fmt.Println()
	}
	fmt.Println("Optional flags:")
	fmt.Println("  --help              Show this help message")
	if command != "" {
		fmt.Println("  --verbose           Activates verbose mode for debugging")
		fmt.Println("  --duration          The number of frames to render (default: 60)")
	}
	if command == "test" {
		fmt.Println("  --baseline          Glob path to the baseline .avi files")
	}
}

func hasDiff(renderedVideo, baselineVideo, outFile string, config Config) (bool, error) {
	args := []string{
		"-i",
		baselineVideo,
		"-i",
		renderedVideo,
		"-filter_complex",
		"blend=all_mode=difference",
		outFile,
	}
	if !config.Verbose {
		args = slices.Insert(args, 0, "-loglevel", "error")
	}
	_, stderr, err := executeCommandUnsafe("ffmpeg", args)

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

	return HasMultiplePixelValues(outFile, config)
}

func generateComparison(sceneName, rendered, baseline string, config Config) (string, error) {
	outFile := fmt.Sprintf("%s%s%s%s%s", config.ResultDir, filepath.Dir(sceneName), "/comparison_", filepath.Base(sceneName), ".avi")
	args := []string{
		"-y",
		"-loglevel",
		"error",
		"-i",
		baseline,
		"-i",
		rendered,
		"-filter_complex",
		"[0:v][1:v]blend=all_mode=difference[diff];[0:v][1:v][diff]hstack=inputs=3",
		outFile,
	}

	_, stderr, err := executeCommandUnsafe(
		"ffmpeg",
		args)

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

func verifyGodotInstallation(godotPath string) error {
	_, err := exec.LookPath(godotPath)
	if err != nil {
		return fmt.Errorf("cannot find Godot installation at %s", godotPath)
	}

	stdout, stderr, err := executeCommandUnsafe(godotPath, []string{"--version", "--headless"})
	if err != nil {
		return fmt.Errorf("error executing Godot: %v %s", err, stderr)
	}
	fmt.Println("Godot version: " + stdout)

	if !strings.HasPrefix(stdout, "4.4.stable") {
		return fmt.Errorf("godot version is currently not supported. Please install Godot 4.4.stable and try again: %v %s", err, stderr)
	}
	return nil
}

func renderScenes(config Config) ([]string, error) {

	// list all sceneFiles at config.Scenes (that's a glob)
	sceneFiles, err := filepath.Glob(config.Scenes)
	if err != nil {
		return nil, fmt.Errorf("error listing sceneFiles: %v", err)
	}
	// for each file, render it
	defaultArgs := []string{
		"--quit-after",
		strconv.Itoa(config.Duration),
	}
	renderedAvis := make([]string, 0, len(sceneFiles))
	for _, file := range sceneFiles {
		b, err := renderScene(file, strings.Replace(file, ".tscn", ".avi", 1), defaultArgs, config)
		if err != nil {
			return nil, fmt.Errorf("error rendering file: %v", err)
		}
		renderedAvis = append(renderedAvis, b)
	}

	return renderedAvis, nil
}

func renderScene(scene, outputFile string, args []string, config Config) (string, error) {
	args = append(args, "--write-movie", outputFile, scene)
	if config.Verbose {
		args = slices.Insert(args, 0, "--verbose")
	}
	stdout, stderr, err := executeCommandUnsafe(config.Godot, args)
	if config.Verbose {
		fmt.Println(stdout)
		fmt.Println(stderr)
	}
	if err != nil {
		return "", fmt.Errorf("error rendering scene: %v %s", err, stderr)
	}
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		return "", fmt.Errorf("error getting rendered file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("error: rendered file is empty")
	}

	return outputFile, nil
}

func HasMultiplePixelValues(videoPath string, config Config) (bool, error) {
	// Create temporary directory for extracted frames
	tempDir, err := os.MkdirTemp(config.TmpDir, "video_frames_")
	if err != nil {
		return false, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract frames using ffmpeg
	args := []string{
		"-i", videoPath,
		"-vf", fmt.Sprintf("select=gte(n\\,0)"),
		"-vframes", strconv.Itoa(config.Duration),
		"-vsync", "0",
		"-f", "image2",
		fmt.Sprintf("%s/frame_%%05d.png", tempDir),
	}

	if !config.Verbose {
		args = slices.Insert(args, 0, "-loglevel", "error")
	}

	_, stderr, err := executeCommandUnsafe("ffmpeg", args)
	if err != nil {
		return false, fmt.Errorf("failed to extract frames: %v - %s", err, stderr)
	}

	frameFiles, err := filepath.Glob(fmt.Sprintf("%s/frame_*.png", tempDir))
	if err != nil {
		return false, fmt.Errorf("failed to list frame files: %v", err)
	}

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

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				r, g, b, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
				if r != 0 || b != 0 || g != 34181 || a != 65535 {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
