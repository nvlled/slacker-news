local LAYOUT = require "layout"

local id = tonumber(form:Get("id"))
local items, err = go:GetThread(id)

if err ~= nil then
    return LAYOUT {
        H1 "error: failed to fetch data",
        I { err },
    }
end

local style = STYLE {
    --CSS ".post-link.reply" { display = "none" },
    --CSS_MEDIA '(min-width: 1250px)' {
    --    CSS ".post-link.reply" { display = "inline-block" },
    --},

    CSS ".post" {
        position = "relative",
        background_color = "#282a2e",
        border = "0.05rem solid #282a2e",
        padding = "0.5rem",
        margin_bottom = "0.3rem",
    },

    CSS ".post.selected" {
        background = '#1d1d21 !important',
        border = '0.0.5rem solid #111 !important',
    },

    CSS ".post.popup" {
        position = "fixed",
        border = "0.05rem solid gray",
        z_index = "500",
        max_width = '50vw',
    },

    CSS ".post.dead .post-body" {
        color = "#121",
    },

    CSS ".post-header" {
        margin_bottom = "0.3rem",
        max_width = "100%",
        word_break = "break-word",
        CSS " > * " {
            margin_right = "0.2rem",
        },
        CSS ".own-id a" {
            color = "inherit",
            text_decoration = "none",
            [":hover"] = { color = LAYOUT.style.linkColor, },
        },
        CSS "a.reply" {
            font_size = "0.7rem",
            margin_right = "0.3rem",
            --display = "inline-block",
        },
    },
    CSS ".post-body" {
        word_break = "break-word",
    },
    CSS ".post-body pre" {
        overflow_x = "auto",
        white_space = "normal",
        word_break = "break-all",
    },

    CSS ".post-link" {
    },
    CSS ".post-link.dead" {
        text_decoration = "line-through underline",
    },

    CSS ".op" {
        background = LAYOUT.style.bgColor,
    },
    CSS ".op-title" {
        CSS "h1" {
            display = "inline-block",
            font_size = "1.5rem",
            margin = 0,
        },
        CSS "a" {
            text_decoration = "none",
            color = LAYOUT.style.textColor,
        },
        CSS "small" {
            color = "gray",
        }
    },
    CSS ".op-details" {
    }
}

local list = {}
local deadPosts = {}

local commentCount = 0
for _, item in items() do
    if not item.Dead and item.By and item.Text and item.Text ~= "" then
        commentCount = commentCount + 1
    else
        -- okay, that sucks, directly using item.ID
        -- without using tostring
        -- here adds a 3 second lag, most likely
        -- caused by gopher-luar
        deadPosts[tostring(item.ID)] = true
    end
end
if commentCount > 0 then commentCount = commentCount - 1 end

local op
for i, item in items() do
    local node
    if i == 1 then
        op = item
    end

    if i == 1 and item.Type == "story" then
        local url = go:ParseURL(item.Url)
        node = DIV {
            class = "post op",
            id = "item-" .. item.ID,
            DIV {
                class = "op-title",
                H1 ^ A { href = item.Url, item.Title },
                url and SMALL { " (", url.Host, ")" },
            },
            DIV {
                class = "op-details",
                SPAN { item.Score, " points" },
                SPAN { "by ", item.By },
                SPAN { " | ", commentCount, " comments", }
            },
            BR,
        }
    elseif not deadPosts[item.ID] then
        node = DIV {
            id = "item-" .. item.ID,
            class = 'post' .. (item.Dead and ' dead' or ''),
            DIV {
                class = "post-header",
                B / (tostring(item.By)),
                item.Dead and SPAN { "[dead post]" },
                SPAN { go:FormatTime(item.Time) },
                SPAN {
                    class = "own-id",
                    A { href = "#", "No." },
                    A { href = "/item?id=" .. item.ID, item.ID },
                },
                SPAN { class = "triangle", "▶" },
                item.Kids and function()
                    local result = {}
                    for _, childID in item.Kids() do
                        table.insert(result,
                            A { class = "reply post-link" .. (deadPosts[tostring(childID)] and " dead" or ""),
                                href = "#item-" .. childID,
                                ">>" .. tostring(childID)
                            }
                        )
                    end
                    return FRAGMENT(result)
                end,
            },
            item.Parent and DIV {
                A {
                    class = "post-link", href = "#item-" .. item.Parent,
                    --data_id = item.ID,
                    ">>" .. item.Parent, (op and item.Parent == op.ID and " (OP)" or nil)
                }
            },
            DIV {
                class = "post-body",
                __noHTMLEscape = true,
                item.Text,
            },
        }
    end
    table.insert(list, DIV {
        node
    })
end


return LAYOUT {
    title = op and op.title,
    style,
    DIV {
        list
    },
    SCRIPT { src = "item.js" },
}
