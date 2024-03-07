local LAYOUT = require "layout"

local id = tonumber(form:Get("id"))
local items, err = go:GetThread(id)


if not Xt.isEmptyString(err) then
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
    CSS_MEDIA '(orientation: portrait)' {
        CSS ".post.popup" {
            max_width = '80vw !important',
        }
    },

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
        position = "absolute",
        border = "0.05rem solid gray",
        max_width = '50vw',
        overflow = "auto"
    },

    CSS ".post.dead .post-body" {
        color = "#121",
    },

    CSS ".post-header" {
        margin_bottom = "0.3rem",
        max_width = "100%",
        overflow_wrap = "wrap",
        --word_break = "keep-all",
        CSS " > * " {
            margin_right = "0.2rem",
        },
        CSS ".own-id a" {
            color = "inherit",
            text_decoration = "none",
            [":hover"] = { color = LAYOUT.style.linkColor, },
        },
        CSS "a.reply" {
            margin_right = "0.4rem",
            font_size = "0.8rem",
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
        display = "inline-block",
        --word_break="keep-all",
    },
    CSS ".post-link.dead" {
        text_decoration = "line-through underline",
    },
    CSS ".post-link.parent" {
        margin_top = 10,
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
    },
    CSS "#top-commenters" {
        font_size = "0.8rem",
        CSS "ul" {
            margin = 0,
            padding = 0,
        },
        CSS "li" {
            display = "inline-block",
            border_left = "2px solid " .. LAYOUT.style.linkColor,
            padding = "0 5px",
            font_style = "italic",
        }
    },
    CSS "#context" {
        border_left = "2px solid gray",
        padding_left = 5,
    },
}

local op
local list = {}
local deadPosts = {}
local kidIDs = {}
local commentTally = {}
local commentCount = 0

for i, item in items() do
    if i == 1 then
        op = item
    end

    local id = tostring(item.ID)
    kidIDs[id] = {}

    if item.Dead or not item.By or not item.Text or item.Text == "" or item.Text == "" then
        deadPosts[tostring(item.ID)] = true
    else
        commentCount = commentCount + 1

        if not commentTally[item.By] then
            commentTally[item.By] = 0
        end
        commentTally[item.By] = commentTally[item.By] + 1

        if item.Kids then
            for _, childID in item.Kids() do
                table.insert(kidIDs[id], childID)
            end
            table.sort(kidIDs[id])
        end
    end
end

local commentChain = {}
if op and op.Type == "comment" and op.Parent then
    local chain, err = go:GetCommentChain(op.Parent)

    if not Xt.isEmptyString(err) then
        return LAYOUT {
            H1 "error: failed to fetch data",
            I { type(err) == "userdata" and err() or err },
        }
    end

    for _, item in chain() do
        table.insert(commentChain, item)
    end
end


local topCommenters = {}
for username, count in pairs(commentTally) do
    if count >= 2 then
        table.insert(topCommenters, { username, count })
    end
end

table.sort(topCommenters, function(a, b)
    return a[2] > b[2]
end)
for i = 11, #topCommenters do
    topCommenters[i] = nil
end

--if commentCount > 0 then commentCount = commentCount - 1 end

local function renderOP(item)
    local url = go:ParseURL(item.Url)
    return DIV {
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
            SPAN { " by ", item.By },
            SPAN { " | ", commentCount, " comments", }
        },
        BR,
    }
end

local function renderItem(item)
    return DIV {
        id = "item-" .. item.ID,
        class = 'post' .. (item.Dead and ' dead' or ''),
        DIV {
            class = "post-header",
            B / (tostring(item.By)),
            item.Dead and SPAN { "[dead post]" },
            SPAN { go:FormatTime(item.Time) },
            SPAN {
                class = "own-id",
                A { href = "#item-" .. item.ID, "No." },
                A { href = "/item?id=" .. item.ID, item.ID },
            },
            item.Level > 2 and SPAN {
                title = "comment depth, or how deep it is the comment heirarchy tree",
                "{", item.Level, "}"
            },
            SPAN { class = "triangle", "▶" },
            Xt.map(kidIDs[tostring(item.ID)] or {}, function(childID)
                return A { class = "reply post-link" .. (deadPosts[tostring(childID)] and " dead" or ""),
                    href = "#item-" .. childID,
                    ">>" .. tostring(childID)
                }
            end)
        },
        item.Parent and DIV {
            A {
                class = "post-link parent", href = "#item-" .. item.Parent,
                --data_id = item.ID,
                ">>" .. item.Parent, (op and item.Parent == op.ID and op.type == "story" and " (OP)" or nil)
            }
        },
        DIV {
            class = "post-body",
            __noHTMLEscape = true,
            item.Text,
        },
    }
end

for i, item in items() do
    local node

    if i == 2 then
        table.insert(list, DIV {
            BR,
            EM "comments"
        })
    end

    if i == 1 and item.Type == "story" then
        node = renderOP(item)
    elseif not deadPosts[item.ID] then
        node = renderItem(item)
    end
    table.insert(list, DIV {
        node
    })
end


return LAYOUT {
    title = op and op.title,
    style,

    commentCount > 10 and DIV {
        id = "top-commenters",
        "Most active commenters",
        UL {
            Xt.map(topCommenters, function(entry)
                return LI { entry[1], "(", entry[2], ")" }
            end),
        },
    },


    op and op.Type == "comment" and DIV {
        A { href = "/item?id=" .. commentChain[1].ID, "←back to thread" },
        BR,
        BR,

        DIV {
            id = "context",
            DETAILS {
                SUMMARY { "Show context", },
                Xt.map(commentChain, function(item)
                    if item.Type == "story" then
                        return renderOP(item)
                    else
                        return renderItem(item)
                    end
                end)
            },
        },
    },


    DIV {
        id = "thread",
        list
    },


    SCRIPT { src = "item.js" },
}
