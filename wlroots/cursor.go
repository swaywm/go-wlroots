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

func (c Cursor) Destroy() {
	C.wlr_cursor_destroy(c.p)
	man.delete(unsafe.Pointer(c.p))
}

func (c Cursor) X() float64 {
	return float64(c.p.x)
}

func (c Cursor) Y() float64 {
	return float64(c.p.y)
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

func (c Cursor) SetSurface(surface Surface, hotspotX int32, hotspotY int32) {
	C.wlr_cursor_set_surface(c.p, surface.p, C.int32_t(hotspotX), C.int32_t(hotspotY))
}

func (c Cursor) OnMotion(cb func(dev InputDevice, time uint32, dx float64, dy float64)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.motion, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_pointer_motion)(data)
		dev := InputDevice{p: event.device}
		cb(dev, uint32(event.time_msec), float64(event.delta_x), float64(event.delta_y))
	})
}

func (c Cursor) OnMotionAbsolute(cb func(dev InputDevice, time uint32, x float64, y float64)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.motion_absolute, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_pointer_motion_absolute)(data)
		dev := InputDevice{p: event.device}
		cb(dev, uint32(event.time_msec), float64(event.x), float64(event.y))
	})
}

func (c Cursor) OnButton(cb func(dev InputDevice, time uint32, button uint32, state ButtonState)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.button, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_pointer_button)(data)
		dev := InputDevice{p: event.device}
		cb(dev, uint32(event.time_msec), uint32(event.button), ButtonState{event.state})
	})
}

func (c Cursor) OnAxis(cb func(dev InputDevice, time uint32, source AxisSource, orientation AxisOrientation, delta float64, deltaDiscrete int32)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.axis, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_pointer_axis)(data)
		dev := InputDevice{p: event.device}
		cb(dev, uint32(event.time_msec), AxisSource(event.source), AxisOrientation(event.orientation), float64(event.delta), int32(event.delta_discrete))
	})
}
