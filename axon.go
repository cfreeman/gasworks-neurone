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

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/huin/goserial"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	waitLength       = 30.0
	waitTimeout      = 400.0
	startupLength    = 20.0
	cooldownLength   = 4.0
	powerupLength    = 3.0
	powerupThreshold = 0.45
	nanoToSeconds    = 1000000000.0
)

type Neurone struct {
	energy   float32
	deltaE   chan float32
	duration float64
	start    int64
	config   Configuration
}

type stateFn func(neurone Neurone, serialPort io.ReadWriteCloser) (sF stateFn, newNeurone Neurone)

// sendArduinoCommand transmits a new command over the numonated serial port to the arduino. Returns an
// error on failure. Each command is identified by a single byte and may take one argument (a float).
func sendArduinoCommand(command byte, argument float32, serialPort io.ReadWriteCloser) error {
	if serialPort == nil {
		return nil
	}

	// Package argument for transmission
	bufOut := new(bytes.Buffer)
	err := binary.Write(bufOut, binary.LittleEndian, argument)
	if err != nil {
		return err
	}

	// Transmit command and argument down the pipe.
	for _, v := range [][]byte{[]byte{command}, bufOut.Bytes()} {
		_, err = serialPort.Write(v)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateArduinoEnergy transmits a new energy level over the nominated serial port to the arduino. Returns an error
// on failure, nil otherwise. Arduino code takes the energy level and turns it into a lighting sequence.
func updateArduinoEnergy(energy float32, serialPort io.ReadWriteCloser) error {
	return sendArduinoCommand('e', energy, serialPort)
}

// cooldownArduino transmits updates the cooldown lighting sequence on the arduino. Returns an error on failure, nil
// otherwise.
func cooldownArduino(energy float32, serialPort io.ReadWriteCloser) error {
	return sendArduinoCommand('c', energy, serialPort)
}

// powerupArduino puts the arduino into a short powerup animation, indicating that the neurone has recieved a
// large burst of energy. Returns an error on failure, nil otherwise.
func powerupArduino(serialPort io.ReadWriteCloser) error {
	return sendArduinoCommand('p', 0.0, serialPort)
}

// findArduino looks for the file that represents the arduino serial connection. Returns the fully qualified path
// to the device if we are able to find a likely candidate for an arduino, otherwise an empty string if unable to
// find an arduino device.
func findArduino() string {
	contents, _ := ioutil.ReadDir("/dev")

	// Look for the arduino device
	for _, f := range contents {
		if strings.Contains(f.Name(), "tty.usbserial") ||
			strings.Contains(f.Name(), "ttyUSB") {
			return "/dev/" + f.Name()
		}
	}

	// Have not been able to find the device.
	return ""
}

// wait puts the neurone in a holding state untill all the raspberry pi's have started up. Then
// puts all the neurones through a non-interactive animated sequence.
func wait(neurone Neurone, serialPort io.ReadWriteCloser) (sF stateFn, newNeurone Neurone) {
	// Calculate how many seconds have elapsed since this cooldown state started.
	dt := float64(time.Now().UnixNano()-neurone.start) / nanoToSeconds

	if neurone.config.MasterNeurone {
		// Drain off an ignore energy from the dendrites.
		select {
		case <-neurone.deltaE:
		case <-time.After(5 * time.Millisecond):
		}

		if dt >= neurone.duration {
			for _, adjacent := range neurone.config.AllNeurones {
				buf := new(bytes.Buffer)
				fmt.Fprintf(buf, "%s?e=%f", adjacent.Address, adjacent.Transfer)

				address := buf.String()
				go http.Get(address)
				fmt.Printf("INFO: S[" + address + "]\n")
			}

			return startup, Neurone{0.0, neurone.deltaE, startupLength, time.Now().UnixNano(), neurone.config}
		}
	} else {
		// Neurone is not the master, wait to be notified by the master before startup.
		de := <-neurone.deltaE
		if de < -0.5 {
			return startup, Neurone{0.0, neurone.deltaE, startupLength, time.Now().UnixNano(), neurone.config}
		} else if dt >= waitTimeout {

			// If for some reason we don't get notified by the master neurone to enter the animation, just jump
			// straight to interactive mode.
			return accumulate, Neurone{0.0, neurone.deltaE, 0.0, time.Now().UnixNano(), neurone.config}
		}
	}

	return wait, Neurone{-2.0, neurone.deltaE, neurone.duration, neurone.start, neurone.config}
}

// startup puts the neurone through a non-interactive animated sequence before entering the animated
// mode.
func startup(neurone Neurone, serialPort io.ReadWriteCloser) (sF stateFn, newNeurone Neurone) {

	// The warmup animation and cooldown animation are the same, just over different durations.
	return cooldown(neurone, serialPort)
}

// accumulate pulls energy off the dendrites and accumulates it within the neurone. When the neurone reaches
// critical it fires into the axon (the web dendrites of adjacent neurones) and enters the cooldown state.
func accumulate(neurone Neurone, serialPort io.ReadWriteCloser) (sF stateFn, newNeurone Neurone) {
	de := <-neurone.deltaE
	newEnergy := neurone.energy + de

	// Neurone has reached threshold. Fire axon.
	if newEnergy > 1.0 {
		// Axon fires into the web dendrites of adjacent neurones.
		for _, adjacent := range neurone.config.AdjacentNeurones {
			buf := new(bytes.Buffer)
			fmt.Fprintf(buf, "%s?e=%f", adjacent.Address, adjacent.Transfer)

			address := buf.String()
			go http.Get(address)
			fmt.Printf("INFO: a[" + address + "]\n")
		}

		fmt.Printf("INFO: cooldown!\n")
		return cooldown, Neurone{newEnergy, neurone.deltaE, cooldownLength, time.Now().UnixNano(), neurone.config}
	}

	// If the energy level jumps by a large amount, another neuron has fired. Run a power
	// up flash animation.
	if de > powerupThreshold {
		fmt.Printf("INFO: powerup!\n")

		powerupArduino(serialPort)
		return powerup, Neurone{newEnergy, neurone.deltaE, powerupLength, time.Now().UnixNano(), neurone.config}
	}

	updateArduinoEnergy(newEnergy, serialPort)
	return accumulate, Neurone{newEnergy, neurone.deltaE, 0.0, time.Now().UnixNano(), neurone.config}
}

// calcDt calculates the change in seconds since an animation was started.
func calcDt(neurone Neurone) float64 {
	// Drain off and ignore changes in energy from the dendrites.
	select {
	case <-neurone.deltaE:
	case <-time.After(250 * time.Millisecond):
	}

	// Calculate how many seconds have elapsed since this cooldown state started.
	return float64(time.Now().UnixNano()-neurone.start) / nanoToSeconds
}

// powerup allows the neurone to display a large jump in energy to the neurone. It pauses the accumlation
// by the nominated duration before starting accumulation of energy from the dendrites again.
func powerup(neurone Neurone, serialPort io.ReadWriteCloser) (sF stateFn, newNeurone Neurone) {
	dt := calcDt(neurone)

	if dt >= neurone.duration {
		return accumulate, Neurone{neurone.energy, neurone.deltaE, 0.0, time.Now().UnixNano(), neurone.config}
	}

	return powerup, neurone
}

// cooldown allows the neurone to cooldown after firing into the axon, it pauses accumulation by the
// nominated duration before starting accumulation of energy from the dendrites again.
func cooldown(neurone Neurone, serialPort io.ReadWriteCloser) (sF stateFn, newNeurone Neurone) {
	dt := calcDt(neurone)

	// LERP neurone energy from -1.0 to 0.0 over the duration of the cooldown.
	newEnergy := float32(dt / neurone.duration)

	// If the time elapsed is longer than the duration of the cooldown, enter the accumulate state.
	if dt >= neurone.duration {
		return accumulate, Neurone{0.0, neurone.deltaE, 0.0, time.Now().UnixNano(), neurone.config}
	}

	cooldownArduino(newEnergy, serialPort)
	return cooldown, Neurone{newEnergy, neurone.deltaE, neurone.duration, neurone.start, neurone.config}
}

// Axon listens to the dentrites on the deltaE channel, and embodies an artificial neurone. When the energy
// of the neurone reaches a maximum, it fires into the axon (the web dendites of adjacent neurones).
func axon(deltaE chan float32, config Configuration) {
	// Find the device that represents the arduino serial connection.
	c := &goserial.Config{Name: findArduino(), Baud: 9600}
	s, _ := goserial.OpenPort(c)

	// When connecting to an older revision arduino, you need to wait a little while it resets.
	time.Sleep(1 * time.Second)

	neurone := Neurone{-2.0, deltaE, waitLength, time.Now().UnixNano(), config}
	state := wait

	for {
		state, neurone = state(neurone, s)

		fmt.Printf("INFO: e[%f]\n", neurone.energy)
	}
}
