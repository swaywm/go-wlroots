package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import (
	"fmt"
	"unsafe"
)

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/types/wlr_input_device.h>
// #include <wlr/types/wlr_keyboard.h>
// #include <wlr/types/wlr_pointer.h>
import "C"

type (
	InputDeviceType   uint32
	ButtonState       uint32
	AxisSource        uint32
	AxisOrientation   uint32
	RelativeDirection uint32
)

var inputDeviceNames = []string{
	InputDeviceTypeKeyboard:   "keyboard",
	InputDeviceTypePointer:    "pointer",
	InputDeviceTypeTouch:      "touch",
	InputDeviceTypeTabletTool: "tablet tool",
	InputDeviceTypeTabletPad:  "tablet pad",
}

const (
	InputDeviceTypeKeyboard   InputDeviceType = C.WLR_INPUT_DEVICE_KEYBOARD
	InputDeviceTypePointer    InputDeviceType = C.WLR_INPUT_DEVICE_POINTER
	InputDeviceTypeTouch      InputDeviceType = C.WLR_INPUT_DEVICE_TOUCH
	InputDeviceTypeTabletTool InputDeviceType = C.WLR_INPUT_DEVICE_TABLET
	InputDeviceTypeTabletPad  InputDeviceType = C.WLR_INPUT_DEVICE_TABLET_PAD
	InputDeviceTypeSwitch     InputDeviceType = C.WLR_INPUT_DEVICE_SWITCH

	ButtonStateReleased ButtonState = C.WLR_BUTTON_RELEASED
	ButtonStatePressed  ButtonState = C.WLR_BUTTON_PRESSED

	AxisSourceWheel      AxisSource = C.WL_POINTER_AXIS_SOURCE_WHEEL
	AxisSourceFinger     AxisSource = C.WL_POINTER_AXIS_SOURCE_FINGER
	AxisSourceContinuous AxisSource = C.WL_POINTER_AXIS_SOURCE_CONTINUOUS
	AxisSourceWheelTilt  AxisSource = C.WL_POINTER_AXIS_SOURCE_WHEEL_TILT

	AxisOrientationVertical   AxisOrientation = C.WL_POINTER_AXIS_VERTICAL_SCROLL
	AxisOrientationHorizontal AxisOrientation = C.WL_POINTER_AXIS_HORIZONTAL_SCROLL

	RelativeDirectionIdentical RelativeDirection = C.WL_POINTER_AXIS_RELATIVE_DIRECTION_IDENTICAL
	RelativeDirectionInverted  RelativeDirection = C.WL_POINTER_AXIS_RELATIVE_DIRECTION_INVERTED
)

type InputDevice struct {
	p *C.struct_wlr_input_device
}

func (d InputDevice) OnDestroy(cb func(InputDevice)) {
	man.add(unsafe.Pointer(d.p), &d.p.events.destroy, func(unsafe.Pointer) {
		cb(d)
	})
}

func (d InputDevice) Type() InputDeviceType { return InputDeviceType(d.p._type) }
func (d InputDevice) Name() string          { return C.GoString(d.p.name) }

func validateInputDeviceType(d InputDevice, fn string, req InputDeviceType) {
	if typ := d.Type(); typ != req {
		if int(typ) >= len(inputDeviceNames) {
			panic(fmt.Sprintf("%s called on input device of type %d", fn, typ))
		} else {
			panic(fmt.Sprintf("%s called on input device of type %s", fn, inputDeviceNames[typ]))
		}
	}
}

func (d InputDevice) Keyboard() Keyboard {
	validateInputDeviceType(d, "Keyboard", InputDeviceTypeKeyboard)
	p := *(*unsafe.Pointer)(unsafe.Pointer(&d.p))
	return Keyboard{p: (*C.struct_wlr_keyboard)(p)}
}

func wrapInputDevice(p unsafe.Pointer) InputDevice {
	return InputDevice{p: (*C.struct_wlr_input_device)(p)}
}
