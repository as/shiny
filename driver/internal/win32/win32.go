// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

// Package win32 implements a partial shiny screen driver using the Win32 API.
// It provides window, lifecycle, key, and mouse management, but no drawing.
// That is left to windriver (using GDI) or gldriver (using DirectX via ANGLE).
package win32 // import "github.com/as/shiny/driver/internal/win32"

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/as/shiny/screen"
	"github.com/as/shiny/event/key"
	"github.com/as/shiny/event/lifecycle"
	"github.com/as/shiny/event/mouse"
	"github.com/as/shiny/event/paint"
	"github.com/as/shiny/event/size"
	"github.com/as/shiny/geom"
)

// screenHWND is the handle to the "Screen window".
// The Screen window encapsulates all screen.Screen operations
// in an actual Windows window so they all run on the main thread.
// Since any messages sent to a window will be executed on the
// main thread, we can safely use the messages below.
var screenHWND syscall.Handle

const (
	msgCreateWindow = _WM_USER + iota
	msgMainCallback
	msgShow
	msgQuit
	msgLast
)

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

var currentUserWM userWM

func newWindow(opts *screen.NewWindowOptions) (syscall.Handle, error) {
	// TODO(brainman): convert windowClass to *uint16 once (in initWindowClass)
	wcname, err := syscall.UTF16PtrFromString(windowClass)
	if err != nil {
		return 0, err
	}
	title, err := syscall.UTF16PtrFromString(opts.GetTitle())
	if err != nil {
		return 0, err
	}
	
	
	// h := syscall.Handle(0)
	// if opts.Overlay{
//		h = GetConsoleWindow()
//	}
	hwnd, err := _CreateWindowEx(0,
		wcname, title,
		_WS_OVERLAPPEDWINDOW,
		_CW_USEDEFAULT, _CW_USEDEFAULT,
		_CW_USEDEFAULT, _CW_USEDEFAULT,
		0, // was console handle in experiment
		0, hThisInstance, 0)
	if err != nil {
		return 0, err
	}
	// TODO(andlabs): use proper nCmdShow
	// TODO(andlabs): call UpdateWindow()

	return hwnd, nil
}

// ResizeClientRect makes hwnd client rectangle opts.Width by opts.Height in size.
func ResizeClientRect(hwnd syscall.Handle, opts *screen.NewWindowOptions) error {
	if opts == nil || opts.Width <= 0 || opts.Height <= 0 {
		return nil
	}
	var cr, wr _RECT
	err := _GetClientRect(hwnd, &cr)
	if err != nil {
		return err
	}
	err = _GetWindowRect(hwnd, &wr)
	if err != nil {
		return err
	}
	w := (wr.Right - wr.Left) - (cr.Right - int32(opts.Width))
	h := (wr.Bottom - wr.Top) - (cr.Bottom - int32(opts.Height))
	return _MoveWindow(hwnd, wr.Left, wr.Top, w, h, false)
}

// Show shows a newly created window.
// It sends the appropriate lifecycle events, makes the window appear
// on the screen, and sends an initial size event.
//
// This is a separate step from NewWindow to give the driver a chance
// to setup its internal state for a window before events start being
// delivered.
func Show(hwnd syscall.Handle) {
	SendMessage(hwnd, msgShow, 0, 0)
}

func Release(hwnd syscall.Handle) {
	// TODO(andlabs): check for errors from this?
	// TODO(andlabs): remove unsafe
	_DestroyWindow(hwnd)
	// TODO(andlabs): what happens if we're still painting?
}

func sendFocus(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	switch uMsg {
	case _WM_SETFOCUS:
		LifecycleEvent(hwnd, lifecycle.StageFocused)
	case _WM_KILLFOCUS:
		LifecycleEvent(hwnd, lifecycle.StageVisible)
	default:
		panic(fmt.Sprintf("unexpected focus message: %d", uMsg))
	}
	return _DefWindowProc(hwnd, uMsg, wParam, lParam)
}

func sendShow(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	LifecycleEvent(hwnd, lifecycle.StageVisible)
	_ShowWindow(hwnd, _SW_SHOWDEFAULT)
	sendSize(hwnd)
	return 0
}

func sendSizeEvent(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	wp := (*_WINDOWPOS)(unsafe.Pointer(lParam))
	if wp.Flags&_SWP_NOSIZE != 0 {
		return 0
	}
	sendSize(hwnd)
	return 0
}

