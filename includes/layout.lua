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
                    --border_right = "1px solid #484a4e",
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

        },
        CSS "html" {
            font_size = "100%",
        },
        CSS "body" {
            background = style.bgColor,
            color = style.textColor,
            text_size_adjust = "none",
        },
        CSS "#wrapper" {
            margin = "auto",
            width = "100%",
            font_size = "100%",
                max_width = "720px !important",
        },

        CSS_MEDIA '(max-width: 1000px)' {
            CSS "#wrapper" {
                max_width = "unset !important",
            }
        },

        CSS_MEDIA '(orientation: portrait)' {
            CSS "html" {
                font_size = "200%",
            },
            CSS "#wrapper" {
                max_width = "unset !important",
            },
            CSS "#site-name, #site-menu" {
                display="none"
            },
        },

        CSS_MEDIA '(orientation: landscape)' {
            CSS "html" {
                font_size = "100%",
            },
            CSS "#wrapper" {
                width = "100%",
                margin = "auto",
            },
        },

        CSS "a" {
            color = style.linkColor,
        },

        CSS "#footer-notice" {
            padding = 2,
            font_size = ".7rem",
            width = "100%",
            margin_top = 20,
            text_align = "right",
            position = "fixed",
            bottom = 0,
            right = 0,
            CSS "i" { background = style.bgColor },
        },
    })

    local menu = UL {
        id = "site-menu",

        LI ^ A { href = "/", "/top/" },
        LI ^ A { href = "#/new", "/new/" },
        LI ^ A { href = "#/best", "/best/" },
        LI ^ A { href = "#/ask", "/ask/" },
        LI ^ A { href = "#/show", "/show/" },
        LI ^ A { href = "#/job", "/job/" },
    }

    local account = DIV {
        DIV {
            id = "account-info",
            not user and LI ^ A { href = "/login", "login" }
            or FRAGMENT {
                LI ^ A { href = "/user?id=" .. user.ID, user.Username },
                LI ^ A { href = "logout", "logout" },
            },
            LI ^ A { href = "/about", "about" },
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

    local contents = DIV {
        id = "wrapper",
        DIV {
            children
        },
    }

    return HTML {
        HEAD {
            TITLE((not Xt.isEmptyString(props.title) and props.title .. " | " or "") .. siteName),
            props.style and STYLE(props.style),
            config.DevMode and not props.noAutoReload and SCRIPT [[
(function() {
var disconnected = false;
var evtSource = new EventSource("/.autoreload");
evtSource.addEventListener("fsevent", function(event) {
    window.location.reload();
});
evtSource.onopen = function() {
    console.log("open", {disconnected});
    if (disconnected) {
        window.location.reload();
    }
}
evtSource.onerror = function() {
    disconnected = true;
}

console.log({evtSource});
window.addEventListener("unload", function() { evtSource.close(); })
})();
            ]]
        },

        BODY {
            navigation,
            contents,
            DIV {
                id = "footer-notice",
                SMALL ^ I {
                    "NOTE: showing cached content from ",
                    math.floor(os.difftime(os.time(), cm.LastUpdate) / 60),
                    " minutes ago, next update will be in ",
                    math.floor(os.difftime(cm:GetNextUpdate(), os.time()) / 60),
                    " minutes",
                }
            }
        },

    }
end


return setmetatable({
    layout = LAYOUT,
    style = style,

}, {
    __call = function(self, args) return LAYOUT(args) end
})
