// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package screen // import "github.com/as/shiny/screen"

import (
	"image"
	"image/color"
	"image/draw"
	"unicode/utf8"

	"golang.org/x/image/math/f64"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

// Screen creates Buffers, Textures and Windows.
type Screen interface {
	NewBuffer(size image.Point) (Buffer, error)
	NewTexture(size image.Point) (Texture, error)
	NewWindow(opts *NewWindowOptions) (Window, error)
}

type Buffer interface {
	Release()
	Size() image.Point
	Bounds() image.Rectangle
	RGBA() *image.RGBA
}

type Texture interface {
	Release()
	Size() image.Point
	Bounds() image.Rectangle
	Uploader
}

// Window is a top-level, double-buffered GUI window.
type Window interface {
	Device() *Dev
	Release()
	Uploader
	Drawer
	Publish() PublishResult
}

type (
	Lifecycle = lifecycle.Event
	Scroll    = mouse.Event
	Mouse     = mouse.Event
	Key       = key.Event
	Size      = size.Event
	Paint     = paint.Event
)

type Dev struct {
	Lifecycle chan Lifecycle
	Scroll    chan Scroll
	Mouse     chan Mouse
	Key       chan Key
	Size      chan Size
	Paint     chan Paint
}

// PublishResult is the result of an Window.Publish call.
type PublishResult struct {
	BackBufferPreserved bool
}

// NewWindowOptions are optional arguments to NewWindow.
type NewWindowOptions struct {
	Width, Height int
	Title         string
}

func (o *NewWindowOptions) GetTitle() string {
	if o == nil {
		return ""
	}
	return sanitizeUTF8(o.Title, 4096)
}

func sanitizeUTF8(s string, n int) string {
	if n < len(s) {
		s = s[:n]
	}
	i := 0
	for i < len(s) {
		r, n := utf8.DecodeRuneInString(s[i:])
		if r == 0 || (r == utf8.RuneError && n == 1) {
			break
		}
		i += n
	}
	return s[:i]
}

// Uploader is something you can upload a Buffer to.
type Uploader interface {
	Upload(dp image.Point, src Buffer, sr image.Rectangle)
	Fill(dr image.Rectangle, src color.Color, op draw.Op)
}

type Drawer interface {
	Draw(src2dst f64.Aff3, src Texture, sr image.Rectangle, op draw.Op, opts *DrawOptions)
	DrawUniform(src2dst f64.Aff3, src color.Color, sr image.Rectangle, op draw.Op, opts *DrawOptions)
	Copy(dp image.Point, src Texture, sr image.Rectangle, op draw.Op, opts *DrawOptions)
	Scale(dr image.Rectangle, src Texture, sr image.Rectangle, op draw.Op, opts *DrawOptions)
}

const (
	Over = draw.Over
	Src  = draw.Src
)

type DrawOptions struct {
}
