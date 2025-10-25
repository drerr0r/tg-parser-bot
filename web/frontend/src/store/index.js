import { createStore } from 'vuex'
import axios from 'axios'

const API_BASE = 'http://localhost:8080/api'

export default createStore({
  state: {
    rules: [],
    posts: [],
    stats: {},
    loading: false
  },
  mutations: {
    SET_RULES(state, rules) {
      state.rules = rules
    },
    SET_POSTS(state, posts) {
      state.posts = posts
    },
    SET_STATS(state, stats) {
      state.stats = stats
    },
    SET_LOADING(state, loading) {
      state.loading = loading
    }
  },
  actions: {
    async fetchRules({ commit }) {
      commit('SET_LOADING', true)
      try {
        const response = await axios.get(`${API_BASE}/rules`)
        commit('SET_RULES', response.data)
      } catch (error) {
        console.error('Error fetching rules:', error)
        throw error
      } finally {
        commit('SET_LOADING', false)
      }
    },

    async createRule({ dispatch }, ruleData) {
      try {
        const response = await axios.post(`${API_BASE}/rules`, ruleData)
        await dispatch('fetchRules') // Обновляем список
        return response.data
      } catch (error) {
        console.error('Error creating rule:', error)
        throw error
      }
    },

    async updateRule({ dispatch }, { id, ruleData }) {
      try {
        const response = await axios.put(`${API_BASE}/rules/${id}`, ruleData)
        await dispatch('fetchRules') // Обновляем список
        return response.data
      } catch (error) {
        console.error('Error updating rule:', error)
        throw error
      }
    },

    async deleteRule({ dispatch }, id) {
      try {
        await axios.delete(`${API_BASE}/rules/${id}`)
        await dispatch('fetchRules') // Обновляем список
      } catch (error) {
        console.error('Error deleting rule:', error)
        throw error
      }
    },

    async fetchPosts({ commit }) {
      commit('SET_LOADING', true)
      try {
        const response = await axios.get(`${API_BASE}/posts`)
        commit('SET_POSTS', response.data)
      } catch (error) {
        console.error('Error fetching posts:', error)
        throw error
      } finally {
        commit('SET_LOADING', false)
      }
    },

    async fetchStats({ commit }) {
      commit('SET_LOADING', true)
      try {
        const response = await axios.get(`${API_BASE}/stats`)
        commit('SET_STATS', response.data)
      } catch (error) {
        console.error('Error fetching stats:', error)
        throw error
      } finally {
        commit('SET_LOADING', false)
      }
    }
  }
})