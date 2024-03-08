local LAYOUT = require "layout"

local page = tonumber(form:Get("page")) or 1
local pageSize = 30

local items, err, hasMore = go:GetTopStories(pageSize, page - 1)

if not Xt.isEmptyString(err) then
    return LAYOUT {
        H1 "error: failed to fetch data",
        I { err },
    }
end


local style = {
    CSS 'ol li' {
        margin_bottom = 5,
    },
    CSS 'ol a' {
        font_size="1.1rem",
        text_decoration = "none",
    },
}

local list = {}

for _, item in items() do
    table.insert(list, LI {
        A { href = "/item?id=" .. tostring(item.ID), item.Title },
        BR,
        SMALL { item.Score, " points ", item.Descendants, " comments" },
    })
end

return LAYOUT {
    STYLE(style),
    OL {
        start = (page - 1) * pageSize + 1,
        list
    },
    hasMore and A { href = "/index?page=" .. page + 1, "more" }
}
