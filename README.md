# inject
golang autowise



#### example

```go
package main

import (
	"fmt"
	"github.com/eatmoreapple/inject"
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

func init() {
	var name = Name{Name: "John"}
	fmt.Println(inject.Repository(&name))
	fmt.Println(inject.Repository("ok", "ok"))
}

func main() {

	var a = new(Name)
	var b Stringer
	var c string
	var d struct {
		B   *Name
		Ok  Stringer
		Aok string `inject:"ok"`
	}
	fmt.Println(inject.Autowired(&a))
	fmt.Println(a)
	fmt.Println(inject.Autowired(&b))
	fmt.Println(b)
	fmt.Println(inject.Autowired(&c, "ok"))
	fmt.Println(c)
	fmt.Println(inject.AutowiredStruct(&d))
	fmt.Println(d)
}
```

