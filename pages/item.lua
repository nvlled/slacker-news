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
local userAlias = {}
local userAliasID = 0
local kidIDs = {}
local commentTally = {}
local commentCount = 0
local firstComment = nil

for i, item in items() do
    if i == 1 then
        op = item
    end
    if (i == 1 and item.Type == "comment") or i == 2 then
        firstComment = item
    end

    if not userAlias[item.By] then
        userAliasID = userAliasID + 1
        userAlias[item.By] = userAliasID
    end

    local id = tostring(item.ID)
    kidIDs[id] = {}

    if item.Dead or not item.By or not item.Text or item.Text == "" or item.Text == "" then
        deadPosts[tostring(item.ID)] = true
    else
        commentCount = commentCount + 1

        local by = tostring(item.By)
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
        if kidIDs[id] then
            goto continue
        end

        if not userAlias[item.By] then
            userAliasID = userAliasID + 1
            userAlias[item.By] = userAliasID
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

        table.insert(commentChain, item)

        ::continue::
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
            not Xt.isEmptyString(url.Host) and SMALL { " (", url.Host, ")" },
        },
        DIV {
            class = "op-details",
            SPAN { item.Score, " points" },
            A { class = "username", href = hnBaseURL .. "/user?id=" .. item.By, " ", "slacker" .. userAlias[item.By] },
            SPAN { " | ", commentCount, " comments | ", },
            SPAN { class = "post-datetime", go:FormatTime(item.Time) },
            SPAN " | ",
            A { href = hnBaseURL .. "/item?id=" .. item.ID, "source" },
            SPAN " | ",
            "HN request time: ",
            (timeEnd - timeStart) / 1000,
            "s",
        },
        BR,
    }
end

local function renderItem(item, num)
    return DIV {
        id = "item-" .. item.ID,
        class = 'post' .. (item.Dead and ' dead' or ''),
        DIV {
            class = "post-header",
            num and {
                SPAN {class="comment-num", num + 1, ". "},
            },
            A {
                class = "username",
                href = hnBaseURL .. "/user?id=" .. item.By,
                title = "HN username: " .. item.By,
                SPAN { class = "alias", "slacker" .. userAlias[item.By] },
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
            A { href = hnBaseURL .. "/item?id=" .. item.ID, "[source]" },
            SPAN { class = "triangle", "▶" },
            Xt.map(kidIDs[tostring(item.ID)] or {}, function(childID)
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
        item.Parent and DIV {
            A {
                class = "post-link parent", href = "#item-" .. item.Parent,
                --data_id = item.ID,
                ">>" .. item.Parent, (op and item.Parent == op.ID and op.type == "story" and " (OP)" or nil)
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
    }
end

local commentNum = 0
for i, item in items() do
    local node


    if i == 1 and item.Type == "story" then
        node = renderOP(item, i)
    elseif not deadPosts[tostring(item.ID)] then
        node = renderItem(item, commentNum)
        commentNum = commentNum+1
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
            INPUT { name = "id", value = op.ID, placeholder = "HN item ID or url" },
        },
        BUTTON { "GO" }
    },

    commentCount > 5 and DIV {
        id = "top",
        class = "thread-nav",
        SPAN { "[", A { href = "#bottom", "bottom" }, "]" },
    },
})



return LAYOUT {
    title = op and op.title,
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

        CSS "#thread-header" {
            display = "flex",
            justify_content = "space-between",
        },
        CSS ".thread-nav" {
            text_align = "right",
            width = "100%",
            CSS "> *" {
                margin_right = 5
            }
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
            padding = 2,
            font_size = ".7rem",
            margin_top = 20,
            text_align = "right",
            position = "fixed",
            bottom = 0,
            right = 0,
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
            color="#777",
            font_size=".5em",
            position="absolute",
            top="-0px",
            right="-0px",
        },
    },


    commentCount > 10 and DIV {
        id = "top-commenters",
        "Most active commenters",
        UL {
            Xt.map(topCommenters, function(entry)
                return LI { "slacker" .. userAlias[entry[1]], "(", entry[2], ")" }
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

    commentCount == 0 and DIV {
        BR,
        EM "(no comments)",
    },


    commentCount > 5 and DIV {
        id = "bottom",
        class = "thread-nav",
        SPAN { "[", A { href = "#item-" .. op.ID, "top" }, "]" },
    },

    --DIV {
    --    id = "floating-nav",
    --    A { href = "#bottom", "←back" },
    --    SPAN " ",
    --    A { href = "#bottom", "forward→" },
    --},

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

    SCRIPT { src = "item.js" },
}
