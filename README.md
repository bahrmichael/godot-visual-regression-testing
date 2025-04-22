# Godot Visual Regression Testing (Godot VRT)

Inspired by [Factorio's visual regression testing](https://www.youtube.com/watch?v=LXnyTZBmfXM), this is 
a test runner for visual regression testing and end-to-end testing with Godot scenes.

You can start using it today!

Run `godot-vrt --help` to see the available commands and their options.

## Concept

The idea is that we can generate videos from Godot scenes, and then compare those videos to a baseline. A baseline is a
video that we generated previously, and which we are confident is correct.

If there's a difference between the generated video and the baseline, we know that something has changed. We can then
use this information to determine if it's an intended change, or if it's something we need to fix.

There's a lot that can go on in a scene. We can look at it manually, but it's a lot of work. Instead, we can let
the computer find pixels that changed, and give us a video with all those changes.

## Quick Start

### Prerequisites

You need to run this on a computer that is equipped with a graphics card, and has Godot as well as ffmpeg installed. 

Headless servers (such as GitHub action runners) are not supported because they lack the required hardware. If you're
interested in paying someone to run the tests on GPU powered servers and integrate them into your CI, [please get in touch](https://forms.gle/VopXGutf3NSKrRXC8).

1. Install [Godot 4.4.1 Stable](https://godotengine.org/download)
2. Install ffmpeg (if you have homebrew on macOS: `brew install ffmpeg`)

### Download the executable

You can get it here: https://github.com/bahrmichael/godot-visual-regression-testing/releases

You can also build it yourself by installing go 1.23 and running `go build .` in the root of this repository.

### Create a baseline video

Assume you have a project with a single scene called `my_scene.tscn`.

Take the executable, move it to the root of your project, and run this command:

```
godot-vrt baseline --godot path_to_godot_binary --scenes my_scene.tscn
```

<details>
  <summary>How do I find the path to godot?</summary>

    Windows: Probably where you unpacked it.

    macOS: Most likely at `/Applications/Godot.app/Contents/MacOS/Godot`. You can just copy paste that into the command line and hit return.

    Linux: If you installed godot into your path, you can run `which godot`.
</details>

This will evaluate the `--scenes` parameter (you can use a glob expression) and generate a video file for each scene.

```
my_scene.tscn
my_scene.avi
```

We recommend that you put the testing scenes into a separate folder to keep them neatly organized, and make it
easier to pick the right scenes. For example, we might have a folder called `vrt` that holds all the scenes used for
visual regression testing:

```
godot-vrt baseline --godot path_to_godot_binary --scenes vrt/*.tscn
```

### Compare against a baseline video

Once you have a baseline video, you can pass it to a test run. When you run the `test` command, it expects to find a baseline
video for each scene. E.g. `test_scene.tscn` requires a baseline video called `test_scene.avi`. Remember that you can generate
baseline videos with the `baseline` command shown above.

```
godot-vrt test --godot path_to_godot_binary --scenes my_scene.tscn --baseline my_scene.avi
```

Again, with test scenes in a separate `vrt` folder the command looks like this:

```
godot-vrt test --godot path_to_godot_binary --scenes vrt/*.tscn --baseline vrt/*.avi
```

If any scene fails the test, it will exit with a non-zero exit code and produce a diff video next to the test scene and baseline video.

```
my_scene.tscn
my_scene.avi
diff/my_scene_<some timestamp>.avi
```

## Example

Below you can see a player character idling on an island. The character has an idling animation, that we want to
make sure doesn't change accidentally.

![Screenshot of a godot scene with a player character standing on an island](docs/img/character_island.png)

We first run this tool to generate a baseline video:

```shell
godot-vrt baseline --godot /Applications/Godot.app/Contents/MacOS/Godot --scenes scenes/test/test_scene_npc_cow.tscn
```
 
This results in a baseline video file being generated:

```text
scenes/test/test_scene_npc_cow.tscn
scenes/test/test_scene_npc_cow.avi
```

We can then run the tool again with the baseline video, and it will compare the generated video to the baseline video:

```shell
godot-vrt test --godot /Applications/Godot.app/Contents/MacOS/Godot --scenes scenes/test/test_scene_npc_cow.tscn --baseline scenes/test/test_scene_npc_cow.avi
```

```text
✅ All tests passed
```

If the generated video is different from the baseline video, the tool will output a diff video:

```text
scenes/test/test_scene_npc_cow.tscn
scenes/test/test_scene_npc_cow.avi
vrt-results/scenes/test/test_scene_npc_cow_12345678.avi
```

You can then open the diff video to see what changed. It shows the two videos side by side, as well as a delta of the two.

![Screenshot of the diff video showing two scenes side by side as well as a diff view](docs/img/character_island_diff.png)

## Supporting this project

You can support this project by
- sharing your experience with me (through an issue on this repository), or with other devs,
- contributing to this project,
- or by [getting in touch for the managed service](https://forms.gle/VopXGutf3NSKrRXC8).
Hello World
