<template>
  <div class="home-page">
    <div class="recommend-card">
      <div class="card-content">
        <h2 class="section-title">推荐歌曲</h2>
        <div class="song-list">
          <div v-if="loading" class="loading">加载中...</div>
          <div v-else-if="songs.length === 0" class="empty">暂无推荐内容</div>
          <div v-else class="songs-grid">
            <div v-for="song in songs" :key="song.id" class="song-item">
              <div class="song-info">
                <div class="song-name">{{ song.name }}</div>
                <div class="song-artist">{{ song.artist }}</div>
              </div>
              <div class="song-actions">
                <button class="btn-play" @click="playSong(song)">播放</button>
                <button class="btn-download" @click="downloadSong(song)">下载</button>
              </div>
            </div>
          </div>
        </div>

        <h2 class="section-title" style="margin-top: 32px;">推荐歌单</h2>
        <div class="playlist-list">
          <div v-if="loading" class="loading">加载中...</div>
          <div v-else-if="playlists.length === 0" class="empty">暂无推荐内容</div>
          <div v-else class="playlists-grid">
            <div v-for="playlist in playlists" :key="playlist.id" class="playlist-item">
              <div @click="viewPlaylist(playlist)">
                <div class="playlist-cover">
                  <img v-if="playlist.cover" :src="playlist.cover" :alt="playlist.name" />
                  <div v-else class="placeholder-cover">♪</div>
                </div>
                <div class="playlist-info">
                  <div class="playlist-name">{{ playlist.name }}</div>
                  <div class="playlist-count">{{ playlist.trackCount }} 首</div>
                </div>
              </div>
              <button v-if="!isCollected(playlist.id)" class="collect-btn" @click.stop="collectPlaylist(playlist)">
                收藏
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { recommendAPI, playlistAPI } from '../api'
import { usePlayerStore } from '../stores/player'
import { useAuthStore } from '../stores/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const playerStore = usePlayerStore()
const authStore = useAuthStore()
const loading = ref(true)
const songs = ref<any[]>([])
const playlists = ref<any[]>([])
const userPlaylistIds = ref<Set<number>>(new Set())

const loadRecommendations = async () => {
  loading.value = true
  try {
    // 获取推荐内容
    const [songsRes, playlistsRes] = await Promise.all([
      recommendAPI.getSongs().catch(() => ({ data: { songs: [] } })),
      recommendAPI.getPlaylists().catch(() => ({ data: { playlists: [] } }))
    ])

    songs.value = songsRes.data.songs || []
    playlists.value = playlistsRes.data.playlists || []
    
    // 获取用户歌单列表，用于判断是否已收藏
    if (authStore.isLoggedIn) {
      try {
        const userPlaylistsRes = await playlistAPI.getUserPlaylists()
        const userPlaylists = userPlaylistsRes.data.playlists || []
        userPlaylistIds.value = new Set(userPlaylists.map((p: any) => p.id))
      } catch (e) {
        console.error('获取用户歌单失败:', e)
      }
    }
  } catch (error) {
    console.error('获取推荐内容失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(loadRecommendations)

// 监听网易云登录状态变化，登录后刷新推荐
watch(() => authStore.isLoggedIn, (newVal, oldVal) => {
  if (newVal && !oldVal) {
    // 从网易云未登录变为已登录，刷新推荐
    loadRecommendations()
  }
})

const playSong = (song: any) => {
  playerStore.play({
    id: song.id,
    name: song.name,
    artist: song.artist,
    album: song.album,
    pic_url: song.pic_url,
    duration: song.duration
  })
}

const downloadSong = (song: any) => {
  router.push({ path: '/search', query: { q: song.name } })
}

const viewPlaylist = (playlist: any) => {
  router.push({ path: '/playlist', query: { id: playlist.id } })
}

const isCollected = (playlistId: number) => {
  return userPlaylistIds.value.has(playlistId)
}

const collectPlaylist = async (playlist: any) => {
  try {
    await playlistAPI.subscribePlaylist(playlist.id)
    ElMessage.success(`已收藏歌单「${playlist.name}」`)
    userPlaylistIds.value.add(playlist.id)
  } catch (e: any) {
    console.error('收藏歌单失败:', e)
    ElMessage.error(e?.response?.data?.error || '收藏失败')
  }
}
</script>

<style scoped>
.home-page {
  padding: 24px;
  min-height: 100%;
}

.recommend-card {
  background: var(--card-bg);
  border: 2px solid var(--primary-color);
  border-radius: 32px;
  padding: 48px;
  min-height: 600px;
  box-shadow: 0 8px 32px var(--shadow-color);
}

.card-content {
  color: var(--text-primary);
}

.section-title {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 24px;
  color: var(--text-primary);
}

.loading, .empty {
  text-align: center;
  padding: 40px;
  font-size: 16px;
  opacity: 0.8;
}

.songs-grid {
  display: grid;
  gap: 16px;
}

.song-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: var(--bg-secondary);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  transition: background 0.2s;
}

.song-item:hover {
  background: var(--primary-color);
  color: #000;
}

.song-info {
  flex: 1;
}

.song-name {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 4px;
}

.song-artist {
  font-size: 14px;
  opacity: 0.8;
}

.song-actions {
  display: flex;
  gap: 12px;
}

.btn-play, .btn-download {
  padding: 8px 20px;
  border: none;
  border-radius: 20px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-play {
  background: var(--primary-color);
  color: #000;
}

.btn-play:hover {
  background: var(--primary-hover);
}

.btn-download {
  background: transparent;
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.btn-download:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.playlists-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 24px;
}

.playlist-item {
  cursor: pointer;
  transition: transform 0.2s;
  position: relative;
}

.playlist-item:hover {
  transform: translateY(-4px);
}

.collect-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 6px 12px;
  background: var(--primary-color);
  color: #000;
  border: none;
  border-radius: 16px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  z-index: 10;
}

.collect-btn:hover {
  background: var(--primary-hover);
  transform: scale(1.05);
}

.playlist-cover {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 12px;
  overflow: hidden;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  margin-bottom: 12px;
}

.playlist-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.placeholder-cover {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  opacity: 0.5;
}

.playlist-name {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.playlist-count {
  font-size: 12px;
  opacity: 0.7;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .home-page {
    padding: 16px;
  }

  .recommend-card {
    padding: 24px;
    border-radius: 16px;
    min-height: auto;
  }

  .section-title {
    font-size: 20px;
    margin-bottom: 16px;
  }

  .song-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 16px;
  }

  .song-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .playlists-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 16px;
  }
}
</style>
