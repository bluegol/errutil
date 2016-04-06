package errutil

import (
	"errors"

	"testing"
	"fmt"
)



func TestAll(t *testing.T) {
	e1 := New(errors.New("Test1"), NoCallStack)
	fmt.Println(e1)
	e2 := New(errors.New("Test2"))
	fmt.Println(e2)
	e3 := New(errors.New("Test3"), FullCallStack, "where", "here")
	fmt.Println(e3)
	e4 := Embed(errors.New("Test4"), errors.New("inner error"),
		"who", "me", MoreInfo, "detail comes first")
	fmt.Println(e4)
	DefaultCallStackLevel = FullCallStack
	e5 := New(errors.New("Test5"))
	fmt.Println(e5)

}
