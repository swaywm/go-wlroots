package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import (
	"log/slog"
	"sync"
	"unsafe"
)

// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/types/wlr_xdg_shell.h>
//
// void _wlr_xdg_surface_for_each_cb(struct wlr_surface *surface, int sx, int sy, void *data);
// static inline void _wlr_xdg_surface_for_each_surface(struct wlr_xdg_surface *surface, void *user_data) {
//		wlr_xdg_surface_for_each_surface(surface, &_wlr_xdg_surface_for_each_cb, user_data);
// }
import "C"

type XDGSurfaceRole uint32

const (
	XDGSurfaceRoleNone     XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_NONE
	XDGSurfaceRoleTopLevel XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_TOPLEVEL
	XDGSurfaceRolePopup    XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_POPUP
)

var (
	xdgSurfaceWalkers      = map[*C.struct_wlr_xdg_surface]XDGSurfaceWalkFunc{}
	xdgSurfaceWalkersMutex sync.RWMutex
)

type XDGShell struct {
	p *C.struct_wlr_xdg_shell
}

type XDGPopup struct {
	p *C.struct_wlr_xdg_popup
}

func (x XDGPopup) Parent() Surface {
	return Surface{p: x.p.parent}
}

func (x XDGPopup) Seat() Seat {
	return Seat{p: x.p.seat}
}

func (x XDGPopup) Base() XDGSurface {
	return XDGSurface{p: x.p.base}
}

type XDGSurfaceWalkFunc func(surface Surface, sx int, sy int)

func (s XDGShell) Version() int {
	return int(s.p.version)
}

func (s XDGShell) PingTimeout() int {
	return int(s.p.ping_timeout)
}

func (s XDGShell) OnDestroy(cb func(XDGShell)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.destroy, func(unsafe.Pointer) {
		cb(s)
	})
}

func (s XDGShell) OnNewSurface(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.new_surface, func(data unsafe.Pointer) {
		surface := XDGSurface{p: (*C.struct_wlr_xdg_surface)(data)}
		man.add(unsafe.Pointer(surface.p), &surface.p.events.destroy, func(data unsafe.Pointer) {
			man.delete(unsafe.Pointer(surface.p))
			man.delete(unsafe.Pointer(surface.TopLevel().p))
		})
		man.add(unsafe.Pointer(surface.p.surface), &surface.p.surface.events.destroy, func(data unsafe.Pointer) {
			man.delete(unsafe.Pointer(surface.p.surface))
		})
		cb(surface)
	})
}

//export _wlr_xdg_surface_for_each_cb
func _wlr_xdg_surface_for_each_cb(surface *C.struct_wlr_surface, sx C.int, sy C.int, data unsafe.Pointer) {
	xdgSurfaceWalkersMutex.RLock()
	cb := xdgSurfaceWalkers[(*C.struct_wlr_xdg_surface)(data)]
	xdgSurfaceWalkersMutex.RUnlock()
	if cb != nil {
		cb(Surface{p: surface}, int(sx), int(sy))
	}
}

type XDGSurface struct {
	p *C.struct_wlr_xdg_surface
}

func (x XDGSurface) Nil() bool {
	return x.p == nil
}

func (x XDGSurface) Walk(visit XDGSurfaceWalkFunc) {
	xdgSurfaceWalkersMutex.Lock()
	xdgSurfaceWalkers[x.p] = visit
	xdgSurfaceWalkersMutex.Unlock()

	C._wlr_xdg_surface_for_each_surface(x.p, unsafe.Pointer(x.p))

	xdgSurfaceWalkersMutex.Lock()
	delete(xdgSurfaceWalkers, x.p)
	xdgSurfaceWalkersMutex.Unlock()
}

/**
 * The lifetime-bound role of the xdg_surface. WLR_XDG_SURFACE_ROLE_NONE
 * if the role was never set.
 */
func (x XDGSurface) Role() XDGSurfaceRole {
	return XDGSurfaceRole(x.p.role)
}

func (x XDGSurface) Popup() XDGPopup {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&x.p.anon0[0]))
	return XDGPopup{p: (*C.struct_wlr_xdg_popup)(p)}
}
func (x XDGSurface) TopLevel() XDGTopLevel {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&x.p.anon0[0]))
	return XDGTopLevel{p: (*C.struct_wlr_xdg_toplevel)(p)}
}

