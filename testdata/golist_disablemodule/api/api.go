package api

/*
#include "foo.h"
*/
import "C"
import (
	"fmt"
)

func PrintInt(i, j int) {
	res := C.add(C.int(i), C.int(j))
	fmt.Println(res)
}
