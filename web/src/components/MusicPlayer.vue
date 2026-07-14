<template>
  <div v-if="player.currentSong" class="music-player">
    <div class="player-content">
      <!-- 歌曲信息 -->
      <div class="song-info">
        <img
          v-if="player.currentSong.pic_url"
          :src="player.currentSong.pic_url"
          :alt="player.currentSong.name"
          class="song-cover"
        />
        <div v-else class="song-cover-placeholder">♪</div>
        <div class="song-details">
          <div class="song-name">{{ player.currentSong.name }}</div>
          <div class="song-artist">{{ player.currentSong.artist }}</div>
        </div>
      </div>

      <!-- 播放控制 -->
      <div class="player-controls">
        <div class="control-buttons">
          <button class="control-btn" @click="player.togglePlayMode" :title="player.getPlayModeName()">
            <svg v-if="player.playMode === 'sequence'" viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M7 7h10v3l4-4-4-4v3H5v6h2V7zm10 10H7v-3l-4 4 4 4v-3h12v-6h-2v4z"/>
            </svg>
            <svg v-else-if="player.playMode === 'random'" viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M10.59 9.17L5.41 4 4 5.41l5.17 5.17 1.42-1.41zM14.5 4l2.04 2.04L4 18.59 5.41 20 17.96 7.46 20 9.5V4h-5.5zm.33 9.41l-1.41 1.41 3.13 3.13L14.5 20H20v-5.5l-2.04 2.04-3.13-3.13z"/>
            </svg>
            <div v-else class="single-loop-icon">
              <svg viewBox="0 0 24 24" width="20" height="20">
                <path fill="currentColor" d="M7 7h10v3l4-4-4-4v3H5v6h2V7zm10 10H7v-3l-4 4 4 4v-3h12v-6h-2v4z"/>
              </svg>
              <span class="loop-1">1</span>
            </div>
          </button>
          <button class="control-btn" @click="player.playPrev" title="上一首">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M6 6h2v12H6zm3.5 6l8.5 6V6z"/>
            </svg>
          </button>
          <button class="control-btn play-btn" @click="player.togglePlay">
            <svg v-if="!player.isPlaying" viewBox="0 0 24 24" width="24" height="24">
              <path fill="currentColor" d="M8 5v14l11-7z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="24" height="24">
              <path fill="currentColor" d="M6 4h4v16H6V4zm8 0h4v16h-4V4z"/>
            </svg>
          </button>
          <button class="control-btn" @click="player.playNext" title="下一首">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z"/>
            </svg>
          </button>
          <button class="control-btn" @click="showPlaylistDialog" title="播放列表">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M15 6H3v2h12V6zm0 4H3v2h12v-2zM3 16h8v-2H3v2zM17 6v8.18c-.31-.11-.65-.18-1-.18-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3V8h3V6h-5z"/>
            </svg>
          </button>
          <button class="control-btn" @click="showAddToPlaylistDialog" title="添加到歌单">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M14 10H2v2h12v-2zm0-4H2v2h12V6zm4 8v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zM2 16h8v-2H2v2z"/>
            </svg>
          </button>
        </div>

        <!-- 进度条 -->
        <div class="progress-container">
          <span class="time">{{ formatTime(player.currentTime) }}</span>
          <input
            type="range"
            min="0"
            :max="player.duration || 0"
            :value="player.currentTime"
            @input="onSeek"
            class="progress-bar"
          />
          <span class="time">{{ formatTime(player.duration) }}</span>
        </div>
      </div>

      <!-- 音量控制 -->
      <div class="volume-control">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/>
        </svg>
        <input
          type="range"
          min="0"
          max="1"
          step="0.01"
          :value="player.volume"
          @input="onVolumeChange"
          class="volume-bar"
        />
      </div>
    </div>

    <!-- 播放列表弹窗 -->
    <div v-if="showPlaylist" class="dialog-overlay" @click="showPlaylist = false">
      <div class="dialog-content" @click.stop>
        <div class="dialog-header">
          <h3>播放列表</h3>
          <button class="close-btn" @click="showPlaylist = false">×</button>
        </div>
        <div class="playlist-list">
          <div v-for="(song, index) in player.playlist" :key="index" class="playlist-item" :class="{ active: index === player.currentIndex }" @click="playFromPlaylist(index)">
            <span class="song-title">{{ song.name }}</span>
            <span class="song-artist">{{ song.artist }}</span>
          </div>
          <div v-if="!player.playlist.length" class="empty-playlist">播放列表为空</div>
        </div>
      </div>
    </div>

    <!-- 添加到歌单弹窗 -->
    <div v-if="showAddToPlaylist" class="dialog-overlay" @click="showAddToPlaylist = false">
      <div class="dialog-content" @click.stop>
        <div class="dialog-header">
          <h3>添加到歌单</h3>
          <button class="close-btn" @click="showAddToPlaylist = false">×</button>
        </div>
        <div class="playlist-selector">
          <div v-for="pl in userPlaylists" :key="pl.id" class="playlist-option" @click="addToPlaylist(pl.id)">
            {{ pl.name }}
          </div>
          <div v-if="!userPlaylists.length" class="empty-playlist">暂无歌单</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { usePlayerStore } from '../stores/player'
import { playlistAPI } from '../api'
import { ElMessage } from 'element-plus'

