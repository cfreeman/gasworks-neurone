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

import "testing"

func TestMissingConfiguration(t *testing.T) {
	config, err := ParseConfiguration("foo")
	if err == nil {
		t.Errorf("error not raised for invalid configuration file.")
	}

	if config.OpticalFlowScale != 1000.0 {
		t.Errorf("incorrect default optical flow scale.")
	}

	if config.MovementThreshold != 1.0 {
		t.Errorf("incorrect default movement threshold.")
	}

	if config.ListenAddress != ":8080" {
		t.Errorf("incorrect default listen address")
	}

	if len(config.AdjacentNeurons) != 0 {
		t.Errorf("incorrect default list of AdjacentNeurons")
	}
}

func TestValidConfiguration(t *testing.T) {
	config, err := ParseConfiguration("testdata/test-config.json")
	if err != nil {
		t.Errorf("returned error when parsing valid configuration file")
	}

	if config.OpticalFlowScale != 0.23 {
		t.Errorf("parsed incorrect value for optical flow scale.")
	}

	if config.MovementThreshold != 0.10 {
		t.Errorf("parsed incorrect value for movement threshold.")
	}

	if config.ListenAddress != "10.1.1.1:8080" {
		t.Errorf("parsed incorrect listen address")
	}

	if len(config.AdjacentNeurons) != 2 {
		t.Errorf("Did not parse enough adjacent neurons from the configuration")
	}

	if config.AdjacentNeurons[0].Address != "http://10.1.1.5:8080/" && config.AdjacentNeurons[0].Transfer != 0.8 {
		t.Errorf("Did not correctly parse the first transfer neuron")
	}

	if config.AdjacentNeurons[1].Address != "http://10.1.1.4:8080/" && config.AdjacentNeurons[1].Transfer != 0.2 {
		t.Errorf("Did not correctly parse the first transfer neuron")
	}
}
