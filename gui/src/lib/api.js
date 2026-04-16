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

export function listWorkflows() { return request('/api/workflows'); }
export function getWorkflow(name) { return request(`/api/workflows/${encodeURIComponent(name)}`); }
export function saveWorkflow(name, source) {
  return request(`/api/workflows/${encodeURIComponent(name)}`, { method: 'PUT', body: JSON.stringify({ source }) });
}
export function runWorkflow(name, params) {
  return request(`/api/workflows/${encodeURIComponent(name)}/run`, { method: 'POST', body: JSON.stringify({ params }) });
}
export function listRuns() { return request('/api/runs'); }
export function getRun(id) { return request(`/api/runs/${id}`); }
export function getResult(path) { return request(`/api/results/${path}`); }
export function getKibanaWorkflow(name) { return request(`/api/kibana/workflow/${encodeURIComponent(name)}`); }
export function getKibanaRun(id) { return request(`/api/kibana/run/${id}`); }

export function saveResult(path, content) {
  return request(`/api/results/${path}`, { method: 'PUT', body: JSON.stringify({ content }) });
}
export function getResultText(path) {
  return fetch(`/api/results/${path}`).then(res => { if (!res.ok) throw new Error(res.statusText); return res.text(); });
}
export function getWorkflowActions(context) {
  return request(`/api/workflows/actions/${context}`);
}
export function getWorkspace() { return request('/api/workspace'); }
export function updateWorkspace(data) {
  return request('/api/workspace', { method: 'PUT', body: JSON.stringify(data) });
}
export function getProviders() { return request('/api/providers'); }
