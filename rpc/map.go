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

	"github.com/ninjasphere/redigo/redis"
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
	method       reflect.Method // receiver method
	argsType     reflect.Type   // type of the request argument
	replyType    reflect.Type   // type of the response argument
	hasRedisConn bool
}

// ----------------------------------------------------------------------------
// serviceMap
// ----------------------------------------------------------------------------

// serviceMap is a registry for services.
type serviceMap struct {
	mutex    sync.Mutex
	services map[string]*service
}

type rpcService interface {
	GetRPCMethods() []string
}

// register adds a new service using reflection to extract its methods.
func (m *serviceMap) register(rcvr interface{}, name string, exportableMethods []string) (methods []string, err error) {

	/*var providedMethods *[]string
	switch rcvr := rcvr.(type) {
	case rpcService:
		providedMethods = rcvr.GetRPCMethods()
	}*/

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
			log.Fatalf("rpc: type %q is not exported", s.name)
		}
	}
	if s.name == "" {
		log.Fatalf("rpc: no service name for type %q", s.rcvrType.String())
	}
	// Setup methods.
	for i := 0; i < s.rcvrType.NumMethod(); i++ {

		method := s.rcvrType.Method(i)
		mtype := method.Type

		//log.Infof("Method: %s", method.Name)

		if !isExported(method.Name) {
			continue
		}

		// Method must be in the list of exportable methods
		if !isValueInList(lowerFirst(method.Name), exportableMethods) {
			//log.Infof("Not exportable: %s", method.Name)

			continue
		}

		var hasRedisConn = false
		if (mtype.NumIn() == 2 || mtype.NumIn() == 3) && mtype.In(mtype.NumIn()-1).Implements(reflect.TypeOf((*redis.Conn)(nil)).Elem()) {
			hasRedisConn = true
		}

		// Method must have no or one arguments (plus optional redis connection)
		if mtype.NumIn() > 2 && !hasRedisConn {
			//log.Infof("Wrong number: %s", method.Name)
			continue
		}

		// Method must be exported.
		if method.PkgPath != "" {
			log.Fatalf("RPC Method '%s' must be exported", method.Name)
			continue
		}

		// The one argument (args) must be a pointer and must be exported, if its there
		var args reflect.Type
		if mtype.NumIn() > 1 {
			args = mtype.In(1)
			if !isExportedOrBuiltin(args) {
				log2.Fatalf("RPC Method %s.%s arguments must be exported", name, method.Name)
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
			method:       method,
			hasRedisConn: hasRedisConn,
		}
		if reply != nil {
			s.methods[method.Name].replyType = reply.Elem()
		}
		if args != nil {
			s.methods[method.Name].argsType = args
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

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
