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
	"time"
)

// We have two different kinds of dendrite. One on a webcam, that increases the energy of the neurone
// when motion is detected in the webcam and another dendrite that listens for when other neurones fire.
//
// We have one axon which transmits energy to adjacent neurones.
func main() {
	fmt.Printf("Gasworks neurone\n")
	configuration, _ := ParseConfiguration("gasworks.json")
	delta_e := make(chan float32)

	go Axon(delta_e, configuration)

	go DendriteWeb(delta_e, configuration)
	DendriteCam(delta_e, configuration)

	for true {
		time.Sleep(1 * time.Second)
	}
}
