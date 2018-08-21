package wlroots

// #include <wlr/types/wlr_box.h>
import "C"

type Box struct {
	b C.struct_wlr_box
}

func (b *Box) Set(x, y, width, height int) {
	b.b.x = C.int(x)
	b.b.y = C.int(y)
	b.b.width = C.int(width)
	b.b.height = C.int(height)
}

func (b *Box) Get() (x, y, width, height int) {
	return int(b.b.x), int(b.b.y), int(b.b.width), int(b.b.height)
}
