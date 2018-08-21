package wlroots

// #include <wlr/types/wlr_keyboard.h>
import "C"
import (
	"unsafe"

	"github.com/alexbakker/go-wlroots/xkbcommon"
)

type Keyboard struct {
	p *C.struct_wlr_keyboard
}

func (k Keyboard) SetKeymap(keymap xkbcommon.Keymap) {
	C.wlr_keyboard_set_keymap(k.p, (*C.struct_xkb_keymap)(keymap.Ptr()))
}

func (k Keyboard) RepeatInfo() (rate int32, delay int32) {
	return int32(k.p.repeat_info.rate), int32(k.p.repeat_info.delay)
}

func (k Keyboard) SetRepeatInfo(rate int32, delay int32) {
	C.wlr_keyboard_set_repeat_info(k.p, C.int32_t(rate), C.int32_t(delay))
}

func (k Keyboard) OnModifiers(cb func()) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb()
	})

	C.wl_signal_add(&k.p.events.modifiers, listener.p)
}

func (k Keyboard) OnKey(cb func()) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb()
	})

	C.wl_signal_add(&k.p.events.key, listener.p)
}
