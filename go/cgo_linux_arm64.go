//go:build linux && arm64

package aclic

/*
#cgo LDFLAGS: ${SRCDIR}/native/linux-aarch64/libaclic.a -lm -ldl -lpthread
*/
import "C"
