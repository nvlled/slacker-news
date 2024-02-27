local function loader(modname)
    local mod
    local content, err = readfile("lua/" .. modname .. ".lua")
    if not err then
        mod, err = loadstring(content, modname)
        if err then
            return error(err)
        end
        return mod
    end

    content, err = readfile("includes/" .. modname .. ".lua")
    if not err then
        mod, err = loadstring(content, modname)
        if err then
            return error(err)
        end
        return mod
    end
    return "module not found: " .. modname
end

table.insert(package.loaders, loader)
