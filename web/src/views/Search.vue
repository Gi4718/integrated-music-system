<template>
  <div class="search-container">
    <div class="search-header">
      <h2>音乐检索</h2>
      <span class="subtitle">SEARCH MUSIC</span>
    </div>

    <div class="search-box">
      <input
        v-model="keyword"
        class="search-input"
        placeholder="输入关键词搜索..."
        @keyup.enter="handleSearch"
      />
      <button class="search-btn" @click="handleSearch" :disabled="searching">
        {{ searching ? '检索中...' : '检索' }}
      </button>
    </div>

    <div v-if="results.length" class="results">
      <table class="result-table">
        <thead>
          <tr>
            <th style="width:40px">#</th>
            <th>歌曲</th>
            <th>歌手</th>
            <th>专辑</th>
            <th style="width:80px">时长</th>
            <th style="width:120px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(song, index) in results" :key="song.id">
            <td>{{ index + 1 }}</td>
            <td>{{ song.name }}</td>
            <td>{{ song.artist }}</td>
            <td>{{ song.album }}</td>
            <td>{{ formatDuration(song.duration) }}</td>
            <td>
              <div class="track-actions">
                <button class="icon-btn" title="播放" @click="playSong(song)">
                  <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M8 5v14l11-7z"/></svg>
                </button>
                <button class="icon-btn" title="下载" @click="downloadSong(song)">
                  <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else-if="searched" class="empty-state">未找到相关曲目</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { searchAPI, downloadAPI } from '../api'
import { ElMessage } from 'element-plus'
import { usePlayerStore } from '../stores/player'

const route = useRoute()
const playerStore = usePlayerStore()
const keyword = ref((route.query.q as string) || '')
const results = ref<any[]>([])
const searching = ref(false)
const searched = ref(false)

const handleSearch = async () => {
  if (!keyword.value) return

  searching.value = true
  searched.value = false

  try {
    const res = await searchAPI.searchSongs(keyword.value)
    results.value = res.data.songs || []
    searched.value = true
  } catch {
    ElMessage.error('搜索失败')
  } finally {
    searching.value = false
  }
}

const playSong = (song: any) => {
  playerStore.play({
    id: song.id,
    name: song.name || '未知歌曲',
    artist: song.artist || '未知歌手',
    album: song.album || '未知专辑',
    pic_url: song.pic_url || song.al?.picUrl,
    duration: song.duration || song.dt
  })
}

const downloadSong = async (song: any) => {
  try {
    await downloadAPI.downloadSong(song.id, 'high')
    ElMessage.success('已加入下载队列')
  } catch {
    ElMessage.error('下载失败')
  }
}

const formatDuration = (ms: number) => {
  if (!ms) return '--:--'
  // 网易云 API 返回的是毫秒，需要转换为秒
  const totalSeconds = Math.floor(ms / 1000)
  const m = Math.floor(totalSeconds / 60)
  const s = totalSeconds % 60
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

onMounted(() => {
  if (keyword.value) handleSearch()
})
</script>

<style scoped>
.search-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px 32px;
  min-height: 100%;
}

.search-header {
  display: flex;
  align-items: baseline;
  gap: 1rem;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid var(--primary-color);
}

.search-header h2 {
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--text-primary);
}

.search-header .subtitle {
  font-size: 0.9rem;
  color: var(--text-secondary);
  letter-spacing: 0.2em;
}

.search-box {
  display: flex;
  gap: 12px;
  margin-bottom: 2rem;
}

.search-input {
  flex: 1;
  padding: 10px 16px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
}

.search-input:focus {
  border-color: #FFFA00;
}

.search-btn {
  padding: 10px 24px;
  background: #FFFA00;
  color: #000;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.search-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.result-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.result-table thead th {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 2px solid var(--border-color);
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 13px;
}

.result-table tbody td {
  padding: 10px 8px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
}

.result-table tbody tr:hover {
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
  background: #FFFA00;
  color: #000;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
  font-size: 16px;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .search-container {
    padding: 16px;
  }

  .search-header h2 {
    font-size: 1.5rem;
  }

  .search-box {
    flex-direction: column;
    gap: 8px;
  }

  .search-btn {
    width: 100%;
  }

  .result-table {
    font-size: 12px;
  }

  .result-table thead th,
  .result-table tbody td {
    padding: 8px 4px;
  }

  .result-table th:nth-child(3),
  .result-table td:nth-child(3),
  .result-table th:nth-child(4),
  .result-table td:nth-child(4) {
    display: none;
  }
}
</style>
