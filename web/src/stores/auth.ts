import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '../api'

export interface NeteaseUser {
  userId: number
  nickname: string
  avatar: string
  vipType: string
  cookieValid: boolean
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<NeteaseUser | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => user.value !== null)

  const checkLoginStatus = async () => {
    try {
      const res = await authAPI.getLoginStatus()
      if (res.data.logged_in && res.data.user) {
        user.value = {
          userId: res.data.user.userId,
          nickname: res.data.user.nickname,
          avatar: res.data.user.avatarUrl,
          vipType: res.data.user.vipType || '',
          cookieValid: res.data.user.cookieValid !== false
        }
      } else {
        user.value = null
      }
    } catch {
      user.value = null
    }
  }

  const clearUser = () => {
    user.value = null
  }

  return { user, loading, isLoggedIn, checkLoginStatus, clearUser }
})
