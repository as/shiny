// +build windows

package win32

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/as/shiny/event/lifecycle"
	"github.com/as/shiny/screen"
)

var windowMsgs = map[uint32]func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr){
	_WM_SETFOCUS:         sendFocus,
	_WM_KILLFOCUS:        sendFocus,
	_WM_PAINT:            sendPaint,
	msgShow:              sendShow,
	_WM_WINDOWPOSCHANGED: sendSizeEvent,
	_WM_CLOSE:            sendClose,

	_WM_LBUTTONDOWN: mousetab[_WM_LBUTTONDOWN].send,
	_WM_LBUTTONUP:   mousetab[_WM_LBUTTONUP].send,
	_WM_MBUTTONDOWN: mousetab[_WM_MBUTTONDOWN].send,
	_WM_MBUTTONUP:   mousetab[_WM_MBUTTONUP].send,
	_WM_RBUTTONDOWN: mousetab[_WM_RBUTTONDOWN].send,
	_WM_RBUTTONUP:   mousetab[_WM_RBUTTONUP].send,
	_WM_MOUSEMOVE:   mousetab[_WM_MOUSEMOVE].send,
	_WM_MOUSEWHEEL:  sendScrollEvent,

	_WM_KEYDOWN: keytab.sendDown,
	_WM_KEYUP:   keytab.sendUp,
	// TODO case _WM_SYSKEYDOWN, _WM_SYSKEYUP:

	// TODO(as): This will probably break something, let's not
	//_WM_INPUTLANGCHANGE: changeLanguage,
}

func NewWindow(opts *screen.NewWindowOptions) (syscall.Handle, error) {
	var p newWindowParams
	p.opts = opts
	SendScreenMessage(msgCreateWindow, 0, uintptr(unsafe.Pointer(&p)))
	return p.w, p.err
}

func AddWindowMsg(fn func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr)) uint32 {
	uMsg := currentUserWM.next()
	windowMsgs[uMsg] = func(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) uintptr {
		fn(hwnd, uMsg, wParam, lParam)
		return 0
	}
	return uMsg
}

func SendMessage(hwnd syscall.Handle, uMsg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) {
	return sendMessage(hwnd, uMsg, wParam, lParam)
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
	_DestroyWindow(hwnd)
}

type newWindowParams struct {
	opts *screen.NewWindowOptions
	w    syscall.Handle
	err  error
}

func windowWndProc(hwnd syscall.Handle, uMsg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) {
	fn := windowMsgs[uMsg]
	if fn != nil {
		return fn(hwnd, uMsg, wParam, lParam)
	}
	return _DefWindowProc(hwnd, uMsg, wParam, lParam)
}

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

func sendClose(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	LifecycleEvent(hwnd, lifecycle.StageDead)
	return 0
}

func sendShow(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	LifecycleEvent(hwnd, lifecycle.StageVisible)
	_ShowWindow(hwnd, _SW_SHOWDEFAULT)
	sendSize(hwnd)
	return 0
}
