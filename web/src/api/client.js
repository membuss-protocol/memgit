// Type-safe REST client for MemGit API & Membuss Network

const BASE = '/api/v1';

async function request(path, options = {}) {
  const url = `${BASE}${path}`;
  const res = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });

  const text = await res.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch (e) {
    throw new Error(`Invalid JSON response: ${text.slice(0, 100)}`);
  }

  if (!res.ok || json.ok === false) {
    throw new Error(json.error || `HTTP ${res.status}`);
  }

  return json.data;
}

export const api = {
  // System & Swarm Telemetry
  getSystemStatus: () => request('/system/status'),
  getActivityFeed: (limit = 30) => request(`/activity/feed?limit=${limit}`),

  // User & Identity
  getCurrentUser: () => request('/user'),
  updateProfile: (data) =>
    request('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  getUserProfile: (username) => request(`/users/${encodeURIComponent(username)}`),
  listUsers: () => request('/users'),

  // Repositories & Discovery
  getRepos: () => request('/repos'),
  getExploreRepos: (filter = 'all', query = '') => {
    const params = new URLSearchParams();
    if (filter) params.append('filter', filter);
    if (query) params.append('q', query);
    const qs = params.toString() ? `?${params.toString()}` : '';
    return request(`/explore/repos${qs}`);
  },
  getRepo: (name) => request(`/repos/${encodeURIComponent(name)}`),
  createRepo: (data) =>
    request('/repos', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  starRepo: (name) =>
    request(`/repos/${encodeURIComponent(name)}/star`, {
      method: 'POST',
    }),
  forkRepo: (name, newName = '') =>
    request(`/repos/${encodeURIComponent(name)}/fork`, {
      method: 'POST',
      body: JSON.stringify({ new_name: newName }),
    }),
  syncRepo: (name) =>
    request(`/repos/${encodeURIComponent(name)}/sync`, {
      method: 'POST',
    }),
  getSwarmStatus: (name) => request(`/repos/${encodeURIComponent(name)}/network`),

  // Git Objects & Tree
  getTree: (name, ref = 'HEAD', path = '') => {
    const sub = path ? `/${path}` : '';
    return request(`/repos/${encodeURIComponent(name)}/tree/${encodeURIComponent(ref)}${sub}`);
  },
  getBlob: (name, ref = 'HEAD', path = '') =>
    request(`/repos/${encodeURIComponent(name)}/blob/${encodeURIComponent(ref)}/${path}`),
  getCommits: (name, ref = 'HEAD', limit = 30) =>
    request(`/repos/${encodeURIComponent(name)}/commits/${encodeURIComponent(ref)}?limit=${limit}`),
  getCommit: (name, sha) =>
    request(`/repos/${encodeURIComponent(name)}/commit/${encodeURIComponent(sha)}`),
  getBranches: (name) => request(`/repos/${encodeURIComponent(name)}/branches`),
  createBranch: (name, branchName, targetSHA) =>
    request(`/repos/${encodeURIComponent(name)}/branches`, {
      method: 'POST',
      body: JSON.stringify({ name: branchName, target_sha: targetSHA }),
    }),
  getTags: (name) => request(`/repos/${encodeURIComponent(name)}/tags`),

  // Issues
  getIssues: (name, state = '') => {
    const q = state ? `?state=${encodeURIComponent(state)}` : '';
    return request(`/repos/${encodeURIComponent(name)}/issues${q}`);
  },
  createIssue: (name, data) =>
    request(`/repos/${encodeURIComponent(name)}/issues`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  getIssue: (name, id) => request(`/repos/${encodeURIComponent(name)}/issues/${id}`),
  addIssueComment: (name, id, body) =>
    request(`/repos/${encodeURIComponent(name)}/issues/${id}/comment`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    }),
  updateIssueState: (name, id, state) =>
    request(`/repos/${encodeURIComponent(name)}/issues/${id}/state`, {
      method: 'PATCH',
      body: JSON.stringify({ state }),
    }),

  // Pull Requests
  getPRs: (name, state = '') => {
    const q = state ? `?state=${encodeURIComponent(state)}` : '';
    return request(`/repos/${encodeURIComponent(name)}/pulls${q}`);
  },
  createPR: (name, data) =>
    request(`/repos/${encodeURIComponent(name)}/pulls`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  getPR: (name, id) => request(`/repos/${encodeURIComponent(name)}/pulls/${id}`),
  mergePR: (name, id) =>
    request(`/repos/${encodeURIComponent(name)}/pulls/${id}/merge`, {
      method: 'POST',
    }),

  // Releases
  getReleases: (name) => request(`/repos/${encodeURIComponent(name)}/releases`),
  createRelease: (name, data) =>
    request(`/repos/${encodeURIComponent(name)}/releases`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
};
