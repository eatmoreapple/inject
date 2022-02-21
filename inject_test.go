package inject

import (
	"testing"
)

type Stringer interface {
	String() string
}

type Name struct {
	Name string
}

func (n Name) String() string {
	return n.Name + "!"
}

func TestInjector(t *testing.T) {
	var name = Name{Name: "John"}
	if err := Repository(&name); err != nil {
		t.Error(err)
	}

	if err := Repository("ok", "ok"); err != nil {
		t.Error(err)
	}

	var a = new(Name)
	var b Stringer
	var c string
	var d struct {
		Name     *Name
		Stringer Stringer
		Aok      string `inject:"ok"`
	}

	if err := Autowired(&a); err != nil {
		t.Error(err)
	}
	if err := Autowired(&b); err != nil {
		t.Error(err)
	}
	if err := Autowired(&c, "ok"); err != nil {
		t.Error(err)
	}
	if err := AutowiredStruct(&d); err != nil {
		t.Error(err)
	}
	t.Log(a == b, c == d.Aok, d.Name == d.Stringer, d.Name == a, c == "ok")
}
