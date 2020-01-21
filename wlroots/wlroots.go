package wlroots

// #include <stdarg.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include <time.h>
// #include <wayland-server.h>
// #include <wlr/backend.h>
// #include <wlr/backend/wayland.h>
// #include <wlr/backend/x11.h>
// #include <wlr/render/wlr_renderer.h>
// #include <wlr/render/wlr_texture.h>
// #include <wlr/types/wlr_box.h>
// #include <wlr/types/wlr_compositor.h>
// #include <wlr/types/wlr_cursor.h>
// #include <wlr/types/wlr_data_device.h>
// #include <wlr/types/wlr_server_decoration.h>
// #include <wlr/types/wlr_linux_dmabuf_v1.h>
// #include <wlr/types/wlr_input_device.h>
// #include <wlr/types/wlr_keyboard.h>
// #include <wlr/types/wlr_matrix.h>
// #include <wlr/types/wlr_output.h>
// #include <wlr/types/wlr_output_layout.h>
// #include <wlr/types/wlr_seat.h>
// #include <wlr/types/wlr_surface.h>
// #include <wlr/types/wlr_xcursor_manager.h>
// #include <wlr/types/wlr_xdg_shell.h>
// #include <wlr/util/edges.h>
// #include <wlr/util/log.h>
// #include <wlr/xwayland.h>
//
// void _wlr_log_cb(enum wlr_log_importance importance, char *msg);
//
// static inline void _wlr_log_inner_cb(enum wlr_log_importance importance, const char *fmt, va_list args) {
//		char *msg = NULL;
//		if (vasprintf(&msg, fmt, args) == -1) {
//			return;
//		}
//
//		_wlr_log_cb(importance, msg);
//		free(msg);
// }
//
// static inline void _wlr_log_set_cb(enum wlr_log_importance verbosity, bool is_set) {
//		wlr_log_init(verbosity, is_set ? &_wlr_log_inner_cb : NULL);
// }
//
// void _wlr_xdg_surface_for_each_cb(struct wlr_surface *surface, int sx, int sy, void *data);
//
// static inline void _wlr_xdg_surface_for_each_surface(struct wlr_xdg_surface *surface, void *user_data) {
//		wlr_xdg_surface_for_each_surface(surface, &_wlr_xdg_surface_for_each_cb, user_data);
// }
//
// void _wl_listener_cb(struct wl_listener *listener, void *data);
//
// static inline void _wl_listener_set_cb(struct wl_listener *listener) {
//		listener->notify = &_wl_listener_cb;
// }
//
// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
import "C"

import (
	"errors"
	"sync"
	"time"
	"unsafe"

	"github.com/swaywm/go-wlroots/xkb"
)

type XWayland struct {
	p *C.struct_wlr_xwayland
}

type XWaylandSurface struct {
	p *C.struct_wlr_xwayland_surface
}

func NewXWayland(display Display, compositor Compositor, lazy bool) XWayland {
	p := C.wlr_xwayland_create(display.p, compositor.p, C.bool(lazy))
	return XWayland{p: p}
}

func (x XWayland) Destroy() {
	C.wlr_xwayland_destroy(x.p)
}

