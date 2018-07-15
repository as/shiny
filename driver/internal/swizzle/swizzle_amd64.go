// Copyright 2018 as
// Copyright 2015 The Go Authors
package swizzle

func haveSSSE3() bool
func haveAVX() bool
func haveAVX2() bool

var (
	useBGRA4 = true
	useSSSE3 = haveSSSE3()
	useAVX = haveAVX()
	useAVX2 = haveAVX2()
)

func init() {
	swizzler = bgra4sd
	if useSSSE3 {
		swizzler = bgra16sd
	}
	if useAVX {
		swizzler = bgra128sd
	}
	if useAVX2{
		swizzler = bgra256sd
	}
}

func BGRASD(p, q []byte) {
	if len(p) < 4 {
		return
	}
	swizzler(p,q)
}
var BGRA = BGRASD

// AVX2
func bgra256sd(p, q []byte)	// swizzle_amd64.s:/bgra256sd/
func bgra128sd(p,q []byte)	// swizzle_amd64.s:/bgra128sd/

// SSSE
func bgra16sd(p,q []byte)	// swizzle_amd64.s:/bgra16sd/

// AMD64
func bgra4sd(p,q []byte)	