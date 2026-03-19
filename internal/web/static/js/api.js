class APIClient {
    async get(endpoint) {
        const response = await fetch(endpoint);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.json();
    }
    async post(endpoint, body) {
        const response = await fetch(endpoint, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(body)
        });
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.json();
    }
    async put(endpoint, body) {
        const response = await fetch(endpoint, {
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(body)
        });
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.json();
    }
    async patch(endpoint, body) {
        const response = await fetch(endpoint, {
            method: 'PATCH',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(body)
        });
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.json();
    }
    async del(endpoint) {
        const response = await fetch(endpoint, { method: 'DELETE' });
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.json();
    }
}
var api = new APIClient();
