package wlroots

// #include <stdlib.h>
// #include <wlr/types/wlr_xcursor_manager.h>
import "C"
import "unsafe"

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

func (m XCursorManager) SetImage(cursor Cursor, name string) {
	s := C.CString(name)
	C.wlr_xcursor_manager_set_cursor_image(m.p, s, cursor.p)
	C.free(unsafe.Pointer(s))
}
