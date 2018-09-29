// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

// Package win32 implements a partial shiny screen driver using the Win32 API.
// It provides window, lifecycle, key, and mouse management, but no drawing.
// That is left to windriver (using GDI) or gldriver (using DirectX via ANGLE).
package win32 // import "github.com/as/shiny/driver/internal/win32"

import (
	"sync"
	"syscall"
)

const (
	msgCreateWindow = _WM_USER + iota
	msgMainCallback
	msgShow
	msgQuit
	msgLast
)

const windowClass = "shiny_Window"
const (
	CS_OWNDC = 32
)

var currentUserWM userWM

var (
	hDefaultIcon   syscall.Handle
	hDefaultCursor syscall.Handle
	hThisInstance  syscall.Handle
)

func initCommon() (err error) {
	hDefaultIcon, err = _LoadIcon(0, _IDI_APPLICATION)
	if err != nil {
		return err
	}
	hDefaultCursor, err = _LoadCursor(0, _IDC_ARROW)
	if err != nil {
		return err
	}
	// TODO(andlabs) hThisInstance
	return nil
}

func initWindowClass() (err error) {
	wcname, err := syscall.UTF16PtrFromString(windowClass)
	if err != nil {
		return err
	}
	_, err = _RegisterClass(&_WNDCLASS{
		Style:         CS_OWNDC,
		LpszClassName: wcname,
		LpfnWndProc:   syscall.NewCallback(windowWndProc),
		HIcon:         hDefaultIcon,
		HCursor:       hDefaultCursor,
		HInstance:     hThisInstance,
		HbrBackground: syscall.Handle(_COLOR_BTNFACE + 1),
	})
	return err
}


// userWM is used to generate private (WM_USER and above) window message IDs
// for use by screenWindowWndProc and windowWndProc.
type userWM struct {
	sync.Mutex
	id uint32
}

func (m *userWM) next() uint32 {
	m.Lock()
	if m.id == 0 {
		m.id = msgLast
	}
	r := m.id
	m.id++
	m.Unlock()
	return r
}
