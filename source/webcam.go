// Example program that uses blakjack/webcam library
// for working with V4L2 devices.
// The application reads frames from device and writes them to stdout
// If your device supports motion formats (e.g. H264 or MJPEG) you can
// use it's output as a video stream.
package source

import (
	//"bytes"
	"errors"
	"fmt"
	"github.com/blackjack/webcam"
	"os"
)

type WebcamSource struct {
	device  string
	done    chan chan bool
	OutChan chan []byte
}

func NewWebcamSource(device string) *WebcamSource {
	w := new(WebcamSource)
	w.device = device
	w.done = make(chan chan bool)
	return w
}

type FrameSizes []webcam.FrameSize

func selectHighestQualityFormat(cam *webcam.Webcam, supported_formats map[webcam.PixelFormat]string) (webcam.PixelFormat, webcam.FrameSize) {
	var format webcam.PixelFormat
	for pf, desc := range supported_formats {
		if desc == "Motion-JPEG" {
			format = pf
		}
	}

	frames := FrameSizes(cam.GetSupportedFrameSizes(format))
	if len(frames) <= 0 {
		panic(errors.New("No supported frame sizes available"))
	}

	size := frames[0]

	for _, value := range frames {
		if value.MaxWidth*value.MaxHeight > size.MaxWidth*size.MaxHeight {
			size.MaxWidth = value.MaxWidth
			size.MaxHeight = value.MaxHeight
		}
	}

	return format, size
}

func (s *WebcamSource) Output() chan []byte {
	s.OutChan = make(chan []byte)

	return s.OutChan
}

func (s *WebcamSource) Close() {
	donechan := make(chan bool)

	s.done <- donechan
	<-donechan
}

func waitForFrame(cam *webcam.Webcam, ready chan error) {
	var err error
	for {
		err = cam.WaitForFrame(1)
		if _, valid := err.(*webcam.Timeout); valid {
			fmt.Fprint(os.Stderr, err.Error())

			continue
		}

		break
	}
	ready <- err
}

func (s *WebcamSource) Open() error {
	if s.OutChan == nil {
		return errors.New("No output channel provided")
	}

	cam, err := webcam.Open(s.device)
	if err != nil {
		return err
	}
	defer cam.Close()

	supported_formats := cam.GetSupportedFormats()
	format, size := selectHighestQualityFormat(cam, supported_formats)

	f, w, h, err := cam.SetImageFormat(format, uint32(size.MaxWidth), uint32(size.MaxHeight))

	if err != nil {
		return err
	} else {
		fmt.Fprintf(os.Stderr, "Resulting image format: %s (%dx%d)\n", supported_formats[f], w, h)
	}

	err = cam.StartStreaming()
	if err != nil {
		return err
	}

	for {

		ready := make(chan error)
		go waitForFrame(cam, ready)

		select {
		case exit := <-s.done:
			err = <-ready
			exit <- true
			return err

		case err = <-ready:
			if err != nil {
				exit := <-s.done
				exit <- true
				return err
			}

			frame, err := cam.ReadFrame()
			if len(frame) != 0 {
				print(".")

				// create a copy in go memory space
				// avoid use after valid on rotating buffers
				local := make([]byte, len(frame))
				copy(local, frame)

				s.OutChan <- local

			} else if err != nil {
				return err
			}
		}

	}
}
