// Copyright 2017 aaharu All rights reserved.
// This source code is licensed under the BSD-style license found in
// the LICENSE file in the root directory of this source tree.

package version

import "fmt"

// Version of schemarshal
const Version = "1.2.0"

// String return `<name> <version>`
func String() string {
	return fmt.Sprintf("schemarshal %s", Version)
}
