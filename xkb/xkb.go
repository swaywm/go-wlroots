package xkb

// #include <xkbcommon/xkbcommon.h>
// #cgo pkg-config: xkbcommon
import "C"
import "unsafe"

type (
	KeySym  uint32
	KeyCode uint32
)

type Context struct {
	p *C.struct_xkb_context
}

type Keymap struct {
	p *C.struct_xkb_keymap
}

type State struct {
	p *C.struct_xkb_state
}

func WrapState(p unsafe.Pointer) State {
	return State{p: (*C.struct_xkb_state)(p)}
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

func (s State) Syms(keyCode uint32) []KeySym {
	var syms *C.xkb_keysym_t
	n := int(C.xkb_state_key_get_syms(s.p, C.uint32_t(keyCode), &syms))
	slice := (*[1 << 30]C.xkb_keysym_t)(unsafe.Pointer(syms))[:n:n]

	res := make([]KeySym, n)
	for i := 0; i < n; i++ {
		res[i] = KeySym(slice[i])
	}

	return res
}
