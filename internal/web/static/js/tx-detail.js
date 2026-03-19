document.addEventListener('DOMContentLoaded', function() {
    initNodeSelector().then(function() {
        var page = document.getElementById('tx-detail-page');
        var txid = page ? page.getAttribute('data-txid') : '';
        var blockhash = page ? page.getAttribute('data-blockhash') : '';
        if (txid && blockhash && getSelectedNode()) {
            fetchTransaction(txid, blockhash);
        } else if (txid && !blockhash) {
            document.getElementById('tx-summary-content').innerHTML =
                '<p class="text-yellow-400">Block hash is required to fetch transaction details.</p>';
        }
    });
});

function handleRefresh() {
    var page = document.getElementById('tx-detail-page');
    var txid = page ? page.getAttribute('data-txid') : '';
    var blockhash = page ? page.getAttribute('data-blockhash') : '';
    if (txid && blockhash) fetchTransaction(txid, blockhash);
}

function fetchTransaction(txid, blockhash) {
    var nodeId = getSelectedNode();
    if (!nodeId) {
        showToast('Please select a node', 'error');
        return;
    }

    api.get('/api/v1/' + nodeId + '/tx/' + txid + '?blockhash=' + blockhash).then(function(resp) {
        renderTransaction(resp.data, blockhash);
    }).catch(function(err) {
        document.getElementById('tx-summary-content').innerHTML =
            '<p class="text-red-400">Error: ' + err.message + '</p>';
    });
}

function renderTransaction(tx, blockhash) {
    var isCoinbase = tx.vin && tx.vin.length > 0 && tx.vin[0].coinbase;

    // Summary
    var html = '';
    html += '<div><span class="text-slate-400 text-sm">TxID</span><br><span class="hash-text cursor-pointer" onclick="copyToClipboard(\'' + tx.txid + '\')">' + tx.txid + '</span></div>';
    html += '<div><span class="text-slate-400 text-sm">Size</span><br>' + formatBytes(tx.size) + ' (vsize: ' + formatBytes(tx.vsize) + ')</div>';
    html += '<div><span class="text-slate-400 text-sm">Weight</span><br>' + formatNumber(tx.weight) + '</div>';
    html += '<div><span class="text-slate-400 text-sm">Version</span><br>' + tx.version + '</div>';
    html += '<div><span class="text-slate-400 text-sm">Locktime</span><br>' + formatNumber(tx.locktime) + '</div>';
    if (isCoinbase) {
        html += '<div><span class="coinbase-indicator">Coinbase Transaction</span></div>';
    }
    if (tx.fee !== undefined && tx.fee !== null) {
        html += '<div><span class="text-slate-400 text-sm">Fee</span><br>' + formatBTC(tx.fee) + '</div>';
    }
    html += '<div><span class="text-slate-400 text-sm">Block</span><br><a href="/block/' + blockhash + '" class="hash-text">' + truncateHash(blockhash, 12) + '</a></div>';
    document.getElementById('tx-summary-content').innerHTML = html;

    // Inputs
    var inputBadge = document.getElementById('input-count-badge');
    if (inputBadge) inputBadge.textContent = tx.vin ? tx.vin.length : 0;

    var inputsHtml = '';
    if (tx.vin && tx.vin.length > 0) {
        tx.vin.forEach(function(vin, i) {
            inputsHtml += '<tr>';
            inputsHtml += '<td class="py-2 px-3">' + i + '</td>';
            if (vin.coinbase) {
                var cbAscii = hexToAscii(vin.coinbase);
                inputsHtml += '<td class="py-2 px-3 coinbase-indicator">Coinbase' + (cbAscii ? ' <span class="text-slate-400 text-xs ml-2">(' + cbAscii + ')</span>' : '') + '</td>';
                inputsHtml += '<td class="py-2 px-3 text-right">-</td>';
                inputsHtml += '<td class="py-2 px-3 hash-text text-xs">' + vin.coinbase + '</td>';
            } else {
                inputsHtml += '<td class="py-2 px-3 hash-text">' + truncateHash(vin.txid, 10) + '</td>';
                inputsHtml += '<td class="py-2 px-3 text-right">' + vin.vout + '</td>';
                inputsHtml += '<td class="py-2 px-3">' + (vin.scriptSig ? 'scriptSig' : 'witness') + '</td>';
            }
            inputsHtml += '</tr>';
        });
    } else {
        inputsHtml = '<tr><td colspan="4" class="py-4 text-center text-slate-500">No inputs.</td></tr>';
    }
    document.getElementById('tx-inputs-body').innerHTML = inputsHtml;

    // Outputs
    var outputBadge = document.getElementById('output-count-badge');
    if (outputBadge) outputBadge.textContent = tx.vout ? tx.vout.length : 0;

    var outputsHtml = '';
    if (tx.vout && tx.vout.length > 0) {
        tx.vout.forEach(function(vout) {
            outputsHtml += '<tr>';
            outputsHtml += '<td class="py-2 px-3">' + vout.n + '</td>';
            outputsHtml += '<td class="py-2 px-3 text-right text-amber-400">' + formatBTC(vout.value) + '</td>';
            if (vout.scriptPubKey.type === 'nulldata') {
                outputsHtml += '<td class="py-2 px-3 text-slate-500 text-xs italic">Witness Commitment</td>';
                outputsHtml += '<td class="py-2 px-3 text-slate-500">' + vout.scriptPubKey.type + '</td>';
            } else {
                outputsHtml += '<td class="py-2 px-3 hash-text">' + (vout.scriptPubKey.address || 'N/A') + '</td>';
                outputsHtml += '<td class="py-2 px-3">' + (vout.scriptPubKey.type || 'N/A') + '</td>';
            }
            outputsHtml += '</tr>';
        });
    } else {
        outputsHtml = '<tr><td colspan="4" class="py-4 text-center text-slate-500">No outputs.</td></tr>';
    }
    document.getElementById('tx-outputs-body').innerHTML = outputsHtml;
}