func (x XDGSurface) TopLevelSetActivated(activated bool) {
	C.wlr_xdg_toplevel_set_activated(x.TopLevel().p, C.bool(activated))
}

func (x XDGSurface) TopLevelSetSize(width uint32, height uint32) {
	C.wlr_xdg_toplevel_set_size(x.TopLevel().p, C.int(width), C.int(height))
}

func (x XDGSurface) TopLevelSetTiled(edges Edges) {
	C.wlr_xdg_toplevel_set_tiled(x.TopLevel().p, C.uint(edges))
}

func (x XDGSurface) SendClose() {
	C.wlr_xdg_toplevel_send_close(x.TopLevel().p)
}

func (x XDGSurface) SceneTree() SceneTree {
	slog.Debug("XDGSurface SceneTree()", "x.p", x.p)
	slog.Debug("XDGSurface SceneTree()", "x.p.data", x.p.data)
	return SceneTree{p: (*C.struct_wlr_scene_tree)(x.p.data)}
}

func (x XDGSurface) Ping() {
	C.wlr_xdg_surface_ping(x.p)
}

func (x XDGSurface) Surface() Surface {
	return Surface{p: x.p.surface}
}

func (x XDGSurface) SurfaceAt(sx float64, sy float64) (surface Surface, subX float64, subY float64) {
	var csubX, csubY C.double
	p := C.wlr_xdg_surface_surface_at(x.p, C.double(sx), C.double(sy), &csubX, &csubY)
	return Surface{p: p}, float64(csubX), float64(csubY)
}

func (x XDGSurface) SetData(tree SceneTree) {
	x.p.data = unsafe.Pointer(tree.p)
	slog.Debug("XDGSurface SetData", "x.p", x.p)
	slog.Debug("XDGSurface SetData", "x.data:", x.p.data)
}

func (x XDGSurface) OnMap(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(x.p), &x.p.surface.events._map, func(data unsafe.Pointer) {
		cb(x)
	})
}

func (x XDGSurface) OnUnmap(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(x.p), &x.p.surface.events.unmap, func(data unsafe.Pointer) {
		cb(x)
	})
}

func (x XDGSurface) OnDestroy(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(x.p), &x.p.events.destroy, func(data unsafe.Pointer) {
		cb(x)
	})
}

func (x XDGSurface) OnPingTimeout(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(x.p), &x.p.events.ping_timeout, func(data unsafe.Pointer) {
		cb(x)
	})
}

func (x XDGSurface) OnNewPopup(cb func(XDGSurface, XDGPopup)) {
	man.add(unsafe.Pointer(x.p), &x.p.events.ping_timeout, func(data unsafe.Pointer) {
		popup := XDGPopup{p: (*C.struct_wlr_xdg_popup)(data)}
		cb(x, popup)
	})
}

func (x XDGSurface) Geometry() GeoBox {
	var cb C.struct_wlr_box
	C.wlr_xdg_surface_get_geometry(x.p, &cb)

	var b GeoBox
	b.fromC(&cb)
	return b
}

type XDGTopLevel struct {
	p *C.struct_wlr_xdg_toplevel
}

func (t XDGTopLevel) OnRequestMove(cb func(client SeatClient, serial uint32)) {
	man.add(unsafe.Pointer(t.p), &t.p.events.request_move, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_xdg_toplevel_move_event)(data)
		client := SeatClient{p: event.seat}
		cb(client, uint32(event.serial))
	})
}

func (t XDGTopLevel) OnRequestResize(cb func(client SeatClient, serial uint32, edges Edges)) {
	man.add(unsafe.Pointer(t.p), &t.p.events.request_resize, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_xdg_toplevel_resize_event)(data)
		client := SeatClient{p: event.seat}
		cb(client, uint32(event.serial), Edges(event.edges))
	})
}

func (t XDGTopLevel) Nil() bool {
	return t.p == nil
}

func (t XDGTopLevel) Title() string {
	return C.GoString(t.p.title)
}

func (t XDGTopLevel) AppId() string {
	return C.GoString(t.p.app_id)
}

func (t XDGTopLevel) Parent() XDGTopLevel {
	return XDGTopLevel{p: t.p.parent}
}

func (t XDGTopLevel) Base() XDGSurface {
	return XDGSurface{p: t.p.base}
}

func (t XDGTopLevel) SetActivated(activated bool) {
	C.wlr_xdg_toplevel_set_activated(t.p, C.bool(activated))
}
