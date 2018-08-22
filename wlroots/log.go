package wlroots

// #define _GNU_SOURCE
//
// #include <stdarg.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include <wlr/util/log.h>
//
// void _wlr_log_cb(enum wlr_log_importance importance, char *msg);
//
// static inline void _wlr_log_inner_cb(enum wlr_log_importance importance, const char *fmt, va_list args) {
//		char *msg = NULL;
//		vasprintf(&msg, fmt, args);
//
//		_wlr_log_cb(importance, msg);
//		free(msg);
// }
//
// static inline void _wlr_log_set_cb(enum wlr_log_importance verbosity, bool is_set) {
//		wlr_log_init(verbosity, is_set ? &_wlr_log_inner_cb : NULL);
// }
import "C"

type (
	LogImportance uint32
	LogFunc       func(importance LogImportance, msg string)
)

const (
	LogImportanceSilent LogImportance = C.WLR_SILENT
	LogImportanceError  LogImportance = C.WLR_ERROR
	LogImportanceInfo   LogImportance = C.WLR_INFO
	LogImportanceDebug  LogImportance = C.WLR_DEBUG
)

var (
	onLog LogFunc
)

//export _wlr_log_cb
func _wlr_log_cb(importance LogImportance, msg *C.char) {
	if onLog != nil {
		onLog(importance, C.GoString(msg))
	}
}

func OnLog(verbosity LogImportance, cb LogFunc) {
	C._wlr_log_set_cb(C.enum_wlr_log_importance(verbosity), cb != nil)
	onLog = cb
}
