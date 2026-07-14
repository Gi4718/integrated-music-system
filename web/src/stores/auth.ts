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

  // 网易云 vipType 数字映射
  const vipTypeMap: Record<number, string> = {
    0: '',
    110: '黑胶VIP',
    100: '音乐包',
    120: '黑胶VIP',
  }

  const checkLoginStatus = async () => {
    try {
      const res = await authAPI.getLoginStatus()
      if (res.data.logged_in && res.data.user) {
        const vipNum = res.data.vipType || res.data.user.vipType || 0
        user.value = {
          userId: res.data.user.user_id,
          nickname: res.data.user.nickname,
          avatar: res.data.user.avatar,
          vipType: vipTypeMap[vipNum] || '',
          cookieValid: res.data.cookie_valid !== false
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
