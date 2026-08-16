import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { systemAPI } from '../api'

// 解析 JWT token（不验证签名，仅解析 payload）
function parseJwtPayload(token: string): any {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
    }).join(''))
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

// 检查 token 是否过期
function isTokenExpired(token: string): boolean {
  const payload = parseJwtPayload(token)
  if (!payload || !payload.exp) return true
  // 提前 60 秒视为过期，避免边界情况
  return Date.now() / 1000 >= payload.exp - 60
}

export const useSystemAuthStore = defineStore('systemAuth', () => {
  const token = ref<string | null>(localStorage.getItem('system_token'))
  const username = ref<string | null>(localStorage.getItem('system_username'))
  const role = ref<string | null>(localStorage.getItem('system_role'))

  // 检查登录状态：token 存在且未过期
  const isSystemLoggedIn = computed(() => {
    if (!token.value) return false
    // 检查 token 是否过期
    if (isTokenExpired(token.value)) {
      // token 已过期，清除登录状态
      logout()
      return false
    }
    return true
  })
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
