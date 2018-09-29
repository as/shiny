// +build windows

package win32

import (
	"syscall"

	"github.com/as/shiny/event/mouse"
	"github.com/as/shiny/screen"
)

type Scroll = mouse.Event

func sendScrollEvent(hwnd syscall.Handle, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
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
