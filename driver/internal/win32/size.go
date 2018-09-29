// +build windows

package win32

import (
	"syscall"
	"unsafe"

	"github.com/as/shiny/event/size"
	"github.com/as/shiny/geom"
	"github.com/as/shiny/screen"
)

type Size = size.Event

var SizeEvent func(hwnd syscall.Handle, e size.Event)

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
