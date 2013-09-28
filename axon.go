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

type NeuronState struct {
	Name     string
	Duration float64
	Start    int64
}

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

func Axon(delta_e chan float32, config Configuration) {
	state := [2]NeuronState{{"ACCUMULATING", 0, 0}, {"COOLDOWN", 4.0, 0}}
	currentState := 0

	// Find the device that represents the arduino serial connection.
	c := &goserial.Config{Name: findArduino(), Baud: 9600}
	s, _ := goserial.OpenPort(c)

	// When connecting to an older revision arduino, you need to wait a little while it resets.
	time.Sleep(1 * time.Second)

	// The energy level of the neuron.
	var energy float32 = 0.0
	var oldEnergy float32 = 0.0

	for i := 0; i < 500; i++ {
		de := <-delta_e

		if currentState == 0 {
			energy += de

			// Neuron has reached threshold. Fire axon.
			if energy > 1.0 {

				// Axon fires into the web dendrites of adjacent neurons.
				for _, n := range config.AdjacentNeurons {
					buf := new(bytes.Buffer)
					fmt.Fprintf(buf, "%s?e=%f", n.Address, n.Transfer)

					address := buf.String()
					go http.Get(address)
					fmt.Printf("Firing into " + address + "\n")
				}

				// Flash this neuron when it fires.
				currentState = 1
				state[currentState].Start = time.Now().UnixNano()
				energy = -1.0
			}
		} else {

			dt := float64(time.Now().UnixNano()-state[currentState].Start) / 1000000000.0
			energy = float32(dt/state[currentState].Duration) - 1.0

			if dt >= state[currentState].Duration {
				currentState = 0
			}
		}

		// If we have a valid serial connection to an arduino, update the energy level.
		if s != nil && oldEnergy != energy {
			updateArduinoEnergy(energy, s)
		}
		oldEnergy = energy

		fmt.Printf("Energy level %f %f\n", energy, de)
	}
}
