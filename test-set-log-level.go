// +build ignore
package main

import (
       "github.com/ninjasphere/go-ninja/support"
)

type fixture struct {
     support.DriverSupport
}

func (self *fixture) logVarious(level string) {
     self.SetLogLevel(level)
     self.Log.Debugf("logging at debug while level is %s", level)
     self.Log.Errorf("logging at error while level is %s", level)
     self.Log.Infof("logging at info while level is %s", level)
     self.Log.Warningf("logging at warning while level is %s", level)
}


func main() {
	obj := &fixture{}

	err := obj.Init(nil);
	_ = err

	obj.logVarious("DEBUG")
	obj.logVarious("INFO")
	obj.logVarious("WARNING")
	obj.logVarious("ERROR")
}
