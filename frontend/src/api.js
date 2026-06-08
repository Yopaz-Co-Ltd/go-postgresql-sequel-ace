async function request(path, options = {}) {
  const response = await fetch(path, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })

  const data = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(data.error || `Request failed with status ${response.status}`)
  }
  return data
}

export function testConnection(payload) {
  return request('/api/test-connection', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function connect(payload) {
  return request('/api/connect', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function getSchemas(sessionId) {
  return request(`/api/schemas?sessionId=${encodeURIComponent(sessionId)}`)
}

export function getTables(sessionId, schema) {
  return request(`/api/tables?sessionId=${encodeURIComponent(sessionId)}&schema=${encodeURIComponent(schema)}`)
}

export function getTableRows(sessionId, schema, table, limit = 200) {
  const params = new URLSearchParams({ sessionId, schema, table, limit: String(limit) })
  return request(`/api/table-rows?${params}`)
}

export function getTableInfo(sessionId, schema, table) {
  const params = new URLSearchParams({ sessionId, schema, table })
  return request(`/api/table-info?${params}`)
}

export function runQuery(sessionId, sql) {
  return request('/api/query', {
    method: 'POST',
    body: JSON.stringify({ sessionId, sql }),
  })
}
