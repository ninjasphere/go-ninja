package support

import (
	"github.com/ninjasphere/go-ninja/api"
)

//
// The AppSupport object is intended to be used as an anonymous member of
// App objects. It contains that part of the App state that is common
// to most, if not all app implementations and allows the app implementor to
// reuse method implementations that typically do not need
// to vary across implementations.
//
// Example Usage
//
//	import "github.com/ninjasphere.com/support"
//
//	type acmeApp {
// 		support.AppSupport
// 		// app specific state ...
// 	}
//
// 	func newAcmeApp() (*acmeApp, error) {
// 		app := &acmeApp{}
//
// 		err := app.Init(info)
// 		if err != nil {
// 			return err
// 		}
//
// 		err = app.Export(app)
// 		if err != nil {
// 			return err
// 		}
//
//		// after this time, the framework will start calling methods on the app.
//
//		return app
// 	}
//
type AppSupport struct {
	ModuleSupport
}

// Export the app to the local message bus. After this call completes the app's Start method
// , and other supported methods, will be called.
//
// The methods parameter is a reference to the full interface of the app which implements any
// optional app methods (such as Start and Stop) not implemented by the AppSupport interface.
// If you do not specify this interface during the export, those methods will not be exposed
// to the RPC subsystem and so will not be called.
//
// This method should not be called until Init has been successfully called. Otherwise, it will
// return a non-nil error.
func (self *AppSupport) Export(methods ninja.App) error {
	err := failIfNotInitialized(&self.ModuleSupport)
	if err == nil {
		return self.Conn.ExportApp(methods)
	} else {
		return err
	}
}
