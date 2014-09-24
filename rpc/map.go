// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2014 Ninja Blocks Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"fmt"
	log2 "log"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

var (
	// Precompute the reflect.Type of error and *rpc.Message
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfRequest = reflect.TypeOf((*Message)(nil))
)

// ----------------------------------------------------------------------------
// service
// ----------------------------------------------------------------------------

type service struct {
	name     string                    // name of service
	rcvr     reflect.Value             // receiver of methods for the service
	rcvrType reflect.Type              // type of the receiver
	methods  map[string]*serviceMethod // registered methods
}

type serviceMethod struct {
	method    reflect.Method // receiver method
	argsType  reflect.Type   // type of the request argument
	replyType reflect.Type   // type of the response argument
}

// ----------------------------------------------------------------------------
// serviceMap
// ----------------------------------------------------------------------------

// serviceMap is a registry for services.
type serviceMap struct {
	mutex    sync.Mutex
	services map[string]*service
}

// register adds a new service using reflection to extract its methods.
func (m *serviceMap) register(rcvr interface{}, name string) (methods []string, err error) {

	// Setup service.
	s := &service{
		name:     name,
		rcvr:     reflect.ValueOf(rcvr),
		rcvrType: reflect.TypeOf(rcvr),
		methods:  make(map[string]*serviceMethod),
	}
	if name == "" {
		s.name = reflect.Indirect(s.rcvr).Type().Name()
		if !isExported(s.name) {
			return nil, fmt.Errorf("rpc: type %q is not exported", s.name)
		}
	}
	if s.name == "" {
		return nil, fmt.Errorf("rpc: no service name for type %q", s.rcvrType.String())
	}
	// Setup methods.
	for i := 0; i < s.rcvrType.NumMethod(); i++ {

		method := s.rcvrType.Method(i)
		mtype := method.Type

		// Method must have one or two arguments,
		if mtype.NumIn() == 1 || mtype.NumIn() > 3 {
			continue
		}

		// and the first argument must be *rpc.Message
		reqType := mtype.In(1)
		if reqType != typeOfRequest {
			continue
		}

		// Method must be exported.
		if method.PkgPath != "" {
			log.Fatalf("RPC Method '%s' must be exported", method.Name)
			continue
		}

		// If there's a second argument, it must be a pointer and must be exported.
		var args reflect.Type
		if mtype.NumIn() == 3 {
			args = mtype.In(2)
			if args.Kind() != reflect.Ptr || !isExportedOrBuiltin(args) {
				log2.Fatalf("RPC Method '%s' second argument '%s' must be a pointer and exported", method.Name, args.Name())
				continue
			}
		}

		// Method needs one or two outs
		if mtype.NumOut() != 1 && mtype.NumOut() != 2 {
			log2.Fatalf("RPC Method '%s' must have one or two outs", method.Name)
			continue
		}

		// If there are two outs, the first must be exported
		var reply reflect.Type
		if mtype.NumOut() == 2 {
			// Third argument must be a pointer and must be exported.
			reply = mtype.Out(0)
			if reply.Kind() != reflect.Ptr || !isExportedOrBuiltin(reply) {
				log2.Fatalf("RPC Method '%s' return type '%s' must be a pointer and exported", method.Name, reply.Name())
				continue
			}
		}

		// The last out needs to be an error
		if lastReturn := mtype.Out(mtype.NumOut() - 1); lastReturn != typeOfError {
			log2.Fatalf("RPC Method '%s' must return only an error", method.Name)
			continue
		}

		s.methods[method.Name] = &serviceMethod{
			method: method,
		}
		if reply != nil {
			s.methods[method.Name].replyType = reply.Elem()
		}
		if args != nil {
			s.methods[method.Name].argsType = args.Elem()
		}

	}
	/*if len(s.methods) == 0 {
		return nil, fmt.Errorf("rpc: %q has no exported methods of suitable type", s.name)
	}*/
	// Add to the map.
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.services == nil {
		m.services = make(map[string]*service)
	} else if _, ok := m.services[s.name]; ok {
		return nil, fmt.Errorf("rpc: service already defined: %q", s.name)
	}
	m.services[s.name] = s

	exportedMethods := make([]string, len(s.methods))
	i := 0
	for name := range s.methods {
		exportedMethods[i] = name
		i++
	}
	return exportedMethods, nil
}

// get returns a registered service given a method name.
func (m *serviceMap) get(topic string, method string) (*service, *serviceMethod, error) {
	m.mutex.Lock()
	service := m.services[topic]
	m.mutex.Unlock()
	if service == nil {
		err := fmt.Errorf("rpc: can't find service %q", topic)
		return nil, nil, err
	}
	serviceMethod := service.methods[method]
	if serviceMethod == nil {
		err := fmt.Errorf("rpc: can't find method %q", method)
		return nil, nil, err
	}
	return service, serviceMethod, nil
}

// isExported returns true of a string is an exported (upper case) name.
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// isExportedOrBuiltin returns true if a type is exported or a builtin.
func isExportedOrBuiltin(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}
