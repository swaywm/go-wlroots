package wlroots

// #include <wlr/types/wlr_box.h>
// #include <wlr/types/wlr_matrix.h>
import "C"

type Matrix struct {
	m [9]C.float
}

func (m *Matrix) ProjectBox(box *Box, transform uint32, rotation float32, projection *Matrix) {
	C.wlr_matrix_project_box(&m.m[0], &box.b, C.enum_wl_output_transform(transform), C.float(rotation), &projection.m[0])
}
