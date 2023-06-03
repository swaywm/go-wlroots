package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import "unsafe"

// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <stdlib.h>
// #include <time.h>
// #include <wlr/types/wlr_xdg_shell.h>
// #include <wlr/xwayland.h>
import "C"

type XWayland struct {
	p *C.struct_wlr_xwayland
}

/**
 * An Xwayland user interface component. It has an absolute position in
 * layout-local coordinates.
 *
 * The inner struct wlr_surface is valid once the associate event is emitted.
 * Compositors can set up e.g. map and unmap listeners at this point. The
 * struct wlr_surface becomes invalid when the dissociate event is emitted.
 */
type XWaylandSurface struct {
	p *C.struct_wlr_xwayland_surface
}

func (x XWayland) Destroy() {
	C.wlr_xwayland_destroy(x.p)
}

func (x XWayland) OnNewSurface(cb func(XWaylandSurface)) {
	man.add(unsafe.Pointer(x.p), &x.p.events.new_surface, func(data unsafe.Pointer) {
		surface := XWaylandSurface{p: (*C.struct_wlr_xwayland_surface)(data)}
		man.track(unsafe.Pointer(surface.p), &surface.p.events.destroy)
		man.add(unsafe.Pointer(surface.p.surface), &surface.p.surface.events.destroy, func(data unsafe.Pointer) {
			man.delete(unsafe.Pointer(surface.p.surface))
		})
		cb(surface)
	})
}

func (x XWayland) SetCursor(img XCursorImage) {
	C.wlr_xwayland_set_cursor(x.p, img.p.buffer, img.p.width*4, img.p.width, img.p.height, C.int32_t(img.p.hotspot_x), C.int32_t(img.p.hotspot_y))
}

func (s XWaylandSurface) Activate(activated bool) {
	C.wlr_xwayland_surface_activate(s.p, C.bool(activated))
}

func (s XWaylandSurface) Surface() Surface {
	return Surface{p: s.p.surface}
}

func (s XWaylandSurface) Geometry() GeoBox {
	return GeoBox{
		X:      int(s.p.x),
		Y:      int(s.p.y),
		Width:  int(s.p.width),
		Height: int(s.p.height),
	}
}

func (s XWaylandSurface) Configure(x int16, y int16, width uint16, height uint16) {
	C.wlr_xwayland_surface_configure(s.p, C.int16_t(x), C.int16_t(y), C.uint16_t(width), C.uint16_t(height))
}

func (s XWaylandSurface) OnMap(cb func(XWaylandSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.surface.events._map, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XWaylandSurface) OnUnmap(cb func(XWaylandSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.surface.events.unmap, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XWaylandSurface) OnDestroy(cb func(XWaylandSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.destroy, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XWaylandSurface) OnRequestMove(cb func(surface XWaylandSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.request_move, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XWaylandSurface) OnRequestResize(cb func(surface XWaylandSurface, edges Edges)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.request_resize, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_xwayland_resize_event)(data)
		cb(s, Edges(event.edges))
	})
}

func (s XWaylandSurface) OnRequestConfigure(cb func(surface XWaylandSurface, x int16, y int16, width uint16, height uint16)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.request_configure, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_xwayland_surface_configure_event)(data)
		cb(s, int16(event.x), int16(event.y), uint16(event.width), uint16(event.height))
	})
}

/** Create an Xwayland server and XWM.
 *
 * The server supports a lazy mode in which Xwayland is only started when a
 * client tries to connect.
 */
func (d Display) XWaylandCreate(compositor Compositor, lazy bool) XWayland {
	p := C.wlr_xwayland_create(d.p, compositor.p, C.bool(lazy))
	return XWayland{p: p}
}
