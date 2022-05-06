# Download Progress
This example was built based on [this](https://github.com/charmbracelet/bubbles/discussions/127) discussion.
This example demonstrates how to download a file from a given URL, and show its progress with a [progress Bubble](https://github.com/charmbracelet/bubbles/).
The status of the download is updated with [`io.TeeReader`](https://pkg.go.dev/io#TeeReader).
This calls `Write` which is where we send the updated status with `Program#Send()` 

## How to Run
`go build .` in this directory on your machine (in examples/download-progress)
then run `./download-progress --url="https://download.blender.org/demo/color_vortex.blend"` this can be whatever file you'd like to download. 
Note: the current version will not show a TUI for downloads that do not provide the ContentLength header field.

