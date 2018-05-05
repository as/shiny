// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.	EX

package swizzle

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
)

var (
	rgbaslice = "rgbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargbargba"
	bgraslice = "bgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgrabgra"
)

const alpha = "abcdefghijklmnopqrstuvwxyz012345ABCDEFGHIJKLMNOPQRSTUVWXYZ6789@"

func testSwizzleNSD(t *testing.T, N int) {
	d := make([]byte, N, N)
	s := []byte(rgbaslice[:N])
	want := bgraslice[:N]
	bgra256sd(s, d)
	if string(d) != want {
		t.Fatalf("have: %s\nwant: %s\n", d, want)
	}
}

func TestSwizzle32SD(t *testing.T)  { testSwizzleNSD(t, 32) }
func TestSwizzle64SD(t *testing.T)  { testSwizzleNSD(t, 64) }
func TestSwizzle96SD(t *testing.T)  { testSwizzleNSD(t, 96) }
func TestSwizzle128SD(t *testing.T) { testSwizzleNSD(t, 128) }
func TestSwizzle160SD(t *testing.T) { testSwizzleNSD(t, 160) }
func TestSwizzle192SD(t *testing.T) { testSwizzleNSD(t, 192) }
func TestSwizzle224SD(t *testing.T) { testSwizzleNSD(t, 224) }

func TestSwizzle256SD(t *testing.T) {
	d := make([]byte, 256, 256)
	s := []byte(rgbaslice)
	want := bgraslice
	bgra256sd(s, d)
	if string(d) != want {
		t.Fatalf("have: %s\nwant: %s\n", d, want)
	}
}

func BenchmarkSwizzle256SD(b *testing.B) {
	d := make([]byte, 256, 256)
	s := []byte(rgbaslice)
	b.SetBytes(int64(len(s)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bgra256sd(s, d)
	}
}

func BenchmarkSwizzle64KSD(b *testing.B) {
	d := make([]byte, 256*256, 256*256)
	s := []byte(strings.Repeat(rgbaslice, 256))
	b.SetBytes(int64(len(s)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bgra256sd(s, d)
	}
}

func TestBGRAShortInput(t *testing.T) {
	const s = "012.456.89A.CDE.GHI.KLM.O"
	testCases := []string{
		0: "012.456.89A.CDE.GHI.KLM.O",
		1: "210.456.89A.CDE.GHI.KLM.O",
		2: "210.654.89A.CDE.GHI.KLM.O",
		3: "210.654.A98.CDE.GHI.KLM.O",
		4: "210.654.A98.EDC.GHI.KLM.O",
		5: "210.654.A98.EDC.IHG.KLM.O",
		6: "210.654.A98.EDC.IHG.MLK.O",
	}
	for i, want := range testCases {
		b := []byte(s)
		BGRA(b[:4*i])
		got := string(b)
		if got != want {
			t.Errorf("i=%d: got %q, want %q", i, got, want)
		}
		changed := got != s
		wantChanged := i != 0
		if changed != wantChanged {
			t.Errorf("i=%d: changed=%t, want %t", i, changed, wantChanged)
		}
	}
}

func TestBGRARandomInput(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	fastBuf := make([]byte, 1024)
	slowBuf := make([]byte, 1024)
	for i := range fastBuf {
		fastBuf[i] = uint8(r.Intn(256))
	}
	copy(slowBuf, fastBuf)

	for i := 0; i < 100000; i++ {
		o := r.Intn(len(fastBuf))
		n := r.Intn(len(fastBuf)-o) &^ 0x03
		BGRA(fastBuf[o : o+n])
		pureGoBGRA(slowBuf[o : o+n])
		if bytes.Equal(fastBuf, slowBuf) {
			continue
		}
		for j := range fastBuf {
			x := fastBuf[j]
			y := slowBuf[j]
			if x != y {
				t.Fatalf("iter %d: swizzling [%d:%d+%d]: bytes differ at offset %d (aka %d+%d): %#02x vs %#02x",
					i, o, o, n, j, o, j-o, x, y)
			}
		}
	}
}

func pureGoBGRA(p []byte) {
	if len(p)%4 != 0 {
		return
	}
	for i := 0; i < len(p); i += 4 {
		p[i+0], p[i+2] = p[i+2], p[i+0]
	}
}

func benchmarkBGRA(b *testing.B, f func([]byte)) {
	const w, h = 1920, 1080 // 1080p RGBA.
	buf := make([]byte, 4*w*h)
	b.SetBytes(int64(len(buf)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(buf)
	}
}

func BenchmarkBGRA(b *testing.B)       { benchmarkBGRA(b, BGRA) }
func BenchmarkPureGoBGRA(b *testing.B) { benchmarkBGRA(b, pureGoBGRA) }
func BenchmarkBGRA256(b *testing.B)    { benchmarkBGRA(b, bgra256) }
func BenchmarkBGRA64(b *testing.B)     { benchmarkBGRA(b, bgra64) }
func BenchmarkBGRA32(b *testing.B)     { benchmarkBGRA(b, bgra32) }
func BenchmarkBGRA16(b *testing.B)     { benchmarkBGRA(b, bgra16) }
func BenchmarkBGRA4(b *testing.B)      { benchmarkBGRA(b, bgra4) }
