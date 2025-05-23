// The UnixCredentials system call is currently only implemented on Linux
// http://golang.org/src/pkg/syscall/sockcmsg_linux.go
// https://golang.org/s/go1.4-syscall
// http://code.google.com/p/go/source/browse/unix/sockcmsg_linux.go?repo=sys

// Local implementation of the UnixCredentials system call for DragonFly BSD

package dbus

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

// http://golang.org/src/pkg/syscall/ztypes_linux_amd64.go
// http://golang.org/src/pkg/syscall/ztypes_dragonfly_amd64.go
type Ucred struct {
	Pid int32
	Uid uint32
	Gid uint32
}

// http://golang.org/src/pkg/syscall/types_linux.go
// http://golang.org/src/pkg/syscall/types_dragonfly.go

// http://golang.org/src/pkg/syscall/sockcmsg_unix.go
func cmsgAlignOf(salen int) int {
	// From http://golang.org/src/pkg/syscall/sockcmsg_unix.go
	//salign := sizeofPtr
	// NOTE: It seems like 64-bit Darwin and DragonFly BSD kernels
	// still require 32-bit aligned access to network subsystem.
	//if darwin64Bit || dragonfly64Bit {
	//	salign = 4
	//}
	salign := 4
	return (salen + salign - 1) & ^(salign - 1)
}

// http://golang.org/src/pkg/syscall/sockcmsg_unix.go
func cmsgData(h *syscall.Cmsghdr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(h)) + uintptr(cmsgAlignOf(syscall.SizeofCmsghdr)))
}

// http://golang.org/src/pkg/syscall/sockcmsg_linux.go
// UnixCredentials encodes credentials into a socket control message
// for sending to another process. This can be used for
// authentication.
func UnixCredentials(ucred *Ucred) []byte {
	// We just need a size >= than the size of the actual structure, so use a giant size to be safe
	// https://github.com/DragonFlyBSD/DragonFlyBSD/blob/master/sys/sys/ucred.h
	const SizeofUcred = 128
	b := make([]byte, syscall.CmsgSpace(SizeofUcred))
	h := (*syscall.Cmsghdr)(unsafe.Pointer(&b[0]))
	h.Level = syscall.SOL_SOCKET
	h.Type = syscall.SCM_CREDS
	h.SetLen(syscall.CmsgLen(SizeofUcred))
	*((*Ucred)(cmsgData(h))) = *ucred
	return b
}

// http://golang.org/src/pkg/syscall/sockcmsg_linux.go
// ParseUnixCredentials decodes a socket control message that contains
// credentials in a Ucred structure. To receive such a message, the
// SO_PASSCRED option must be enabled on the socket.
func ParseUnixCredentials(m *syscall.SocketControlMessage) (*Ucred, error) {
	if m.Header.Level != syscall.SOL_SOCKET {
		return nil, syscall.EINVAL
	}
	if m.Header.Type != syscall.SCM_CREDS {
		return nil, syscall.EINVAL
	}
	ucred := *(*Ucred)(unsafe.Pointer(&m.Data[0]))
	return &ucred, nil
}

func (t *unixTransport) SendNullByte() error {
	ucred := &Ucred{Pid: int32(os.Getpid()), Uid: uint32(os.Getuid()), Gid: uint32(os.Getgid())}
	b := UnixCredentials(ucred)
	_, oobn, err := t.UnixConn.WriteMsgUnix([]byte{0}, b, nil)
	if err != nil {
		return err
	}
	if oobn != len(b) {
		return io.ErrShortWrite
	}
	return nil
}
