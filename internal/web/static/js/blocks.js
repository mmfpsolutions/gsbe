var blocksData = [];
var blocksCount = 20;

document.addEventListener('DOMContentLoaded', function() {
    initNodeSelector().then(function() {
        if (getSelectedNode()) {
            fetchBlocks();
        }
    });
});

function handleRefresh() {
    blocksData = [];
    fetchBlocks();
}

function fetchBlocks() {
    var nodeId = getSelectedNode();
    if (!nodeId) {
        showToast('Please select a node', 'error');
        return;
    }

    api.get('/api/v1/' + nodeId + '/blocks/recent?count=' + blocksCount).then(function(resp) {
        blocksData = resp.data || [];
        renderBlocksTable();
        var loadMoreBtn = document.getElementById('load-more-btn');
        if (loadMoreBtn) loadMoreBtn.style.display = 'inline-block';
    }).catch(function(err) {
        document.getElementById('blocks-table-body').innerHTML =
            '<tr><td colspan="6" class="py-4 text-center text-red-400">Error: ' + err.message + '</td></tr>';
    });
}

function renderBlocksTable() {
    if (!blocksData || blocksData.length === 0) {
        document.getElementById('blocks-table-body').innerHTML =
            '<tr><td colspan="6" class="py-4 text-center text-slate-500">No blocks found.</td></tr>';
        return;
    }

    var html = '';
    blocksData.forEach(function(block) {
        html += '<tr class="hover:bg-slate-800/30 cursor-pointer" onclick="window.location.href=\'/block/' + block.hash + '\'">';
        html += '<td class="py-2 px-3 font-medium text-amber-400">' + formatNumber(block.height) + '</td>';
        html += '<td class="py-2 px-3 hash-text">' + truncateHash(block.hash, 10) + '</td>';
        html += '<td class="py-2 px-3">' + formatTimestamp(block.time) + '</td>';
        html += '<td class="py-2 px-3 text-right">' + block.nTx + '</td>';
        html += '<td class="py-2 px-3 text-right">' + formatBytes(block.size) + '</td>';
        html += '<td class="py-2 px-3 text-right">' + formatDifficulty(block.difficulty) + '</td>';
        html += '</tr>';
    });
    document.getElementById('blocks-table-body').innerHTML = html;
}

function loadMoreBlocks() {
    blocksCount += 20;
    fetchBlocks();
}
