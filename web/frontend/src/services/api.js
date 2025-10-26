import axios from 'axios'

const API_BASE = 'http://localhost:8080/api'

// Создаем экземпляр axios с базовой конфигурацией
const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Интерцептор для добавления токена к запросам
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Интерцептор для обработки ошибок авторизации
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Токен невалидный или просрочен
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_data')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth service
export const authService = {
  async login(credentials) {
    const response = await api.post('/auth/login', credentials)
    return response.data
  },

  async getCurrentUser() {
    const response = await api.get('/auth/me')
    return response.data
  },

  logout() {
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_data')
    window.location.href = '/login'
  },

  isAuthenticated() {
    return !!localStorage.getItem('auth_token')
  },

  getToken() {
    return localStorage.getItem('auth_token')
  },

  getUser() {
    const userData = localStorage.getItem('user_data')
    return userData ? JSON.parse(userData) : null
  }
}

// Rules service
export const rulesService = {
  async getRules() {
    const response = await api.get('/rules')
    return response.data
  },

  async createRule(ruleData) {
    const response = await api.post('/rules', ruleData)
    return response.data
  },

  async updateRule(id, ruleData) {
    const response = await api.put(`/rules/${id}`, ruleData)
    return response.data
  },

  async deleteRule(id) {
    const response = await api.delete(`/rules/${id}`)
    return response.data
  }
}

// Posts service
export const postsService = {
  async getPosts(limit = 50, offset = 0) {
    const response = await api.get('/posts', {
      params: { limit, offset }
    })
    return response.data
  }
}

// Stats service
export const statsService = {
  async getStats() {
    const response = await api.get('/stats')
    return response.data
  }
}

// Logs service
export const logsService = {
  async getLogs(params = {}) {
    const response = await api.get('/logs', { 
      params: {
        limit: params.limit || 50,
        offset: params.offset || 0,
        level: params.level || '',
        search: params.search || '',
        service: params.service || ''
      }
    })
    return response.data
  }
}

export default api
