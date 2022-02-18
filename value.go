package inject

import (
	"errors"
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
	Name    string
	Value   interface{}
	rv      reflect.Value
	mapping map[string]*Instance
	base    bool
	inited  bool
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

func (i *Instance) init() error {
	if i.inited {
		return nil
	}
	i.inited = true
	i.mapping = make(map[string]*Instance)
	if i.base {
		return nil
	}
	if i.Value == nil {
		return errors.New("inject: instance value is nil")
	}
	i.rv = reflect.ValueOf(i.Value)
	ri := reflect.Indirect(i.rv)
	rt := ri.Type()
	if i.Name == "" {
		i.Name = strings.Replace(rt.PkgPath(), "/", ".", -1) + "." + rt.Name()
	}
	i.mapping[i.Name] = i
	// is a struct
	if ri.Kind() == reflect.Struct {
		for j := 0; j < ri.NumField(); j++ {
			field := ri.Field(j)
			field = reflect.Indirect(field)
			if field.Kind() == reflect.Struct {
				instance := Instance{Value: field.Interface(), Name: i.Name + "." + rt.Field(j).Name}
				if err := instance.init(); err != nil {
					return err
				}
				for k, v := range instance.mapping {
					i.mapping[k] = v
				}
				continue
			}
			i.mapping[i.Name+"."+rt.Field(j).Name] = &Instance{Value: field.Interface(), Name: i.Name + "." + rt.Field(j).Name}
		}
	}
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
		for k, v := range s.mapping {
			i.mapping[k] = v
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
			if field.CanSet() {
				inject := t.Field(j).Tag.Get("inject")
				if inject == "" {
					continue
				}
				v, ok := i.mapping[inject]
				if !ok {
					return errors.New("inject: auto wired value not found")
				}
				field.Set(reflect.ValueOf(v.Value))
			}
		}
	}
	return nil
}
