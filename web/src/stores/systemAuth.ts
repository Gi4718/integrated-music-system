import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { systemAPI } from '../api'

export const useSystemAuthStore = defineStore('systemAuth', () => {
  const token = ref<string | null>(localStorage.getItem('system_token'))
  const username = ref<string | null>(localStorage.getItem('system_username'))
  const role = ref<string | null>(localStorage.getItem('system_role'))

  const isSystemLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => !role.value || role.value === 'admin')

  const login = async (usernameInput: string, password: string) => {
    const res = await systemAPI.login({
      username: usernameInput,
      password
    })
    if (res.data.token) {
      token.value = res.data.token
      username.value = res.data.username
      role.value = res.data.role || 'user'
      localStorage.setItem('system_token', res.data.token)
      localStorage.setItem('system_username', res.data.username)
      localStorage.setItem('system_role', role.value || 'user')
      return true
    }
    return false
  }

  const logout = () => {
    token.value = null
    username.value = null
    role.value = null
    localStorage.removeItem('system_token')
    localStorage.removeItem('system_username')
    localStorage.removeItem('system_role')
  }

  const checkSystemUser = async () => {
    try {
      const res = await systemAPI.check()
      return res.data.has_user
    } catch {
      return false
    }
  }

  return {
    token,
    username,
    role,
    isSystemLoggedIn,
    isAdmin,
    login,
    logout,
    checkSystemUser
  }
})
