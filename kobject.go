package uevent

import (
	"io/ioutil"
	"regexp"
	"strings"
	//	"path/filepath"
	"os"
	//	"fmt"
)

var reg *regexp.Regexp

type KObject struct {
	attr   map[string]string
	uevent map[string]string
}

func (o *KObject) GetAttr(name string) string {
	res, _ := o.attr[name]
	return res
}

func (o *KObject) GetEvent(name string) string {
	res, _ := o.uevent[name]
	return res
}

func (o *KObject) Attrs() map[string]string {
	return o.attr
}

func (o *KObject) Uevent() map[string]string {
	return o.uevent
}

func (o *KObject) Path() string {
	res, _ := o.uevent["DEVPATH"]
	return res
}

func (o *KObject) Action() string {
	res, _ := o.uevent["ACTION"]
	return res
}

func newKObject() *KObject {
	return &KObject{make(map[string]string, 20), make(map[string]string, 20)}
}

func parseKObject(buff []byte, length int, path string) (*KObject, error) {
	obj := newKObject()
	if path != "" {
		obj.uevent["ACTION"] = "add"
		obj.uevent["DEVPATH"] = path
	}
	action_remove := false
	for _, v := range reg.FindAllSubmatch(buff[:length], -1) {
		key, val := string(v[1]), string(v[2])
		if key == "ACTION" && val == "remove" {
			action_remove = true
		}
		obj.uevent[key] = val
	}
	if val, ok := obj.uevent["DEVPATH"]; ok && !action_remove {
		files, e := ioutil.ReadDir("/sys/" + val)
		if e == nil {
			for _, f := range files {
				if f.Name() != "uevent" && f.Name() != "subsystem" && f.Name() != "descriptors" && f.Mode().IsRegular() && (f.Mode().Perm()&044 > 0) {
					tmp, e := ioutil.ReadFile("/sys/" + val + "/" + f.Name())
					if e == nil {
						obj.attr[f.Name()] = strings.TrimRight(string(tmp), "\n")
					}
				}
				if f.Name() == "subsystem" {
					if obj.uevent["SUBSYSTEM"] == "" {
						name, _ := os.Readlink("/sys/" + val + "/" + f.Name())
						//						fmt.Println(name)
						divs := strings.Split(name, "/")
						obj.uevent["SUBSYSTEM"] = divs[len(divs)-1]
					}
				}
			}
		}
	}
	return obj, nil
}

func traverse(dir string, acc []*KObject) []*KObject {
	files, e := ioutil.ReadDir("/sys/" + dir)
	if e != nil {
		return acc
	}
	for _, file := range files {
		if file.Mode().IsDir() {
			acc = traverse(dir+file.Name()+"/", acc)
		} else {
			if file.Name() == "uevent" {
				data, e := ioutil.ReadFile("/sys/" + dir + file.Name())
				if e == nil && len(data) != 0 {
					obj, e := parseKObject(data, len(data), strings.TrimRight(dir, "/"))
					if e == nil {
						acc = append(acc, obj)
					}
				}
			}
		}
	}
	return acc
}

func Coldplug() []*KObject {
	res := make([]*KObject, 0, 512)
	res = traverse("/devices/", res)
	return res
}

func init() {
	reg = regexp.MustCompile("([a-zA-Z0-9-\"#¤%&/()=?\\<>;:{}!_. ]+)=([a-zA-Z0-9-\"#¤%&/()=?\\<>;:{}!_. ]+)")
}