func (x XWayland) Display() int {
	return int(x.p.display)
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

func (s XWaylandSurface) Surface() Surface {
	return Surface{p: s.p.surface}
}

func (s XWaylandSurface) Geometry() Box {
	return Box{
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

type XDGSurfaceRole uint32

const (
	XDGSurfaceRoleNone     XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_NONE
	XDGSurfaceRoleTopLevel XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_TOPLEVEL
	XDGSurfaceRolePopup    XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_POPUP
)

var (
	// TODO: guard this with a mutex
	xdgSurfaceWalkers      = map[*C.struct_wlr_xdg_surface]XDGSurfaceWalkFunc{}
	xdgSurfaceWalkersMutex sync.RWMutex
)

type XDGShell struct {
	p *C.struct_wlr_xdg_shell
}

type XDGSurface struct {
	p *C.struct_wlr_xdg_surface
}

type XDGPopup struct {
	p *C.struct_wlr_xdg_popup
}

type XDGSurfaceWalkFunc func(surface Surface, sx int, sy int)

type XDGTopLevel struct {
	p *C.struct_wlr_xdg_toplevel
}

func NewXDGShell(display Display) XDGShell {
	p := C.wlr_xdg_shell_create(display.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return XDGShell{p: p}
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

func (s XDGSurface) Nil() bool {
	return s.p == nil
}

func (s XDGSurface) Walk(visit XDGSurfaceWalkFunc) {
	xdgSurfaceWalkersMutex.Lock()
	xdgSurfaceWalkers[s.p] = visit
	xdgSurfaceWalkersMutex.Unlock()

	C._wlr_xdg_surface_for_each_surface(s.p, unsafe.Pointer(s.p))

	xdgSurfaceWalkersMutex.Lock()
	delete(xdgSurfaceWalkers, s.p)
	xdgSurfaceWalkersMutex.Unlock()
}

func (s XDGSurface) Role() XDGSurfaceRole {
	return XDGSurfaceRole(s.p.role)
}

func (s XDGSurface) TopLevel() XDGTopLevel {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&s.p.anon0[0]))
	return XDGTopLevel{p: (*C.struct_wlr_xdg_toplevel)(p)}
}

func (s XDGSurface) TopLevelSetActivated(activated bool) {
	C.wlr_xdg_toplevel_set_activated(s.p, C.bool(activated))
}

func (s XDGSurface) TopLevelSetSize(width uint32, height uint32) {
	C.wlr_xdg_toplevel_set_size(s.p, C.uint32_t(width), C.uint32_t(height))
}

func (s XDGSurface) TopLevelSetTiled(edges Edges) {
	C.wlr_xdg_toplevel_set_tiled(s.p, C.uint32_t(edges))
}

func (s XDGSurface) SendClose() {
	C.wlr_xdg_toplevel_send_close(s.p)
}

func (s XDGSurface) Ping() {
	C.wlr_xdg_surface_ping(s.p)
}

func (s XDGSurface) Surface() Surface {
	return Surface{p: s.p.surface}
}

func (s XDGSurface) SurfaceAt(sx float64, sy float64) (surface Surface, subX float64, subY float64) {
	var csubX, csubY C.double
	p := C.wlr_xdg_surface_surface_at(s.p, C.double(sx), C.double(sy), &csubX, &csubY)
	return Surface{p: p}, float64(csubX), float64(csubY)
}

func (s XDGSurface) OnMap(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events._map, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XDGSurface) OnUnmap(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.unmap, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XDGSurface) OnDestroy(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.destroy, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XDGSurface) OnPingTimeout(cb func(XDGSurface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.ping_timeout, func(data unsafe.Pointer) {
		cb(s)
	})
}

func (s XDGSurface) OnNewPopup(cb func(XDGSurface, XDGPopup)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.ping_timeout, func(data unsafe.Pointer) {
		popup := XDGPopup{p: (*C.struct_wlr_xdg_popup)(data)}
		cb(s, popup)
	})
}

func (s XDGSurface) Geometry() Box {
	var cb C.struct_wlr_box
	C.wlr_xdg_surface_get_geometry(s.p, &cb)

	var b Box
	b.fromC(&cb)
	return b
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

func (s XDGTopLevel) Nil() bool {
	return s.p == nil
}

func (t XDGTopLevel) Title() string {
	return C.GoString(t.p.title)
}

type XCursor struct {
	p *C.struct_wlr_xcursor
}

type XCursorImage struct {
	p *C.struct_wlr_xcursor_image
}

type XCursorManager struct {
	p *C.struct_wlr_xcursor_manager
}

func NewXCursorManager() XCursorManager {
	p := C.wlr_xcursor_manager_create(nil, 24)
	return XCursorManager{p: p}
}

func (m XCursorManager) Destroy() {
	C.wlr_xcursor_manager_destroy(m.p)
}

func (m XCursorManager) Load() {
	C.wlr_xcursor_manager_load(m.p, 1)
}

func (m XCursorManager) SetCursorImage(cursor Cursor, name string) {
	s := C.CString(name)
	C.wlr_xcursor_manager_set_cursor_image(m.p, s, cursor.p)
	C.free(unsafe.Pointer(s))
}

func (m XCursorManager) XCursor(name string, scale float32) XCursor {
	s := C.CString(name)
	p := C.wlr_xcursor_manager_get_xcursor(m.p, s, C.float(scale))
	C.free(unsafe.Pointer(s))
	return XCursor{p: p}
}

func (c XCursor) Image(i int) XCursorImage {
	n := c.ImageCount()
	slice := (*[1 << 30]*C.struct_wlr_xcursor_image)(unsafe.Pointer(c.p.images))[:n:n]
	return XCursorImage{p: slice[i]}
}

func (c XCursor) Images() []XCursorImage {
	images := make([]XCursorImage, 0, c.ImageCount())
	for i := 0; i < cap(images); i++ {
		images = append(images, c.Image(i))
	}
	return images
}

func (c XCursor) ImageCount() int {
	return int(c.p.image_count)
}

func (c XCursor) Name() string {
	return C.GoString(c.p.name)
}

type Edges uint32

const (
	EdgeNone   Edges = C.WLR_EDGE_NONE
	EdgeTop    Edges = C.WLR_EDGE_TOP
	EdgeBottom Edges = C.WLR_EDGE_BOTTOM
	EdgeLeft   Edges = C.WLR_EDGE_LEFT
	EdgeRight  Edges = C.WLR_EDGE_RIGHT
)

type Texture struct {
	p *C.struct_wlr_texture
}

func (t Texture) Destroy() {
	C.wlr_texture_destroy(t.p)
}

func (t Texture) Nil() bool {
	return t.p == nil
}

type SurfaceType uint32

const (
	SurfaceTypeNone SurfaceType = iota
	SurfaceTypeXDG
	SurfaceTypeXWayland
)

type Surface struct {
	p *C.struct_wlr_surface
}

type SurfaceState struct {
	s C.struct_wlr_surface_state
}

func (s Surface) Nil() bool {
	return s.p == nil
}

func (s Surface) OnDestroy(cb func(Surface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.destroy, func(unsafe.Pointer) {
		cb(s)
	})
}

func (s Surface) Type() SurfaceType {
	if C.wlr_surface_is_xdg_surface(s.p) {
		return SurfaceTypeXDG
	} else if C.wlr_surface_is_xwayland_surface(s.p) {
		return SurfaceTypeXWayland
	}

	return SurfaceTypeNone
}

func (s Surface) SurfaceAt(sx float64, sy float64) (surface Surface, subX float64, subY float64) {
	var csubX, csubY C.double
	p := C.wlr_surface_surface_at(s.p, C.double(sx), C.double(sy), &csubX, &csubY)
	return Surface{p: p}, float64(csubX), float64(csubY)
}

func (s Surface) Texture() Texture {
	p := C.wlr_surface_get_texture(s.p)
	return Texture{p: p}
}

func (s Surface) CurrentState() SurfaceState {
	return SurfaceState{s: s.p.current}
}

func (s Surface) Walk(visit func()) {
	panic("not implemented")
}

func (s Surface) SendFrameDone(when time.Time) {
	t := C.struct_timespec{}
	C.wlr_surface_send_frame_done(s.p, &t)
}

func (s Surface) XDGSurface() XDGSurface {
	p := C.wlr_xdg_surface_from_wlr_surface(s.p)
	return XDGSurface{p: p}
}

func (s Surface) XWaylandSurface() XWaylandSurface {
	p := C.wlr_xwayland_surface_from_wlr_surface(s.p)
	return XWaylandSurface{p: p}
}

func (s SurfaceState) Width() int {
	return int(s.s.width)
}

func (s SurfaceState) Height() int {
	return int(s.s.height)
}

func (s SurfaceState) Transform() uint32 {
	return uint32(s.s.transform)
}

type Seat struct {
	p *C.struct_wlr_seat
}

type SeatClient struct {
	p *C.struct_wlr_seat_client
}

type SeatKeyboardState struct {
	s C.struct_wlr_seat_keyboard_state
}

type SeatPointerState struct {
	s C.struct_wlr_seat_pointer_state
}

type SeatCapability uint32

const (
	SeatCapabilityPointer  SeatCapability = C.WL_SEAT_CAPABILITY_POINTER
	SeatCapabilityKeyboard SeatCapability = C.WL_SEAT_CAPABILITY_KEYBOARD
	SeatCapabilityTouch    SeatCapability = C.WL_SEAT_CAPABILITY_TOUCH
)

func NewSeat(display Display, name string) Seat {
	s := C.CString(name)
	p := C.wlr_seat_create(display.p, s)
	C.free(unsafe.Pointer(s))
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Seat{p: p}
}

func (s Seat) Destroy() {
	C.wlr_seat_destroy(s.p)
}

func (s Seat) OnDestroy(cb func(Seat)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.destroy, func(unsafe.Pointer) {
		cb(s)
	})
}

func (s Seat) OnSetCursorRequest(cb func(client SeatClient, surface Surface, serial uint32, hotspotX int32, hotspotY int32)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.request_set_cursor, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_seat_pointer_request_set_cursor_event)(data)
		client := SeatClient{p: event.seat_client}
		surface := Surface{p: event.surface}
		cb(client, surface, uint32(event.serial), int32(event.hotspot_x), int32(event.hotspot_y))
	})
}

