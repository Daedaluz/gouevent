package uevent

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

type Event struct {
	Header    string
	Action    string
	Devpath   string
	Subsystem string
	Seq       uint64
	Values    map[string]string
}

func (e *Event) String() string {
	x, _ := json.MarshalIndent(e, "", "    ")
	return string(x)
}

func newUEvent() *Event {
	return &Event{
		Values: make(map[string]string, 20),
	}
}

func (e *Event) step(b *bufio.Reader, delim byte) (string, error) {
	s, err := b.ReadString(delim)
	if err != nil {
		return "", err
	}
	return s[:len(s)-1], nil
}

func (e *Event) decode(b *bufio.Reader, header bool, delim byte) (n int, err error) {
	if header {
		e.Header, err = e.step(b, delim)
		if err != nil {
			return 0, err
		}
	}
parser:
	for {
		line, err := e.step(b, delim)
		if err != nil {
			return n, err
		}
		i := strings.Index(line, "=")
		if i < 0 {
			return n, errors.New("unknown event format")
		}
		k, v := line[:i], line[i+1:]
		e.Values[k] = v
		switch k {
		case "ACTION":
			e.Action = v
		case "DEVPATH":
			e.Devpath = v
		case "SUBSYSTEM":
			e.Subsystem = v
		case "SEQNUM":
			e.Seq, _ = strconv.ParseUint(v, 10, 64)
			break parser
		}
		n++
	}
	return n, nil
}

func (e *Event) ParentNth(n int) (*Event, error) {
	return GetEvent(path.Clean(e.Devpath + strings.Repeat("/..", n)))
}

func (e *Event) SubDevices(filter Filter) []*Event {
	devs := make([]*Event, 0, 5)
	f := func(e *Event) bool {
		if filter == nil || filter(e) {
			devs = append(devs, e)
		}
		return true
	}
	meta, err := getMeta(e.Devpath)
	if err != nil {
		return nil
	}
	for _, v := range meta.subDirs {
		devices(e.Devpath+"/"+v.Name(), f)
	}
	return devs
}

func (e *Event) OpenAttr(attr string, flag int) (*os.File, error) {
	return os.OpenFile(filePath(e.Devpath, attr), flag, 0)
}

func (e *Event) ReadAttrString(attr string) (string, error) {
	f, err := e.OpenAttr(attr, os.O_RDONLY)
	if err != nil {
		return "", err
	}
	defer f.Close()
	all, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(all), "\n"), nil
}

func (e *Event) WriteAttrString(attr, value string) error {
	f, err := os.OpenFile(filePath(e.Devpath, attr), os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(value)
	return err
}
