package wlroots

// #include <wlr/xwayland.h>
import "C"
import "unsafe"

type XWayland struct {
	p *C.struct_wlr_xwayland
}

type XWaylandSurface struct {
	p *C.struct_wlr_xwayland_surface
}

func NewXWayland(display Display, compositor Compositor, lazy bool) XWayland {
	p := C.wlr_xwayland_create(display.p, compositor.p, C.bool(lazy))
	//man.track(unsafe.Pointer(p), &p.events.destroy)
	return XWayland{p: p}
}

func (x XWayland) Destroy() {
	C.wlr_xwayland_destroy(x.p)
}

func (x XWayland) OnDestroy(cb func(XWayland)) {
	/*man.add(unsafe.Pointer(x.p), &x.p.events.destroy, func(unsafe.Pointer) {
		cb(x)
	})*/
}

func (x XWayland) Display() int {
	return int(x.p.display)
}

func (x XWayland) OnNewSurface(cb func(XWaylandSurface)) {
	man.add(unsafe.Pointer(x.p), &x.p.events.new_surface, func(data unsafe.Pointer) {
		surface := XWaylandSurface{p: (*C.struct_wlr_xwayland_surface)(data)}
		man.track(unsafe.Pointer(surface.p), &surface.p.events.destroy)
		/*man.add(unsafe.Pointer(surface.p.surface), &surface.p.surface.events.destroy, func(data unsafe.Pointer) {
			man.delete(unsafe.Pointer(surface.p.surface))
		})*/
		cb(surface)
	})
}

func (x XWayland) SetCursor(img XCursorImage) {
	C.wlr_xwayland_set_cursor(x.p, img.p.buffer, img.p.width*4, img.p.width, img.p.height, C.int32_t(img.p.hotspot_x), C.int32_t(img.p.hotspot_y))
}

func (s XWaylandSurface) Surface() Surface {
	return Surface{p: s.p.surface}
}

func (s XWaylandSurface) OnMap(cb func(XWaylandSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events._map, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XWaylandSurface) OnUnmap(cb func(XWaylandSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.unmap, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XWaylandSurface) OnDestroy(cb func(XWaylandSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.destroy, func(data unsafe.Pointer) {
		cb(s)
	})
}
