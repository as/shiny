// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package win32

import (
	"syscall"

	"github.com/as/shiny/event/mouse"
	"github.com/as/shiny/screen"
)

type mouseevent struct {
	dir mouse.Direction
	but mouse.Button
}

func (m *mouseevent) send(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) (lResult uintptr) {
	screen.SendMouse(mouse.Event{
		Direction: m.dir,
		Button:    m.but,
		X:         float32(_GET_X_LPARAM(lParam)),
		Y:         float32(_GET_Y_LPARAM(lParam)),
		Modifiers: keyModifiers(),
	})
	return 0
}

func sendMouseEvent(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) (lResult uintptr) {
	s := mousetab[msg]
	screen.SendMouse(mouse.Event{
		Direction: s.dir,
		Button:    s.but,
		X:         float32(_GET_X_LPARAM(lParam)),
		Y:         float32(_GET_Y_LPARAM(lParam)),
		Modifiers: keyModifiers(),
	})
	return 0
}

var mousetab = [...]mouseevent{
	_WM_LBUTTONDOWN: {mouse.DirPress, mouse.ButtonLeft},
	_WM_MBUTTONDOWN: {mouse.DirPress, mouse.ButtonMiddle},
	_WM_RBUTTONDOWN: {mouse.DirPress, mouse.ButtonRight},
	_WM_LBUTTONUP:   {mouse.DirRelease, mouse.ButtonLeft},
	_WM_MBUTTONUP:   {mouse.DirRelease, mouse.ButtonMiddle},
	_WM_RBUTTONUP:   {mouse.DirRelease, mouse.ButtonRight},
	_WM_MOUSEMOVE:   {},
}
