local LAYOUT = require "layout"

local items, err = go:GetTopStories(10, 0)

if err ~= nil then
    return LAYOUT {
        H1 "error: failed to fetch data",
        I { err },
    }
end


local list = {}

for _, item in items() do
    print(item.Title)
    table.insert(list, LI {
        item.Title
    })
end

return LAYOUT {
    noAutoReload = true,
    UL {
        list
    }
}
