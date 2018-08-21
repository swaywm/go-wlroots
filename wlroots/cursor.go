package wlroots

// #include <wlr/types/wlr_cursor.h>
import "C"
import "unsafe"

type Cursor struct {
	p *C.struct_wlr_cursor
}

func NewCursor() Cursor {
	p := C.wlr_cursor_create()
	return Cursor{p: p}
}

func (c Cursor) AttachOutputLayout(layout OutputLayout) {
	C.wlr_cursor_attach_output_layout(c.p, layout.p)
}

func (c Cursor) AttachInputDevice(dev InputDevice) {
	C.wlr_cursor_attach_input_device(c.p, dev.p)
}

func (c Cursor) Move(dev InputDevice, dx float64, dy float64) {
	C.wlr_cursor_move(c.p, dev.p, C.double(dx), C.double(dy))
}

func (c Cursor) WarpAbsolute(dev InputDevice, x float64, y float64) {
	C.wlr_cursor_warp_absolute(c.p, dev.p, C.double(x), C.double(y))
}

func (c Cursor) OnMotion(cb func(dev InputDevice, time uint32, dx float64, dy float64)) {
	listener := NewListener(func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_pointer_motion)(data)
		dev := InputDevice{p: event.device}
		cb(dev, uint32(event.time_msec), float64(event.delta_x), float64(event.delta_y))
	})

	C.wl_signal_add(&c.p.events.motion, listener.p)
}

func (c Cursor) OnMotionAbsolute(cb func(dev InputDevice, time uint32, x float64, y float64)) {
	listener := NewListener(func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_pointer_motion_absolute)(data)
		dev := InputDevice{p: event.device}
		cb(dev, uint32(event.time_msec), float64(event.x), float64(event.y))
	})

	C.wl_signal_add(&c.p.events.motion_absolute, listener.p)
}
