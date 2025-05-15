package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import "unsafe"

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <stdlib.h>
// #include <wlr/xcursor.h>
import "C"

type XCursor struct {
	p *C.struct_wlr_xcursor
}

type XCursorImage struct {
	p *C.struct_wlr_xcursor_image
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
