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

const COOLDOWN_LENGTH = 4.0
const NANO_TO_SECONDS = 1000000000.0

type Neuron struct {
	energy   float32
	deltaE   chan float32
	duration float64
	start    int64
	config   Configuration
}

type stateFn func(neuron Neuron) (sF stateFn, newNeuron Neuron)

// updateArduinoEnergy transmits a new energy level over the nominated serial port to the arduino. Returns an error
// on failure, nil otherwise. Arduino code takes the energy level and turns it into a lighting sequence.
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

// accumulate pulls energy off the dendrites and accumulates it within the neuron. When the neuron reaches
// critical it fires into the axon (the web dendrites of adjacent neurons) and enters the cooldown state.
func accumulate(neuron Neuron) (sF stateFn, newNeuron Neuron) {
	newEnergy := neuron.energy + <-neuron.deltaE

	// Neuron has reached threshold. Fire axon.
	if newEnergy > 1.0 {
		// Axon fires into the web dendrites of adjacent neurons.
		for _, adjacent := range neuron.config.AdjacentNeurons {
			buf := new(bytes.Buffer)
			fmt.Fprintf(buf, "%s?e=%f", adjacent.Address, adjacent.Transfer)

			address := buf.String()
			go http.Get(address)
			fmt.Printf("Firing into " + address + "\n")
		}

		return cooldown, Neuron{-1.0, neuron.deltaE, COOLDOWN_LENGTH, time.Now().UnixNano(), neuron.config}
	}

	return accumulate, Neuron{newEnergy, neuron.deltaE, 0.0, time.Now().UnixNano(), neuron.config}
}

// cooldown allows the neuron to cooldown after firing into the axon, it pauses accumulation by the
// nominated duration before starting accumulation of energy from the dendrites again.
func cooldown(neuron Neuron) (sF stateFn, newNeuron Neuron) {
	<-neuron.deltaE //drain off and ignore changes in energy from the dendrites.

	// Calculate how many seconds have elapsed since this cooldown state started.
	dt := float64(time.Now().UnixNano()-neuron.start) / NANO_TO_SECONDS

	// LERP neuron energy from -1.0 to 0.0 over the duration of the cooldown.
	newEnergy := float32(dt/neuron.duration) - 1.0

	// If the time elapsed is longer than the duration of the cooldown, enter the accumulate state.
	if dt >= neuron.duration {
		return accumulate, Neuron{newEnergy, neuron.deltaE, 0.0, time.Now().UnixNano(), neuron.config}
	}

	return cooldown, Neuron{newEnergy, neuron.deltaE, COOLDOWN_LENGTH, neuron.start, neuron.config}
}

// Axon listens to the dentrites on the deltaE channel, and embodies an artificial neuron. When the energy
// of the neuron reaches a maximum, it fires into the axon (the web dendites of adjacent neurons).
func Axon(deltaE chan float32, config Configuration) {
	// Find the device that represents the arduino serial connection.
	c := &goserial.Config{Name: findArduino(), Baud: 9600}
	s, _ := goserial.OpenPort(c)

	// When connecting to an older revision arduino, you need to wait a little while it resets.
	time.Sleep(1 * time.Second)

	newNeuron := Neuron{0.0, deltaE, 0.0, time.Now().UnixNano(), config}
	oldNeuron := newNeuron
	state := accumulate

	for true {
		state, newNeuron = state(oldNeuron)

		// If we have a valid serial connection to an arduino, update the energy level.
		if s != nil && newNeuron.energy != oldNeuron.energy {
			updateArduinoEnergy(newNeuron.energy, s)
		}

		fmt.Printf("e: %f\n", newNeuron.energy)
		oldNeuron = newNeuron
	}
}
