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
	var cr, wr Rect32
	err := GetClientRect(hwnd, &cr)
	if err != nil {
		return err
	}
	err = GetWindowRect(hwnd, &wr)
	if err != nil {
		return err
	}
	w := wr.Dx() - (cr.Max.X - int32(opts.Width))
	h := wr.Dy()- (cr.Max.Y - int32(opts.Height))
	return MoveWindow(hwnd, wr.Min.X, wr.Min.Y, w, h, false)
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
	var r Rect32
	if err := GetClientRect(hwnd, &r); err != nil {
		panic(err) // TODO(andlabs)
	}

	dx, dy := int(r.Dx()), int(r.Dy())
	screen.SendSize(size.Event{
		WidthPx:     dx,
		HeightPx:    dy,
		WidthPt:     geom.Pt(dx),
		HeightPt:    geom.Pt(dy),
		PixelsPerPt: 1,
	})
}
