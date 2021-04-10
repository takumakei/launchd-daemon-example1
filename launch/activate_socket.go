package launch

/*
#include <stdlib.h>
#include <launch.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

func activateSocket(name string) ([]int, error) {
	c_name := C.CString(name)
	var c_fds *C.int
	var c_cnt C.size_t

	err := C.launch_activate_socket(c_name, &c_fds, &c_cnt)
	if err != 0 {
		return nil, errors.New(C.GoString(C.strerror(err)))
	}

	ptr := unsafe.Pointer(c_fds)
	defer C.free(ptr)
	fds := (*[1 << 30]C.int)(ptr)
	cnt := int(c_cnt)

	result := make([]int, cnt)
	for i := 0; i < cnt; i++ {
		result[i] = int(fds[i])
	}
	return result, nil
}
