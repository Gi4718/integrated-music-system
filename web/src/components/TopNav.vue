<template>
  <nav class="top-nav">
    <div class="nav-left">
      <div class="logo">
        <img :src="isDark ? '/icon-black.ico' : '/icon-light.ico'" class="logo-icon" alt="Logo" />
        <span class="logo-text">集成音乐系统</span>
      </div>
    </div>

    <div class="nav-center">
      <div class="search-box">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索歌曲 / 粘贴歌单链接或ID"
          @keyup.enter="handleSearch"
          @paste="handlePaste"
        />
        <button @click="handleSearch">
          <svg viewBox="0 0 24 24" width="18" height="18">
            <path fill="currentColor" d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
          </svg>
        </button>
      </div>
    </div>

    <div class="nav-right">
      <button @click="navigateTo('/')" class="nav-link" :class="{ 'router-link-active': $route.path === '/' }">首页</button>
      <button @click="navigateTo('/playlist')" class="nav-link" :class="{ 'router-link-active': $route.path === '/playlist' }">歌单</button>
      <button @click="navigateTo('/login')" class="nav-link" :class="{ 'router-link-active': $route.path === '/login' }">账号</button>
      <button @click="navigateTo('/settings')" class="nav-link" :class="{ 'router-link-active': $route.path === '/settings' }">设置</button>
      <button class="message-btn" @click="showMessagePanel = !showMessagePanel">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M12 22c1.1 0 2-.9 2-2h-4c0 1.1.9 2 2 2zm6-6v-5c0-3.07-1.63-5.64-4.5-6.32V4c0-.83-.67-1.5-1.5-1.5s-1.5.67-1.5 1.5v.68C7.64 5.36 6 7.92 6 11v5l-2 2v1h16v-1l-2-2zm-2 1H8v-6c0-2.48 1.51-4.5 4-4.5s4 2.02 4 4.5v6z"/>
        </svg>
        <span v-if="unreadCount > 0" class="badge">{{ unreadCount }}</span>
      </button>
      <button class="theme-toggle-btn" @click="toggleTheme" :title="isDark ? '切换到明亮模式' : '切换到黑暗模式'">
        <svg v-if="isDark" viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M12 7c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5zM2 13h2c.55 0 1-.45 1-1s-.45-1-1-1H2c-.55 0-1 .45-1 1s.45 1 1 1zm18 0h2c.55 0 1-.45 1-1s-.45-1-1-1h-2c-.55 0-1 .45-1 1s.45 1 1 1zM11 2v2c0 .55.45 1 1 1s1-.45 1-1V2c0-.55-.45-1-1-1s-1 .45-1 1zm0 18v2c0 .55.45 1 1 1s1-.45 1-1v-2c0-.55-.45-1-1-1s-1 .45-1 1zM5.99 4.58c-.39-.39-1.03-.39-1.42 0-.39.39-.39 1.03 0 1.42l1.42 1.42c.39.39 1.03.39 1.42 0 .38-.39.39-1.03 0-1.42L5.99 4.58zm12.37 12.37c-.39-.39-1.03-.39-1.42 0-.39.39-.39 1.03 0 1.42l1.42 1.42c.39.39 1.03.39 1.42 0 .39-.39.39-1.03 0-1.42l-1.42-1.42zm1.42-12.37c.39-.39.39-1.03 0-1.42-.39-.39-1.03-.39-1.42 0l-1.42 1.42c-.39.39-.39 1.03 0 1.42.39.39 1.03.39 1.42 0l1.42-1.42zM6.05 18.36c.39-.39.39-1.03 0-1.42-.39-.39-1.03-.39-1.42 0l-1.42 1.42c-.39.39-.39 1.03 0 1.42.39.39 1.03.39 1.42 0l1.42-1.42z"/>
        </svg>
        <svg v-else viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.06-.9-.1-1.36-.1z"/>
        </svg>
      </button>
    </div>

    <MessagePanel :visible="showMessagePanel" @close="showMessagePanel = false" />
  </nav>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { downloadAPI, taskAPI } from '../api'
import { ElMessage } from 'element-plus'
import MessagePanel from './MessagePanel.vue'
import { useThemeStore } from '../stores/theme'

const router = useRouter()
const themeStore = useThemeStore()
const searchQuery = ref('')
const showMessagePanel = ref(false)
const unreadCount = ref(0)
let pollTimer: number | null = null

const isDark = ref(false)
const toggleTheme = () => {
  themeStore.toggleTheme()
  isDark.value = themeStore.isDark
}

// 获取 App.vue 提供的导航方法
const navigateWithAnimation = inject('navigateWithAnimation') as ((targetPath: string) => Promise<void>) | undefined

