# Godot Visual Regression Testing (Godot VRT)

Inspired by [Factorio's visual regression testing](https://www.youtube.com/watch?v=LXnyTZBmfXM), this is 
a test runner for visual regression testing and end-to-end testing with Godot scenes.

You can start using it today!

## Concept

The idea is that we can generate videos from Godot scenes, and then compare those videos to a baseline (i.e. a previously
generated video that we know is correct).

## Quick Start

### Prerequisites

This currently doesn't work on headless machines (such as GitHub action runners) because we need to open a window to render the video.

1. Install [Godot 4.4 Stable](https://godotengine.org/download)
2. Install ffmpeg (if you have homebrew on macOS: `brew install ffmpeg`)

### Download the executable

You can get it here: https://github.com/bahrmichael/godot-visual-regression-testing/releases

You can also build it yourself by installing go 1.23 and running `go build .` in the root of this repository.

### Create a baseline video

Run this at the root of your Godot project:

```
godot-vrt-mac --command baseline --godot path_to_godot_binary --scenes my_scene.tscn
```

This will evaluate the `--scenes` path (you can use a glob path) and generate a video file for each scene.

```
my_scene.tscn
my_scene.avi
```

We recommend that you put the testing scenes into a separate folder to keep them neatly organized, and make it
easier to pick the right scenes. For example, we might have a folder called `vrt`:

```
godot-vrt-mac --command baseline --godot path_to_godot_binary --scenes vrt/*.tscn
```

### Compare against a baseline video

Once you have a baseline video, you can pass it to a test run.

```
godot-vrt-mac --command test --godot path_to_godot_binary --scenes my_scene.tscn --baseline my_scene.avi
```

Again, with test scenes in a separate `vrt` folder it looks like this:

```
godot-vrt-mac --command test --godot path_to_godot_binary --scenes vrt/*.tscn --baseline vrt/*.avi
```

If any scene fails the test, it will exit with a non-zero exit code and produce a diff video next to the test scene and baseline video.

```
my_scene.tscn
my_scene.avi
comparison_my_scene.avi
```

## Example

Below you can see a player character idling on an island. The character has an idling animation, that we want to
make sure doesn't change accidentally.

![Screenshot of a godot scene with a player character standing on an island](docs/img/character_island.png)

We first run this tool to generate a baseline video:

```shell
./godot-vrt-mac --command baseline --godot /Applications/Godot.app/Contents/MacOS/Godot --scenes scenes/test/test_scene_npc_cow.tscn
```
 
This results in a baseline video file being generated:

```text
scenes/test/test_scene_npc_cow.tscn
scenes/test/test_scene_npc_cow.avi
```

We can then run the tool again with the baseline video, and it will compare the generated video to the baseline video:

```shell
./godot-vrt-mac --command test --godot /Applications/Godot.app/Contents/MacOS/Godot --scenes scenes/test/test_scene_npc_cow.tscn --baseline scenes/test/test_scene_npc_cow.avi
```

```text
No difference between the baseline and the scene.
```

If the generated video is different from the baseline video, the tool will output a diff video:

```text
scenes/test/test_scene_npc_cow.tscn
scenes/test/test_scene_npc_cow.avi
scenes/test/comparison_test_scene_npc_cow.avi
```

You can then open the diff video to see what changed. It shows the two videos side by side, as well as a delta of the two.

![Screenshot of the diff video showing two scenes side by side as well as a diff view](docs/img/character_island_diff.png)

## Open questions

### Seeding randomness

- How do we seed randomness so that NPC movement is deterministic?

### More complex scenarios

- Can we script behavior into scenes? E.g. make the player character move around in a deterministic way?

### Behavior of libraries

- Does ffmpeg always generate a green screen when there is no difference? Do the pixel values ever change?

## Tools used

- ffmpeg
- go 1.23
- godot 4.4.stable

## Want an easier testing experience?

If you don't want to manage the open‑source CLI on your own, or if you just want to support this project, consider joining the early access of [Godot Foresight](https://github.com/bahrmichael/vrt-marketing)! With Godot Foresight, enjoy a hassle‑free, fully managed solution for visual regression and end-to-end testing — so you can focus on creating games.
