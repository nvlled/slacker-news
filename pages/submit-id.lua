local LAYOUT = require "layout"

local inputID = form:Get("id")
local id = tonumber(inputID)

if inputID and Xt.isEmptyString(id) then
    local idStr = string.match(inputID, "item%?id=(%d+)")
    print("idStr", idStr)
    if idStr then
        id = tonumber(idStr)
    end
end

if id then
    go:Redirect("/item?id=" .. id)
    return nil
end

return LAYOUT {
    not id and EM {
        "invalid ID or URL",
    },


    BR,
    BR,

    FORM {
        action = "submit-id",
        id = "thread-id-input",
        LABEL {
            "ID: ",
            INPUT { name = "id", value = "", placeholder = "HN item ID or url" },
        },
        BUTTON { "GO" }
    },
}
