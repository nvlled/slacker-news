package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"strings"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func respondInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "error: %v", err)
	log.Print(err)
	debug.PrintStack()
}

func handleServeLuaPage(
	config Config,
	fsys fs.FS,
) http.Handler {
	initState := *initLuaState(fsys)
	pageDir := http.FileServer(http.Dir("./pages"))
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
		L.SetGlobal("request", luar.New(L, r))
		L.SetGlobal("go", luar.New(L, NewGoLuaBindings(r)))
		L.SetGlobal("go", luar.New(L, NewGoLuaBindings(r)))

		pagePath := path.Clean(r.URL.Path)

		var filename string
		if pagePath == "/" {
			filename = path.Join("pages", "index.lua")
		} else {
			filename = path.Join("pages", pagePath)
		}

		stat, err := os.Stat(filename)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			respondInternalError(w, err)
			return
		}

		if stat != nil && stat.IsDir() {
			filename = path.Join(filename, "index.lua")
		} else if !strings.HasSuffix(filename, ".lua") {
			filename += ".lua"
		}


		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			pageDir.ServeHTTP(w, r)
			return
		}

		if err := DoFile(L, fsys, filename); err != nil {
			respondInternalError(w, err)
			return
		}
		lv := L.Get(-1)
		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("tostring"),
			NRet:    1,
			Protect: true,
		}, lv); err != nil {
			respondInternalError(w, err)
			return
		}

		ret := L.Get(-1)

		//fmt.Fprintf(w, ret.String()) // % is actually parsed...
		_, err = w.Write([]byte(ret.String()))
		if err != nil {
			respondInternalError(w, err)
			return
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
	if err := DoFile(L, fsys, "includes/init.lua"); err != nil {
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
