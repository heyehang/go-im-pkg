package hack

import (
	"reflect"
	"unsafe"
)

// no copy to change slice to string
// use your own risk
func String(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}

// no copy to change string to slice
// use your own risk
func Slice(s string) (b []byte) {
	if s == "" {
		return nil
	}
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	b = *(*[]byte)(unsafe.Pointer(&bh))
	return
}
