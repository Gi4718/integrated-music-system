<template>
  <div class="playlist-page">
    <!-- 左侧歌单列表 -->
    <aside class="playlist-sidebar">
      <div
        v-for="pl in userPlaylists"
        :key="pl.id"
        class="sidebar-item"
        :class="{ active: selectedPlaylist?.id === pl.id }"
        @click="selectPlaylist(pl)"
      >
        <img class="sidebar-cover" :src="pl.cover || defaultCover" :alt="pl.name" />
        <span class="sidebar-name">{{ pl.name }}</span>
      </div>
      <div v-if="!userPlaylists.length" class="sidebar-empty">
        暂无歌单，请先登录
      </div>
    </aside>

    <!-- 右侧歌单详情 -->
    <main class="playlist-main" v-if="selectedPlaylist">
      <!-- 歌单头部（固定） -->
      <div class="playlist-info">
        <img class="playlist-cover-large" :src="selectedPlaylist.cover || defaultCover" />
        <div class="playlist-meta">
          <h2 class="playlist-title">{{ selectedPlaylist.name }}</h2>
          <p class="playlist-stats">
            全部: {{ totalTracks || selectedPlaylist.track_count || 0 }}首
          </p>
          <div class="playlist-actions">
            <button class="action-btn" @click="playAll">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M8 5v14l11-7z"/></svg>
              播放全部列表
            </button>
            <button class="action-btn" @click="syncToLocal">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
              同步到服务器本地
            </button>
            <button class="action-btn" @click="verifyMetadata">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
              验证补全
            </button>
          </div>
        </div>
      </div>

      <!-- 歌曲列表容器（独立滚动） -->
      <div class="tracks-container">
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <p>正在加载歌曲列表，请稍候...</p>
        </div>
        <template v-else>
          <table class="song-table">
            <thead>
              <tr>
                <th style="width:40px">#</th>
                <th>歌曲</th>
                <th>歌手</th>
                <th>专辑</th>
                <th>时长</th>
                <th style="width:120px">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(track, index) in tracks" :key="track.id">
                <td>{{ currentOffset + index + 1 }}</td>
                <td>{{ track.name || '未知歌曲' }}</td>
                <td>{{ track.artist || '未知歌手' }}</td>
                <td>{{ track.album || '未知专辑' }}</td>
                <td>{{ formatDuration(track.duration) }}</td>
                <td>
                  <div class="track-actions">
                    <button class="icon-btn" title="播放" @click="playTrack(track)">
                      <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M8 5v14l11-7z"/></svg>
                    </button>
                    <button class="icon-btn" title="下载" @click="downloadTrack(track)">
                      <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>

          <!-- 分页控制（底部） -->
          <div class="pagination" v-if="totalTracks > 0">
            <div class="pagination-info">
              共 {{ totalTracks }} 首，当前第 {{ currentPage }} / {{ totalPages }} 页（{{ currentOffset + 1 }} - {{ Math.min(currentOffset + tracks.length, totalTracks) }}）
            </div>
            <div class="pagination-controls">
              <button class="page-btn" :disabled="currentPage <= 1" @click="loadPage(1)">首页</button>
              <button class="page-btn" :disabled="currentPage <= 1" @click="loadPage(currentPage - 1)">上一页</button>
              <select v-model.number="pageSize" @change="onPageSizeChange" class="page-size-select">
                <option :value="50">每页50首</option>
                <option :value="100">每页100首</option>
                <option :value="200">每页200首</option>
                <option :value="500">每页500首</option>
              </select>
              <button class="page-btn" :disabled="currentPage >= totalPages" @click="loadPage(currentPage + 1)">下一页</button>
              <button class="page-btn" :disabled="currentPage >= totalPages" @click="loadPage(totalPages)">末页</button>
            </div>
          </div>
        </template>
      </div>
    </main>

    <main class="playlist-main empty-state" v-else>
      <p>请从左侧选择一个歌单</p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { playlistAPI, downloadAPI } from '../api'
import { ElMessage } from 'element-plus'
import { usePlayerStore } from '../stores/player'

const route = useRoute()
const router = useRouter()
const defaultCover = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect fill="%23eee" width="100" height="100"/><text x="50" y="55" text-anchor="middle" font-size="30" fill="%23999">♪</text></svg>'

const playerStore = usePlayerStore()
const userPlaylists = ref<any[]>([])
const selectedPlaylist = ref<any>(null)
const tracks = ref<any[]>([])
const loading = ref(false)