func (s Seat) SetCapabilities(caps SeatCapability) {
	C.wlr_seat_set_capabilities(s.p, C.uint32_t(caps))
}

func (s Seat) SetKeyboard(dev InputDevice) {
	C.wlr_seat_set_keyboard(s.p, dev.p)
}

func (s Seat) NotifyPointerButton(time uint32, button uint32, state ButtonState) {
	C.wlr_seat_pointer_notify_button(s.p, C.uint32_t(time), C.uint32_t(button), uint32(state))
}

func (s Seat) NotifyPointerAxis(time uint32, orientation AxisOrientation, delta float64, deltaDiscrete int32, source AxisSource) {
	C.wlr_seat_pointer_notify_axis(s.p, C.uint32_t(time), C.enum_wlr_axis_orientation(orientation), C.double(delta), C.int32_t(deltaDiscrete), C.enum_wlr_axis_source(source))
}

func (s Seat) NotifyPointerEnter(surface Surface, sx float64, sy float64) {
	C.wlr_seat_pointer_notify_enter(s.p, surface.p, C.double(sx), C.double(sy))
}

func (s Seat) NotifyPointerMotion(time uint32, sx float64, sy float64) {
	C.wlr_seat_pointer_notify_motion(s.p, C.uint32_t(time), C.double(sx), C.double(sy))
}

