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

		if mtype.NumIn() == 1 {
			continue
		}

		if mtype.NumIn() >= 2 {
			reqType := mtype.In(1)
			if reqType != typeOfRequest {
				continue
			}
		}

		// Method must be exported.
		if method.PkgPath != "" {
			log.Fatalf("RPC Method '%s' must be exported", method.Name)
			continue
		}
		// Method needs four ins: receiver, *rpc.Message, *args, *reply.
		if mtype.NumIn() != 4 {
			log2.Fatalf("RPC Method '%s' must have three arguments (*rpc.Message, *args, *reply)", method.Name)
			continue
		}
		// First argument must be a pointer and must be rpc.Message.
		//reqType := mtype.In(1)
		//if reqType != typeOfRequest {
		//	continue
		//}
		// Second argument must be a pointer and must be exported.
		args := mtype.In(2)
		if args.Kind() != reflect.Ptr || !isExportedOrBuiltin(args) {
			log2.Fatalf("RPC Method '%s' second argument '%s' must be a pointer and exported", method.Name, args.Name())
			continue
		}
		// Third argument must be a pointer and must be exported.
		reply := mtype.In(3)
		if reply.Kind() != reflect.Ptr || !isExportedOrBuiltin(reply) {
			log2.Fatalf("RPC Method '%s' third argument '%s' must be a pointer and exported", method.Name, reply.Name())
			continue
		}
		// Method needs one out: error.
		if mtype.NumOut() != 1 {
			log2.Fatalf("RPC Method '%s' must return only an error", method.Name)
			continue
		}
		if returnType := mtype.Out(0); returnType != typeOfError {
			log2.Fatalf("RPC Method '%s' must return only an error", method.Name)
			continue
		}
		s.methods[method.Name] = &serviceMethod{
			method:    method,
			argsType:  args.Elem(),
			replyType: reply.Elem(),
		}
	}
	if len(s.methods) == 0 {
		return nil, fmt.Errorf("rpc: %q has no exported methods of suitable type", s.name)
	}
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
