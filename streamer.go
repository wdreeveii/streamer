// Example program that uses blakjack/webcam library
// for working with V4L2 devices.
// The application reads frames from device and writes them to stdout
// If your device supports motion formats (e.g. H264 or MJPEG) you can
// use it's output as a video stream.
// Example usage: go run stdout_streamer.go | vlc -
package main

import (
	"flag"
	"fmt"
	"streamer/source"
	"time"
)

var device string

func init() {
	flag.StringVar(&device, "device", "/dev/video0", "path to device: /dev/video0")
}

func main() {
	flag.Parse()

	var src source.VideoSource
	src = source.NewWebcamSource(device)

	println("Press Enter to start streaming")
	fmt.Scanf("\n")

	stream := src.Output()
	go src.Open()

	go func() {
		for {
			data := <-stream
			fmt.Println(data)
		}
	}()

	t := time.After(5 * time.Second)

	<-t
	fmt.Println("Timeout done")
	src.Close()
	return

}
