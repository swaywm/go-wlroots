package wlroots

// #include <wlr/types/wlr_input_device.h>
import "C"

import (
	"unsafe"
)

const (
	InputDeviceKeyboard   = C.WLR_INPUT_DEVICE_KEYBOARD
	InputDevicePointer    = C.WLR_INPUT_DEVICE_POINTER
	InputDeviceTouch      = C.WLR_INPUT_DEVICE_TOUCH
	InputDeviceTabletTool = C.WLR_INPUT_DEVICE_TABLET_TOOL
	InputDeviceTabletPad  = C.WLR_INPUT_DEVICE_TABLET_PAD
)

type InputDevice struct {
	p *C.struct_wlr_input_device
}

func wrapInputDevice(p unsafe.Pointer) InputDevice {
	return InputDevice{p: (*C.struct_wlr_input_device)(p)}
}

func (d InputDevice) Type() uint32 {
	return d.p._type
}

func (d InputDevice) Keyboard() Keyboard {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&d.p.anon0[0]))
	return Keyboard{p: (*C.struct_wlr_keyboard)(p)}
}
