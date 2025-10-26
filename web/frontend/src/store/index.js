import { createStore } from 'vuex'
import { authService, rulesService, postsService, statsService, logsService } from '../services/api' 

export default createStore({
  state: {
    user: null,
    isAuthenticated: false,
    loading: false,
    rules: [],
    posts: [],
    stats: {},
    logs: {} 
  },
  mutations: {
    SET_USER(state, user) {
      state.user = user
      state.isAuthenticated = !!user
    },
    SET_LOADING(state, loading) {
      state.loading = loading
    },
    SET_RULES(state, rules) {
      state.rules = rules
    },
    SET_POSTS(state, posts) {
      state.posts = posts
    },
    SET_STATS(state, stats) {
      state.stats = stats
    },
    SET_LOGS(state, logs) { 
    state.logs = logs
  },
    CLEAR_AUTH(state) {
      state.user = null
      state.isAuthenticated = false
    }
  },
  actions: {
    // Auth actions 
    async login({ commit }, credentials) {
      commit('SET_LOADING', true)
      try {
        const response = await authService.login(credentials)
        const { token, user } = response
        
        localStorage.setItem('auth_token', token)
        localStorage.setItem('user_data', JSON.stringify(user))
        commit('SET_USER', user)
        
        return response
      } catch (error) {
        commit('CLEAR_AUTH')
        throw error
      } finally {
        commit('SET_LOADING', false)
      }
    },

    async checkAuth({ commit }) {
      if (authService.isAuthenticated()) {
        try {
          const user = await authService.getCurrentUser()
          commit('SET_USER', user)
        } catch (error) {
          authService.logout()
          commit('CLEAR_AUTH')
        }
      }
    },

    logout({ commit }) {
      authService.logout()
      commit('CLEAR_AUTH')
    },

    // Rules actions 
    async fetchRules({ commit }) {
      commit('SET_LOADING', true)
      try {
        const rules = await rulesService.getRules()
        commit('SET_RULES', rules)
      } catch (error) {
        console.error('Error fetching rules:', error)
        throw error
      } finally {
        commit('SET_LOADING', false)
      }
    },

    async createRule({ dispatch }, ruleData) {
      try {
        const result = await rulesService.createRule(ruleData)
        await dispatch('fetchRules')
        return result
      } catch (error) {
        console.error('Error creating rule:', error)
        throw error
      }
    },

    async updateRule({ dispatch }, { id, ruleData }) {
      try {
        const result = await rulesService.updateRule(id, ruleData)
        await dispatch('fetchRules')
        return result
      } catch (error) {
        console.error('Error updating rule:', error)
        throw error
      }
    },

    async deleteRule({ dispatch }, id) {
      try {
        await rulesService.deleteRule(id)
        await dispatch('fetchRules')
      } catch (error) {
        console.error('Error deleting rule:', error)
        throw error
      }
    },

    // Posts actions 
    async fetchPosts({ commit }) {
      commit('SET_LOADING', true)
      try {
        const posts = await postsService.getPosts()
        commit('SET_POSTS', posts)
      } catch (error) {
        console.error('Error fetching posts:', error)
        throw error
      } finally {
        commit('SET_LOADING', false)
      }
    },

    // Stats actions 
    async fetchStats({ commit }) {
      commit('SET_LOADING', true)
      try {
        const stats = await statsService.getStats()
        commit('SET_STATS', stats)
      } catch (error) {
        console.error('Error fetching stats:', error)
        throw error
      } finally {
        commit('SET_LOADING', false)
      }
    },

   
// ДОБАВЬТЕ ЭТИ ДЕЙСТВИЯ ДЛЯ ЛОГОВ
async fetchLogs({ commit }, params = {}) {
  commit('SET_LOADING', true)
  try {
    const logsData = await logsService.getLogs(params)
    commit('SET_LOGS', logsData)
  } catch (error) {
    console.error('Error fetching logs:', error)
    throw error
  } finally {
    commit('SET_LOADING', false)
  }
}
  }
})