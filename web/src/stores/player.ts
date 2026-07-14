import { defineStore } from 'pinia'
import { ref } from 'vue'
import { playerAPI } from '../api'
import { ElMessage } from 'element-plus'

interface Song {
  id: number
  name: string
  artist: string
  album: string
  pic_url?: string
  duration?: number
}

type PlayMode = 'sequence' | 'random' | 'single'

export const usePlayerStore = defineStore('player', () => {
  const currentSong = ref<Song | null>(null)
  const isPlaying = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)
  const volume = ref(0.7)
  const audio = ref<HTMLAudioElement | null>(null)
  
  // 播放列表相关
  const playlist = ref<Song[]>([])
  const currentIndex = ref(-1)
  const playMode = ref<PlayMode>('sequence')

  const play = async (song: Song) => {
    if (!audio.value) {
      audio.value = new Audio()
      audio.value.volume = volume.value

      audio.value.addEventListener('timeupdate', () => {
        currentTime.value = audio.value?.currentTime || 0
      })

      audio.value.addEventListener('loadedmetadata', () => {
        duration.value = audio.value?.duration || 0
      })

      audio.value.addEventListener('ended', () => {
        isPlaying.value = false
      })

      audio.value.addEventListener('error', () => {
        console.error('音频加载错误, src:', audio.value?.src)
        ElMessage.error('音频加载失败，可能是版权限制或网络问题')
        isPlaying.value = false
      })
    }

    currentSong.value = song
    duration.value = 0
    currentTime.value = 0
    isPlaying.value = false
    const token = localStorage.getItem('system_token') || ''
    audio.value.src = playerAPI.getStreamUrl(song.id, token)

    try {
      await new Promise<void>((resolve, reject) => {
        const onCanPlay = () => {
          audio.value!.removeEventListener('canplay', onCanPlay)
          audio.value!.removeEventListener('error', onError)
          resolve()
        }
        const onError = () => {
          audio.value!.removeEventListener('canplay', onCanPlay)
          audio.value!.removeEventListener('error', onError)
          reject(new Error('音频加载失败'))
        }
        audio.value!.addEventListener('canplay', onCanPlay)
        audio.value!.addEventListener('error', onError)
        audio.value!.load()
      })
      await audio.value.play()
      isPlaying.value = true
    } catch (error: any) {
      console.error('播放失败:', error)
      ElMessage.error(error?.message || '播放失败，请检查网络或歌曲版权')
      isPlaying.value = false
    }
  }

  const togglePlay = async () => {
    if (!audio.value || !currentSong.value) return

    if (isPlaying.value) {
      audio.value.pause()
      isPlaying.value = false
    } else {
      try {
        await audio.value.play()
        isPlaying.value = true
      } catch (error: any) {
        console.error('播放失败:', error)
        ElMessage.error(error?.message || '播放失败，请检查网络或歌曲版权')
      }
    }
  }

  const seek = (time: number) => {
    if (audio.value) {
      audio.value.currentTime = time
      currentTime.value = time
    }
  }

  const setVolume = (vol: number) => {
    volume.value = vol
    if (audio.value) {
      audio.value.volume = vol
    }
  }

  const stop = () => {
    if (audio.value) {
      audio.value.pause()
      audio.value.currentTime = 0
      isPlaying.value = false
    }
    currentSong.value = null
  }

  // 播放列表相关方法
  const setPlaylist = (songs: Song[], startIndex: number = 0) => {
    playlist.value = songs
    currentIndex.value = startIndex
    if (songs.length > 0 && startIndex >= 0 && startIndex < songs.length) {
      play(songs[startIndex])
    }
  }

  const addToPlaylist = (song: Song) => {
    playlist.value.push(song)
  }

  const playNext = () => {
    if (playlist.value.length === 0) return
    
    let nextIndex = currentIndex.value + 1
    
    if (playMode.value === 'random') {
      nextIndex = Math.floor(Math.random() * playlist.value.length)
    } else if (playMode.value === 'single') {
      nextIndex = currentIndex.value
    } else {
      // sequence
      if (nextIndex >= playlist.value.length) {
        nextIndex = 0
      }
    }
    
    currentIndex.value = nextIndex
    if (nextIndex < playlist.value.length) {
      play(playlist.value[nextIndex])
    }
  }

  const playPrev = () => {
    if (playlist.value.length === 0) return
    
    let prevIndex = currentIndex.value - 1
    
    if (playMode.value === 'random') {
      prevIndex = Math.floor(Math.random() * playlist.value.length)
    } else if (prevIndex < 0) {
      prevIndex = playlist.value.length - 1
    }
    
    currentIndex.value = prevIndex
    if (prevIndex >= 0) {
      play(playlist.value[prevIndex])
    }
  }

  const togglePlayMode = () => {
    const modes: PlayMode[] = ['sequence', 'random', 'single']
    const currentModeIndex = modes.indexOf(playMode.value)
    const nextModeIndex = (currentModeIndex + 1) % modes.length
    playMode.value = modes[nextModeIndex]
    
    const modeNames = {
      sequence: '顺序播放',
      random: '随机播放',
      single: '单曲循环'
    }
    ElMessage.success(modeNames[playMode.value])
  }

  const getPlayModeName = () => {
    const modeNames = {
      sequence: '顺序播放',
      random: '随机播放',
      single: '单曲循环'
    }
    return modeNames[playMode.value]
  }

  return {
    currentSong,
    isPlaying,
    currentTime,
    duration,
    volume,
    playlist,
    currentIndex,
    playMode,
    play,
    togglePlay,
    seek,
    setVolume,
    stop,
    setPlaylist,
    addToPlaylist,
    playNext,
    playPrev,
    togglePlayMode,
    getPlayModeName
  }
})
