document.addEventListener('DOMContentLoaded', function() {
    initNodeSelector().then(function() {
        if (getSelectedNode()) {
            fetchMempool();
        }
    });
});

function handleRefresh() {
    fetchMempool();
}

function fetchMempool() {
    var nodeId = getSelectedNode();
    if (!nodeId) {
        showToast('Please select a node', 'error');
        return;
    }

    api.get('/api/v1/' + nodeId + '/mempool').then(function(resp) {
        renderMempoolStats(resp.data);
    }).catch(function(err) {
        document.getElementById('mempool-stats').innerHTML =
            '<div class="card p-5"><p class="text-red-400">Error: ' + err.message + '</p></div>';
    });
}

function renderMempoolStats(info) {
    var container = document.getElementById('mempool-stats');
    if (!info) {
        container.innerHTML = '<div class="card p-5"><p class="text-slate-500">No mempool data.</p></div>';
        return;
    }

    var labels = {
        'loaded': 'Loaded',
        'size': 'TX Count',
        'bytes': 'Size (bytes)',
        'usage': 'Memory Usage',
        'total_fee': 'Total Fee',
        'maxmempool': 'Max Mempool',
        'mempoolminfee': 'Min Fee Rate',
        'minrelaytxfee': 'Min Relay Fee',
        'incrementalrelayfee': 'Incremental Fee',
        'unbroadcastcount': 'Unbroadcast Count',
        'fullrbf': 'Full RBF'
    };

    var html = '';
    var keys = Object.keys(info);
    keys.forEach(function(key) {
        var label = labels[key] || key;
        var value = info[key];
        var displayValue = '';

        if (key === 'bytes' || key === 'usage' || key === 'maxmempool') {
            displayValue = formatBytes(value);
        } else if (key === 'total_fee' || key === 'mempoolminfee' || key === 'minrelaytxfee' || key === 'incrementalrelayfee') {
            displayValue = formatBTC(value);
        } else if (typeof value === 'boolean') {
            displayValue = value ? 'Yes' : 'No';
        } else if (typeof value === 'number') {
            displayValue = formatNumber(value);
        } else {
            displayValue = String(value);
        }

        html += '<div class="card p-4 text-center">';
        html += '<div class="text-slate-400 text-sm mb-1">' + label + '</div>';
        html += '<div class="stat-value text-base">' + displayValue + '</div>';
        html += '</div>';
    });

    container.innerHTML = html;
}
