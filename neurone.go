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
	"fmt"
	"os"
)

// We have two different kinds of dendrite. One on a webcam, that increases the energy of the neurone
// when motion is detected in the webcam and another dendrite that listens for when other neurones fire.
//
// We have one axon which transmits energy to adjacent neurones.
func main() {
	fmt.Printf("Gasworks neurone\n")

	configFile := "/home/pi/gasworks/neurone/bin/gasworks.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	configuration, _ := parseConfiguration(configFile)
	deltaE := make(chan float32)

	fmt.Println("Starting Axon")
	go axon(deltaE, configuration)

	fmt.Println("Starting Web Dendrite")
	go dendriteWeb(deltaE, configuration)

	fmt.Println("Starting Camera Dendrite")
	dendriteCam(deltaE, configuration)

	// Make sure we block if no webcam is found and DendriteCam returns straight away.
	select {}
}