func (s Seat) NotifyPointerFrame() {
	C.wlr_seat_pointer_notify_frame(s.p)
}

func (s Seat) NotifyKeyboardEnter(surface Surface, k Keyboard) {
	C.wlr_seat_keyboard_notify_enter(s.p, surface.p, &k.p.keycodes[0], k.p.num_keycodes, &k.p.modifiers)
}

func (s Seat) NotifyKeyboardModifiers(k Keyboard) {
	C.wlr_seat_keyboard_notify_modifiers(s.p, &k.p.modifiers)
}

func (s Seat) NotifyKeyboardKey(time uint32, keyCode uint32, state KeyState) {
	C.wlr_seat_keyboard_notify_key(s.p, C.uint32_t(time), C.uint32_t(keyCode), C.uint32_t(state))
}

func (s Seat) ClearPointerFocus() {
	C.wlr_seat_pointer_clear_focus(s.p)
}

func (s Seat) Keyboard() Keyboard {
	p := C.wlr_seat_get_keyboard(s.p)
	return Keyboard{p: p}
}

func (s Seat) KeyboardState() SeatKeyboardState {
	return SeatKeyboardState{s: s.p.keyboard_state}
}

func (s Seat) PointerState() SeatPointerState {
	return SeatPointerState{s: s.p.pointer_state}
}

func (s SeatKeyboardState) FocusedSurface() Surface {
	return Surface{p: s.s.focused_surface}
}

func (s SeatPointerState) FocusedSurface() Surface {
	return Surface{p: s.s.focused_surface}
}

func (s SeatPointerState) FocusedClient() SeatClient {
	return SeatClient{p: s.s.focused_client}
}

type Renderer struct {
	p *C.struct_wlr_renderer
}

func (r Renderer) Destroy() {
	C.wlr_renderer_destroy(r.p)
}

func (r Renderer) OnDestroy(cb func(Renderer)) {
	man.add(unsafe.Pointer(r.p), &r.p.events.destroy, func(unsafe.Pointer) {
		cb(r)
	})
}

func (r Renderer) InitDisplay(display Display) {
	C.wlr_renderer_init_wl_display(r.p, display.p)
}

func (r Renderer) Begin(output Output, width int, height int) {
	C.wlr_renderer_begin(r.p, C.int(width), C.int(height))
}

func (r Renderer) Clear(color *Color) {
	c := color.toC()
	C.wlr_renderer_clear(r.p, &c[0])
}

func (r Renderer) End() {
	C.wlr_renderer_end(r.p)
}

func (r Renderer) RenderTextureWithMatrix(texture Texture, matrix *Matrix, alpha float32) {
	m := matrix.toC()
	C.wlr_render_texture_with_matrix(r.p, texture.p, &m[0], C.float(alpha))
}

func (r Renderer) RenderRect(box *Box, color *Color, projection *Matrix) {
	b := box.toC()
	c := color.toC()
	pm := projection.toC()
	C.wlr_render_rect(r.p, &b, &c[0], &pm[0])
}