func sendSize(hwnd syscall.Handle) {
	var r _RECT
	if err := _GetClientRect(hwnd, &r); err != nil {
		panic(err) // TODO(andlabs)
	}

	width := int(r.Right - r.Left)
	height := int(r.Bottom - r.Top)
	screen.SendSize(size.Event{
		WidthPx:     width,
		HeightPx:    height,
		WidthPt:     geom.Pt(width),
		HeightPt:    geom.Pt(height),
		PixelsPerPt: 1,
	})
}

type Lifecycle = lifecycle.Event
type Scroll = mouse.Event
type Mouse = mouse.Event
type Key = key.Event
type Size = size.Event
type Paint = paint.Event

func sendClose(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	LifecycleEvent(hwnd, lifecycle.StageDead)
	return 0
}

func sendScrollEvent(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	if uMsg != _WM_MOUSEWHEEL {
		panic("sendScrollEvent: not a scroll message")
	}
	e := mouse.Event{
		X:         float32(_GET_X_LPARAM(lParam)),
		Y:         float32(_GET_Y_LPARAM(lParam)),
		Modifiers: keyModifiers(),
	}
	e.Direction = mouse.DirStep
	// Convert from screen to window coordinates.
	p := _POINT{
		int32(e.X),
		int32(e.Y),
	}
	_ScreenToClient(hwnd, &p)
	e.X = float32(p.X)
	e.Y = float32(p.Y)
	delta := _GET_WHEEL_DELTA_WPARAM(wParam) / _WHEEL_DELTA
	switch {
	case delta > 0:
		e.Button = mouse.ButtonWheelUp
	case delta < 0:
		e.Button = mouse.ButtonWheelDown
		delta = -delta
	default:
		return
	}
	screen.SendScroll(e)
	return
}

var mousetab = [...]mouseevent{
	_WM_LBUTTONDOWN: {mouse.DirPress, mouse.ButtonLeft},
	_WM_MBUTTONDOWN: {mouse.DirPress, mouse.ButtonMiddle},
	_WM_RBUTTONDOWN: {mouse.DirPress, mouse.ButtonRight},
	_WM_LBUTTONUP: {mouse.DirRelease, mouse.ButtonLeft},
	_WM_MBUTTONUP: {mouse.DirRelease, mouse.ButtonMiddle},
	_WM_RBUTTONUP: {mouse.DirRelease, mouse.ButtonRight},
	_WM_MOUSEMOVE: {},
}

type mouseevent struct{
	dir mouse.Direction
	but mouse.Button
}

func (m mouseevent) send(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) (lResult uintptr) {
	screen.SendMouse(mouse.Event{
		Direction: m.dir,
		Button: m.but,
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
		Button: s.but,
		X:         float32(_GET_X_LPARAM(lParam)),
		Y:         float32(_GET_Y_LPARAM(lParam)),
		Modifiers: keyModifiers(),
	})
	return 0
}

// Precondition: this is called in immediate response to the message that triggered the event (so not after w.Send).
var (
	MouseEvent     func(hwnd syscall.Handle, e mouse.Event)
	PaintEvent     func(hwnd syscall.Handle, e paint.Event)
	SizeEvent      func(hwnd syscall.Handle, e size.Event)
	KeyEvent       func(hwnd syscall.Handle, e key.Event)
	LifecycleEvent func(hwnd syscall.Handle, e lifecycle.Stage)
)

func sendPaint(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	screen.SendPaint(Paint{})
	return _DefWindowProc(hwnd, uMsg, wParam, lParam)
}

var screenMsgs = map[uint32]func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr){}

func AddScreenMsg(fn func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr)) uint32 {
	uMsg := currentUserWM.next()
	screenMsgs[uMsg] = func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) uintptr {
		fn(hwnd, uMsg, wParam, lParam)
		return 0
	}
	return uMsg
}

func screenWindowWndProc(hwnd syscall.Handle, uMsg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) {
	switch uMsg {
	case msgCreateWindow:
		p := (*newWindowParams)(unsafe.Pointer(lParam))
		p.w, p.err = newWindow(p.opts)
	case msgMainCallback:
		go func() {
			mainCallback()
			SendScreenMessage(msgQuit, 0, 0)
		}()
	case msgQuit:
		_PostQuitMessage(0)
	}
	fn := screenMsgs[uMsg]
	if fn != nil {
		return fn(hwnd, uMsg, wParam, lParam)
	}
	return _DefWindowProc(hwnd, uMsg, wParam, lParam)
}

//go:uintptrescapes

func SendScreenMessage(uMsg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) {
	return SendMessage(screenHWND, uMsg, wParam, lParam)
}

