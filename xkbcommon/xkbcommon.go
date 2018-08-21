package xkbcommon

// #include <xkbcommon/xkbcommon.h>
// #cgo pkg-config: xkbcommon
import "C"
import "unsafe"

type Context struct {
	p *C.struct_xkb_context
}

type Keymap struct {
	p *C.struct_xkb_keymap
}

func NewContext() Context {
	p := C.xkb_context_new(C.XKB_CONTEXT_NO_FLAGS)
	return Context{p: p}
}

func (c Context) Destroy() {
	C.xkb_context_unref(c.p)
}

func (c Context) Map() Keymap {
	var rules C.struct_xkb_rule_names
	p := C.xkb_keymap_new_from_names(c.p, &rules, C.XKB_KEYMAP_COMPILE_NO_FLAGS)
	return Keymap{p: p}
}

func (m Keymap) Ptr() unsafe.Pointer {
	return unsafe.Pointer(m.p)
}

func (m Keymap) Destroy() {
	C.xkb_keymap_unref(m.p)
}
