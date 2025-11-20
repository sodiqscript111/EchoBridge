const API_BASE = 'http://localhost:8000';

export const api = {
    async post(endpoint, data, token = null) {
        const headers = {
            'Content-Type': 'application/json',
        };
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(`${API_BASE}${endpoint}`, {
            method: 'POST',
            headers,
            body: JSON.stringify(data),
        });

        const json = await response.json();
        if (!response.ok) {
            throw new Error(json.error || 'Request failed');
        }
        return json;
    },

    async get(endpoint, token = null) {
        const headers = {};
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        console.log(`[API] GET ${API_BASE}${endpoint}`);
        try {
            const response = await fetch(`${API_BASE}${endpoint}`, {
                method: 'GET',
                headers,
            });

            const json = await response.json();
            if (!response.ok) {
                console.error('[API] Error:', json);
                throw new Error(json.error || 'Request failed');
            }
            return json;
        } catch (error) {
            console.error('[API] Fetch failed:', error);
            throw error;
        }
    },

    async patch(endpoint, data, token = null) {
        const headers = {
            'Content-Type': 'application/json',
        };
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(`${API_BASE}${endpoint}`, {
            method: 'PATCH',
            headers,
            body: JSON.stringify(data),
        });

        const json = await response.json();
        if (!response.ok) {
            throw new Error(json.error || 'Request failed');
        }
        return json;
    },
};
