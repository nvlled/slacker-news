package main

import (
	"bufio"
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
	parse "github.com/yuin/gopher-lua/parse"
	luar "layeh.com/gopher-luar"
)

type CompiledLuaModules map[string]*lua.FunctionProto

func handleServeLuaPage(
	config Config,
	fsys fs.FS,
	cacheManager *CacheManager,
) http.Handler {
	var pageDir http.Handler
	pageDir = http.FileServer(http.FS(fsys))

	var modules CompiledLuaModules
	if !config.DevMode {
		modules = compileLuaModules(fsys)
	} else {
		modules = CompiledLuaModules{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		L := initLuaState(config, fsys, modules)
		defer L.Close()

		r.ParseForm()
		L.SetGlobal("config", luar.New(L, config))
		L.SetGlobal("form", luar.New(L, r.Form))
		L.SetGlobal("request", luar.New(L, r))
		L.SetGlobal("go", luar.New(L, NewGoLuaBindings(w, r)))
		L.SetGlobal("cm", luar.New(L, cacheManager))

		if err := DoFile(L, modules, fsys, filename); err != nil {
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
	modules CompiledLuaModules,
) *lua.LState {
	L := lua.NewState(lua.Options{
            SkipOpenLibs: true,
        })

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
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

	L.SetGlobal("loadmodule", L.NewFunction(func(L *lua.LState) int {
		modname := L.ToString(1)
		filename1 := path.Join("lua", modname+".lua")
		filename2 := path.Join("includes", modname+".lua")

		var fn *lua.LFunction
		var err error
		if fsExists(fsys, filename1) {
			fn, err = LoadFile(L, modules, fsys, filename1)
		} else if fsExists(fsys, filename2) {
			fn, err = LoadFile(L, modules, fsys, filename2)
		} else {
			panic(fmt.Errorf("module not found: %v", modname))
		}

		if err != nil {
			panic(err)
		}

		L.Push(fn)
		return 1
	}))

	if err := DoFile(L, modules, fsys, "lua/loader.lua"); err != nil {
		panic(err)
	}
	if err := DoFile(L, modules, fsys, "includes/init.lua"); err != nil {
		panic(err)
	}

	return L
}

func LoadFile(
	L *lua.LState,
	modules CompiledLuaModules,
	fsys fs.FS,
	filename string,
) (*lua.LFunction, error) {
	bytes, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return nil, err
	}

	source := string(bytes)

	if proto, ok := modules[filename]; ok {
		fn := L.NewFunctionFromProto(proto)
		return fn, nil
	} else {
		fn, err := L.Load(strings.NewReader(source), filename)
		if err != nil {
			return nil, err
		}
		return fn, nil
	}

}

func DoFile(
	L *lua.LState,
	modules CompiledLuaModules,
	fsys fs.FS,
	filename string,
) error {
	bytes, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return err
	}

	source := string(bytes)

	if proto, ok := modules[filename]; ok {
		fn := L.NewFunctionFromProto(proto)
		L.Push(fn)
		return L.PCall(0, lua.MultRet, nil)
	} else {
		fn, err := L.Load(strings.NewReader(source), filename)
		if err != nil {
			return err
		}
		L.Push(fn)
		return L.PCall(0, lua.MultRet, nil)
	}

}

func compileLuaModules(fsys fs.FS) CompiledLuaModules {
	modules := CompiledLuaModules{}

	dirFS, ok := fsys.(fs.ReadDirFS)
	if !ok {
		panic("cannot cast os.DirFS to fs.ReadDirFS")
	}

	err := fs.WalkDir(dirFS, ".", func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path.Ext(filename) != ".lua" {
			return nil
		}

		proto, err := CompileLua(fsys, filename)
		if err != nil {
			return err
		}
		modules[filename] = proto

		return nil
	})

	if err != nil {
		panic(err)
	}

	return modules
}

func CompileLua(fsys fs.FS, filePath string) (*lua.FunctionProto, error) {
	file, err := fsys.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

func respondInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "error: %v", err)
	log.Print(err)
	debug.PrintStack()
}

func preloadModules(L *lua.LState, fsys fs.FS, modules CompiledLuaModules) error {
	dirFS, ok := fsys.(fs.ReadDirFS)
	if !ok {
		log.Print("cannot preload modules, fsys cannot be cast to fs.ReadDirFS")
		return nil
	}

	for _, dir := range []string{"includes", "lua"} {
		entries, err := dirFS.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			name := entry.Name()
			filename := path.Join(dir, name)
			if path.Ext(name) != ".lua" {
				continue
			}

			i := strings.Index(entry.Name(), ".")
			if i < 0 {
				continue
			}

			moduleName := name[:i]

			L.PreloadModule(moduleName, func(L *lua.LState) int {
				if err := DoFile(L, modules, fsys, filename); err != nil {
					panic(err)
				}
				return 1
			})
		}
	}

	return nil
}
