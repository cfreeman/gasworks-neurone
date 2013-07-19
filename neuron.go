/*
 * Copyright (c) Clinton Freeman 2013
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
 * associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute,
 * sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or
 * substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
 * NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */
package main

// #cgo CFLAGS: -Wno-error -I/opt/local/include -I/opt/local/include/opencv
// #cgo LDFLAGS: -L/opt/local/lib -lopencv_highgui -lopencv_core -lopencv_video
// #include "cv.h"
// #include "highgui.h"
import "C"

import (
		"fmt"
		"log"
		"github.com/huin/goserial"
		"bytes"
		"encoding/binary"
		// "time"
		"io"
		"unsafe"
		"math"
)

// updateArduinoEnergy transmits a new energy level over the nominated serial port to the arduino. Returns an error
// on failure, nil otherwise.
func updateArduinoEnergy(energy float32, serialPort io.ReadWriteCloser) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, energy)
	if err != nil {
		return err
	}

	_, err = serialPort.Write(buf.Bytes())
    if err != nil {
		return err
    }

    return nil
}

func calcDeltaEnergy(flow *C.IplImage) float64 {
	var i C.int
	var dx, dy float64

	// Accumulate the change in flow across all the pixels.
	totalPixels := flow.width * flow.height
	for i = 0; i < totalPixels; i++ {
		value := C.cvGet2D(unsafe.Pointer(flow), i / flow.width, i % flow.width)
		dx += math.Abs(float64(value.val[0]))
		dy += math.Abs(float64(value.val[1]))
	}

	// The magnitude of accumulated flow forms our change in energy for the frame.
	deltaE := math.Sqrt((dx * dx) + (dy * dy))

	// Clamp the energy to start at 0 for 'still' frames with little/no motion.
	//fmt.Printf("%f, %f: %f\n", dx, dy, deltaE)
	deltaE = math.Max(0.0, (deltaE - 250000))

	// Scale the flow to be less than 0.1 for 'active' frames with lots of motion.
	//fmt.Printf("%f\n", deltaE)
	deltaE = deltaE / 10000000
	//fmt.Printf("%f\n", deltaE)

	return deltaE
}

func main() {
	fmt.Printf("Gasworks neruon.\n")

	c := &goserial.Config{Name: "/dev/tty.usbserial-A1017HU2", Baud: 9600}
	s, err := goserial.OpenPort(c)
    if err != nil {
		log.Fatal(err)
    }

    var energy float32
    energy = 0.0

 //    // When connecting to an arduino, you need to wait a little while it resets.
	// time.Sleep(1 * time.Second)
	// updateArduinoEnergy(0.8, s)

	// prev = cvQueryFrame
	// while not done.
	// 	next = cvQueryFrame
	// 	cvCalcOpticalFlowFarneback
	//	normalise optical flow.
	//	use optical flow as an increase to energy level.
	//  update arduino energy level.
	//  prev = next
	//
	//
	//
	// Calculate optical flow for each pixel
	// http://docs.opencv.org/modules/video/doc/motion_analysis_and_object_tracking.html#void calcOpticalFlowFarneback(InputArray prev, InputArray next, InputOutputArray flow, double pyr_scale, int levels, int winsize, int iterations, int poly_n, double poly_sigma, int flags)
	// Farneback.
	//
	// normalise optical flow for each pixel into a single scalar value for whole frame.
	// this is the change in energy level for the neuron.
	//

	// file := C.CString("foo.png")
	// C.cvSaveImage(file, unsafe.Pointer(frame), nil)
	// C.free(unsafe.Pointer(file))

	camera := C.cvCaptureFromCAM(-1)
	prev := C.cvCloneImage(C.cvQueryFrame(camera))

	flow := C.cvCreateImage(C.cvSize(prev.width, prev.height), C.IPL_DEPTH_32F, 2)
	prevG := C.cvCreateImage(C.cvSize(prev.width, prev.height), C.IPL_DEPTH_8U, 1)
	nextG := C.cvCreateImage(C.cvSize(prev.width, prev.height), C.IPL_DEPTH_8U, 1)

	for i := 0; i < 500; i++ {
		C.cvGrabFrame(camera)
		next := C.cvCloneImage(C.cvQueryFrame(camera))

		// Convert the captured frames to greyscale.
		C.cvConvertImage(unsafe.Pointer(prev), unsafe.Pointer(prevG), 0)
		C.cvConvertImage(unsafe.Pointer(next), unsafe.Pointer(nextG), 0)

		C.cvCalcOpticalFlowFarneback(unsafe.Pointer(prevG), unsafe.Pointer(nextG), unsafe.Pointer(flow), 0.5, 2, 5, 2, 5, 1.1, 0)

		delta := float32(calcDeltaEnergy(flow))
		energy += delta
		fmt.Printf("energy: %f %f\n", energy, delta)
		updateArduinoEnergy(energy, s)

		C.cvReleaseImage(&prev)
		prev = next
	}

    C.cvReleaseImage(&prev)

    C.cvReleaseImage(&nextG)
    C.cvReleaseImage(&prevG)
    C.cvReleaseImage(&flow)
    C.cvReleaseCapture(&camera)


	// Make sure the port stays open, otherwise the arduino will reset as soon as it discconects.
	// file := C.CString("a.png")
	// C.cvSaveImage(file, unsafe.Pointer(prev), nil)
	// C.free(unsafe.Pointer(file))

	// file = C.CString("b.png")
	// C.cvSaveImage(file, unsafe.Pointer(next), nil)
	// C.free(unsafe.Pointer(file))
}
