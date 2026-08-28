import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export interface EQBand {
  frequency: number
  gain: number
  type: BiquadFilterType
}

export interface EQPreset {
  name: string
  bands: number[] // 8 gain values in dB
}

// 8个频段
export const EQ_FREQUENCIES = [60, 170, 310, 600, 1000, 3000, 6000, 14000]
export const EQ_LABELS = ['60', '170', '310', '600', '1K', '3K', '6K', '14K']
export const EQ_MIN = -12
export const EQ_MAX = 12

// 内置预设
export const BUILTIN_PRESETS: EQPreset[] = [
  { name: '平坦', bands: [0, 0, 0, 0, 0, 0, 0, 0] },
  { name: '低音增强', bands: [6, 5, 3, 1, 0, 0, 0, 0] },
  { name: '高音增强', bands: [0, 0, 0, 0, 1, 3, 5, 6] },
  { name: '摇滚', bands: [5, 3, -1, -3, 1, 3, 5, 6] },
  { name: '流行', bands: [-1, 2, 4, 5, 3, 0, -1, -2] },
  { name: '爵士', bands: [3, 2, 0, 2, -1, -1, 0, 2] },
  { name: '古典', bands: [4, 3, 2, 1, -1, -1, 0, 2] },
  { name: '电子', bands: [5, 4, 1, 0, -2, 2, 4, 5] },
  { name: '人声', bands: [-2, -1, 2, 5, 5, 3, 1, -1] },
  { name: '低音削减', bands: [-5, -3, -1, 0, 0, 0, 0, 0] },
]

const STORAGE_KEY = 'equalizer_config'
const MAX_CUSTOM = 5

function loadConfig(): { activePreset: string; customSlots: EQPreset[]; enabled: boolean } {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw)
  } catch {}
  return { activePreset: '平坦', customSlots: [], enabled: false }
}

function saveConfig(activePreset: string, customSlots: EQPreset[], enabled: boolean) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({ activePreset, customSlots, enabled }))
}

export const useEqualizerStore = defineStore('equalizer', () => {
  const config = loadConfig()
  const enabled = ref(config.enabled)
  const activePresetName = ref(config.activePreset)
  const customSlots = ref<EQPreset[]>(config.customSlots)

  // Web Audio API nodes
  let audioContext: AudioContext | null = null
  let sourceNode: MediaElementAudioSourceNode | null = null
  let filters: BiquadFilterNode[] = []
  let destinationNode: AudioDestinationNode | null = null
  let connectedAudio: HTMLAudioElement | null = null

  const getCurrentBands = (): number[] => {
    // 检查自定义槽位
    const custom = customSlots.value.find(s => s.name === activePresetName.value)
    if (custom) return [...custom.bands]
    // 检查内置预设
    const builtin = BUILTIN_PRESETS.find(p => p.name === activePresetName.value)
    if (builtin) return [...builtin.bands]
    return [0, 0, 0, 0, 0, 0, 0, 0]
  }

  const initAudioContext = (audio: HTMLAudioElement) => {
    if (audioContext && connectedAudio === audio) return

    // 断开旧连接
    disconnect()

    try {
      audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
      sourceNode = audioContext.createMediaElementSource(audio)
      destinationNode = audioContext.destination

      // 创建8个BiquadFilterNode
      filters = EQ_FREQUENCIES.map((freq, i) => {
        const filter = audioContext!.createBiquadFilter()
        filter.type = i === 0 ? 'lowshelf' : i === EQ_FREQUENCIES.length - 1 ? 'highshelf' : 'peaking'
        filter.frequency.value = freq
        filter.Q.value = 1.4
        filter.gain.value = 0
        return filter
      })

      // 连接: source -> filter0 -> filter1 -> ... -> filter7 -> destination
      let prevNode: AudioNode = sourceNode
      filters.forEach(filter => {
        prevNode.connect(filter)
        prevNode = filter
      })
      prevNode.connect(destinationNode)

      connectedAudio = audio
      // 应用当前预设
      applyCurrentPreset()
    } catch (e) {
      console.error('Failed to init equalizer:', e)
    }
  }

  const disconnect = () => {
    if (sourceNode) {
      try { sourceNode.disconnect() } catch {}
    }
    filters.forEach(f => { try { f.disconnect() } catch {} })
    if (audioContext && audioContext.state !== 'closed') {
      try { audioContext.close() } catch {}
    }
    audioContext = null
    sourceNode = null
    filters = []
    connectedAudio = null
  }

  const applyCurrentPreset = () => {
    const bands = getCurrentBands()
    filters.forEach((filter, i) => {
      if (filter) {
        filter.gain.value = enabled.value ? bands[i] : 0
      }
    })
  }

  const setEnabled = (val: boolean) => {
    enabled.value = val
    applyCurrentPreset()
    saveConfig(activePresetName.value, customSlots.value, enabled.value)
  }

  const selectPreset = (name: string) => {
    activePresetName.value = name
    applyCurrentPreset()
    saveConfig(activePresetName.value, customSlots.value, enabled.value)
  }

  const updateCustomBand = (slotIndex: number, bandIndex: number, gain: number) => {
    if (slotIndex >= 0 && slotIndex < customSlots.value.length) {
      customSlots.value[slotIndex].bands[bandIndex] = gain
      if (activePresetName.value === customSlots.value[slotIndex].name) {
        applyCurrentPreset()
      }
      saveConfig(activePresetName.value, customSlots.value, enabled.value)
    }
  }

  const addCustomSlot = (name: string): boolean => {
    if (customSlots.value.length >= MAX_CUSTOM) return false
    if (customSlots.value.some(s => s.name === name)) return false
    customSlots.value.push({ name, bands: [0, 0, 0, 0, 0, 0, 0, 0] })
    saveConfig(activePresetName.value, customSlots.value, enabled.value)
    return true
  }

  const removeCustomSlot = (index: number) => {
    if (index >= 0 && index < customSlots.value.length) {
      const removed = customSlots.value.splice(index, 1)[0]
      if (activePresetName.value === removed.name) {
        activePresetName.value = '平坦'
        applyCurrentPreset()
      }
      saveConfig(activePresetName.value, customSlots.value, enabled.value)
    }
  }

  const renameCustomSlot = (index: number, newName: string): boolean => {
    if (index >= 0 && index < customSlots.value.length) {
      if (customSlots.value.some((s, i) => i !== index && s.name === newName)) return false
      const oldName = customSlots.value[index].name
      customSlots.value[index].name = newName
      if (activePresetName.value === oldName) {
        activePresetName.value = newName
      }
      saveConfig(activePresetName.value, customSlots.value, enabled.value)
      return true
    }
    return false
  }

  const isCustom = (name: string) => customSlots.value.some(s => s.name === name)

  return {
    enabled,
    activePresetName,
    customSlots,
    initAudioContext,
    disconnect,
    setEnabled,
    selectPreset,
    updateCustomBand,
    addCustomSlot,
    removeCustomSlot,
    renameCustomSlot,
    getCurrentBands,
    isCustom,
  }
})