type Output struct {
	p *C.struct_wlr_output
}

type OutputMode struct {
	p *C.struct_wlr_output_mode
}

func wrapOutput(p unsafe.Pointer) Output {
	return Output{p: (*C.struct_wlr_output)(p)}
}

func (o Output) OnDestroy(cb func(Output)) {
	man.add(unsafe.Pointer(o.p), &o.p.events.destroy, func(unsafe.Pointer) {
		cb(o)
	})
}

func (o Output) Name() string {
	return C.GoString(&o.p.name[0])
}

func (o Output) Scale() float32 {
	return float32(o.p.scale)
}

func (o Output) TransformMatrix() Matrix {
	var matrix Matrix
	matrix.fromC(&o.p.transform_matrix)
	return matrix
}

func (o Output) OnFrame(cb func(Output)) {
	man.add(unsafe.Pointer(o.p), &o.p.events.frame, func(data unsafe.Pointer) {
		cb(o)
	})
}

func (o Output) RenderSoftwareCursors() {
	C.wlr_output_render_software_cursors(o.p, nil)
}

func (o Output) TransformedResolution() (int, int) {
	var width, height C.int
	C.wlr_output_transformed_resolution(o.p, &width, &height)
	return int(width), int(height)
}

func (o Output) EffectiveResolution() (int, int) {
	var width, height C.int
	C.wlr_output_effective_resolution(o.p, &width, &height)
	return int(width), int(height)
}

func (o Output) AttachRender() (int, error) {
	var bufferAge C.int
	if !C.wlr_output_attach_render(o.p, &bufferAge) {
		return 0, errors.New("can't make output context current")
	}

	return int(bufferAge), nil
}

func (o Output) Rollback() {
	C.wlr_output_rollback(o.p)
}

func (o Output) CreateGlobal() {
	C.wlr_output_create_global(o.p)
}

func (o Output) DestroyGlobal() {
	C.wlr_output_destroy_global(o.p)
}

func (o Output) Commit() {
	C.wlr_output_commit(o.p)
}

func (o Output) Modes() []OutputMode {
	// TODO: figure out what to do with this ridiculous for loop
	// perhaps this can be refactored into a less ugly hack that uses reflection
	var modes []OutputMode
	var mode *C.struct_wlr_output_mode
	for mode := (*C.struct_wlr_output_mode)(unsafe.Pointer(uintptr(unsafe.Pointer(o.p.modes.next)) - unsafe.Offsetof(mode.link))); &mode.link != &o.p.modes; mode = (*C.struct_wlr_output_mode)(unsafe.Pointer(uintptr(unsafe.Pointer(mode.link.next)) - unsafe.Offsetof(mode.link))) {
		modes = append(modes, OutputMode{p: mode})
	}

	return modes
}

func (o Output) SetMode(mode OutputMode) {
	C.wlr_output_set_mode(o.p, mode.p)
}

func (o Output) Enable(enable bool) {
	C.wlr_output_enable(o.p, C.bool(enable))
}

func (o Output) SetTitle(title string) error {
	if C.wlr_output_is_wl(o.p) {
		C.wlr_wl_output_set_title(o.p, C.CString(title))
	} else if C.wlr_output_is_x11(o.p) {
		C.wlr_x11_output_set_title(o.p, C.CString(title))
	} else {
		return errors.New("this output type cannot have a title")
	}

	return nil
}

type OutputLayout struct {
	p *C.struct_wlr_output_layout
}

func NewOutputLayout() OutputLayout {
	p := C.wlr_output_layout_create()
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return OutputLayout{p: p}
}

func (l OutputLayout) Destroy() {
	C.wlr_output_layout_destroy(l.p)
}

func (l OutputLayout) AddOutputAuto(output Output) {
	C.wlr_output_layout_add_auto(l.p, output.p)
}

func (l OutputLayout) Coords(output Output) (x float64, y float64) {
	var ox, oy C.double
	C.wlr_output_layout_output_coords(l.p, output.p, &ox, &oy)
	return float64(ox), float64(oy)
}

func OutputTransformInvert(transform uint32) uint32 {
	return uint32(C.wlr_output_transform_invert(C.enum_wl_output_transform(transform)))
}

func (m OutputMode) Width() int32 {
	return int32(m.p.width)
}

func (m OutputMode) Height() int32 {
	return int32(m.p.height)
}

func (m OutputMode) RefreshRate() int32 {
	return int32(m.p.refresh)
}

type Matrix [9]float32

