package uevent

import (
	"fmt"
)

const (
	sysfsDir = "/sys"
)

func sysDir(devPath string) string {
	return fmt.Sprintf("%s%s", sysfsDir, devPath)
}

func filePath(devPath, fileName string) string {
	return sysDir(fmt.Sprintf("%s/%s", devPath, fileName))
}

const (
	ActionAdd    = "add"
	ActionBind   = "bind"
	ActionRemove = "remove"
	ActionUnbind = "unbind"
)
