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
// #cgo LDFLAGS: -L/opt/local/lib -lopencv_highgui -lopencv_core
// #include "cv.h"
// #include "highgui.h"
import "C"

import (
		"fmt"
		"log"
		"github.com/huin/goserial"
		"bytes"
		"encoding/binary"
		"time"
		"io"
		// "unsafe"
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

func main() {
	fmt.Printf("Gasworks neruon.\n")

	c := &goserial.Config{Name: "/dev/tty.usbserial-A1017HU2", Baud: 9600}
	s, err := goserial.OpenPort(c)
    if err != nil {
		log.Fatal(err)
    }

    // When connecting to an arduino, you need to wait a little while it resets.
	time.Sleep(1 * time.Second)
	updateArduinoEnergy(0.8, s)

	camera := C.cvCaptureFromCAM(-1)
	C.cvQueryFrame(camera)

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

	C.cvReleaseCapture(&camera)

    for {
    	// Make sure the port stays open, otherwise the arduino will reset as soon as it discconects.
    }
}