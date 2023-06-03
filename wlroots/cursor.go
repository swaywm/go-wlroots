package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import "unsafe"

// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <stdlib.h>
// #include <wlr/types/wlr_cursor.h>
// #include <wlr/types/wlr_pointer.h>
// #include <wlr/xwayland.h>
import "C"

/**
 * wlr_cursor implements the behavior of the "cursor", that is, the image on the
 * screen typically moved about with a mouse or so. It provides tracking for
 * this in global coordinates, and integrates with struct wlr_output,
 * struct wlr_output_layout, and struct wlr_input_device. You can use it to
 * abstract multiple input devices over a single cursor, constrain cursor
 * movement to the usable area of a struct wlr_output_layout and communicate
 * position updates to the hardware cursor, constrain specific input devices to
 * specific outputs or regions of the screen, and so on.
 */
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

/**
 * Uses the given layout to establish the boundaries and movement semantics of
 * this cursor. Cursors without an output layout allow infinite movement in any
 * direction and do not support absolute input events.
 */
func (c Cursor) AttachOutputLayout(layout OutputLayout) {
	C.wlr_cursor_attach_output_layout(c.p, layout.p)
}

/**
 * Attaches this cursor to the given output, which must be among the outputs in
 * the current output_layout for this cursor. This call is invalid for a cursor
 * without an associated output layout.
 */
func (c Cursor) MapToOutput(output Output) {
	C.wlr_cursor_map_to_output(c.p, output.p)
}

/**
 * Maps all input from a specific input device to a given output. The input
 * device must be attached to this cursor and the output must be among the
 * outputs in the attached output layout.
 */
func (c Cursor) MapInputToOutput(input InputDevice, output Output) {
	C.wlr_cursor_map_input_to_output(c.p, input.p, output.p)
}

/**
 * Maps this cursor to an arbitrary region on the associated
 * struct wlr_output_layout.
 */
func (c Cursor) MapToRegion(box GeoBox) {
	C.wlr_cursor_map_to_region(c.p, box.p)
}

/**
 * Maps inputs from this input device to an arbitrary region on the associated
 * struct wlr_output_layout.
 */
func (c Cursor) MapInputToRegion(dev InputDevice, box GeoBox) {
	C.wlr_cursor_map_input_to_region(c.p, dev.p, box.p)
}

/**
 * Attaches this input device to this cursor. The input device must be one of:
 *
 * - WLR_INPUT_DEVICE_POINTER
 * - WLR_INPUT_DEVICE_TOUCH
 * - WLR_INPUT_DEVICE_TABLET_TOOL
 */
func (c Cursor) AttachInputDevice(dev InputDevice) {
	C.wlr_cursor_attach_input_device(c.p, dev.p)
}

func (c Cursor) DetachInputDevice(dev InputDevice) {
	C.wlr_cursor_detach_input_device(c.p, dev.p)
}

/**
 * Set the cursor image from an XCursor theme.
 *
 * The image will be loaded from the struct wlr_xcursor_manager.
 */
func (c Cursor) SetXCursor(cm XCursorManager, name string) {
	C.wlr_cursor_set_xcursor(c.p, cm.p, C.CString(name))
}

/**
 * Move the cursor in the direction of the given x and y layout coordinates. If
 * one coordinate is NAN, it will be ignored.
 *
 * `dev` may be passed to respect device mapping constraints. If `dev` is NULL,
 * device mapping constraints will be ignored.
 */
func (c Cursor) Move(dev InputDevice, dx float64, dy float64) {
	C.wlr_cursor_move(c.p, dev.p, C.double(dx), C.double(dy))
}

/**
 * Warp the cursor to the given x and y in absolute 0..1 coordinates. If the
 * given point is out of the layout boundaries or constraints, the closest point
 * will be used. If one coordinate is NAN, it will be ignored.
 *
 * `dev` may be passed to respect device mapping constraints. If `dev` is NULL,
 * device mapping constraints will be ignored.
 */
func (c Cursor) WarpAbsolute(dev InputDevice, x float64, y float64) {
	C.wlr_cursor_warp_absolute(c.p, dev.p, C.double(x), C.double(y))
}

/**
 * Set the cursor surface. The surface can be committed to update the cursor
 * image. The surface position is subtracted from the hotspot. A NULL surface
 * commit hides the cursor.
 */
func (c Cursor) SetSurface(surface Surface, hotspotX int32, hotspotY int32) {
	C.wlr_cursor_set_surface(c.p, surface.p, C.int32_t(hotspotX), C.int32_t(hotspotY))
}

/**
 * Hide the cursor image.
 */
func (c Cursor) UnsetImage() {
	C.wlr_cursor_unset_image(c.p)
}

func (c Cursor) OnMotion(cb func(dev InputDevice, time uint32, dx float64, dy float64)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.motion, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_pointer_motion_event)(data)
		dev := InputDevice{p: &event.pointer.base}
		cb(dev, uint32(event.time_msec), float64(event.delta_x), float64(event.delta_y))
	})
}

func (c Cursor) OnMotionAbsolute(cb func(dev InputDevice, time uint32, x float64, y float64)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.motion_absolute, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_pointer_motion_absolute_event)(data)
		dev := InputDevice{p: &event.pointer.base}
		cb(dev, uint32(event.time_msec), float64(event.x), float64(event.y))
	})
}

func (c Cursor) OnButton(cb func(dev InputDevice, time uint32, button uint32, state ButtonState)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.button, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_pointer_button_event)(data)
		dev := InputDevice{p: &event.pointer.base}
		cb(dev, uint32(event.time_msec), uint32(event.button), ButtonState(event.state))
	})
}

func (c Cursor) OnAxis(cb func(dev InputDevice, time uint32, source AxisSource, orientation AxisOrientation, delta float64, deltaDiscrete int32)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.axis, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_pointer_axis_event)(data)
		dev := InputDevice{p: &event.pointer.base}
		cb(dev, uint32(event.time_msec), AxisSource(event.source), AxisOrientation(event.orientation), float64(event.delta), int32(event.delta_discrete))
	})
}

func (c Cursor) OnFrame(cb func()) {
	man.add(unsafe.Pointer(c.p), &c.p.events.frame, func(data unsafe.Pointer) {
		cb()
	})
}
