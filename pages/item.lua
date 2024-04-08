local LAYOUT = require "layout"

local timeStart = go:UnixMilli()
local hnBaseURL = "https://news.ycombinator.com"
local id = tonumber(form:Get("id"))
local items, err = go:GetThread(id)
local timeEnd = go:UnixMilli()

if not Xt.isEmptyString(err) then
    return LAYOUT {
        H1 "error: failed to fetch data",
        I { err },
    }
end

local op
local list = {}
local deadPosts = {}
local kidIDs = {}
local commentTally = {}
local replyTally = {}
local commentCount = 0
local firstComment = nil

for i, item in items() do
    if i == 1 then
        op = item
    end
    if (i == 1 and item.Type == "comment") or i == 2 then
        firstComment = item
    end

    local id = tostring(item.ID)
    kidIDs[id] = {}


    if item.Dead or not item.By or not item.Text or item.Text == "" or item.Text == "" then
        deadPosts[tostring(item.ID)] = true
    else
        commentCount = commentCount + 1

        local by = tostring(item.By)
        if not commentTally[by] then
            commentTally[by] = 0
        end
        commentTally[by] = commentTally[by] + 1

        if item.Kids then
            for _, childID in item.Kids() do
                table.insert(kidIDs[id], childID)
            end
            table.sort(kidIDs[id])
            if #kidIDs[id] >= 3 and op.ID ~= item.ID then
                table.insert(replyTally, { id = id, count = #kidIDs[id] })
            end
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
        if kidIDs[id] then
            goto continue
        end

        local id = tostring(item.ID)
        kidIDs[id] = {}

        if item.Dead or not item.By or not item.Text or item.Text == "" or item.Text == "" then
            deadPosts[tostring(item.ID)] = true
        else
            --commentCount = commentCount + 1

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

        table.insert(commentChain, item)

        ::continue::
    end
end


local topCommenters = {}
for username, count in pairs(commentTally) do
    if count >= 3 then
        table.insert(topCommenters, { username, count })
    end
end

table.sort(topCommenters, function(a, b)
    return a[2] > b[2]
end)

table.sort(replyTally, function(a, b)
    return a.count > b.count
end)

--if commentCount > 0 then commentCount = commentCount - 1 end

local function renderOP(item)
    local url = go:ParseURL(item.Url)
    return DIV {
        class = "post op",
        id = "item-" .. item.ID,
        DIV {
            class = "op-title",
            H1 ^ A { href = item.Url, item.Title },
            not Xt.isEmptyString(url.Host) and SMALL { " (", url.Host, ")" },
        },
        DIV {
            class = "op-details",
            SPAN { item.Score, " points" },
            A { class = "username", href = hnBaseURL .. "/user?id=" .. item.By, " ", item.By },
            SPAN { " | ", commentCount, " comments | ", },
            SPAN { class = "post-datetime", go:FormatTime(item.Time) },
            SPAN " | ",
            "HN request time: ",
            (timeEnd - timeStart) / 1000,
            "s",
            SPAN " | ",
            A { class = "source", href = hnBaseURL .. "/item?id=" .. item.ID, "source" },
            commentCount > 5 and FRAGMENT {
                id = "top",
                SPAN " | ",
                SPAN { A { href = "#bottom", "bottom" }, },
            },
        },
        item.Text and DIV {
            __noHTMLEscape = true,
            not Xt.isEmptyString(item.Text) and BR,
            item.Text
        },
        DIV { id = "post-footer" },
    }
end

local function renderItem(item, num)
    local replies = kidIDs[tostring(item.ID)] or {}
    local parentLinkSuffix = ""
    if op then
    if (item.Parent == op.ID and op.Type == "story") or (commentChain[1] and item.Parent == commentChain[1].ID) then
        parentLinkSuffix = " (OP)"
    elseif item.Parent == op.ID and op.Type == "comment" then
        parentLinkSuffix = " (TP)"
    end
end


    return DIV {
        id = "item-" .. item.ID,
        class = 'post' .. (item.Dead and ' dead' or ''),
        DIV {
            class = "post-header",
            num and {
                SPAN { class = "comment-num", num + 1, ". " },
            },
            A {
                class = "username",
                href = hnBaseURL .. "/user?id=" .. item.By,
                SPAN { class = "alias", item.By },
            },

            item.Dead and SPAN { "[dead post]" },
            " ◴[", SPAN { class = "post-datetime", go:FormatTime(item.Time) }, "] ",
            SPAN {
                class = "own-id",
                A { href = "#item-" .. item.ID, "No." },
                A { href = "/item?id=" .. item.ID, item.ID },
            },
            item.Level > 2 and SPAN {
                title = "comment depth, or how deep it is the comment heirarchy tree",
                "{", item.Level, "}"
            },
            A { class = "source", href = hnBaseURL .. "/item?id=" .. item.ID, "[source]" },
            SPAN { class = "triangle", "▶" },
            BR,
        },
        item.Parent and DIV {
            A {
                class = "post-link parent", href = "#item-" .. item.Parent,
                --data_id = item.ID,
                ">>" .. item.Parent, parentLinkSuffix
            },
            A { class = "post-link-hash" .. (deadPosts[tostring(item.Parent)] and " dead" or ""),
                href = "#item-" .. item.Parent,
                " #",
            }
        },
        DIV {
            class = "post-body",
            __noHTMLEscape = true,
            item.Text,
        },
        #replies > 0 and DIV {
            class = "post-replies-container",
            I { " replies(", #replies, "): " },
            Xt.map(replies, function(childID)
                return SPAN {
                    class = "reply-post-link-container",
                    A { class = "reply post-link" .. (deadPosts[tostring(childID)] and " dead" or ""),
                        href = "#item-" .. childID,
                        ">>" .. tostring(childID)
                    },
                    A { class = "post-link-hash" .. (deadPosts[tostring(childID)] and " dead" or ""),
                        href = "#item-" .. childID,
                        " #",
                    }
                }
            end)
        },
        DIV { id = "post-footer" },
    }
end

local commentNum = 0
for i, item in items() do
    local node

    if i == 1 and item.Type == "story" then
        node = renderOP(item, i)
    elseif not deadPosts[tostring(item.ID)] then
        node = renderItem(item, commentNum)
        commentNum = commentNum + 1
    end

    table.insert(list, DIV {
        node
    })
end

table.insert(list, 2, DIV {
    id = "thread-header",
    FORM {
        action = "submit-id",
        id = "thread-id-input",
        LABEL {
            "ID: ",
            INPUT { name = "id", value = op and op.ID or 0, placeholder = "HN item ID or url" },
        },
        BUTTON { "GO" }
    },

})



return LAYOUT {
    title = #commentChain > 0 and commentChain[1].Title or (op and op.Title),
    style = {
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
            border_left = "2px solid " .. LAYOUT.style.linkColor,
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
                text_decoration = "underline",
                font_size = "0.8rem",
                --display = "inline-block",
            },
        },

        CSS ".reply-post-link-container" {
            white_space = "nowrap",
            margin_right = "0.7rem",
        },

        CSS ".post-link-hash" {
            display = "none",
        },
        CSS ".m .post-link-hash" {
            display = "inline",
        },

        CSS ".post-body" {
            word_break = "break-word",
            CSS "pre" {
                overflow_x = "auto",
                white_space = "normal",
                word_break = "break-all",
            },
            CSS "p:first-child" {
                margin_top = 0,
            },
            CSS "p:last-child" {
                margin_bottom = 0,
            },
            CSS "pre" {
                background = "#333",
                margin = 0,
                padding = 10,
            }
        },

        CSS ".post-link" {
            display = "inline-block",
            [":active"] = { color = "red" },
            [":hover"] = { color = "orange" }
            --word_break="keep-all",
        },
        CSS ".post-link.dead" {
            text_decoration = "line-through underline !important",
        },
        CSS ".post-link.parent" {
            margin_top = 10,
        },
        CSS ".post-replies-container" {
            margin_top = 8,
            font_size = ".7em",
            line_height = "1.9",
            CSS "i" {
                margin_right = 5,
            }
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
        CSS "#most-popular" {
            font_size = "0.8rem",
            CSS "ul" {
                margin = 0,
                padding = 0,
            },
            CSS "li" {
                display = "inline-block",
                font_style = "italic",
            }
        },
        CSS "#context" {
            border_left = "2px solid gray",
            padding_left = 5,
        },

        CSS "#thread-header" {
            display = "flex",
            justify_content = "space-between",
        },
        CSS "#back-to-top" {
            font_size = "1.8em",
            text_size_adjust = "none",
            position = "fixed",
            bottom = 10,
            right = 10,
            text_align = "center",
            CSS "a" {
                background = "#6419",
                color = LAYOUT.style.textColor, text_decoration = "none",
                width = "1.2em",
                height = "1.2em",
                display = "flex",
                align_items = "center",
                justify_content = "center",
            },
        },
        CSS "#floating-nav" {
            padding = 3,
            background = "#111",
            position = "fixed",
            bottom = "0",
            right = 0,
            text_align = "right",
            CSS "> *" {
                margin_right = 5
            }
        },
        CSS "#thread-id-input" {
            display = "flex",
            CSS "label" {
                display = "flex",
            }
        },
        CSS "#footer-notice" {
            display = "inline-block",
            padding = 2,
            font_size = ".7rem",
            CSS "i" { background = LAYOUT.style.bgColor },
        },

        CSS ".username" {
            text_decoration = "none",
            color = LAYOUT.style.textColor,
        },
        CSS ".post-header .username" {
            text_decoration = "none",
            color = LAYOUT.style.textColor,
            display = "inline-block",
        },
        CSS ".comment-num" {
            color = "#777",
            font_size = ".5em",
            position = "absolute",
            top = "-0px",
            right = "-0px",
        },
        CSS "pre code" {
            white_space = "break-spaces",
            text_wrap = "nowrap !important",
        },
        CSS ".green-text" {
            color = "#b5bd68",
        },
    },

    SCRIPT { src = "item.js" },


    commentCount > 10 and DIV {
        id = "top-commenters",
        "Most active commenters",
        UL {
            Xt.mapSlice(1, 10, topCommenters, function(entry)
                return LI { entry[1], "(", entry[2], ")" }
            end)
        },
    },

    BR,

    #replyTally > 0 and DIV {
        id = "most-popular",
        "Popular/hot comments",
        UL {
            Xt.mapSlice(1, 20, replyTally, function(entry)
                return LI {
                    class = "reply-post-link-container",
                    A { class = "reply post-link" .. (deadPosts[tostring(entry.id)] and " dead" or ""),
                        href = "#item-" .. entry.id,
                        ">>" .. tostring(entry.id)
                    },
                    A { class = "post-link-hash" .. (deadPosts[tostring(entry.id)] and " dead" or ""),
                        href = "#item-" .. entry.id,
                        " #",
                    }
                }
            end),
        },
    },

    op and op.Type == "comment" and DIV {
        A { href = "/item?id=" .. commentChain[1].ID .. "#item-" .. op.ID, "←back to thread" },
        BR,
        BR,

        DIV {
            renderOP(commentChain[1]),
        },

        #commentChain > 1 and DIV {
            id = "context",
            DETAILS {
                SUMMARY { "Show context", },
                Xt.mapSlice(2, #commentChain, commentChain, function(item)
                    return renderItem(item)
                end)
            },
        },
    },


    DIV {
        id = "thread",
        list
    },

    commentCount == 0 and DIV {
        BR,
        EM "(no comments)",
    },

    DIV {
        id = "back-to-top",
        class = "thread-nav",
        title = "back to top",
        SPAN { A { href = "#site-nav", " ↑ " }, },
    },

    firstComment and firstComment.FetchTime > 0 and DIV {
        id = "footer-notice",

        SMALL ^ I {
            "NOTE: showing cached content from ",
            math.floor(os.difftime(os.time(), firstComment.FetchTime) / 60),
            " minutes ago, next update will be in ",
            math.floor(os.difftime(cm:GetNextUpdate(), os.time()) / 60),
            " minutes",
        }
    },

    DIV {
        id = "bottom",
    }
}
