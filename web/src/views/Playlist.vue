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
      <!-- 歌单头部 -->
      <div class="playlist-info">
        <img class="playlist-cover-large" :src="selectedPlaylist.cover || defaultCover" />
        <div class="playlist-meta">
          <h2 class="playlist-title">{{ selectedPlaylist.name }}</h2>
          <p class="playlist-stats">
            全部: {{ selectedPlaylist.trackCount || 0 }}首
          </p>
          <div class="playlist-actions">
            <button class="action-btn" @click="playAll">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M8 5v14l11-7z"/></svg>
              播放全部列表
            </button>
            <button v-if="!isPlaylistCollected" class="action-btn collect-btn" @click="collectCurrentPlaylist">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
              收藏歌单
            </button>
            <button v-else class="action-btn collected-btn" disabled>
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
              已收藏
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

      <!-- 歌曲表格 -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>正在加载歌曲列表，请稍候...</p>
      </div>
      <table class="song-table" v-else>
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
            <td>{{ index + 1 }}</td>
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
    </main>

    <main class="playlist-main empty-state" v-else>
      <p>请从左侧选择一个歌单</p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { playlistAPI, downloadAPI } from '../api'
import { ElMessage } from 'element-plus'
import { usePlayerStore } from '../stores/player'

const route = useRoute()
const defaultCover = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect fill="%23eee" width="100" height="100"/><text x="50" y="55" text-anchor="middle" font-size="30" fill="%23999">♪</text></svg>'

const playerStore = usePlayerStore()
const userPlaylists = ref<any[]>([])
const selectedPlaylist = ref<any>(null)
const tracks = ref<any[]>([])
const loading = ref(false)

// 判断当前歌单是否已收藏（在用户歌单列表中）
const isPlaylistCollected = computed(() => {
  if (!selectedPlaylist.value) return false
  return userPlaylists.value.some(p => String(p.id) === String(selectedPlaylist.value.id))
})

const loadPlaylists = async () => {
  try {
    const res = await playlistAPI.getUserPlaylists()
    userPlaylists.value = res.data.playlists || []

    // 检查 URL 参数中是否有歌单 ID
    const playlistId = route.query.id
    if (playlistId) {
      // 先检查是否在用户歌单列表中
      const targetPlaylist = userPlaylists.value.find(p => String(p.id) === String(playlistId))
      if (targetPlaylist) {
        selectPlaylist(targetPlaylist)
      } else {
        // 如果不在用户歌单列表中（比如从推荐歌单跳转过来），直接加载歌单详情
        const detailRes = await playlistAPI.getPlaylistDetail(Number(playlistId))
        if (detailRes.data.playlist) {
          const playlist = {
            id: detailRes.data.playlist.id,
            name: detailRes.data.playlist.name,
            cover: detailRes.data.playlist.coverImgUrl,
            trackCount: detailRes.data.playlist.trackCount
          }
          // 不添加到用户歌单列表，只设置 selectedPlaylist
          selectedPlaylist.value = playlist
          // 加载歌曲列表
          tracks.value = detailRes.data.tracks || []
        }
      }
      return
    }

    // 默认选择第一个
    if (userPlaylists.value.length) {
      selectPlaylist(userPlaylists.value[0])
    }
  } catch {
    ElMessage.error('获取歌单失败，请先登录')
  }
}

const selectPlaylist = async (pl: any) => {
  selectedPlaylist.value = pl
  loading.value = true
  try {
    const res = await playlistAPI.getPlaylistDetail(pl.id)
    tracks.value = res.data.tracks || []
    // 合并歌单信息（保留已处理的 tracks）
    if (res.data.playlist) {
      const { tracks: _, ...playlistInfo } = res.data.playlist
      selectedPlaylist.value = { ...pl, ...playlistInfo }
    }
  } catch {
    tracks.value = []
  } finally {
    loading.value = false
  }
}

const playAll = () => {
  if (tracks.value.length > 0) {
    playerStore.setPlaylist(tracks.value, 0)
    ElMessage.success('开始播放歌单')
  }
}

const syncToLocal = async () => {
  if (!selectedPlaylist.value) {
    ElMessage.warning('请先选择歌单')
    return
  }
  try {
    const res = await downloadAPI.downloadPlaylist(selectedPlaylist.value.id, 'high')
    console.log('downloadPlaylist response:', res)
    ElMessage.success(`歌单「${selectedPlaylist.value.name}」已加入下载队列`)
  } catch (e: any) {
    console.error('downloadPlaylist error:', e)
    console.error('response data:', e?.response?.data)
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
  if (!selectedPlaylist.value) {
    ElMessage.warning('请先选择歌单')
    return
  }
  try {
    const res = await downloadAPI.verifyMetadata(selectedPlaylist.value.id)
    ElMessage.success(`歌单「${selectedPlaylist.value.name}」验证补全任务已创建`)
    console.log('verifyMetadata response:', res)
  } catch (e: any) {
    console.error('verifyMetadata error:', e)
    const errorMsg = e?.response?.data?.error || e?.message || '验证补全失败'
    ElMessage.error(errorMsg)
  }
}

const collectCurrentPlaylist = async () => {
  if (!selectedPlaylist.value) return
  try {
    await playlistAPI.subscribePlaylist(selectedPlaylist.value.id)
    ElMessage.success(`已收藏歌单「${selectedPlaylist.value.name}」`)
    // 刷新用户歌单列表
    await loadPlaylists()
  } catch (e: any) {
    console.error('收藏歌单失败:', e)
    ElMessage.error(e?.response?.data?.error || '收藏失败')
  }
}

onMounted(loadPlaylists)
</script>

<style scoped>
.playlist-page {
  display: flex;
  min-height: 100%;
}

/* 左侧歌单列表 */
.playlist-sidebar {
  width: 220px;
  min-width: 220px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  overflow-y: auto;
  padding: 8px 0;
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
  overflow-y: auto;
  padding: 24px 32px;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: 16px;
}

/* 歌单头部信息 */
.playlist-info {
  display: flex;
  gap: 24px;
  align-items: flex-start;
  margin-bottom: 32px;
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
  gap: 16px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 20px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  background: transparent;
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.action-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.collect-btn {
  border-color: #ff6b6b;
  color: #ff6b6b;
}

.collect-btn:hover {
  background: #ff6b6b;
  color: #fff;
  border-color: #ff6b6b;
}

.collected-btn {
  border-color: #51cf66;
  color: #51cf66;
  cursor: default;
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
}

.song-table tbody td {
  padding: 10px 8px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
}

.song-table tbody tr:hover {
  background: var(--bg-secondary);
}

.status-icon {
  display: flex;
  align-items: center;
  color: var(--text-secondary);
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

  .sidebar-item {
    padding: 6px 12px;
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

  .playlist-title {
    font-size: 18px;
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

  .song-table th:nth-child(3),
  .song-table td:nth-child(3),
  .song-table th:nth-child(4),
  .song-table td:nth-child(4) {
    display: none;
  }
}
</style>