func (m *Matrix) ProjectBox(box *Box, transform uint32, rotation float32, projection *Matrix) {
	cm := m.toC()
	b := box.toC()
	pm := projection.toC()
	C.wlr_matrix_project_box(&cm[0], &b, C.enum_wl_output_transform(transform), C.float(rotation), &pm[0])
	m.fromC(&cm)
}

func (m *Matrix) toC() [9]C.float {
	var cm [9]C.float
	for i := range m {
		cm[i] = C.float(m[i])
	}
	return cm
}

func (m *Matrix) fromC(cm *[9]C.float) {
	for i := range cm {
		m[i] = float32(cm[i])
	}
}

type (
	LogImportance uint32
	LogFunc       func(importance LogImportance, msg string)
)

const (
	LogImportanceSilent LogImportance = C.WLR_SILENT
	LogImportanceError  LogImportance = C.WLR_ERROR
	LogImportanceInfo   LogImportance = C.WLR_INFO
	LogImportanceDebug  LogImportance = C.WLR_DEBUG
)

var (
	onLog LogFunc
)

//export _wlr_log_cb
func _wlr_log_cb(importance LogImportance, msg *C.char) {
	if onLog != nil {
		onLog(importance, C.GoString(msg))
	}
}

func OnLog(verbosity LogImportance, cb LogFunc) {
	C._wlr_log_set_cb(C.enum_wlr_log_importance(verbosity), cb != nil)
	onLog = cb
}

type (
	KeyState         uint32
	KeyboardModifier uint32
)

const (
	KeyStateReleased KeyState = C.WLR_KEY_RELEASED
	KeyStatePressed  KeyState = C.WLR_KEY_PRESSED

	KeyboardModifierShift KeyboardModifier = C.WLR_MODIFIER_SHIFT
	KeyboardModifierCaps  KeyboardModifier = C.WLR_MODIFIER_CAPS
	KeyboardModifierCtrl  KeyboardModifier = C.WLR_MODIFIER_CTRL
	KeyboardModifierAlt   KeyboardModifier = C.WLR_MODIFIER_ALT
	KeyboardModifierMod2  KeyboardModifier = C.WLR_MODIFIER_MOD2
	KeyboardModifierMod3  KeyboardModifier = C.WLR_MODIFIER_MOD3
	KeyboardModifierLogo  KeyboardModifier = C.WLR_MODIFIER_LOGO
	KeyboardModifierMod5  KeyboardModifier = C.WLR_MODIFIER_MOD5
)

type Keyboard struct {
	p *C.struct_wlr_keyboard
}

func (k Keyboard) SetKeymap(keymap xkb.Keymap) {
	C.wlr_keyboard_set_keymap(k.p, (*C.struct_xkb_keymap)(keymap.Ptr()))
}

func (k Keyboard) RepeatInfo() (rate int32, delay int32) {
	return int32(k.p.repeat_info.rate), int32(k.p.repeat_info.delay)
}

func (k Keyboard) SetRepeatInfo(rate int32, delay int32) {
	C.wlr_keyboard_set_repeat_info(k.p, C.int32_t(rate), C.int32_t(delay))
}

func (k Keyboard) XKBState() xkb.State {
	return xkb.WrapState(unsafe.Pointer(k.p.xkb_state))
}

func (k Keyboard) Modifiers() KeyboardModifier {
	return KeyboardModifier(C.wlr_keyboard_get_modifiers(k.p))
}

func (k Keyboard) OnModifiers(cb func(keyboard Keyboard)) {
	man.add(unsafe.Pointer(k.p), &k.p.events.modifiers, func(data unsafe.Pointer) {
		cb(k)
	})
}

func (k Keyboard) OnKey(cb func(keyboard Keyboard, time uint32, keyCode uint32, updateState bool, state KeyState)) {
	man.add(unsafe.Pointer(k.p), &k.p.events.key, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_keyboard_key)(data)
		cb(k, uint32(event.time_msec), uint32(event.keycode), bool(event.update_state), KeyState(event.state))
	})
}

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

func (d InputDevice) OnDestroy(cb func(InputDevice)) {
	man.add(unsafe.Pointer(d.p), &d.p.events.destroy, func(unsafe.Pointer) {
		cb(d)
	})
}

func (d InputDevice) Type() InputDeviceType {
	return InputDeviceType(d.p._type)
}

func (d InputDevice) Keyboard() Keyboard {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&d.p.anon0[0]))
	return Keyboard{p: (*C.struct_wlr_keyboard)(p)}
}

