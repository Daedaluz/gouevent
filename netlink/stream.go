package netlink

import (
	"io"
	"os"
	"sync/atomic"
	"syscall"
)

type Stream struct {
	fd atomic.Value
}

func (r *Stream) Read(buff []byte) (n int, err error) {
	fd := r.fd.Load()
	n, err = syscall.Read(fd.(int), buff)
	if fd.(int) == -1 {
		return 0, io.EOF
	}
	return
}

func (r *Stream) Close() error {
	fd := r.fd.Swap(-1)
	return syscall.Close(fd.(int))
}

func Open() (*Stream, error) {
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_KOBJECT_UEVENT)
	if err != nil {
		return nil, err
	}
	syscall.CloseOnExec(fd)
	addr := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Pid:    uint32(os.Getpid()),
		Groups: 1,
	}
	err = syscall.Bind(fd, addr)
	if err != nil {
		syscall.Close(fd)
		return nil, err
	}
	s := &Stream{}
	s.fd.Store(fd)
	return s, nil
}
