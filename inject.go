package inject

import (
	"fmt"
	"reflect"
)

type Injector interface {
	Repository(name string, v interface{}) error
	Autowired(name string, v interface{}) error
}

var instance = New()

func RepositoryWithInjector(obj Injector, name string, v interface{}) error {
	return obj.Repository(name, v)
}

func Repository(v interface{}, names ...string) error {
	var name string
	if len(names) > 0 {
		name = names[0]
	}
	return RepositoryWithInjector(instance, name, v)
}

func AutowiredWithInjector(obj Injector, name string, v interface{}) error {
	return obj.Autowired(name, v)
}

func Autowired(v interface{}, names ...string) error {
	var name string
	if len(names) > 0 {
		name = names[0]
	}
	return AutowiredWithInjector(instance, name, v)
}

func AutowiredStruct(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("%v is not a pointer", v)
	}
	rv = reflect.Indirect(rv)
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not a struct", v)
	}
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		rf := rt.Field(i)
		// is unexported
		// PkgPath is the package path that qualifies a lower case (unexported)
		// field name. It is empty for upper case (exported) field names.
		// See https://golang.org/ref/spec#Uniqueness_of_identifiers
		if rf.PkgPath != "" {
			continue
		}
		tag := rf.Tag.Get("inject")
		if tag == "-" {
			continue
		}
		field := rv.Field(i)
		if err := Autowired(field.Addr().Interface(), tag); err != nil {
			return err
		}
	}
	return nil
}

type injector struct {
	namedRepository   map[string]interface{}
	unnamedRepository []interface{}
}

func New() Injector {
	return &injector{}
}

func (i *injector) Repository(name string, v interface{}) error {
	if i.namedRepository == nil {
		i.namedRepository = make(map[string]interface{})
	}
	if i.unnamedRepository == nil {
		i.unnamedRepository = make([]interface{}, 0)
	}
	if name == "" {
		i.unnamedRepository = append(i.unnamedRepository, v)
	} else {
		if _, exists := i.namedRepository[name]; exists {
			return fmt.Errorf("%v is already registered", name)
		}
		i.namedRepository[name] = v
	}
	return nil
}

func (i *injector) Autowired(name string, v interface{}) error {
	if v == nil {
		return fmt.Errorf("inject: nil value")
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("inject: %v is not a pointer", rv)
	}
	ri := reflect.Indirect(rv)
	if name == "" {
		var found bool
		for _, r := range i.unnamedRepository {
			vpv := reflect.ValueOf(r)
			if vpv.CanConvert(ri.Type()) {
				if found {
					return fmt.Errorf("inject: %v is ambiguous", ri.Type())
				}
				found = true
				ri.Set(vpv.Convert(ri.Type()))
			}
		}
		if found {
			return nil
		}
	} else {
		if r, exists := i.namedRepository[name]; exists {
			vpv := reflect.ValueOf(r)
			if vpv.CanConvert(ri.Type()) {
				ri.Set(vpv.Convert(ri.Type()))
				return nil
			}
		}
	}
	return fmt.Errorf("inject: value can not convert to %s", ri.Type())
}
