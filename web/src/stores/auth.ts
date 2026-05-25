import { defineStore } from 'pinia'

import { http } from '../api/http'

export interface CurrentUser {
  id?: number
  username: string
  nickname: string
  avatar: string
  gender?: number
  signature?: string
  publicKey?: string
  publicKeyAlgorithm?: string
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null as CurrentUser | null,
    bootstrapped: false
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.user)
  },
  actions: {
    setSession(token: string, user: CurrentUser) {
      this.token = token
      this.user = user
      this.bootstrapped = true
      localStorage.setItem('token', token)
    },
    clearSession() {
      this.token = ''
      this.user = null
      this.bootstrapped = true
      localStorage.removeItem('token')
    },
    async bootstrap() {
      if (!this.token) {
        this.bootstrapped = true
        return
      }
      try {
        const { data } = await http.get('/users/me')
        this.user = data
      } catch {
        this.clearSession()
      } finally {
        this.bootstrapped = true
      }
    }
  }
})
