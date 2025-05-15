package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import (
	"unsafe"

	"github.com/swaywm/go-wlroots/xkb"
)

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/types/wlr_input_device.h>
// #include <wlr/types/wlr_keyboard.h>
// #include <wlr/types/wlr_pointer.h>
import "C"

type (
	KeyState          uint32
	KeyboardModifier  uint32
	KeyboardModifiers uint32
	KeyboardLed       uint32
)

const (
	KeyboardLedCount               = C.WLR_LED_COUNT
	KeyboardModifierCount          = C.WLR_MODIFIER_COUNT
	KeyboardKeysCap                = C.WLR_KEYBOARD_KEYS_CAP
	KeyStateReleased      KeyState = C.WL_KEYBOARD_KEY_STATE_RELEASED
	KeyStatePressed       KeyState = C.WL_KEYBOARD_KEY_STATE_PRESSED

	KeyboardModifierShift KeyboardModifier = C.WLR_MODIFIER_SHIFT
	KeyboardModifierCaps  KeyboardModifier = C.WLR_MODIFIER_CAPS
	KeyboardModifierCtrl  KeyboardModifier = C.WLR_MODIFIER_CTRL
	KeyboardModifierAlt   KeyboardModifier = C.WLR_MODIFIER_ALT
	KeyboardModifierMod2  KeyboardModifier = C.WLR_MODIFIER_MOD2
	KeyboardModifierMod3  KeyboardModifier = C.WLR_MODIFIER_MOD3
	KeyboardModifierLogo  KeyboardModifier = C.WLR_MODIFIER_LOGO
	KeyboardModifierMod5  KeyboardModifier = C.WLR_MODIFIER_MOD5

	KeyboardLedNumLock    KeyboardLed = C.WLR_LED_NUM_LOCK
	KeyboardLedCapsLock   KeyboardLed = C.WLR_LED_CAPS_LOCK
	KeyboardLedScrollLock KeyboardLed = C.WLR_LED_SCROLL_LOCK
)

/**
 * Get a struct wlr_keyboard from a struct wlr_input_device.
 *
 * Asserts that the input device is a keyboard.
 */
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

func (k Keyboard) Base() InputDevice {
	return InputDevice{p: &k.p.base}
}

func (k Keyboard) XKBState() xkb.State {
	return xkb.WrapState(unsafe.Pointer(k.p.xkb_state))
}

func (k Keyboard) Leds() int {
	return int(k.p.leds)
}

func (k Keyboard) Modifiers() KeyboardModifier {
	return KeyboardModifier(C.wlr_keyboard_get_modifiers(k.p))
}

func (k Keyboard) OnModifiers(cb func(keyboard Keyboard)) {
	man.add(unsafe.Pointer(k.p), &k.p.events.modifiers, func(data unsafe.Pointer) {
		cb(k)
	})
}

func (k Keyboard) OnDestroy(cb func(keyboard Keyboard)) {
	man.add(unsafe.Pointer(k.p), &k.p.base.events.destroy, func(data unsafe.Pointer) {
		cb(k)
	})
}

func (k Keyboard) OnKey(cb func(keyboard Keyboard, time uint32, keyCode uint32, updateState bool, state KeyState)) {
	man.add(unsafe.Pointer(k.p), &k.p.events.key, func(data unsafe.Pointer) {
		event := (*C.struct_wlr_keyboard_key_event)(data)
		cb(k, uint32(event.time_msec), uint32(event.keycode), bool(event.update_state), KeyState(event.state))
	})
}