func wrapInputDevice(p unsafe.Pointer) InputDevice {
	return InputDevice{p: (*C.struct_wlr_input_device)(p)}
}

type DMABuf struct {
	p *C.struct_wlr_linux_dmabuf_v1
}

func NewDMABuf(display Display, renderer Renderer) DMABuf {
	p := C.wlr_linux_dmabuf_v1_create(display.p, renderer.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return DMABuf{p: p}
}

func (b DMABuf) OnDestroy(cb func(DMABuf)) {
	man.add(unsafe.Pointer(b.p), &b.p.events.destroy, func(unsafe.Pointer) {
		cb(b)
	})
}

type Display struct {
	p *C.struct_wl_display
}

func NewDisplay() Display {
	p := C.wl_display_create()
	d := Display{p: p}
	d.OnDestroy(func(Display) {
		man.delete(unsafe.Pointer(p))
	})
	return d
}

func (d Display) Destroy() {
	C.wl_display_destroy(d.p)
}

func (d Display) OnDestroy(cb func(Display)) {
	l := man.add(unsafe.Pointer(d.p), nil, func(data unsafe.Pointer) {
		cb(d)
	})
	C.wl_display_add_destroy_listener(d.p, l.p)
}

func (d Display) Run() {
	C.wl_display_run(d.p)
}

func (d Display) Terminate() {
	C.wl_display_terminate(d.p)
}

func (d Display) AddSocketAuto() (string, error) {
	socket := C.wl_display_add_socket_auto(d.p)
	if socket == nil {
		return "", errors.New("can't auto add wayland socket")
	}

	return C.GoString(socket), nil
}

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

type DataDeviceManager struct {
	p *C.struct_wlr_data_device_manager
}

func NewDataDeviceManager(display Display) DataDeviceManager {
	p := C.wlr_data_device_manager_create(display.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return DataDeviceManager{p: p}
}

func (m DataDeviceManager) OnDestroy(cb func(DataDeviceManager)) {
	man.add(unsafe.Pointer(m.p), &m.p.events.destroy, func(unsafe.Pointer) {
		cb(m)
	})
}

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
		cb(dev, uint32(event.time_msec), uint32(event.button), ButtonState(event.state))
	})
}

func (c Cursor) OnAxis(cb func(dev InputDevice, time uint32, source AxisSource, orientation AxisOrientation, delta float64, deltaDiscrete int32)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.axis, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_pointer_axis)(data)
		dev := InputDevice{p: event.device}
		cb(dev, uint32(event.time_msec), AxisSource(event.source), AxisOrientation(event.orientation), float64(event.delta), int32(event.delta_discrete))
	})
}

func (c Cursor) OnFrame(cb func()) {
	man.add(unsafe.Pointer(c.p), &c.p.events.frame, func(data unsafe.Pointer) {
		cb()
	})
}

type Compositor struct {
	p *C.struct_wlr_compositor
}

func NewCompositor(display Display, renderer Renderer) Compositor {
	p := C.wlr_compositor_create(display.p, renderer.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Compositor{p: p}
}

func (c Compositor) OnDestroy(cb func(Compositor)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.destroy, func(unsafe.Pointer) {
		cb(c)
	})
}

type Color struct {
	R, G, B, A float32
}

func (c *Color) Set(r, g, b, a float32) {
	c.R = r
	c.G = g
	c.B = b
	c.A = a
}

func (c *Color) toC() [4]C.float {
	return [...]C.float{
		C.float(c.R),
		C.float(c.G),
		C.float(c.B),
		C.float(c.A),
	}
}

type Box struct {
	X, Y, Width, Height int
}

func (b *Box) Set(x, y, width, height int) {
	b.X = x
	b.Y = y
	b.Width = width
	b.Height = height
}

func (b *Box) toC() C.struct_wlr_box {
	return C.struct_wlr_box{
		x:      C.int(b.X),
		y:      C.int(b.Y),
		width:  C.int(b.Width),
		height: C.int(b.Height),
	}
}

func (b *Box) fromC(cb *C.struct_wlr_box) {
	b.X = int(cb.x)
	b.Y = int(cb.y)
	b.Width = int(cb.width)
	b.Height = int(cb.height)
}

type Backend struct {
	p *C.struct_wlr_backend
}

func NewBackend(display Display) Backend {
	p := C.wlr_backend_autocreate(display.p, nil)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Backend{p: p}
}

func (b Backend) Destroy() {
	C.wlr_backend_destroy(b.p)
}

