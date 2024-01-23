package xkb

// #include <stdlib.h>
// #include <xkbcommon/xkbcommon.h>
// #cgo pkg-config: xkbcommon
import "C"
import "unsafe"

type (
	KeyCode     uint32
	KeySymFlags uint32
)

const (
	KeySymFlagNoFlags         KeySymFlags = C.XKB_KEYSYM_NO_FLAGS
	KeySymFlagCaseInsensitive KeySymFlags = C.XKB_KEYSYM_CASE_INSENSITIVE
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

func NewContext(ksf KeySymFlags) Context {
	p := C.xkb_context_new(uint32(ksf))
	return Context{p: p}
}

func SymFromName(name string, flags KeySymFlags) KeySym {
	s := C.CString(name)
	sym := C.xkb_keysym_from_name(s, C.enum_xkb_keysym_flags(flags))
	C.free(unsafe.Pointer(s))
	return KeySym(sym)
}

func (c Context) Destroy() {
	C.xkb_context_unref(c.p)
}

func (c Context) KeyMap() Keymap {
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

func (s State) Syms(keyCode KeyCode) []KeySym {
	var syms *C.xkb_keysym_t
	n := int(C.xkb_state_key_get_syms(s.p, C.uint32_t(keyCode), &syms))
	if n == 0 || syms == nil {
		return nil
	}
	slice := (*[1 << 30]C.xkb_keysym_t)(unsafe.Pointer(syms))[:n:n]

	res := make([]KeySym, n)
	for i := 0; i < n; i++ {
		res[i] = KeySym(slice[i])
	}

	return res
}
