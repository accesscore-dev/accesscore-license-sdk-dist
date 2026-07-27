//go:build darwin && arm64

package aclic

// The core is linked statically, so a Go binary carries it rather than needing a shared
// library installed beside it. One file per platform, because the archive lives in a
// per-platform directory and cgo directives cannot be computed.

/*
#cgo LDFLAGS: ${SRCDIR}/native/darwin-aarch64/libaclic.a
*/
import "C"
