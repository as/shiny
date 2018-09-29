// +build windows

package win32

import (
	"syscall"

	"github.com/as/shiny/event/lifecycle"
)

type Lifecycle = lifecycle.Event

var LifecycleEvent func(hwnd syscall.Handle, e lifecycle.Stage)
