// Copyright 2019-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// see the example at https://golang.org/pkg/image/png/#Decode

package main

import "github.com/u-root/u-root/pkg/fb"
import "github.com/u-root/tpmtotp/pkg/token"

func main() {
  _, qrCode, _ := token.CreateQRSecretTOTP()
  fb.DrawScaledImageAt(qrCode, 200, 160, 3)
}
