package wlroots

// #include <wlr/types/wlr_input_device.h>
import "C"

import (
	"unsafe"
)

type (
	InputDeviceType uint32
	ButtonState     uint32
	AxisSource      uint32
	AxisOrientation uint32
)

const (
	InputDeviceTypeKeyboard   InputDeviceType = C.WLR_INPUT_DEVICE_KEYBOARD
	InputDeviceTypePointer    InputDeviceType = C.WLR_INPUT_DEVICE_POINTER
	InputDeviceTypeTouch      InputDeviceType = C.WLR_INPUT_DEVICE_TOUCH
	InputDeviceTypeTabletTool InputDeviceType = C.WLR_INPUT_DEVICE_TABLET_TOOL
	InputDeviceTypeTabletPad  InputDeviceType = C.WLR_INPUT_DEVICE_TABLET_PAD

	ButtonStateReleased ButtonState = C.WLR_BUTTON_RELEASED
	ButtonStatePressed  ButtonState = C.WLR_BUTTON_PRESSED

	AxisSourceWheel      AxisSource = C.WLR_AXIS_SOURCE_WHEEL
	AxisSourceFinger     AxisSource = C.WLR_AXIS_SOURCE_FINGER
	AxisSourceContinuous AxisSource = C.WLR_AXIS_SOURCE_CONTINUOUS
	AxisSourceWheelTilt  AxisSource = C.WLR_AXIS_SOURCE_WHEEL_TILT

	AxisOrientationVertical   AxisOrientation = C.WLR_AXIS_ORIENTATION_VERTICAL
	AxisOrientationHorizontal AxisOrientation = C.WLR_AXIS_ORIENTATION_HORIZONTAL
)

type InputDevice struct {
	p *C.struct_wlr_input_device
}

func wrapInputDevice(p unsafe.Pointer) InputDevice {
	return InputDevice{p: (*C.struct_wlr_input_device)(p)}
}

func (d InputDevice) Type() InputDeviceType {
	return InputDeviceType(d.p._type)
}

func (d InputDevice) Keyboard() Keyboard {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&d.p.anon0[0]))
	return Keyboard{p: (*C.struct_wlr_keyboard)(p)}
}
