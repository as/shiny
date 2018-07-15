// Copyright 2018 as
// Copyright 2015 The Go Authors

package swizzle

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
)

var (
	rgbaslice = "abcdefghijklmnopqrstuvwxyz012345ABCDEFGHIJKLMNOPQRSTUVWXYZ6789@="
	bgraslice = "cbadgfehkjilonmpsrqtwvux0zy14325CBADGFEHKJILONMPSRQTWVUX6ZY7@98="

	supported = map[string]func(p, q []byte){
		"pure":     pureBGRA,
		"amd64.4":  bgra4sd,
		"ssse.16":  bgra16sd,
		"avx2.128": bgra128sd,
		"avx2.256": bgra256sd,
	}
)

func TestMain(m *testing.M) {
	const safe = 1024
	rgbaslice = strings.Repeat(rgbaslice, safe)
	bgraslice = strings.Repeat(bgraslice, safe)

	if !haveAVX2() {
		delete(supported, "avx2.256")
	}
	if !haveAVX() {
		delete(supported, "avx2.128")
	}
	if !haveSSSE3() {
		delete(supported, "ssse.16")
	}
	os.Exit(m.Run())
}

func TestSwizzleDistinct(t *testing.T) {
	for _, v := range []int{32, 64, 96, 128, 160, 192, 224, 256} {
		t.Run(fmt.Sprint(v), func(t *testing.T) {
			testSwizzle1(t, v, true)
		})
	}
}
func TestSwizzleOverlap(t *testing.T) {
	for _, v := range []int{32, 64, 96, 128, 160, 192, 224, 256} {
		t.Run(fmt.Sprint(v), func(t *testing.T) {
			testSwizzle1(t, v, false)
		})
	}
}

func testSwizzle1(t *testing.T, N int, distinct bool) {
	t.Helper()
	s := []byte(rgbaslice[:N])
	d := s
	if distinct {
		d = make([]byte, N, N)
	}
	want := bgraslice[:N]
	bgra256sd(s, d)
	if string(d) != want {
		t.Fatalf("have: %s\nwant: %s\n", d, want)
	}
}

func makeRGBA(len int) []byte {
	if len%4 != 0 {
		panic("makeRGBA: len % 4 != 0")
	}
	return []byte(strings.Repeat("rgba", len/4))
}

func TestBGRAShort(t *testing.T) {
	var lens = []int{
		4, 8, 12, 16, 24, 32, 48, 64, 128, 192, 256, 512,
	}
	var tab []string
	for _, v := range lens{
		tab = append(tab, rgbaslice[:v])
	}
	
	for name, fn := range supported {
		t.Run(name, func(t *testing.T) {
			for i, in := range tab {
				want := append([]byte{}, in...)
				pureBGRA(append([]byte{}, in...), want)
				p := []byte(in)
				q := make([]byte, len(p))
				fn(p, q)
				have := string(q)
				if want := string(want); have != want {
					t.Errorf("len=%d: have %q, want %q", lens[i], have, want)
				}
			}
		})
	}
}

func TestBGRARandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	var (
		p0, q0,
		p1, q1 [1024]byte
	)
	for i := range q0 {
		p0[i] = byte(r.Intn(256))
		p1[i] = p0[0]
	}
	exp := BGRASD
	ctl := pureBGRA

	for i := 0; i < 100000; i++ {
		o := r.Intn(len(p1))
		n := r.Intn(len(p1)-o) &^ 0x03
		exp(p1[o:o+n], q1[o:o+n])
		ctl(p0[o:o+n], q0[o:o+n])
		if bytes.Equal(q0[:], q1[:]) {
			continue
		}
		for j := range q1 {
			x := q1[j]
			y := q1[j]
			if x != y {
				t.Fatalf("iter %d: swizzling [%d:%d+%d]: bytes differ at offset %d (aka %d+%d): %#02x vs %#02x",
					i, o, o, n, j, o, j-o, x, y)
			}
		}
	}
}

func BenchmarkSwizzle(b *testing.B) {
	sizes := []string{"64B", "64K", "64M"}
	for name, fn := range supported {
		for i, v := range sizes {
			b.Run(name+"/"+v, func(b *testing.B) {
				d := makeRGBA(64 * 1 << uint(i))
				s := make([]byte, len(d))
				b.SetBytes(int64(len(s)))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					fn(s, d)
				}

			})
		}
	}
}
