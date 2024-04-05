table.insert(package.loaders, loadmodule)

function respond(obj)
    if obj.print then
        obj:print()
    elseif write then
        write(obj)
    else
        io.write(obj)
    end
end
