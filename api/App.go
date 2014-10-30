package ninja

import ()

//
// App implementors should provide an implementation of this interface.
//
// The methods of this interface are used by the RPC framework to learn about
// the app and also to configure the app with a callback that the
// app can use to emit events that comply with the app service
// specification (http://schema.ninjablocks.com/service/app)
//
// Implementors can inherit default implementations of these methods
// by including an anonymous struct member of type support.AppSupport.
//
type App interface {
	Module
}