var windowMsgs = map[uint32]func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr){
	_WM_SETFOCUS:         sendFocus,
	_WM_KILLFOCUS:        sendFocus,
	_WM_PAINT:            sendPaint,
	msgShow:              sendShow,
	_WM_WINDOWPOSCHANGED: sendSizeEvent,
	_WM_CLOSE:            sendClose,

	_WM_LBUTTONDOWN: mousetab[_WM_LBUTTONDOWN].send,
	_WM_LBUTTONUP: mousetab[_WM_LBUTTONUP].send,
	_WM_MBUTTONDOWN: mousetab[_WM_MBUTTONDOWN].send,
	_WM_MBUTTONUP: mousetab[_WM_MBUTTONUP].send,
	_WM_RBUTTONDOWN: mousetab[_WM_RBUTTONDOWN].send,
	_WM_RBUTTONUP: mousetab[_WM_RBUTTONUP].send,
	_WM_MOUSEMOVE:    mousetab[_WM_MOUSEMOVE].send,

	_WM_KEYDOWN: sendKeyEvent,
	_WM_KEYUP:   sendKeyEvent,
	// TODO case _WM_SYSKEYDOWN, _WM_SYSKEYUP:
	
	// TODO(as): This will probably break something, let's not
	//_WM_INPUTLANGCHANGE: changeLanguage,
}

func AddWindowMsg(fn func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr)) uint32 {
	uMsg := currentUserWM.next()
	windowMsgs[uMsg] = func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) uintptr {
		fn(hwnd, uMsg, wParam, lParam)
		return 0
	}
	return uMsg
}

func windowWndProc(hwnd syscall.Handle, uMsg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) {
	fn := windowMsgs[uMsg]
	if fn != nil {
		return fn(hwnd, uMsg, wParam, lParam)
	}
	return _DefWindowProc(hwnd, uMsg, wParam, lParam)
}

type newWindowParams struct {
	opts *screen.NewWindowOptions
	w    syscall.Handle
	err  error
}

func NewWindow(opts *screen.NewWindowOptions) (syscall.Handle, error) {
	var p newWindowParams
	p.opts = opts
	SendScreenMessage(msgCreateWindow, 0, uintptr(unsafe.Pointer(&p)))
	return p.w, p.err
}

const windowClass = "shiny_Window"
const (
	CS_OWNDC = 32
)

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

func initScreenWindow() (err error) {
	const screenWindowClass = "shiny_ScreenWindow"
	swc, err := syscall.UTF16PtrFromString(screenWindowClass)
	if err != nil {
		return err
	}
	empty, err := syscall.UTF16PtrFromString("")
	if err != nil {
		return err
	}

	wc := _WNDCLASS{
		LpszClassName: swc,
		LpfnWndProc:   syscall.NewCallback(screenWindowWndProc),
		HIcon:         hDefaultIcon,
		HCursor:       hDefaultCursor,
		HInstance:     hThisInstance,
		HbrBackground: syscall.Handle(_COLOR_BTNFACE + 1),
	}
	_, err = _RegisterClass(&wc)
	if err != nil {
		return err
	}
	
	const(
		//style = _WS_OVERLAPPEDWINDOW | _WS_VISIBLE | _WS_CHILD
		style = _WS_OVERLAPPEDWINDOW
		def = int32(_CW_USEDEFAULT)
	)
	//screenHWND, err = _CreateWindowEx(0, swc, empty, style, def, def, def, def, GetConsoleWindow(), 0, hThisInstance, 0)
	screenHWND, err = _CreateWindowEx(0, swc, empty, style, def, def, def, def, _HWND_MESSAGE, 0, hThisInstance, 0)
	return err
}

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

//go:uintptrescapes

func SendMessage(hwnd syscall.Handle, uMsg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) {
	return sendMessage(hwnd, uMsg, wParam, lParam)
}

var mainCallback func()

func Main(f func()) (retErr error) {
	runtime.LockOSThread()

	if err := initCommon(); err != nil {
		return err
	}

	if err := initScreenWindow(); err != nil {
		return err
	}
	defer _DestroyWindow(screenHWND)

	if err := initWindowClass(); err != nil {
		return err
	}

	// Prime the pump.
	mainCallback = f
	_PostMessage(screenHWND, msgMainCallback, 0, 0)

	// Main message pump.
	var m _MSG
	for {
		done, err := _GetMessage(&m, 0, 0, 0)
		if err != nil {
			return fmt.Errorf("win32 GetMessage failed: %v", err)
		}
		if done == 0 { // WM_QUIT
			break
		}
		_TranslateMessage(&m)
		_DispatchMessage(&m)
	}

	return nil
}
