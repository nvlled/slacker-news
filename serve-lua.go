package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"
	"runtime/debug"
	"strings"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func handleServeLuaPage(
	config Config,
	fsys fs.FS,
) http.Handler {
	initState := *initLuaState(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var L *lua.LState
		if config.DevMode {
			L = initLuaState(fsys)
		} else {
			copyState := initState
			L = &copyState
			if L == &initState {
				panic("copy failed")
			}
		}
		defer L.Close()

		r.ParseForm()
		L.SetGlobal("form", luar.New(L, r.Form))
		L.SetGlobal("go", luar.New(L, NewGoLuaBindings(r)))

		pagePath := path.Clean(r.URL.Path)

		filename := path.Join("pages", pagePath)
		if !strings.HasSuffix(filename, ".lua") {
			filename += ".lua"
		}

                // TODO: replace with pages/index.lua with actual path
		if err := DoFile(L, fsys, "pages/index.lua"); err != nil {
			fmt.Fprintf(w, "error: %v", err)
			log.Print(err)
			debug.PrintStack()
		} else {
			lv := L.Get(-1)
			if err := L.CallByParam(lua.P{
				Fn:      L.GetGlobal("tostring"),
				NRet:    1,
				Protect: true,
			}, lv); err != nil {
				panic(err)
			}
			ret := L.Get(-1)
			fmt.Fprintf(w, ret.String())
		}
	})
}

func initLuaState(fsys fs.FS) *lua.LState {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}

	L.SetGlobal("readfile", luar.New(L, func(filename string) (string, error) {
		file, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return "", err
		}
		return string(file), nil
	}))

	if err := DoFile(L, fsys, "lua/loader.lua"); err != nil {
		panic(err)
	}
	if err := DoFile(L, fsys, "pages/init.lua"); err != nil {
		panic(err)
	}

	return L
}

func DoFile(
	L *lua.LState,
	fsys fs.FS,
	filename string,
) error {
	bytes, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return err
	}

	source := string(bytes)

        // TODO: cache fn on PROD
        // fn.Env

	if fn, err := L.Load(strings.NewReader(source), filename); err != nil {
		return err
	} else {
		L.Push(fn)
		return L.PCall(0, lua.MultRet, nil)
	}
}
