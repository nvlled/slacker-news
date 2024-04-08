let selectedPost;
const posts = {};
const isMobile = detectMobile()

window.onhashchange = highlightAnchoredPost;


function formatRFC8222(date) {
    return date.toUTCString()
}

function highlightAnchoredPost() {
    if (!window.location.hash) return;
    const node = document.querySelector(window.location.hash)
    if (node && node.classList.contains("post")) {
        node.classList.add("selected")
        selectedPost = node;
    }
}

function highlightPost(node) {
    if (node && node.classList.contains("post")) {
        node.classList.add("selected")
        selectedPost = node;
    }
}

function removeHighlightedPost() {
    if (selectedPost) {
        selectedPost.classList.remove("selected")
        selectedPost = null
    }
}

function isVisibleInViewport(node) {
    const rect = node.getBoundingClientRect();
    const winH = window.innerHeight;

    if (rect.top + rect.height < 0) {
        return false;
    }
    if (rect.top + rect.height > winH) {
        return false;
    }

    return true;

}


// `navigator.userAgentData.mobile` doesn't work on my phone so...
function detectMobile() {
    // Source: https://stackoverflow.com/a/11381730
    const toMatch = [
        /Android/i,
        /webOS/i,
        /iPhone/i,
        /iPad/i,
        /iPod/i,
        /BlackBerry/i,
        /Windows Phone/i
    ];

    return toMatch.some((toMatchItem) => {
        return navigator.userAgent.match(toMatchItem);
    });
}

let popup;
function setupLink(link) {
    if (link.getAttribute("initialized")) return;
    link.setAttribute("initialized", "1");

    const m = link.href.match(/#.*$/)
    let idFrag = m && m[0];

    if (!idFrag) {
        idFrag = "#item-" + new URL(link.href).searchParams.get("id");
    }

    if (isMobile) {
        link.onclick = function(e) { e.preventDefault(); }
        link.parentNode.classList.add("m");
    }


    link.onmouseover = function() {
        removeHighlightedPost();
        if (popup) return;

        const winW = window.innerWidth;
        const winH = window.innerHeight;

        if (idFrag) {
            const post = document.querySelector(idFrag);
            highlightPost(post);

            if (isVisibleInViewport(post) && !isMobile) {
                return;
            }

            if (post) {
                popup = post.cloneNode(true)
                popup.classList.add("popup");
                popup.classList.remove("selected");

                popup.style.left = 0;
                popup.style.top = 0;
                document.querySelector("#thread").appendChild(popup)

                const linkRect = link.getBoundingClientRect();
                const postRect = popup.getBoundingClientRect();
                const topOffset = (linkRect.top + linkRect.height / 2) / winH
                const leftOffset = (linkRect.left + linkRect.height / 2) / winW

                popup.style.width = postRect.width + "px"; // the fix

                let left = 0;
                let top = 0;

                if (leftOffset >= 0.5) {
                    left = scrollX + linkRect.right - postRect.width;
                    if (left < 0) {
                        left = 0;
                    }
                } else {
                    left = scrollX + linkRect.left
                }

                if (left + postRect.width > winW) {
                    left = winW - postRect.width;
                }


                if (topOffset > 0.5) {
                    top = scrollY + Math.floor(linkRect.top) - Math.ceil(postRect.height) - 1;
                } else {
                    top = scrollY + linkRect.bottom + 1;
                }

                if (top < scrollY) {
                    top = scrollY
                    popup.style.height = ((scrollY + linkRect.top) - scrollY - linkRect.height) + "px"
                }


                popup.style.left = left + "px"
                popup.style.top = top + "px"
                popup.id = "";
            }
        }
    }

    link.onmouseout = function() {
        removeHighlightedPost();
        if (popup) {
            popup.remove();
            popup = null;
        }
    }
}

function setupPopups(node) {
    for (const link of node.querySelectorAll("a.post-link")) {
        setupLink(link);
    }
}

function setupDate(node) {
    const date = node.querySelector(".post-datetime")
    if (date) {
        date.textContent = new Date(date.textContent).toLocaleString()
    }
}

function setupExternalLinks(post) {
    for (const node of post.querySelectorAll(".post-body a")) {
        if (!node.classList.contains("source") && node.href.includes("news.ycombinator.com")) {
            const url = new URL(node.href);
            const id = url.searchParams.get("id");
            if (url.pathname === "/item" && id) {
                node.textContent = ">>" + id + " →";
                if (posts[id]) {
                    node.textContent = ">>" + id;
                    node.classList.add("post-link");
                    node.href = location.origin + location.pathname + location.search + "#item-" + id;
                    setupLink(node);
                } else {
                    node.href = location.origin + "/item?id=" + id
                }
            }
        }
    }
}

function setupGreenTexts(post) {
    const postBody = post.querySelector(".post-body");
    if (!postBody) return;
    for (let node of postBody.childNodes) {
        if (node.tagName !== "P" && node.tagName !== "PRE") {
            const p = document.createElement("p");
            const nodes = [];
            let current = node;
            while (current) {
                if (current.tagName === "P" || current.tagName == "PRE") {
                    break;
                }

                nodes.push(current);
                current = current.nextSibling;
            }
            if (nodes.length > 0) {
                postBody.replaceChild(p, node);
                for (const elem of nodes) {
                    p.appendChild(elem);
                }
                node = p;
            }
        }

        const firstChild = node.childNodes[0];
        if (!firstChild) continue;

        if ((firstChild.nodeType == Element.TEXT_NODE && firstChild.textContent[0] == ">")
            || node.childNodes.length === 1 && firstChild.tagName == "I") {
            node.classList.add("green-text");
        }
    }

}

const unfinishedNodes = {};

function setupNode(node) {
    if (!node.querySelector("#post-footer")) {
        // node hasn't finished rendering, go back to it later
        unfinishedNodes[node.id] = node;
        return;
    }
    delete unfinishedNodes[node.id];

    if (node.id) {
        if (node.id.startsWith("item-"))
            posts[node.id.slice(5)] = true;
        else
            posts[node.id] = true;
    }

    setupDate(node)
    setupPopups(node);
    setupGreenTexts(node);
    setupExternalLinks(node);
}

const observer = new MutationObserver(function(mutations) {
    mutations.forEach(function(mutation) {
        for (let i = 0; i < mutation.addedNodes.length; i++) {
            let node = mutation.addedNodes[i];

            if (node.nodeType != Element.ELEMENT_NODE) continue;

            if (node.classList.contains("post") && !node.classList.contains("popup")) {
                setupNode(node);
            } else if (node.classList.contains("post-link")) {
                setupLink(node);
            }
        }
    })
});

observer.observe(document.body, {
    childList: true,
    subtree: true,
});

window.onload = function() {
    for (const node of Object.values(unfinishedNodes)) {
        setupNode(node);
    }
}
