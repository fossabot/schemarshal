// Copyright 2017 aaharu All rights reserved.
// This source code is licensed under the BSD-style license found in
// the LICENSE file in the root directory of this source tree.

package jsonschema

import (
	"os"
	"testing"
)

func Test(t *testing.T) {
	file, _ := os.Open("../../test_data/a.json")
	defer file.Close()
	js, _ := New(file)
	jsType, _ := js.GetType()
	actual := jsType.Generate()
	if len(actual) < 1 {
		t.Errorf("got %v\n", string(actual))
	}
}
