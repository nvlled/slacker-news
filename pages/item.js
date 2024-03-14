let selectedPost;

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
        selectedPost.classList.remove("selected");
        selectedPost = null;
    }
}

function isVisibleInViewport(node) {
    const rect = node.getBoundingClientRect();
    //const winW = window.innerWidth;
    const winH = window.innerHeight;

    if (rect.top + rect.height * 0.20 < 0) {
        return false;
    }
    if (rect.top + rect.height * 0.30 > winH) {
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

function setupPopups() {
    const isMobile = detectMobile()

    for (const link of document.querySelectorAll("a.post-link")) {
        removeHighlightedPost();

        const m = link.href.match(/#.*$/)
        const idFrag = m && m[0];

        if ((isMobile)) {
            link.onclick = function(e) { e.preventDefault(); }
            link.parentNode.classList.add("m");
        }

        let popup;

        link.onmouseover = function() {
            if (popup) return;

            const winW = window.innerWidth;
            const winH = window.innerHeight;

            if (idFrag) {
                const post = document.querySelector(idFrag);
                highlightPost(post);
                if (isVisibleInViewport(post)) {
                    return;
                }

                if (post) {
                    popup = post.cloneNode(true)
                    popup.classList.add("popup");
                    popup.classList.remove("selected");
                    document.querySelector("#thread").appendChild(popup)

                    const linkRect = link.getBoundingClientRect();
                    const postRect = popup.getBoundingClientRect();
                    const topOffset = (linkRect.top + linkRect.height / 2) / winH
                    const leftOffset = (linkRect.left + linkRect.height / 2) / winW

                    let left = 0;
                    let top = 0;

                    if (leftOffset > 0.5) {
                        left = scrollX + linkRect.right - postRect.width;
                        if (left + postRect.width >= winW) {
                        }
                        if (left < 0) {
                            left = 0;
                        }
                    } else {
                        left = scrollX + linkRect.left
                    }


                    if (topOffset > 0.5) {
                        top = scrollY + linkRect.top - postRect.height
                    } else {
                        top = scrollY + linkRect.bottom
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
}

function setupDates() {
    for (const elem of document.querySelectorAll(".post-datetime")) {
        elem.textContent = new Date(elem.textContent).toLocaleString()
    }
}

window.onhashchange = highlightAnchoredPost;

window.onload = function() {
    highlightAnchoredPost();
    setupPopups();
    setupDates();
}

