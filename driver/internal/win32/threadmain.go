// +build windows

package win32

import (
	"fmt"
	"runtime"
)

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
