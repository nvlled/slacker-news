package main

import (
	"io/fs"
	"os"
	"testing"

	luar "layeh.com/gopher-luar"
)

func TestLuaExec(t *testing.T) {
	var modules CompiledLuaModules
	var config = Config{DevMode: isDevMode}
	d := os.DirFS(".")
	fsys, _ := d.(fs.ReadFileFS)
	L := initLuaState(config, fsys, modules)
	L.SetGlobal("write", luar.New(L, func(s string) {
		print(s)
	}))
	err := L.DoString(`
            local node = HTML{
                H1 "test"
            }
            blah(node)
        `)
	if err != nil {
		panic(err)
	}
	println()
}
