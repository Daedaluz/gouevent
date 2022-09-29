package uevent

import (
	"bufio"
	"github.com/daedaluz/gouevent/netlink"
)

type Filter func(ev *Event) bool

type Reader struct {
	stream *netlink.Stream
	r      *bufio.Reader
}

func (r *Reader) Next() (*Event, error) {
	ev := newUEvent()
	if _, err := ev.decode(r.r, true, 0); err != nil {
		return nil, err
	}
	return ev, nil
}

func (r *Reader) Find(filter Filter) (*Event, error) {
	for {
		ev, err := r.Next()
		if err != nil {
			return nil, err
		}
		if filter(ev) {
			return ev, nil
		}
	}
}

func (r *Reader) Close() error {
	return r.stream.Close()
}

func Open() (*Reader, error) {
	stream, err := netlink.Open()
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(stream)
	return &Reader{
		stream: stream,
		r:      r,
	}, nil
}
