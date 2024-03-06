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
	"sync"

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
	cacheManager *CacheManager,
) http.Handler {

	var pageDir http.Handler
	//pageDir = http.FileServer(http.Dir("./pages"))
	pageDir = http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		L := initLuaState(config, fsys)
		defer L.Close()

		pagePath := path.Clean(r.URL.Path)

		var filename string
		if pagePath == "/" {
			filename = path.Join("pages", "index.lua")
		} else {
			filename = path.Join("pages", pagePath)
		}

		stat, err := fsStat(fsys, filename)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			respondInternalError(w, err)
			return
		}

		if stat != nil && stat.IsDir() {
			filename = path.Join(filename, "index.lua")
		} else if !strings.HasSuffix(filename, ".lua") {
			filename += ".lua"
		}

		if !fsExists(fsys, filename) {
			r.URL.Path = path.Join("pages", r.URL.Path)
			pageDir.ServeHTTP(w, r)
			return
		}

		r.ParseForm()
		L.SetGlobal("form", luar.New(L, r.Form))
		L.SetGlobal("request", luar.New(L, r))
		L.SetGlobal("go", luar.New(L, NewGoLuaBindings(r)))
		L.SetGlobal("cm", luar.New(L, cacheManager))

		if err := DoFile(L, config, fsys, filename); err != nil {
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

		_, err = w.Write([]byte(ret.String()))
		if err != nil {
			respondInternalError(w, err)
			return
		}
	})
}

func initLuaState(
	config Config,
	fsys fs.FS,
) *lua.LState {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.OsLibName, lua.OpenOs},
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

	if err := DoFile(L, config, fsys, "lua/loader.lua"); err != nil {
		panic(err)
	}
	if err := DoFile(L, config, fsys, "includes/init.lua"); err != nil {
		panic(err)
	}

	return L
}

var modCache = struct {
	mu sync.Mutex
	m  map[string]*lua.LFunction
}{
	mu: sync.Mutex{},
	m:  map[string]*lua.LFunction{},
}

func DoFile(
	L *lua.LState,
	config Config,
	fsys fs.FS,
	filename string,
) error {
	bytes, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return err
	}

	source := string(bytes)

	var fn *lua.LFunction
	var ok bool

	if !config.DevMode {
		modCache.mu.Lock()
		fn, ok = modCache.m[filename]
		modCache.mu.Unlock()
	}

	if !ok || fn == nil {
		if fn, err = L.Load(strings.NewReader(source), filename); err != nil {
			return err
		}

		if !config.DevMode {
			modCache.mu.Lock()
			modCache.m[filename] = fn
			modCache.mu.Unlock()
		}
	}

	g := L.NewFunctionFromProto(fn.Proto)

	L.Push(g)
	err = L.PCall(0, lua.MultRet, nil)

	return err
}
