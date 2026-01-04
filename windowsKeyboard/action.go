package windowskeyboard

import (
	"runtime"
	"unsafe"

	"github.com/moutend/go-hook/pkg/types"
	"golang.org/x/sys/windows"
)

const (
	KEYBOARD_INPUT_TYPE = 1
	KEYBOARD_UP_EVENT   = 0x0002
)

// Represent the keyboard input send to the Windows API
type KEYBDINPUT struct {
	// Windows Virtual Key
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type INPUT struct {
	// Input type: keyboard, hardware or mouse. Limited to keyboard for this project
	Type uint32

	// The only type needed for this project
	Ki KEYBDINPUT
}

// Get the size of the INPUT struct based on the GOARCH variable
func getSizeOfInput() int {
	if runtime.GOARCH == "amd64" {
		return 40
	}

	return 28
}

// Low level function to send a input to the Windows API
func sendInput(inputLength uint, inputs []INPUT) (uint32, error) {
	sendInputProc := windows.NewLazySystemDLL("user32.dll").NewProc("SendInput")

	sizeOfInput := getSizeOfInput()

	ret, _, err := sendInputProc.Call(
		uintptr(inputLength),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(sizeOfInput),
	)

	if ret == 0 {
		return 0, err
	}

	return uint32(ret), nil
}

// Presses a key
func KeyPress(key types.VKCode) (err error) {
	inputs := []INPUT{
		{
			Type: KEYBOARD_INPUT_TYPE,
			Ki: KEYBDINPUT{ // without dwFlags correspond to the down event
				WVk: uint16(key),
			},
		},
		{
			Type: KEYBOARD_INPUT_TYPE,
			Ki: KEYBDINPUT{
				WVk:     uint16(key),
				DwFlags: uint32(KEYBOARD_UP_EVENT),
			},
		},
	}

	_, err = sendInput(uint(len(inputs)), inputs)

	return
}

// Press and hold a key while executing a callback
func KeyHold(key types.VKCode, cb func()) (err error) {
	_, err = sendInput(1, []INPUT{{
		Type: KEYBOARD_INPUT_TYPE,
		Ki: KEYBDINPUT{
			WVk: uint16(key),
		},
	}})
	if err != nil {
		return
	}

	cb()

	_, err = sendInput(1, []INPUT{{
		Type: KEYBOARD_INPUT_TYPE,
		Ki: KEYBDINPUT{
			WVk:     uint16(key),
			DwFlags: uint32(KEYBOARD_UP_EVENT),
		},
	}})

	return
}
