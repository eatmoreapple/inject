package inject

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Injector interface {
	Register(is ...*Instance) error
	AutoWired(vs ...interface{}) error
}

func New() Injector {
	return &Instance{base: true}
}

type Instance struct {
	// Name of the instance
	// If empty, the instance will be registered as the pkg and type name
	Name string
	// Value of the instance
	Value      interface{}
	container  map[string]*Instance
	base       bool
	initialize bool
}

func (i *Instance) init() error {
	if i.initialize {
		return nil
	}
	i.initialize = true
	i.container = make(map[string]*Instance)
	if i.base {
		return nil
	}
	if i.Value == nil {
		return errors.New("inject: instance value is nil")
	}
	rv := reflect.ValueOf(i.Value)
	ri := reflect.Indirect(rv)
	rt := ri.Type()
	if i.Name == "" {
		i.Name = strings.Replace(rt.PkgPath(), "/", ".", -1) + "." + rt.Name()
	}
	i.container[i.Name] = i
	return nil
}

func (i *Instance) Register(is ...*Instance) error {
	if err := i.init(); err != nil {
		return err
	}
	for _, s := range is {
		if err := s.init(); err != nil {
			return err
		}
		for k, v := range s.container {
			i.container[k] = v
		}
	}
	return nil
}

func (i *Instance) AutoWired(vs ...interface{}) error {
	for _, v := range vs {
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			return errors.New("inject: auto wired value must be a pointer")
		}
		rv = reflect.Indirect(rv)
		t := reflect.TypeOf(v).Elem()
		if rv.Kind() != reflect.Struct {
			return errors.New("inject: auto wired value must be a struct")
		}
		for j := 0; j < rv.NumField(); j++ {
			field := rv.Field(j)
			tf := t.Field(j)
			// is unexported
			// PkgPath is the package path that qualifies a lower case (unexported)
			// field name. It is empty for upper case (exported) field names.
			// See https://golang.org/ref/spec#Uniqueness_of_identifiers
			if tf.PkgPath != "" {
				continue
			}
			inject := tf.Tag.Get("inject")
			if inject == "-" {
				continue
			}
			if field.CanSet() {
				if field.Kind() == reflect.Interface {
					// auto wired interface
					if inject == "" {
						var ok bool
						for _, v := range i.container {
							rv := reflect.ValueOf(v.Value)
							// check if the type of the instance is assignable to the interface
							if rv.Type().Implements(field.Type()) {
								field.Set(rv)
								ok = true
								break
							}
						}
						if !ok {
							return fmt.Errorf("inject: can not find a instance for %s", field.Type().String())
						}
					} else {
						// get the instance by name
						v, ok := i.container[inject]
						// check if the instance is registered
						if !ok {
							return errors.New("inject: auto wired value not found")
						}
						rv := reflect.ValueOf(v.Value)
						// check if the type of the instance is assignable to the interface
						if !rv.Type().Implements(field.Type()) {
							field.Set(rv)
							continue
						}
						return fmt.Errorf("inject: type of %s is not assignable to %s", rv.Type().String(), field.Type().String())
					}
				} else {
					// auto wired value
					v, ok := i.container[inject]
					// check if the instance is registered
					if !ok {
						return errors.New("inject: auto wired value not found")
					}
					rv := reflect.ValueOf(v.Value)
					if rv.Type().AssignableTo(field.Type()) {
						field.Set(rv)
						continue
					}
					return fmt.Errorf("inject: type of %s is not assignable to %s", rv.Type().String(), field.Type().String())
				}
			}
		}

	}
	return nil
}

var defaultInstance = New()

func RegisterInstance(is ...*Instance) error {
	return defaultInstance.Register(is...)
}

func AutoWired(vs ...interface{}) error {
	return defaultInstance.AutoWired(vs...)
}

func Register(vs ...interface{}) error {
	for _, v := range vs {
		instance := &Instance{Value: v}
		if err := RegisterInstance(instance); err != nil {
			return err
		}
	}
	return nil
}

func AutoWiredAndRegister(vs ...interface{}) error {
	for _, v := range vs {
		if err := Register(v); err != nil {
			return err
		}
		if err := AutoWired(v); err != nil {
			return err
		}
	}
	return nil
}
