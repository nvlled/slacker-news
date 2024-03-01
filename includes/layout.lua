local siteName = "slacker news"

local style = {
    bgColor = "#1d1f21 none",
    textColor = "#c5c8c6",
    linkColor = "#5f89ac",
}

local function LAYOUT(args)
    local props, children = GetComponentArgs(args)
    if not props.style then props.style = {} end

    local user = go:GetCurrentUser()

    table.insert(props.style, CSS {
        CSS "#site-nav" {
            display = "flex",
            align_items = "center",
            column_gap = 15,
            CSS "#site-menu" {
                flex_grow = "1",
            },
        },
        CSS "#site-logo" {
            font_size = "1.4rem",
            padding = 10,
            display = "inline-block",
            border = "3px solid #484a4e",
            CSS "a" {
                color = "inherit",
                text_decoration = "none",
            },
        },
        CSS "#site-name" {
            display = "inline-block",
            font_weight = "800",
            font_size = "1.4rem",
            text_decoration = "none",
            color = "inherit",
        },
        CSS "#site-menu, #account-info" {
            display = "flex",
            align_items = "center",
            CSS 'li' {
                {
                    margin = 0,
                    list_style_type = "none",
                    border_right = "1px solid #484a4e",
                    padding = "0 10px",
                    border_collapse = "collapse",
                },
                [":last-child"] = {
                    border_right = "0",
                },
            },
            CSS 'a' {
                text_decoration = "none",
                color = "white",
            },

        }
    })

    local menu = UL {
        id = "site-menu",
        LI ^ A { href = "/new", "new" },
        LI ^ A { href = "/submit", "submit" },
    }

    local account = DIV {
        not user and A { href = "/login", "login" } or DIV {
            id = "account-info",
            LI ^ A { href = "/user?id=" .. user.ID, user.Username },
            LI ^ A { href = "logout", "logout" },
        }
    }

    local navigation = FRAGMENT {
        DIV {
            id = "site-nav",
            DIV {
                id = "site-logo",
                A { href = "/", "𐲤" }
            },
            A { id = "site-name", href = "/", siteName },
            menu,
            account,
        },
    }

    local body = DIV {
        id = "wrapper",
        navigation,
        DIV {
            children
        },
    }

    return HTML {
        HEAD {
            TITLE((props.title and props.title .. " | " or "") .. siteName),
            props.style and STYLE(props.style),
            STYLE {
                CSS "html" {
                    font_size = "100%",
                },
                CSS "body" {
                    background = style.bgColor,
                    color = style.textColor,
                },
                CSS "#wrapper" {
                    margin = "auto",
                },
                CSS_MEDIA '(width >= 1000px) or (orientation: landscape)' {
                    CSS "html" {
                        font_size = "120%",
                    },
                    CSS "#wrapper" {
                        max_width = "1000px",
                        width = "100%",
                        margin = "auto",
                    },
                    CSS "#site-menu" {
                        display = "none"
                    },
                },
                CSS "a" {
                    color = style.linkColor,
                },
            },
            not props.noAutoReload and SCRIPT [[
            (function() {
                var evtSource = new EventSource("/.autoreload");
                evtSource.addEventListener("fsevent", function(event) {
                    //console.log("reload");
                    window.location.reload();
                });
                window.addEventListener("unload", function() { evtSource.close(); })
            })();
            ]]
        },
        BODY(body),

    }
end


return setmetatable({
    layout = LAYOUT,
    style = style,

}, {
    __call = function(self, args) return LAYOUT(args) end
})
