package uevent

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type dir []os.DirEntry

type meta struct {
	subDirs   []os.DirEntry
	uEvent    os.DirEntry
	links     []os.DirEntry
	files     []os.DirEntry
	devPath   string
	subsystem string
}

func (d dir) sort(devPath string) *meta {
	res := &meta{devPath: devPath}
	for i, v := range d {
		if v.Name() == "uevent" {
			res.uEvent = d[i]
			continue
		}
		if v.Name() == "subsystem" {
			link, _ := os.Readlink(filePath(devPath, v.Name()))
			res.subsystem = filepath.Base(link)
			res.links = append(res.links, d[i])
			continue
		}
		if v.IsDir() {
			res.subDirs = append(res.subDirs, d[i])
			continue
		}
		if v.Type()&os.ModeSymlink == os.ModeSymlink {
			res.links = append(res.links, d[i])
			continue
		}
		res.files = append(res.files, d[i])
	}
	return res
}

func getMeta(devPath string) (*meta, error) {
	files, e := os.ReadDir(sysDir(devPath))
	if e != nil {
		return nil, e
	}
	return dir(files).sort(devPath), nil
}

func createEvent(data *meta, devPath string) (*Event, error) {
	ev := newUEvent()
	ev.Action = ActionAdd
	ev.Devpath = devPath
	ev.Subsystem = data.subsystem
	ev.Header = "add@" + devPath
	ev.Values["ACTION"] = ActionAdd
	ev.Values["DEVPATH"] = devPath
	ev.Values["SUBSYSTEM"] = data.subsystem
	f, err := os.Open(filePath(devPath, data.uEvent.Name()))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	n, err := ev.decode(r, false, '\n')
	if err != io.EOF {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("empty uevent")
	}
	return ev, nil
}

func GetEvent(devPath string) (*Event, error) {
	files, e := os.ReadDir(sysDir(devPath))
	if e != nil {
		return nil, e
	}
	data := dir(files).sort(devPath)
	if data.uEvent == nil {
		return nil, os.ErrNotExist
	}
	return createEvent(data, devPath)
}

func FindDevices(filter Filter) []*Event {
	res := make([]*Event, 0, 10)
	f := func(ev *Event) bool {
		if filter == nil || filter(ev) {
			res = append(res, ev)
		}
		return true
	}
	devices("/devices", f)
	return res
}

func devices(devPath string, filter Filter) {
	meta, err := getMeta(devPath)
	if err != nil {
		return
	}
	if meta.uEvent != nil {
		ev, err := createEvent(meta, devPath)
		if err == nil {
			filter(ev)
		}
	}
	for _, dir := range meta.subDirs {
		nextPath := devPath + "/" + dir.Name()
		devices(nextPath, filter)
	}
}
