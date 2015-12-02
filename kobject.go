package uevent

import(
	"io/ioutil"
	"strings"
	"regexp"
)

var reg *regexp.Regexp

type KObject struct {
	attr map[string] string
	uevent map[string] string
}

func (o *KObject) GetAttr(name string) (string, bool) {
	res, ok := o.attr[name]
	return res, ok
}

func (o *KObject) GetEvent(name string) (string, bool) {
	res, ok := o.uevent[name]
	return res, ok
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

func parseKObject(buff []byte, length int) (*KObject, error) {
	obj := newKObject()
	action_remove := false
	for _, v := range reg.FindAllSubmatch(buff[:length], -1) {
		key, val := string(v[1]), string(v[2])
		if key == "ACTION" && val == "remove" {
			action_remove = true
		}
		obj.uevent[key] = val
		if string(v[1]) == "DEVPATH" && !action_remove {
			files, e := ioutil.ReadDir("/sys/" + val)
			if e == nil {
				for _, f := range(files) {
					if f.Name() != "uevent" && f.Name() != "descriptors" && f.Mode().IsRegular() && (f.Mode().Perm() & 044 > 0) {
						tmp, e := ioutil.ReadFile("/sys/" + val + "/" + f.Name())
						if e == nil {
							obj.attr[f.Name()] = strings.TrimRight(string(tmp), "\n")
						}
					}
				}
			}
		}
	}
	return obj, nil
}


func init() {
	reg = regexp.MustCompile("([a-zA-Z0-9-\"#¤%&/()=?\\<>;:{}!_. ]+)=([a-zA-Z0-9-\"#¤%&/()=?\\<>;:{}!_. ]+)")
}
