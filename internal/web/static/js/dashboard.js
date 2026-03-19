document.addEventListener('DOMContentLoaded', function() {
    initNodeSelector().then(function() {
        if (getSelectedNode()) {
            fetchDashboard();
        }
    });
});

function handleRefresh() {
    fetchDashboard();
}

function fetchDashboard() {
    var nodeId = getSelectedNode();
    if (!nodeId) {
        showToast('Please select a node', 'error');
        return;
    }

    // Fetch chain info
    api.get('/api/v1/' + nodeId + '/chain').then(function(resp) {
        renderChainInfo(resp.data);
    }).catch(function(err) {
        document.getElementById('chain-info-content').innerHTML =
            '<p class="text-red-400">Error: ' + err.message + '</p>';
    });

    // Fetch mempool
    api.get('/api/v1/' + nodeId + '/mempool').then(function(resp) {
        renderMempoolSummary(resp.data);
    }).catch(function(err) {
        document.getElementById('mempool-summary-content').innerHTML =
            '<p class="text-red-400">Error: ' + err.message + '</p>';
    });

    // Fetch recent blocks
    api.get('/api/v1/' + nodeId + '/blocks/recent?count=5').then(function(resp) {
        renderRecentBlocks(resp.data);
    }).catch(function(err) {
        document.getElementById('recent-blocks-body').innerHTML =
            '<tr><td colspan="5" class="py-4 text-center text-red-400">Error: ' + err.message + '</td></tr>';
    });
}

function renderChainInfo(info) {
    var html = '';
    html += '<div class="flex justify-between"><span class="text-slate-400">Chain</span><span class="font-medium">' + (info.chain || 'N/A') + '</span></div>';
    html += '<div class="flex justify-between"><span class="text-slate-400">Height</span><span class="stat-value text-base">' + formatNumber(info.blocks) + '</span></div>';
    html += '<div class="flex justify-between"><span class="text-slate-400">Difficulty</span><span>' + formatDifficulty(info.difficulty) + '</span></div>';
    html += '<div><span class="text-slate-400">Best Block</span><br><span class="hash-text clickable-hash" onclick="window.location.href=\'/block/' + info.bestblockhash + '\'">' + truncateHash(info.bestblockhash, 16) + '</span></div>';
    if (info.warnings) {
        html += '<div class="text-yellow-400 text-sm">' + info.warnings + '</div>';
    }
    document.getElementById('chain-info-content').innerHTML = html;
}

function renderMempoolSummary(info) {
    var html = '';
    html += '<div class="flex justify-between"><span class="text-slate-400">Transactions</span><span class="stat-value text-base">' + formatNumber(info.size) + '</span></div>';
    html += '<div class="flex justify-between"><span class="text-slate-400">Memory Usage</span><span>' + formatBytes(info.bytes) + '</span></div>';
    if (info.total_fee !== undefined) {
        html += '<div class="flex justify-between"><span class="text-slate-400">Total Fee</span><span>' + formatBTC(info.total_fee) + '</span></div>';
    }
    if (info.mempoolminfee !== undefined) {
        html += '<div class="flex justify-between"><span class="text-slate-400">Min Fee</span><span>' + formatBTC(info.mempoolminfee) + '</span></div>';
    }
    document.getElementById('mempool-summary-content').innerHTML = html;
}

function renderRecentBlocks(blocks) {
    if (!blocks || blocks.length === 0) {
        document.getElementById('recent-blocks-body').innerHTML =
            '<tr><td colspan="5" class="py-4 text-center text-slate-500">No blocks found.</td></tr>';
        return;
    }

    var html = '';
    blocks.forEach(function(block) {
        html += '<tr class="hover:bg-slate-800/30 cursor-pointer" onclick="window.location.href=\'/block/' + block.hash + '\'">';
        html += '<td class="py-2 px-3 font-medium text-amber-400">' + formatNumber(block.height) + '</td>';
        html += '<td class="py-2 px-3 hash-text">' + truncateHash(block.hash, 10) + '</td>';
        html += '<td class="py-2 px-3">' + formatTimeAgo(block.time) + '</td>';
        html += '<td class="py-2 px-3 text-right">' + block.nTx + '</td>';
        html += '<td class="py-2 px-3 text-right">' + formatBytes(block.size) + '</td>';
        html += '</tr>';
    });
    document.getElementById('recent-blocks-body').innerHTML = html;
}
