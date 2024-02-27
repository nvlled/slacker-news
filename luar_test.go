package main

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestArray(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	a := [...]string{"x", "y"}
	b := [...]string{"x", "y"}

        _ = a
        _ = b
}
