local siteName = "slacker news"

return function(args)
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
            font_size = 20,
            padding = 10,
            display = "inline-block",
            border = "2px solid white",
        },
        CSS "#site-name" {
            display = "inline-block",
            font_weight = "800",
            font_size = 28,
        },
        CSS "#site-menu, #account-info" {
            display = "flex",
            align_items = "center",
            CSS 'li' {
                {
                    margin = 0,
                    list_style_type = "none",
                    border_right = "1px solid white",
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
                "X"
            },
            SPAN { id = "site-name", siteName },
            menu,
            account,
        },
    }

    local body = DIV {
        id = "wrapper",
        navigation,
        HR,
        DIV {
            children
        },
    }

    return HTML {
        HEAD {
            TITLE(props.title or "*"),
            props.style and STYLE(props.style),
            STYLE {
                CSS "body" {
                    background = "#111",
                    color = "#eee",
                    font_size = 24,
                },
                CSS "#wrapper" {
                    max_width = 1024,
                    margin = "auto",
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
