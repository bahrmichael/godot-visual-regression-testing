package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/png" // This registers PNG format via init() for pixel comparison
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Config struct {
	GodotPath          string
	ProjectPath        string
	Scene              string
	Duration           int
	BaselineOutputFile string
	DiffOutputFile     string
	Baseline           string
	Headless           bool
	Verbose            bool
}

var defaultConfig = Config{
	Duration:           60,
	BaselineOutputFile: fmt.Sprintf("baseline_%d.avi", rand.Int31()),
	DiffOutputFile:     fmt.Sprintf("diff_%d.avi", rand.Int31()),
}

func main() {
	config := parseFlags()

	err := verifyGodotInstallation(config.GodotPath)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir)

	renderedScene, err := renderScene(dir, config)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if config.Baseline == "" {
		err := os.Rename(renderedScene, config.BaselineOutputFile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("Rendered scene at " + config.BaselineOutputFile)
	} else {
		h, err := hasDiff(dir, renderedScene, config)
		if !h {
			fmt.Println("No difference between the baseline and the scene.")
			return
		}

		d, err := generateComparison(dir, config.Baseline, renderedScene)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		err = os.Rename(d, config.DiffOutputFile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("Diff rendered at " + config.DiffOutputFile)
	}
}

func parseFlags() Config {
	config := defaultConfig

	flag.StringVar(&config.GodotPath, "godot", config.GodotPath, "Path to Godot executable")
	flag.StringVar(&config.ProjectPath, "project", config.ProjectPath, "Path to Godot project")
	flag.StringVar(&config.Scene, "scene", config.Scene, "Scene to render")
	flag.StringVar(&config.Baseline, "baseline", config.Baseline, "Baseline to compare the render against")
	flag.IntVar(&config.Duration, "duration", config.Duration, "Duration of video capture in frames")
	flag.StringVar(&config.BaselineOutputFile, "baseline-output", config.BaselineOutputFile, "Output file path for the render or diff video")
	flag.StringVar(&config.DiffOutputFile, "diff-output", config.DiffOutputFile, "Output file path for the render or diff video")
	flag.BoolVar(&config.Headless, "headless", config.Headless, "Running headless")
	flag.BoolVar(&config.Verbose, "verbose", config.Verbose, "Running verbose")
	helpFlag := flag.Bool("help", false, "Show help information")

	flag.Parse()

	if *helpFlag {
		showHelp()
	}

	var missingFlags []string

	if config.GodotPath == "" {
		missingFlags = append(missingFlags, "godot")
	}
	if config.ProjectPath == "" {
		missingFlags = append(missingFlags, "project")
	}
	if config.Scene == "" {
		missingFlags = append(missingFlags, "scene")
	}

	if len(missingFlags) > 0 {
		fmt.Println("Error: The following required flags were not provided:")
		for _, flagName := range missingFlags {
			fmt.Printf("  --%s\n", flagName)
		}
		fmt.Println("\nUse --help for more information.")
		os.Exit(1)
	}

	return config
}

func showHelp() {
	fmt.Println("Godot Scene Diff Tool")
	fmt.Println("Renders two Godot scenes and generates a visual diff video")
	fmt.Println()
	fmt.Println("Required flags:")
	fmt.Println("  --godot    Path to Godot executable")
	fmt.Println("  --project  Path to the project root")
	fmt.Println("  --scene    Scene to render from the project root")
	fmt.Println()
	fmt.Println("Optional flags:")
	fmt.Println()
	fmt.Println("  --help              Show this help message")
	fmt.Println("  --verbose           Activates verbose mode for debugging")
	fmt.Println("  --duration          The number of frames to render (default: 60)")
	fmt.Println("  --baseline          A baseline video to compare the render against")
	fmt.Println("  --headless          Runs godot in headless mode (doesn't work on macos)")
	fmt.Println("  --baseline-output   Output file path for the new baseline video")
	fmt.Println("  --diff-output       Output file path for the diff video")
	os.Exit(0)
}

func hasDiff(dir, f2 string, config Config) (bool, error) {
	outFile := fmt.Sprintf("%s%d.avi", dir, rand.Int31())
	args := []string{
		"-i",
		config.Baseline,
		"-i",
		f2,
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

	return HasMultiplePixelValues(outFile, config)
}

func generateComparison(dir, f1, f2 string) (string, error) {
	outFile := fmt.Sprintf("%s%d%s", dir, rand.Int31(), ".avi")
	_, stderr, err := executeCommandUnsafe(
		"ffmpeg",
		[]string{
			"-y",
			"-loglevel",
			"error",
			"-i",
			f1,
			"-i",
			f2,
			"-filter_complex",
			"[0:v][1:v]blend=all_mode=difference[diff];[0:v][1:v][diff]hstack=inputs=3",
			outFile,
		})

	if err != nil {
		return "", fmt.Errorf("generating comparison video: %v %s", err, stderr)
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

func renderScene(dir string, config Config) (string, error) {

	outputFile := fmt.Sprintf("%s%d.avi", dir, rand.Int31())
	args := []string{
		"--write-movie",
		outputFile,
		"--quit-after",
		strconv.Itoa(config.Duration),
		"--path",
		fmt.Sprintf("%s", config.ProjectPath),
		fmt.Sprintf("%s", config.Scene),
	}
	if config.Headless {
		args = slices.Insert(args, 0, "--headless")
	}
	if config.Verbose {
		args = slices.Insert(args, 0, "--verbose")
	}
	_, stderr, err := executeCommandUnsafe(config.GodotPath, args)
	if err != nil {
		return "", fmt.Errorf("error writing movie: %v %s", err, stderr)
	}

	return outputFile, nil
}

func HasMultiplePixelValues(videoPath string, config Config) (bool, error) {
	// Create temporary directory for extracted frames
	tempDir, err := os.MkdirTemp("", "video_frames_")
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
