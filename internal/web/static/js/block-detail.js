var currentBlock = null;

document.addEventListener('DOMContentLoaded', function() {
    initNodeSelector().then(function() {
        var page = document.getElementById('block-detail-page');
        var hash = page ? page.getAttribute('data-block-hash') : '';
        if (hash && getSelectedNode()) {
            fetchBlock(hash);
        }
    });
});

function handleRefresh() {
    var page = document.getElementById('block-detail-page');
    var hash = page ? page.getAttribute('data-block-hash') : '';
    if (hash) fetchBlock(hash);
}

function fetchBlock(hash) {
    var nodeId = getSelectedNode();
    if (!nodeId) {
        showToast('Please select a node', 'error');
        return;
    }

    api.get('/api/v1/' + nodeId + '/block/' + hash).then(function(resp) {
        currentBlock = resp.data;
        renderBlockHeader(currentBlock);
        renderBlockTransactions(currentBlock);
    }).catch(function(err) {
        document.getElementById('block-header-content').innerHTML =
            '<p class="text-red-400">Error: ' + err.message + '</p>';
    });
}

function renderBlockHeader(block) {
    var heightDisplay = document.getElementById('block-height-display');
    if (heightDisplay) heightDisplay.textContent = 'Block #' + formatNumber(block.height);

    var prevBtn = document.getElementById('prev-block-btn');
    var nextBtn = document.getElementById('next-block-btn');
    if (prevBtn) prevBtn.style.display = block.previousblockhash ? 'inline-block' : 'none';
    if (nextBtn) nextBtn.style.display = block.nextblockhash ? 'inline-block' : 'none';

    var html = '';
    html += '<div><span class="text-slate-400 text-sm">Hash</span><br><span class="hash-text cursor-pointer" onclick="copyToClipboard(\'' + block.hash + '\')">' + block.hash + '</span></div>';
    html += '<div><span class="text-slate-400 text-sm">Confirmations</span><br><span class="badge badge-green">' + formatNumber(block.confirmations) + '</span></div>';
    html += '<div><span class="text-slate-400 text-sm">Timestamp</span><br>' + formatTimestamp(block.time) + ' (' + formatTimeAgo(block.time) + ')</div>';
    html += '<div><span class="text-slate-400 text-sm">Difficulty</span><br>' + formatDifficulty(block.difficulty) + '</div>';
    html += '<div><span class="text-slate-400 text-sm">Merkle Root</span><br><span class="hash-text">' + truncateHash(block.merkleroot, 16) + '</span></div>';
    html += '<div><span class="text-slate-400 text-sm">Size</span><br>' + formatBytes(block.size) + '</div>';
    html += '<div><span class="text-slate-400 text-sm">Weight</span><br>' + formatNumber(block.weight) + '</div>';
    html += '<div><span class="text-slate-400 text-sm">Nonce</span><br>' + formatNumber(block.nonce) + '</div>';
    if (block.pow_algo) {
        html += '<div><span class="text-slate-400 text-sm">PoW Algo</span><br>' + block.pow_algo + '</div>';
    }
    html += '<div><span class="text-slate-400 text-sm">Version</span><br>0x' + block.version.toString(16) + '</div>';

    document.getElementById('block-header-content').innerHTML = html;

    var countBadge = document.getElementById('tx-count-badge');
    if (countBadge) countBadge.textContent = block.tx ? block.tx.length : 0;
}

function renderBlockTransactions(block) {
    var txs = block.tx || [];
    if (txs.length === 0) {
        document.getElementById('tx-table-body').innerHTML =
            '<tr><td colspan="5" class="py-4 text-center text-slate-500">No transactions.</td></tr>';
        return;
    }

    var html = '';
    txs.forEach(function(tx, i) {
        var isCoinbase = tx.vin && tx.vin.length > 0 && tx.vin[0].coinbase;
        html += '<tr class="hover:bg-slate-800/30 cursor-pointer" onclick="window.location.href=\'/tx/' + tx.txid + '?blockhash=' + block.hash + '\'">';
        html += '<td class="py-2 px-3">' + i;
        if (isCoinbase) {
            var cbText = hexToAscii(tx.vin[0].coinbase);
            html += ' <span class="coinbase-indicator">CB</span>';
            if (cbText) html += ' <span class="text-slate-400 text-xs">(' + cbText + ')</span>';
        }
        html += '</td>';
        html += '<td class="py-2 px-3 hash-text">' + truncateHash(tx.txid, 12) + '</td>';
        html += '<td class="py-2 px-3 text-right">' + (tx.vin ? tx.vin.length : 0) + '</td>';
        html += '<td class="py-2 px-3 text-right">' + (tx.vout ? tx.vout.length : 0) + '</td>';
        html += '<td class="py-2 px-3 text-right">' + formatBytes(tx.size) + '</td>';
        html += '</tr>';
    });
    document.getElementById('tx-table-body').innerHTML = html;
}

function navigateBlock(direction) {
    if (!currentBlock) return;
    var hash = direction === 'prev' ? currentBlock.previousblockhash : currentBlock.nextblockhash;
    if (hash) {
        window.location.href = '/block/' + hash;
    }
}
