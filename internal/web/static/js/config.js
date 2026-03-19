document.addEventListener('DOMContentLoaded', function() {
    fetchConfig();
});

function handleRefresh() {
    fetchConfig();
}

function fetchConfig() {
    api.get('/api/v1/config').then(function(resp) {
        renderNodeList(resp.data.nodes || []);
    }).catch(function(err) {
        showToast('Failed to load config: ' + err.message, 'error');
    });
}

function renderNodeList(nodes) {
    var container = document.getElementById('node-list');
    if (!nodes || nodes.length === 0) {
        container.innerHTML = '<p class="text-slate-500">No nodes configured. Add one below.</p>';
        return;
    }

    var html = '';
    nodes.forEach(function(node) {
        html += '<div class="flex items-center justify-between p-3 rounded" style="background:rgba(30,41,59,0.4); border:1px solid rgba(71,85,105,0.2);">';
        html += '<div>';
        html += '<span class="font-medium">' + node.name + '</span>';
        html += ' <span class="badge badge-blue">' + node.symbol + '</span>';
        html += ' <span class="text-slate-500 text-sm">' + node.host + ':' + node.port + '</span>';
        html += ' <span class="badge ' + (node.network === 'mainnet' ? 'badge-green' : 'badge-yellow') + '">' + node.network + '</span>';
        html += '</div>';
        html += '<div class="flex items-center gap-2">';
        html += '<button class="refresh-btn text-xs px-2 py-1" onclick="editNode(\'' + node.id + '\')">Edit</button>';
        html += '<button class="text-xs px-2 py-1 rounded border cursor-pointer" style="background:rgba(239,68,68,0.1); border-color:rgba(239,68,68,0.3); color:#ef4444;" onclick="deleteNode(\'' + node.id + '\')">Delete</button>';
        html += '</div>';
        html += '</div>';
    });
    container.innerHTML = html;
}

function handleNodeSubmit(event) {
    event.preventDefault();
    var editId = document.getElementById('node-edit-id').value;

    var node = {
        name: document.getElementById('node-name').value,
        symbol: document.getElementById('node-symbol').value,
        host: document.getElementById('node-host').value,
        port: parseInt(document.getElementById('node-port').value),
        network: document.getElementById('node-network').value,
        rest_enabled: document.getElementById('node-rest').checked
    };

    var promise;
    if (editId) {
        promise = api.put('/api/v1/config/nodes/' + editId, node);
    } else {
        promise = api.post('/api/v1/config/nodes', node);
    }

    promise.then(function(resp) {
        showToast(editId ? 'Node updated' : 'Node created', 'success');
        resetNodeForm();
        fetchConfig();
    }).catch(function(err) {
        showToast('Failed: ' + err.message, 'error');
    });
}

function editNode(nodeId) {
    api.get('/api/v1/config').then(function(resp) {
        var nodes = resp.data.nodes || [];
        var node = nodes.find(function(n) { return n.id === nodeId; });
        if (!node) return;

        document.getElementById('node-edit-id').value = node.id;
        document.getElementById('node-name').value = node.name;
        document.getElementById('node-symbol').value = node.symbol;
        document.getElementById('node-host').value = node.host;
        document.getElementById('node-port').value = node.port;
        document.getElementById('node-network').value = node.network;
        document.getElementById('node-rest').checked = node.rest_enabled;
        document.getElementById('node-form-title').textContent = 'Edit Node';
        document.getElementById('cancel-edit-btn').style.display = 'inline-block';
    });
}

function deleteNode(nodeId) {
    if (!confirm('Delete this node?')) return;

    api.del('/api/v1/config/nodes/' + nodeId).then(function() {
        showToast('Node deleted', 'success');
        fetchConfig();
    }).catch(function(err) {
        showToast('Failed: ' + err.message, 'error');
    });
}

function cancelNodeEdit() {
    resetNodeForm();
}

function resetNodeForm() {
    document.getElementById('node-form').reset();
    document.getElementById('node-edit-id').value = '';
    document.getElementById('node-form-title').textContent = 'Add Node';
    document.getElementById('cancel-edit-btn').style.display = 'none';
    document.getElementById('test-result').textContent = '';
}

function testNodeConnection() {
    var node = {
        host: document.getElementById('node-host').value,
        port: parseInt(document.getElementById('node-port').value),
        rest_enabled: true
    };

    if (!node.host || !node.port) {
        showToast('Host and port are required', 'error');
        return;
    }

    var resultEl = document.getElementById('test-result');
    resultEl.textContent = 'Testing...';
    resultEl.style.color = '#94a3b8';

    api.post('/api/v1/config/nodes/test', node).then(function(resp) {
        if (resp.data.success) {
            resultEl.textContent = 'Connection successful!';
            resultEl.style.color = '#22c55e';
        } else {
            resultEl.textContent = 'Failed: ' + resp.data.message;
            resultEl.style.color = '#ef4444';
        }
    }).catch(function(err) {
        resultEl.textContent = 'Error: ' + err.message;
        resultEl.style.color = '#ef4444';
    });
}
