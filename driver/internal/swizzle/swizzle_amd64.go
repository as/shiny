// Copyright 2018 as
// Copyright 2015 The Go Authors
package swizzle

// haveSSSE3 is true on CPUs with SSSE3 (i.e. PSHUFB).
func haveSSSE3() bool

// haveAVX2 is true on CPUs with AVX2 (i.e. VPSHUFB).
func haveAVX2() bool

var (
	useAVX2 = haveAVX2()
	useSSSE3 = haveSSSE3()
	useBGRA4 = true
	
	swizzler func(p, q []byte)
)

func init() {
	swizzler = bgra4sd
	if useSSSE3 {
		swizzler = bgra16sd
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