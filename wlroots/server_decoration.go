package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import (
	"unsafe"
)

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/types/wlr_server_decoration.h>
import "C"

type ServerDecorationManagerMode uint32

const (
	ServerDecorationManagerModeNone   ServerDecorationManagerMode = C.WLR_SERVER_DECORATION_MANAGER_MODE_NONE
	ServerDecorationManagerModeClient ServerDecorationManagerMode = C.WLR_SERVER_DECORATION_MANAGER_MODE_CLIENT
	ServerDecorationManagerModeServer ServerDecorationManagerMode = C.WLR_SERVER_DECORATION_MANAGER_MODE_SERVER
)

type ServerDecorationManager struct {
	p *C.struct_wlr_server_decoration_manager
}

type ServerDecoration struct {
	p *C.struct_wlr_server_decoration
}

func NewServerDecorationManager(display Display) ServerDecorationManager {
	p := C.wlr_server_decoration_manager_create(display.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return ServerDecorationManager{p: p}
}

func (m ServerDecorationManager) OnDestroy(cb func(ServerDecorationManager)) {
	man.add(unsafe.Pointer(m.p), &m.p.events.destroy, func(unsafe.Pointer) {
		cb(m)
	})
}

func (m ServerDecorationManager) SetDefaultMode(mode ServerDecorationManagerMode) {
	C.wlr_server_decoration_manager_set_default_mode(m.p, C.uint32_t(mode))
}

func (m ServerDecorationManager) OnNewMode(cb func(ServerDecorationManager, ServerDecoration)) {
	man.add(unsafe.Pointer(m.p), &m.p.events.new_decoration, func(data unsafe.Pointer) {
		dec := ServerDecoration{
			p: (*C.struct_wlr_server_decoration)(data),
		}
		man.track(unsafe.Pointer(dec.p), &dec.p.events.destroy)
		cb(m, dec)
	})
}

func (d ServerDecoration) OnDestroy(cb func(ServerDecoration)) {
	man.add(unsafe.Pointer(d.p), &d.p.events.destroy, func(unsafe.Pointer) {
		cb(d)
	})
}

func (d ServerDecoration) OnMode(cb func(ServerDecoration)) {
	man.add(unsafe.Pointer(d.p), &d.p.events.mode, func(unsafe.Pointer) {
		cb(d)
	})
}

func (d ServerDecoration) Mode() ServerDecorationManagerMode {
	return ServerDecorationManagerMode(d.p.mode)
}
