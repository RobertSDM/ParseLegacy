package windowskeyboard

import (
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	KEYBOARD_INPUT_TYPE = 1
	KEYBOARD_UP_EVENT   = 0x0002
)

const (
	VK_A       = 0x41
	VK_C       = 0x43
	VK_CONTROL = 0x11
	VK_F8      = 0x77
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
func KeyPress(key uint16) (err error) {
	inputs := []INPUT{
		{
			Type: KEYBOARD_INPUT_TYPE,
			Ki: KEYBDINPUT{
				WVk: key,
			},
		},
		{
			Type: KEYBOARD_INPUT_TYPE,
			Ki: KEYBDINPUT{
				WVk:     key,
				DwFlags: uint32(KEYBOARD_UP_EVENT),
			},
		},
	}

	_, err = sendInput(uint(len(inputs)), inputs)

	return
}

// Press and hold a key while executing a callback
func KeyHold(key uint16, cb func()) (err error) {
	_, err = sendInput(1, []INPUT{{
		Type: KEYBOARD_INPUT_TYPE,
		Ki: KEYBDINPUT{
			WVk: key,
		},
	}})
	if err != nil {
		return
	}

	cb()

	_, err = sendInput(1, []INPUT{{
		Type: KEYBOARD_INPUT_TYPE,
		Ki: KEYBDINPUT{
			WVk:     key,
			DwFlags: uint32(KEYBOARD_UP_EVENT),
		},
	}})

	return
}
