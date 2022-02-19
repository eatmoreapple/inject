package inject

import (
	"testing"
)

type Duck interface {
	Quack() string
}

type RealDuck struct {
	Name  string
	Other struct {
		Age int
	}
}

func (r RealDuck) Quack() string {
	return r.Name
}

func TestValue(t *testing.T) {
	name := "gaga"
	var test struct {
		Duck  Duck   `inject:"real_duck"`
		Name  string `inject:"real_duck.Name"`
		Age   int    `inject:"real_duck.Other.Age"`
		Other struct {
			Age int
		} `inject:"real_duck.Other"`
	}
	realDuck := &RealDuck{Name: name, Other: struct{ Age int }{Age: 10}}
	instance := &Instance{Value: realDuck, Name: "real_duck"}
	if err := RegisterInstance(instance); err != nil {
		t.Error(err)
		return
	}
	if err := AutoWired(&test); err != nil {
		t.Error(err)
		return
	}
	if test.Duck.Quack() != name {
		t.Errorf("Expected %s, got %s", name, test.Duck.Quack())
	}
	if test.Name != name {
		t.Errorf("Expected %s, got %s", name, test.Name)
	}
	if test.Age != 10 {
		t.Errorf("Expected %d, got %d", 10, test.Age)
	}
	if test.Other.Age != 10 {
		t.Errorf("Expected %d, got %d", 10, test.Other.Age)
	}
}