// 分页相关
const totalTracks = ref(0)
const currentPage = ref(1)
const pageSize = ref(parseInt(localStorage.getItem('playlist_page_size') || '100'))

// 保存 pageSize 到 localStorage
watch(pageSize, (newVal) => {
  localStorage.setItem('playlist_page_size', newVal.toString())
})

const currentOffset = computed(() => (currentPage.value - 1) * pageSize.value)
const totalPages = computed(() => Math.ceil(totalTracks.value / pageSize.value) || 1)

let isLoadingTracks = false

const loadPlaylists = async () => {
  try {
    const res = await playlistAPI.getUserPlaylists()
    userPlaylists.value = res.data.data || res.data.playlists || []

    // 从 URL 参数恢复状态
    const playlistId = route.params.id || route.query.id
    const pageParam = route.params.page || route.query.page
    const page = pageParam ? parseInt(pageParam as string) : 1

    if (playlistId) {
      currentPage.value = page
      const targetPlaylist = userPlaylists.value.find((p: any) => String(p.id) === String(playlistId))
      if (targetPlaylist) {
        selectedPlaylist.value = targetPlaylist
        await loadTracks()
      } else {
        // 从推荐歌单跳转过来，直接加载
        await loadPlaylistById(Number(playlistId), page)
      }
      return
    }

    // 默认选择第一个
    if (userPlaylists.value.length) {
      selectedPlaylist.value = userPlaylists.value[0]
      await loadTracks()
    }
  } catch {
    ElMessage.error('获取歌单失败，请先登录')
  }
}

const loadPlaylistById = async (id: number, page: number) => {
  const offset = (page - 1) * pageSize.value
  try {
    const res = await playlistAPI.getPlaylistDetail(id, offset, pageSize.value)
    if (res.data.playlist) {
      selectedPlaylist.value = {
        id: res.data.playlist.id,
        name: res.data.playlist.name,
        cover: res.data.playlist.coverImgUrl,
        track_count: res.data.playlist.trackCount
      }
    }
    tracks.value = res.data.tracks || []
    totalTracks.value = res.data.total || 0
    currentPage.value = page
  } catch {
    ElMessage.error('加载歌单失败')
  }
}

const selectPlaylist = async (pl: any) => {
  selectedPlaylist.value = pl
  currentPage.value = 1
  await loadTracks()
}

const loadTracks = async () => {
  if (!selectedPlaylist.value || isLoadingTracks) return
  isLoadingTracks = true
  loading.value = true
  try {
    const offset = currentOffset.value
    const res = await playlistAPI.getPlaylistDetail(selectedPlaylist.value.id, offset, pageSize.value)
    tracks.value = res.data.tracks || []
    totalTracks.value = res.data.total || 0
    // 合并歌单信息
    if (res.data.playlist) {
      const { tracks: _, ...playlistInfo } = res.data.playlist
      selectedPlaylist.value = { ...selectedPlaylist.value, ...playlistInfo }
    }
    // 更新 URL
    updateURL()
  } catch (e) {
    console.error('加载歌单失败:', e)
    tracks.value = []
  } finally {
    loading.value = false
    isLoadingTracks = false
  }
}

const updateURL = () => {
  if (!selectedPlaylist.value) return
  const newPath = `/playlist/${selectedPlaylist.value.id}/${currentPage.value}`
  if (route.path !== newPath) {
    router.replace({ path: newPath })
  }
}

const loadPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  loadTracks()
  // 滚动到歌曲列表顶部
  const container = document.querySelector('.tracks-container')
  if (container) container.scrollTop = 0
}

const onPageSizeChange = () => {
  currentPage.value = 1
  loadTracks()
}

// 监听路由变化
watch(() => route.params, (newParams) => {
  if (!newParams.id) return
  const id = String(newParams.id)
  const page = parseInt(newParams.page as string) || 1

  if (selectedPlaylist.value && String(selectedPlaylist.value.id) === id) {
    // 同一歌单，页码变化
    if (currentPage.value !== page) {
      currentPage.value = page
      loadTracks()
    }
  } else {
    // 歌单变化
    const target = userPlaylists.value.find((p: any) => String(p.id) === id)
    if (target) {
      selectedPlaylist.value = target
      currentPage.value = page
      loadTracks()
    }
  }
})

const playAll = () => {
  if (tracks.value.length > 0) {
    playerStore.setPlaylist(tracks.value, 0)
    ElMessage.success('开始播放当前页歌曲')
  }
}

