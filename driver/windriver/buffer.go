// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package windriver

import (
	"image"
	"image/draw"
	"sync"
	"syscall"
)

type bufferImpl struct {
	hbitmap syscall.Handle
	buf     []byte
	rgba    image.RGBA
	size    image.Point

	mu        sync.Mutex
	nUpload   uint32
	released  bool
	cleanedUp bool
}

func (b *bufferImpl) Size() image.Point       { return b.size }
func (b *bufferImpl) Bounds() image.Rectangle { return image.Rectangle{Max: b.size} }
func (b *bufferImpl) RGBA() *image.RGBA       { return &b.rgba }
func (b *bufferImpl) Release() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.released && b.nUpload == 0 {
		go b.cleanUp()
	}
	b.released = true
}

func (b *bufferImpl) cleanUp() {
	b.mu.Lock()
	if b.cleanedUp {
		b.mu.Unlock()
		panic("windriver: Buffer clean-up occurred twice")
	}
	b.cleanedUp = true
	b.mu.Unlock()

	b.rgba.Pix = nil
	_DeleteObject(b.hbitmap)
}

func (b *bufferImpl) blitToDC(dc syscall.Handle, dp image.Point, sr image.Rectangle) error {
	err := copyBitmapToDC(dc, sr.Add(dp.Sub(sr.Min)), b.hbitmap, sr, draw.Src)
	return err
}
