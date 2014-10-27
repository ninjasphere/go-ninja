package ninja

import ()

//
// Driver implementors should provide an implementation of this interface.
//
// The methods of this interface are used by the RPC framework to learn about
// the driver and also to configure the driver with a callback that the
// driver can use to emit events that comply with the driver service
// specification (http://schema.ninjablocks.com/service/driver)
//
// Implementors can inherit default implementations of these methods
// by including an anonymous struct member of type support.DriverSupport.
//
type Driver interface {
	Module
}
