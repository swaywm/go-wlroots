package wlroots

// #include <stdlib.h>
// #include <wlr/types/wlr_xcursor_manager.h>
import "C"
import "unsafe"

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