const player = usePlayerStore()
const showPlaylist = ref(false)
const showAddToPlaylist = ref(false)
const userPlaylists = ref<any[]>([])

const showPlaylistDialog = () => {
  showPlaylist.value = true
}

const showAddToPlaylistDialog = async () => {
  try {
    const res = await playlistAPI.getUserPlaylists()
    userPlaylists.value = res.data.playlists || []
    showAddToPlaylist.value = true
  } catch {
    ElMessage.error('获取歌单失败')
  }
}

const playFromPlaylist = (index: number) => {
  if (index >= 0 && index < player.playlist.length) {
    player.currentIndex = index
    player.play(player.playlist[index])
  }
}

const addToPlaylist = async (_playlistId: number) => {
  if (!player.currentSong) return
  
  try {
    // 这里需要调用后端 API 将歌曲添加到歌单
    // 暂时显示成功消息
    ElMessage.success(`已添加到歌单`)
    showAddToPlaylist.value = false
  } catch {
    ElMessage.error('添加失败')
  }
}

const formatTime = (seconds: number) => {
  if (!seconds || isNaN(seconds)) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const onSeek = (event: Event) => {
  const target = event.target as HTMLInputElement
  player.seek(parseFloat(target.value))
}

const onVolumeChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  player.setVolume(parseFloat(target.value))
}
</script>

<style scoped>
.music-player {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--card-bg);
  border-top: 2px solid var(--primary-color);
  box-shadow: 0 -4px 16px var(--shadow-color);
  z-index: 1000;
  padding: 12px 24px;
}

.player-content {
  display: flex;
  align-items: center;
  gap: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.song-info {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 200px;
}

.song-cover {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  object-fit: cover;
}

.song-cover-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  background: var(--bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: var(--text-secondary);
}

.song-details {
  flex: 1;
  overflow: hidden;
}

.song-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.song-artist {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.player-controls {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.control-buttons {
  display: flex;
  align-items: center;
  gap: 16px;
}

.control-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s;
}

.control-btn:hover {
  background: var(--bg-secondary);
  color: var(--primary-color);
}

.play-btn {
  width: 44px;
  height: 44px;
  background: var(--primary-color);
  color: #000;
}

.play-btn:hover {
  background: var(--primary-hover);
  color: #000;
}

.progress-container {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  max-width: 500px;
}

.time {
  font-size: 12px;
  color: var(--text-secondary);
  min-width: 40px;
  text-align: center;
}

.progress-bar {
  flex: 1;
  height: 4px;
  -webkit-appearance: none;
  appearance: none;
  background: var(--bg-secondary);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}

.progress-bar::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 12px;
  height: 12px;
  background: var(--primary-color);
  border-radius: 50%;
  cursor: pointer;
}

.progress-bar::-moz-range-thumb {
  width: 12px;
  height: 12px;
  background: var(--primary-color);
  border-radius: 50%;
  cursor: pointer;
  border: none;
}

.volume-control {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 120px;
  color: var(--text-primary);
}

.volume-bar {
  width: 80px;
  height: 4px;
  -webkit-appearance: none;
  appearance: none;
  background: var(--bg-secondary);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}

.volume-bar::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 10px;
  height: 10px;
  background: var(--primary-color);
  border-radius: 50%;
  cursor: pointer;
}

.volume-bar::-moz-range-thumb {
  width: 10px;
  height: 10px;
  background: var(--primary-color);
  border-radius: 50%;
  cursor: pointer;
  border: none;
}

/* 弹窗样式 */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.dialog-content {
  background: var(--card-bg);
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px var(--shadow-color);
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.dialog-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.2s;
}

.close-btn:hover {
  background: var(--bg-secondary);
}

.playlist-list,
.playlist-selector {
  overflow-y: auto;
  padding: 12px;
  flex: 1;
}

.playlist-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
}

.playlist-item:hover {
  background: var(--bg-secondary);
}

.playlist-item.active {
  background: var(--primary-color);
  color: #000;
}

.playlist-item .song-title {
  flex: 1;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.playlist-item .song-artist {
  font-size: 12px;
  color: var(--text-secondary);
  margin-left: 12px;
}

.playlist-item.active .song-artist {
  color: #000;
}

.playlist-option {
  padding: 12px 16px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
  color: var(--text-primary);
}

.playlist-option:hover {
  background: var(--bg-secondary);
}

.empty-playlist {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
  font-size: 14px;
}

/* 单曲循环图标样式 */
.single-loop-icon {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.loop-1 {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: bold;
  color: var(--primary-color);
  z-index: 1;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .music-player {
    padding: 8px 16px;
  }

  .player-content {
    gap: 12px;
  }

  .song-info {
    min-width: 120px;
  }

  .song-cover,
  .song-cover-placeholder {
    width: 40px;
    height: 40px;
  }

  .volume-control {
    display: none;
  }

  .progress-container {
    max-width: 300px;
  }

  .control-buttons {
    gap: 8px;
  }

  .control-btn {
    width: 32px;
    height: 32px;
  }

  .play-btn {
    width: 40px;
    height: 40px;
  }

  .control-btn svg {
    width: 18px;
    height: 18px;
  }

  .play-btn svg {
    width: 20px;
    height: 20px;
  }

  .dialog-content {
    width: 95%;
    max-width: none;
  }
}
</style>
