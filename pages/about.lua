local LAYOUT = require "layout"

return LAYOUT {
    H2 "What's this?",
    PP [[
        This is an alternative frontend for ]] / A { href = "https://news.ycombinator.com", "Hacker News" } / [[,
        with *chan like user-interface.    
    ]],
    H2 "Why is it slow?",
    PP [[
        The primary slowness comes from the recursive HTTP requests for each HN comment.
        So a thread with a thousand comments will also require a thousand requests. Ideally,
        it could be done in a one or two request, but the official HN API offers no such feature
        for now.

        I do some caching so that subsequent page load will be significantly faster.

        The alternative is to paginate, and defer other page request later, but
        this isn't an option since the HN API has a different comment sorting,
        and I want to sort the comments by date.

        The other weird thing I did is stream the HTTP page response. This results
        in longer overall page load time, but faster initial page load time.
        That is to say, I could start browsing the page even while
        the page is still downloading.
    ]],
}
