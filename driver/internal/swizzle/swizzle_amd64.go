// Copyright 2018 as
// Copyright 2015 The Go Authors
package swizzle

// haveSSSE3 is true on CPUs with SSSE3 (i.e. PSHUFB).
func haveSSSE3() bool

// haveAVX2 is true on CPUs with AVX2 (i.e. VPSHUFB).
func haveAVX2() bool

var (
	useBGRA32 = haveAVX2()
	useBGRA16 = haveSSSE3()
	useBGRA4 = true
	
	swizzler func(p, q []byte)
)

func init() {
	swizzler = bgra4sd
	if useBGRA16 {
		swizzler = bgra16sd
	}
	if useBGRA32 {
		swizzler = bgra256sd
	}
}

func BGRASD(p, q []byte) {
	if len(p) < 4 {
		return
	}
	swizzler(p,q)
}

// AVX2
func bgra256sd(p, q []byte)
func bgra128sd(p,q []byte)

// SSSE
func bgra16sd(p,q []byte)

// AMD64
func bgra4sd(p,q []byte)