package main

import (
	"fmt"
	"github.com/eatmoreapple/inject"
)

type Quacker interface {
	Quack()
}

type RealQuacker struct {
	Name string
}

func (r RealQuacker) Quack() {
	fmt.Printf("%s: Quack!\n", r.Name)
}

func main() {
	injecter := inject.New()

	var bird struct {
		Quacker Quacker `inject:"main.RealQuacker"`
	}
	quacker := RealQuacker{Name: "Daffy Duck"}

	// Inject the quacker into the struct
	// if Name is empty, use pkg path instead
	instance := inject.Instance{Value: &quacker, Name: ""}
	if err := injecter.Register(&instance); err != nil {
		panic(err)
	}
	if err := injecter.AutoWired(&bird); err != nil {
		panic(err)
	}

	bird.Quacker.Quack() // Daffy Duck: Quack!
}
