package wlroots

// #include <wlr/types/wlr_keyboard.h>
import "C"
import (
	"unsafe"

	"github.com/swaywm/go-wlroots/xkb"
)

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
	listener := NewListener(func(data unsafe.Pointer) {
		cb(k)
	})

	C.wl_signal_add(&k.p.events.modifiers, listener.p)
}

func (k Keyboard) OnKey(cb func(keyboard Keyboard, time uint32, keyCode uint32, updateState bool, state KeyState)) {
	listener := NewListener(func(data unsafe.Pointer) {
		event := (*C.struct_wlr_event_keyboard_key)(data)
		cb(k, uint32(event.time_msec), uint32(event.keycode), bool(event.update_state), KeyState(event.state))
	})

	C.wl_signal_add(&k.p.events.key, listener.p)
}
