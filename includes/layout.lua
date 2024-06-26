local siteName = "slacker news"

local feed     = form:Get("feed") or ""
local reqpath  = request.URL.Path
local isHome   = reqpath == "/" or reqpath == "/index"

local style    = {
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
            CSS "#site-nav-spacing" {
                flex_grow="1",
            }
        },
        CSS "#site-logo" {
            padding = "5px 10px",
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
        CSS "#site-menu-container" {
            display = "flex",
            justify_content = "center",
        },
        CSS "#site-menu" {
            margin = 0,
        },
        CSS "#site-menu" {
            ["@media"] = "(max-width: 1000px)",
            margin = 0,
            margin_bottom = "1.6em",
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

            CSS "a.selected" {
                color = "orange",
            }

        },
        CSS "html" {
            font_size = "100%",
            overflow_x="hidden",
        },
        CSS "body" {
            overflow_x="hidden",
            background = style.bgColor,
            color = style.textColor,
            text_size_adjust = "none",
        },
        CSS "#wrapper" {
            margin = "auto",
            width = "100%",
            font_size = "100%",
            max_width = "45em !important",
        },

        CSS_MEDIA '(max-width: 900px)' {
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

    })

    local menu = DIV {
        id = "site-menu-container",
        UL {
            id = "site-menu",

            LI ^ A { class = ((isHome and feed == "") or feed == "top") and "selected" or "", href = "/", "/top/" },
            LI ^ A { class = feed == "new" and "selected" or "", href = "/?feed=new", "/new/" },
            LI ^ A { class = feed == "best" and "selected" or "", href = "/?feed=best", "/best/" },
            LI ^ A { class = feed == "ask" and "selected" or "", href = "/?feed=ask", "/ask/" },
            LI ^ A { class = feed == "show" and "selected" or "", href = "/?feed=show", "/show/" },
            LI ^ A { class = feed == "job" and "selected" or "", href = "/?feed=job", "/job/" },
        }
    }

    local account = DIV {
        DIV {
            id = "account-info",
            not user and LI ^ A { href = "#/login", "login" }
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
                A { href = "/", "^" }
            },
            A { id = "site-name", href = "/", siteName },
            DIV { id = "site-nav-spacing" },
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
            menu,
            navigation,
            contents,
            BR,
            BR,
        },
    }
end


return setmetatable({
    layout = LAYOUT,
    style = style,

}, {
    __call = function(self, args) return LAYOUT(args) end
})