func (b Backend) OnDestroy(cb func(Backend)) {
	man.add(unsafe.Pointer(b.p), &b.p.events.destroy, func(unsafe.Pointer) {
		cb(b)
	})
}

func (b Backend) Start() error {
	if !C.wlr_backend_start(b.p) {
		return errors.New("can't start backend")
	}

	return nil
}

func (b Backend) OnNewOutput(cb func(Output)) {
	man.add(unsafe.Pointer(b.p), &b.p.events.new_output, func(data unsafe.Pointer) {
		output := wrapOutput(data)
		man.track(unsafe.Pointer(output.p), &output.p.events.destroy)
		cb(output)
	})
}

func (b Backend) OnNewInput(cb func(InputDevice)) {
	man.add(unsafe.Pointer(b.p), &b.p.events.new_input, func(data unsafe.Pointer) {
		dev := wrapInputDevice(data)
		man.add(unsafe.Pointer(dev.p), &dev.p.events.destroy, func(data unsafe.Pointer) {
			// delete the underlying device type first
			man.delete(*(*unsafe.Pointer)(unsafe.Pointer(&dev.p.anon0[0])))
			// then delete the wlr_input_device itself
			man.delete(unsafe.Pointer(dev.p))
		})
		cb(dev)
	})
}

func (b Backend) Renderer() Renderer {
	p := C.wlr_backend_get_renderer(b.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Renderer{p: p}
}

// This whole mess has to exist for a number of reasons:
//
// 1. We need to allocate all instances of wl_listener on the heap as storing Go
// pointers in C after a cgo call returns is not allowed.
//
// 2. The wlroots library implicitly destroys objects when wl_display is
// destroyed. So, we need to keep track of all objects (and their listeners)
// manually and listen for the destroy signal to be able to free everything.
//
// 3 (TODO). As we're keeping track of all objects anyway, we might as well
// store a Go pointer to the wrapper struct along with them in order to be able
// to pass the same Go pointer through callbacks every time. This will also
// allow calling runtime.SetFinalizer on some of them to clean them up early
// when the GC notices it has gone out of scope.
//
// Send help.

type (
	listenerCallback func(data unsafe.Pointer)
)

type manager struct {
	mutex     sync.RWMutex
	objects   map[unsafe.Pointer][]*listener
	listeners map[*C.struct_wl_listener]*listener
}

type listener struct {
	p   *C.struct_wl_listener
	s   *C.struct_wl_signal
	cbs []listenerCallback
}

var (
	man = &manager{
		objects:   map[unsafe.Pointer][]*listener{},
		listeners: map[*C.struct_wl_listener]*listener{},
	}
)

//export _wl_listener_cb
func _wl_listener_cb(listener *C.struct_wl_listener, data unsafe.Pointer) {
	man.mutex.RLock()
	l := man.listeners[listener]
	man.mutex.RUnlock()
	for _, cb := range l.cbs {
		cb(data)
	}
}

func (m *manager) add(p unsafe.Pointer, signal *C.struct_wl_signal, cb listenerCallback) *listener {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// if a listener for this object and signal already exists, add the callback
	// to the existing listener
	if signal != nil {
		for _, l := range m.objects[p] {
			if l.s != nil && l.s == signal {
				l.cbs = append(l.cbs, cb)
				return l
			}
		}
	}

	lp := (*C.struct_wl_listener)(C.calloc(C.sizeof_struct_wl_listener, 1))
	C._wl_listener_set_cb(lp)
	if signal != nil {
		C.wl_signal_add((*C.struct_wl_signal)(signal), lp)
	}

	l := &listener{
		p:   lp,
		s:   signal,
		cbs: []listenerCallback{cb},
	}
	m.listeners[lp] = l
	m.objects[p] = append(m.objects[p], l)

	return l
}

func (m *manager) has(p unsafe.Pointer) bool {
	m.mutex.RLock()
	_, found := m.objects[p]
	m.mutex.RUnlock()
	return found
}

func (m *manager) delete(p unsafe.Pointer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, l := range m.objects[p] {
		delete(m.listeners, l.p)

		// remove the listener from the signal
		C.wl_list_remove(&l.p.link)

		// free the listener
		C.free(unsafe.Pointer(l.p))
	}

	delete(m.objects, p)
}

func (m *manager) track(p unsafe.Pointer, destroySignal *C.struct_wl_signal) {
	m.add(p, destroySignal, func(data unsafe.Pointer) { m.delete(p) })
}
