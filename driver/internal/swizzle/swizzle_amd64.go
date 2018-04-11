// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package swizzle

// haveSSSE3 returns whether the CPU supports SSSE3 instructions (i.e. PSHUFB).
//
// Note that this is SSSE3, not SSE3.
func haveSSSE3() bool
func haveAVX2() bool

var useBGRA32 = haveAVX2()
var useBGRA16 = haveSSSE3()

const useBGRA4 = true

// go test
func bgra256sd(p, q []byte)

func bgra256(p []byte)
func bgra64(p []byte)
func bgra32(p []byte)
func bgra16(p []byte)
func bgra4(p []byte)
