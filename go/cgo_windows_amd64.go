//go:build windows && amd64

package aclic

/*
#cgo LDFLAGS: ${SRCDIR}/native/windows-x86_64/aclic.lib -lbcrypt -lntdll -luserenv -lws2_32
*/
import "C"
