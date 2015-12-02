package uevent

import (
	"syscall"
	"os"
)

type UeventSocket struct {
	file *os.File
	refs map[string]*KObject
	tmpbuff []byte
}

func (s *UeventSocket) Next() (*KObject, error) {
	n, e := s.Read(s.tmpbuff)
	if e != nil {
		return nil, e
	}
	obj, e := parseKObject(s.tmpbuff, n)
	if e != nil {
		return nil, e
	}

	return obj, nil
}

func (s *UeventSocket) Read(buff []byte) (n int, e error) {
	n, e = s.file.Read(buff)
	return
}

func NewSocket(groups uint32) (file *UeventSocket, e error) {
	file = nil
	fd, e := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM, syscall.NETLINK_KOBJECT_UEVENT)
	if e != nil {
		return
	}
	sockaddr := &syscall.SockaddrNetlink {Family: syscall.AF_NETLINK, Pid: 0, Groups: groups}
	e = syscall.Bind(fd, sockaddr)
	if e != nil {
		return
	}
	tmp := os.NewFile(uintptr(fd), "NetLink - KOBJECT")
	file = &UeventSocket{
		file: tmp,
		refs: make(map[string]*KObject, 50),
		tmpbuff: make([]byte, syscall.Getpagesize()),
	}
	return
}