const syncToLocal = async () => {
  if (!selectedPlaylist.value) return
  try {
    await downloadAPI.downloadPlaylist(selectedPlaylist.value.id, 'high')
    ElMessage.success(`歌单「${selectedPlaylist.value.name}」已加入下载队列`)
  } catch (e: any) {
    const errorMsg = e?.response?.data?.error || e?.message || '加入下载队列失败'
    ElMessage.error(errorMsg)
  }
}

const playTrack = (track: any) => {
  playerStore.play({
    id: track.id,
    name: track.name || '未知歌曲',
    artist: track.artist || '未知歌手',
    album: track.album || '未知专辑',
    pic_url: track.pic_url || track.al?.picUrl,
    duration: track.duration || track.dt
  })
}

const downloadTrack = async (track: any) => {
  try {
    await downloadAPI.downloadSong(track.id, 'high')
    ElMessage.success(`${track.name} 已加入下载队列`)
  } catch {
    ElMessage.error('下载失败')
  }
}

const formatDuration = (seconds: number) => {
  if (!seconds) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

const verifyMetadata = async () => {
  if (!selectedPlaylist.value) return
  try {
    await downloadAPI.verifyMetadata(selectedPlaylist.value.id)
    ElMessage.success(`歌单「${selectedPlaylist.value.name}」验证补全任务已创建`)
  } catch (e: any) {
    const errorMsg = e?.response?.data?.error || e?.message || '验证补全失败'
    ElMessage.error(errorMsg)
  }
}

onMounted(loadPlaylists)
</script>

<style scoped>
.playlist-page {
  display: flex;
  height: calc(100vh - 120px);
  overflow: hidden;
}

/* 左侧歌单列表（独立滚动） */
.playlist-sidebar {
  width: 220px;
  min-width: 220px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  overflow-y: auto;
  padding: 8px 0;
  flex-shrink: 0;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background 0.15s;
}

.sidebar-item:hover {
  background: var(--border-color);
}

.sidebar-item.active {
  background: var(--primary-color);
}

.sidebar-cover {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  object-fit: cover;
  flex-shrink: 0;
}

.sidebar-name {
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-empty {
  padding: 24px 12px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 13px;
}

/* 右侧主内容 */
.playlist-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 24px 32px;
  min-width: 0;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: 16px;
}

/* 歌单头部信息（固定） */
.playlist-info {
  display: flex;
  gap: 24px;
  align-items: flex-start;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.playlist-cover-large {
  width: 160px;
  height: 160px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
  box-shadow: 0 4px 12px var(--shadow-color);
}

.playlist-meta {
  flex: 1;
}

.playlist-title {
  font-size: 22px;
  font-weight: bold;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.playlist-stats {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.playlist-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: 20px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  background: transparent;
  color: var(--text-primary);
}

.action-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

/* 歌曲列表容器（独立滚动） */
.tracks-container {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

/* 歌曲表格 */
.song-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.song-table thead th {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 2px solid var(--border-color);
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 13px;
  position: sticky;
  top: 0;
  background: var(--bg-primary);
  z-index: 1;
}

.song-table tbody td {
  padding: 10px 8px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
}

.song-table tbody tr:hover {
  background: var(--bg-secondary);
}

.track-actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}

.icon-btn:hover {
  background: var(--primary-color);
  color: #000;
}

/* 分页控制（底部固定） */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-top: 1px solid var(--border-color);
  margin-top: 12px;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 12px;
}

.pagination-info {
  font-size: 13px;
  color: var(--text-secondary);
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  padding: 6px 14px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: transparent;
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-size-select {
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
}

/* 加载状态 */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 移动端适配 */
@media (max-width: 768px) {
  .playlist-page {
    flex-direction: column;
    height: auto;
  }

  .playlist-sidebar {
    width: 100%;
    min-width: 100%;
    max-height: 200px;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
  }

  .playlist-main {
    padding: 16px;
  }

  .playlist-info {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .playlist-cover-large {
    width: 120px;
    height: 120px;
  }

  .playlist-actions {
    justify-content: center;
  }

  .song-table {
    font-size: 12px;
  }

  .song-table thead th,
  .song-table tbody td {
    padding: 8px 4px;
  }

  .song-table th:nth-child(4),
  .song-table td:nth-child(4) {
    display: none;
  }

  .pagination {
    flex-direction: column;
    align-items: stretch;
    text-align: center;
  }

  .pagination-controls {
    justify-content: center;
    flex-wrap: wrap;
  }
}
</style>
