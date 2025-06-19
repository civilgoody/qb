package main

import (
	"errors"
	"fmt"
)
// Great Fast Tutorial: https://www.youtube.com/watch?v=8uiZC0l4Ajw&t=399s

// Structs
type Person struct {
	name string
	age int
}

func (p *Person) namePerson(age int) {
	p.name = "Gin"
	p.age = age
}

type User struct{
	*Person
	id string
}

func main() {
	println("working")
	// Arrays
	arr := []int{1, 2, 3}
	arr2 := []int{4, 5, 6}
	fmt.Println(append(arr, arr2...))

	// Functions
	d, r, err := divide(6, 0)
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("Result: %v . %v\n", d, r)

	// Map
	dict := map[string]int{"a": 2, "g": 3}
	fmt.Println(dict["a"])
	fmt.Println(dict["k"])
	val, ok := dict["e"]

	if !ok {
		println("No exists")
	} else {
		print(val)
	}

	// Loops

	for i, v := range arr {
		println(i, v)
	}

	for k, v := range dict {
		println(k, v)
	}

	// Iterating String - Runes
	str := "hällö "
	println("Strings are byte-indexed")
	for i, v := range str {
		println(i, v)
	}
	fmt.Println("Len:", len(str))
	

	str2 := []rune("hällö ")
	println("Use runes to iterate strings")
	for i, v := range str2 {
		fmt.Printf("%v %c\n", i, v)
	}
	fmt.Println("Len:", len(str2))

	// Simpler syntax
	str3 := "ö"
	fmt.Println(str3)
	
	// Structs
	p := Person{name: "Go"}
	p.namePerson(30)

	fmt.Println(p)

	p2 := Person{name: "Lin"}
	u:= User{Person: &p2, id: "1"}
	u.name = "Loki"
	fmt.Println(u.name, u.id)
	fmt.Println(p2.name)
}

func divide(a int, b int) (int, int, error) {
	var err error
	if (b == 0) {
		err = errors.New("Cannot divide by zero")
		return 0, 0, err
	}
	division := a/b
	remainder := a%b
	return division, remainder, err
}
