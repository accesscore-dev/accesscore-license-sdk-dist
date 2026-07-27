//go:build linux && amd64

package aclic

/*
#cgo LDFLAGS: ${SRCDIR}/native/linux-x86_64/libaclic.a -lm -ldl -lpthread
*/
import "C"
