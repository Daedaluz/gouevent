package uevent

import (
	"os"
	"path/filepath"
	"syscall"
	//	"fmt"
)

const (
	None   = 0
	Kernel = 1
	Udev   = 2
)

type UeventSocket struct {
	file    *os.File
	refs    map[string]*KObject
	tmpbuff []byte
	prenum  []chan *KObject
}

func (s *UeventSocket) Next() (*KObject, error) {
	n, e := s.Read(s.tmpbuff)
	if e != nil {
		return nil, e
	}
	obj, e := parseKObject(s.tmpbuff, n, "")
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
	sockaddr := &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK, Pid: 0, Groups: groups}
	e = syscall.Bind(fd, sockaddr)
	if e != nil {
		return
	}
	tmp := os.NewFile(uintptr(fd), "NetLink - KOBJECT")
	file = &UeventSocket{
		file:    tmp,
		refs:    make(map[string]*KObject, 50),
		tmpbuff: make([]byte, syscall.Getpagesize()),
		prenum:  make([]chan *KObject, 20),
	}
	return
}

func (s *UeventSocket) Coldplug() {
	objects := Coldplug()
	for _, obj := range objects {
		s.refs[obj.Path()] = obj
	}
}

func (s *UeventSocket) GetKObject(path string, rel string) *KObject {
	name := filepath.Clean(path + "/" + rel)
	//	fmt.Println("PATH:", name)
	return s.refs[name]
}

func (s *UeventSocket) Prenum() chan *KObject {
	c := make(chan *KObject)
	go func() {
		for _, obj := range s.refs {
			c <- obj
		}
		for obj, ok := s.Next(); ok == nil; obj, ok = s.Next() {
			switch obj.Action() {
			case "add":
				s.refs[obj.Path()] = obj
				c <- obj
			case "remove":
				prev := s.refs[obj.Path()]
				if prev == nil {
					continue
				}
				obj.attr = prev.attr
				delete(s.refs, obj.Path())
				c <- obj
			default:
				c <- obj
			}
		}
	}()
	return c
}
