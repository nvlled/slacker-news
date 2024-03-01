let selectedPost;

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
    //if (!window.location.hash) return;
    //for (const node of document.querySelectorAll(".post.selected")) {
    //    node.classList.remove("selected")
    //}
    //window.location.hash = "";
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

    console.log(rect)
}

window.onhashchange = highlightAnchoredPost;

window.onload = function() {
    highlightAnchoredPost();

    for (const link of document.querySelectorAll("a.post-link")) {
        removeHighlightedPost();

        const m = link.href.match(/#.*$/)
        const idFrag = m && m[0];

        let popup;
        link.onmouseover = function() {
            if (popup) return;

            const winW = window.innerWidth;
            const winH = window.innerHeight;


            if (winW < 1000) return;

            if (idFrag) {
                const post = document.querySelector(idFrag);
                highlightPost(post);
                if (isVisibleInViewport(post)) {
                    console.log("visible", post);
                    return;
                }

                console.log("not visible");

                if (post) {
                    popup = post.cloneNode(true)
                    popup.classList.add("popup");
                    popup.classList.remove("selected");

                    if (link.classList.contains("reply")) {
                        link.parentNode.parentNode.appendChild(popup);
                    } else {
                        link.parentNode.appendChild(popup);
                    }

                    const linkRect = link.getBoundingClientRect();
                    const postRect = popup.getBoundingClientRect();


                    const topOffset = (linkRect.top+linkRect.height/2) / winH
                    const leftOffset = (linkRect.left+linkRect.height/2) / winW


                    let left = 0;
                    let top = 0;

                    if (leftOffset > 0.5) {
                        left = (linkRect.x - postRect.width - 2)
                    } else {
                        left = (linkRect.left + linkRect.width + 2)
                    }

                    if (left < 0) {
                        left = 0;
                    }
                    if (left + postRect.width > winW) {
                        left = winW - postRect.width + 2;
                    }

                    if (topOffset > 0.5) {
                        top = linkRect.top - postRect.height;
                    } else {
                        top = linkRect.bottom
                    }

                    if (top + postRect.height >= winH) {
                        top = winH - postRect.height;
                    }
                    if (top < 0) {
                        top = 0;
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
                console.log(popup)
                popup.remove();
                popup = null;
            }
        }
    }
}