const navigateTo = async (targetPath: string) => {
  if (navigateWithAnimation) {
    await navigateWithAnimation(targetPath)
  } else {
    await router.push(targetPath)
  }
}

const loadUnreadCount = async () => {
  try {
    const res = await taskAPI.getTasks()
    const tasks = res.data.tasks || []
    unreadCount.value = tasks.filter((t: any) => t.status === 'running' || t.status === 'pending').length
  } catch (e) {
    console.error('加载任务状态失败', e)
  }
}

onMounted(() => {
  isDark.value = themeStore.isDark
  loadUnreadCount()
  pollTimer = window.setInterval(loadUnreadCount, 5000)
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})

const extractPlaylistId = (input: string): number | null => {
  const trimmed = input.trim()

  // 纯数字 ID
  if (/^\d+$/.test(trimmed)) {
    return parseInt(trimmed, 10)
  }

  // 网易云歌单链接: https://music.163.com/#/playlist?id=xxx
  const urlMatch = trimmed.match(/[?&]id=(\d+)/)
  if (urlMatch) {
    return parseInt(urlMatch[1], 10)
  }

  // music.163.com/playlist/xxx
  const pathMatch = trimmed.match(/playlist\/(\d+)/)
  if (pathMatch) {
    return parseInt(pathMatch[1], 10)
  }

  return null
}

const handleSearch = () => {
  const query = searchQuery.value.trim()
  if (!query) return

  const playlistId = extractPlaylistId(query)
  if (playlistId) {
    // 识别为歌单，直接下载
    downloadPlaylistById(playlistId)
  } else {
    // 普通搜索
    router.push({ path: '/search', query: { q: query } })
  }
  searchQuery.value = ''
}

const handlePaste = (_e: ClipboardEvent) => {
  // 粘贴后延迟检测，等值写入 input
  setTimeout(() => {
    const query = searchQuery.value.trim()
    if (query && extractPlaylistId(query)) {
      // 不自动触发，等用户按回车
    }
  }, 50)
}

const downloadPlaylistById = async (id: number) => {
  try {
    await downloadAPI.downloadPlaylist(id, 'high')
    ElMessage.success(`歌单 ${id} 已加入下载队列`)
    router.push('/downloads')
  } catch {
    ElMessage.error('歌单下载失败，请确认已登录')
  }
}
</script>

<style scoped>
.top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;
  background-color: #FFD700;
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-left {
  display: flex;
  align-items: center;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  width: 40px;
  height: 40px;
}

.logo-text {
  font-size: 20px;
  font-weight: bold;
  color: #000;
}

.nav-center {
  flex: 1;
  max-width: 500px;
  margin: 0 40px;
}

.search-box {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 20px;
  padding: 8px 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.search-box input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 14px;
  background: transparent;
}

.search-box button {
  background: none;
  border: none;
  cursor: pointer;
  color: #666;
  display: flex;
  align-items: center;
  padding: 0;
}

.search-box button:hover {
  color: #000;
}

.nav-right {
  display: flex;
  gap: 24px;
}

.nav-link {
  background: none;
  border: none;
  cursor: pointer;
  color: #000;
  font-size: 16px;
  font-weight: 500;
  transition: opacity 0.2s;
  padding: 0;
  display: flex;
  align-items: center;
}

.nav-link:hover {
  opacity: 0.7;
}

.nav-link.router-link-active {
  font-weight: bold;
}

.message-btn {
  position: relative;
  background: none;
  border: none;
  cursor: pointer;
  color: #000;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background 0.2s;
}

.message-btn:hover {
  background: rgba(0, 0, 0, 0.1);
}

.badge {
  position: absolute;
  top: 0;
  right: 0;
  background: #f44336;
  color: #fff;
  font-size: 10px;
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
  font-weight: 600;
}

.theme-toggle-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #000;
  border-radius: 50%;
  transition: background 0.2s, transform 0.3s;
}

.theme-toggle-btn:hover {
  background: rgba(0, 0, 0, 0.1);
  transform: rotate(15deg);
}

[data-theme="dark"] .top-nav {
  background-color: #1A1A1A;
  border-bottom: 1px solid var(--border-color);
}

[data-theme="dark"] .logo-text,
[data-theme="dark"] .nav-link,
[data-theme="dark"] .message-btn,
[data-theme="dark"] .theme-toggle-btn {
  color: var(--text-primary);
}

[data-theme="dark"] .search-box {
  background: var(--bg-secondary);
}

[data-theme="dark"] .search-box input {
  color: var(--text-primary);
}

[data-theme="dark"] .message-btn:hover,
[data-theme="dark"] .theme-toggle-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}
</style>
