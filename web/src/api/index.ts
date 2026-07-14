import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截器：添加认证token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('system_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器：处理401未授权（仅系统接口401才清除登录状态）
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const url = error.config?.url || ''
      if (url.startsWith('/system/') || url.startsWith('/auth/')) {
        localStorage.removeItem('system_token')
        localStorage.removeItem('system_username')
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

// 认证相关
export const authAPI = {
  getQRKey() {
    return api.get('/auth/qr-key')
  },
  getQRCode(key: string) {
    return api.get(`/auth/qr-code?key=${key}`)
  },
  checkQRStatus(key: string) {
    return api.get(`/auth/qr-check?key=${key}`)
  },
  saveCookie(cookie: string) {
    return api.post('/auth/save-cookie', { cookie })
  },
  getLoginStatus() {
    return api.get('/auth/status')
  },
  logout() {
    return api.post('/auth/logout')
  },
  sendSMSCode(phone: string) {
    return api.post('/auth/sms/send', { phone })
  },
  loginByPhone(phone: string, password: string) {
    return api.post('/auth/phone', { phone, captcha: password })
  },
  loginByPhonePassword(phone: string, password: string) {
    return api.post('/auth/phone/password', { phone, password })
  },
  secondVerify(code: string) {
    return api.post('/auth/second-verify', { code })
  },
  loginByEmail(email: string, password: string) {
    return api.post('/auth/email', { email, password })
  },
  loginByQQ(code: string) {
    return api.post('/auth/qq', { code })
  },
  getQQAuthURL() {
    return api.get('/auth/qq/url')
  }
}

// 搜索相关
export const searchAPI = {
  searchSongs(keyword: string, limit = 30, offset = 0) {
    return api.get('/search/songs', { params: { keyword, limit, offset } })
  }
}

// 下载相关
export const downloadAPI = {
  downloadSong(songId: number, quality: string) {
    return api.post('/download/song', { song_id: songId, quality })
  },
  downloadPlaylist(playlistId: number, quality: string) {
    return api.post('/download/playlist', { playlist_id: playlistId, quality })
  },
  getHistory() {
    return api.get('/download/history')
  },
  getProgress() {
    return api.get('/download/progress')
  },
  verifyMetadata(playlistId: number) {
    return api.post('/download/verify-metadata', { playlist_id: playlistId })
  }
}

// 歌单相关
export const playlistAPI = {
  getUserPlaylists() {
    return api.get('/playlist/user')
  },
  getPlaylistDetail(playlistId: number) {
    return api.get(`/playlist/detail?id=${playlistId}`)
  },
  subscribePlaylist(playlistId: number) {
    return api.post('/playlist/subscribe', { playlist_id: playlistId })
  }
}

// 播放器相关
export const playerAPI = {
  getStreamUrl(songId: number, token: string = '') {
    return token ? `/api/player/stream/${songId}?token=${token}` : `/api/player/stream/${songId}`
  },
  checkAvailable(songId: number) {
    return api.get(`/player/check/${songId}`)
  }
}

// 设置相关
export const settingsAPI = {
  getSettings() {
    return api.get('/settings')
  },
  updateSettings(settings: any) {
    return api.post('/settings', settings)
  },
  uploadSSLCert(cert: File, key: File) {
    const formData = new FormData()
    formData.append('cert', cert)
    formData.append('key', key)
    return api.post('/settings/ssl/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  uploadSSLCertFile(formData: FormData) {
    return api.post('/settings/ssl/upload-file', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  validateSSLCert(config: { cert_path: string; key_path: string }) {
    return api.post('/settings/ssl/validate', config)
  },
  applyACME(config: any) {
    return api.post('/settings/ssl/acme', config)
  },
  getACMEPlugins() {
    return api.get('/settings/ssl/acme-plugins')
  }
}

// 任务相关
export const taskAPI = {
  getTasks() {
    return api.get('/tasks')
  },
  getTaskProgress(taskId: string) {
    return api.get(`/tasks/${taskId}/progress`)
  },
  cancelTask(taskId: string) {
    return api.post(`/tasks/${taskId}/cancel`)
  }
}

// 系统认证相关
export const systemAPI = {
  register(data: { username: string; password: string }) {
    return api.post('/system/register', data)
  },
  login(data: { username: string; password: string }) {
    return api.post('/system/login', data)
  },
  check() {
    return api.get('/system/check')
  }
}

// 推荐相关
export const recommendAPI = {
  getSongs() {
    return api.get('/recommend/songs')
  },
  getPlaylists() {
    return api.get('/recommend/playlists')
  }
}

export default api
