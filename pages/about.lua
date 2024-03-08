local LAYOUT = require "layout"

return LAYOUT {
    H2 "What's this?",
    PP [[
        This is an alternative frontend for ]] / A { href = "https://news.ycombinator.com", "Hacker News" } / [[,
        with *chan like user-interface. As of now, it's still very much a work-in-progress, but quite useable
        for me.
    ]]
}
