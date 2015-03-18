package support

import (
	"fmt"
	"reflect"
)

// Types that need to do something post injecction should implement a PostConstruct method as per this
// interface and will then my notified post constructionn via a call to that method.
type PostConstruct interface {
	PostConstruct() error
}

// Given a list of injectable things, call the PostConstruct method of each if there is one. Fail on the first
// one that fails.
func PostConstructAll(injectables ...interface{}) error {
	// Run PostConstruct on any objects that have it
	for _, node := range injectables {
		if pc, ok := node.(PostConstruct); ok {
			if err := pc.PostConstruct(); err != nil {
				return fmt.Errorf("Failed PostConstruct on object %s: %s", reflect.TypeOf(pc).String(), err)
			}
		}
	}
	return nil
}
