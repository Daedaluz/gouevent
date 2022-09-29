package main

import (
	"github.com/daedaluz/gouevent"
	"log"
)

func main() {
	r, err := uevent.Open()
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Close()

	f := func(ev *uevent.Event) bool {
		if ev.Action == uevent.ActionAdd && ev.Subsystem == "tty" {
			if parent, err := ev.ParentNth(3); err == nil {
				if parent.Subsystem == "usb" {
					return true
				}
			}
		}
		return false
	}

	log.Println(uevent.FindDevices(f))

	var ev *uevent.Event
	for ev, err = r.Find(f); err == nil; ev, err = r.Find(f) {
		log.Println(ev)
	}
	log.Println(err)
}
