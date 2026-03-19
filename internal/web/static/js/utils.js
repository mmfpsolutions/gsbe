function getSelectedNode() {
    return localStorage.getItem('gsbe_selected_node') || '';
}

function setSelectedNode(nodeId) {
    localStorage.setItem('gsbe_selected_node', nodeId);
}

function truncateHash(hash, len) {
    if (!hash) return '';
    len = len || 8;
    if (hash.length <= len * 2) return hash;
    return hash.substring(0, len) + '...' + hash.substring(hash.length - len);
}

function formatTimestamp(unix) {
    if (!unix) return 'N/A';
    return new Date(unix * 1000).toLocaleString();
}

function formatTimeAgo(unix) {
    if (!unix) return 'N/A';
    var seconds = Math.floor(Date.now() / 1000 - unix);
    if (seconds < 0) return 'just now';
    if (seconds < 60) return seconds + 's ago';
    if (seconds < 3600) return Math.floor(seconds / 60) + 'm ago';
    if (seconds < 86400) return Math.floor(seconds / 3600) + 'h ago';
    return Math.floor(seconds / 86400) + 'd ago';
}

function formatNumber(n) {
    if (n === null || n === undefined) return 'N/A';
    return Number(n).toLocaleString();
}

function formatBytes(bytes) {
    if (!bytes && bytes !== 0) return 'N/A';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + ' MB';
    return (bytes / 1073741824).toFixed(2) + ' GB';
}

function formatDifficulty(d) {
    if (!d && d !== 0) return 'N/A';
    var num = Number(d);
    if (num >= 1e12) return (num / 1e12).toFixed(2) + ' T';
    if (num >= 1e9) return (num / 1e9).toFixed(2) + ' G';
    if (num >= 1e6) return (num / 1e6).toFixed(2) + ' M';
    if (num >= 1e3) return (num / 1e3).toFixed(2) + ' K';
    return num.toFixed(2);
}

function formatBTC(value) {
    if (value === null || value === undefined) return 'N/A';
    return Number(value).toFixed(8);
}

function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(function() {
        showToast('Copied to clipboard', 'success');
    }).catch(function() {
        showToast('Failed to copy', 'error');
    });
}

function showToast(message, type) {
    var container = document.getElementById('toast-container');
    if (!container) return;

    var toast = document.createElement('div');
    toast.className = 'card px-4 py-2 text-sm';
    if (type === 'error') {
        toast.style.borderColor = 'rgba(239,68,68,0.5)';
        toast.style.color = '#ef4444';
    } else if (type === 'success') {
        toast.style.borderColor = 'rgba(34,197,94,0.5)';
        toast.style.color = '#22c55e';
    } else {
        toast.style.color = '#60a5fa';
    }
    toast.textContent = message;
    container.appendChild(toast);

    setTimeout(function() {
        toast.remove();
    }, 3000);
}

function initNodeSelector() {
    var selector = document.getElementById('node-selector');
    if (!selector) return Promise.resolve();

    return api.get('/api/v1/nodes').then(function(resp) {
        var nodes = resp.data || [];
        selector.innerHTML = '<option value="">Select Node...</option>';
        nodes.forEach(function(node) {
            var opt = document.createElement('option');
            opt.value = node.id;
            opt.textContent = node.name + ' (' + node.symbol + ')' + (node.online ? '' : ' [offline]');
            selector.appendChild(opt);
        });

        var saved = getSelectedNode();
        if (saved && nodes.some(function(n) { return n.id === saved; })) {
            selector.value = saved;
        } else if (nodes.length > 0) {
            selector.value = nodes[0].id;
            setSelectedNode(nodes[0].id);
        }

        selector.addEventListener('change', function() {
            setSelectedNode(this.value);
            if (typeof handleRefresh === 'function') {
                handleRefresh();
            }
        });
    }).catch(function(err) {
        console.error('Failed to load nodes:', err);
    });
}

function handleSearch(event) {
    event.preventDefault();
    var query = document.getElementById('search-input').value.trim();
    var nodeId = getSelectedNode();
    if (!query || !nodeId) {
        showToast('Enter a search query and select a node', 'error');
        return;
    }

    api.get('/api/v1/' + nodeId + '/search?q=' + encodeURIComponent(query)).then(function(resp) {
        var result = resp.data;
        if (result.type === 'block') {
            window.location.href = '/block/' + result.hash;
        }
    }).catch(function(err) {
        showToast('Search failed: ' + err.message, 'error');
    });
}

function hexToAscii(hex) {
    if (!hex) return '';
    var ascii = '';
    for (var i = 0; i < hex.length; i += 2) {
        var code = parseInt(hex.substr(i, 2), 16);
        if (code >= 32 && code <= 126) {
            ascii += String.fromCharCode(code);
        }
    }
    return ascii;
}
