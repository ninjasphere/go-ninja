package ninja

//
// Channel implementors should provide an implementation of this interface
// for each channel a device exports.
//
// FIXME: consider adding a ChannelSupport object
//
type Channel interface {
	GetProtocol() string
	SetEventHandler(func(event string, payload interface{}) error)
}
