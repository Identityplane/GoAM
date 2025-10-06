export function initNodeHistory() {

    console.log('initializing NodeHistory')
    
    // Get the current node name from the main content
    // <div class="main-content" data-node="{{ if .Node }}{{ .Node.Use }}{{ end }}" data-node-current="{{ .NodeName }}">
    const nodeCurrent = document.querySelector('.main-content')
    if (!nodeCurrent) {
        console.log('no main content found')
        return;
    }

    const loginuri = nodeCurrent.dataset.loginuri;
    if (!loginuri) {
        console.log('no loginuri found')
        return;
    }

    // Push the current node name to the history
    window.history.pushState({}, "", loginuri);

    // Force reload when navigating back
    window.addEventListener('popstate', function(event) {
        location.reload(true); // Forces a hard reload from the server, bypassing cache
    });

}