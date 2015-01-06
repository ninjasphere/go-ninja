// +build ignore
package main

import (
	"github.com/ninjasphere/go-ninja/support"
)

type fixture struct {
	support.DriverSupport
}

func (f *fixture) logVarious(level string) {
	f.SetLogLevel(level)
	f.Log.Debugf("logging at debug while level is %s", level)
	f.Log.Errorf("logging at error while level is %s", level)
	f.Log.Infof("logging at info while level is %s", level)
	f.Log.Warningf("logging at warning while level is %s", level)
}

func main() {
	obj := &fixture{}

	err := obj.Init(nil)
	_ = err

	obj.logVarious("DEBUG")
	obj.logVarious("INFO")
	obj.logVarious("WARNING")
	obj.logVarious("ERROR")
}
