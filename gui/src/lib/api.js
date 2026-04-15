const BASE = ''

async function request(path, opts = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(`${res.status}: ${text}`)
  }
  const ct = res.headers.get('content-type') || ''
  if (ct.includes('json')) return res.json()
  return res.text()
}

export const api = {
  listWorkflows: () => request('/api/workflows'),
  getWorkflow: (name) => request(`/api/workflows/${encodeURIComponent(name)}`),
  saveWorkflow: (name, source) =>
    request(`/api/workflows/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify({ source }),
    }),
  runWorkflow: (name, params) =>
    request(`/api/workflows/${encodeURIComponent(name)}/run`, {
      method: 'POST',
      body: JSON.stringify({ params }),
    }),
  listRuns: () => request('/api/runs'),
  getRun: (id) => request(`/api/runs/${id}`),
  getResult: (path) => request(`/api/results/${path}`),
  getKibanaWorkflow: (name) => request(`/api/kibana/workflow/${encodeURIComponent(name)}`),
  getKibanaRun: (id) => request(`/api/kibana/run/${id}`),
}
