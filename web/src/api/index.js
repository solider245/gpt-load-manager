import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

export function handleError(err) {
  const msg = err.response?.data?.error || err.message || '请求失败'
  return msg
}

export default {
  // Server CRUD
  listServers() { return api.get('/servers') },
  getServer(id)  { return api.get(`/servers/${id}`) },
  createServer(data) { return api.post('/servers', data) },
  updateServer(id, data) { return api.put(`/servers/${id}`, data) },
  deleteServer(id) { return api.delete(`/servers/${id}`) },

  // Health check
  checkServerHealth(id) { return api.post(`/servers/${id}/check`) },
}
