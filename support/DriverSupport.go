package support

import (
	"github.com/ninjasphere/go-ninja/api"
)

//
// The DriverSupport object is intended to be used as an anonymous member of
// Driver objects. It contains that part of the Driver state that is common
// to most, if not all driver implementations and allows the driver implementor to
// reuse method implementations that typically do not need
// to vary across implementations.
//
// Example Usage
//
//	import "github.com/ninjasphere.com/support"
//
//	type acmeDriver {
// 		support.DriverSupport
// 		// driver specific state ...
// 	}
//
// 	func newAcmeDriver() (*acmeDriver, error) {
// 		driver := &acmeDriver{}
//
// 		err := driver.Init(info)
// 		if err != nil {
// 			return err
// 		}
//
// 		err = driver.Export(driver)
// 		if err != nil {
// 			return err
// 		}
//
//		// after this time, the framework will start calling methods on the driver.
//
//		return driver
// 	}
//
// For another example of how to use this type, refer to the FakeDriver module of the github.com/ninjasphere/go-ninja/fakedriver package.
//
type DriverSupport struct {
	ModuleSupport
}

// Export the driver to the local message bus. After this call completes the driver's Start method
// , and other supported methods, will be called.
//
// The methods parameter is a reference to the full interface of the driver which implements any
// optional driver methods (such as Start and Stop) not implemented by the DriverSupport interface.
// If you do not specify this interface during the export, those methods will not be exposed
// to the RPC subsystem and so will not be called.
//
// This method should not be called until Init has been successfully called. Otherwise, it will
// return a non-nil error.
func (d *DriverSupport) Export(methods ninja.Driver) error {
	err := failIfNotInitialized(&d.ModuleSupport)
	if err == nil {
		return d.Conn.ExportDriver(methods)
	} else {
		return err
	}
}
