package wlroots

// #include <stdlib.h>
// #include <wlr/types/wlr_seat.h>
import "C"
import "unsafe"

type Seat struct {
	p *C.struct_wlr_seat
}

const (
	SeatCapabilityPointer  = C.WL_SEAT_CAPABILITY_POINTER
	SeatCapabilityKeyboard = C.WL_SEAT_CAPABILITY_KEYBOARD
	SeatCapabilityTouch    = C.WL_SEAT_CAPABILITY_TOUCH
)

func NewSeat(display Display, name string) Seat {
	s := C.CString(name)
	p := C.wlr_seat_create(display.p, s)
	C.free(unsafe.Pointer(s))
	return Seat{p: p}
}

func (s Seat) OnSetCursor(cb func()) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb()
	})

	C.wl_signal_add(&s.p.events.request_set_cursor, listener.p)
}

func (s Seat) SetCapabilities(caps uint32) {
	C.wlr_seat_set_capabilities(s.p, C.uint32_t(caps))
}

func (s Seat) SetKeyboard(dev InputDevice) {
	C.wlr_seat_set_keyboard(s.p, dev.p)
}
