<template>
  <div class="eq-panel">
    <div class="eq-header">
      <span class="eq-title">均衡器</span>
      <label class="eq-toggle">
        <input type="checkbox" :checked="eq.enabled" @change="eq.setEnabled(!eq.enabled)" />
        <span class="eq-toggle-slider"></span>
      </label>
    </div>

    <!-- 预设按钮 -->
    <div class="eq-section">
      <div class="eq-section-label">预设</div>
      <div class="eq-presets">
        <button
          v-for="preset in BUILTIN_PRESETS"
          :key="preset.name"
          class="eq-preset-btn"
          :class="{ active: eq.activePresetName === preset.name }"
          @click="eq.selectPreset(preset.name)"
        >{{ preset.name }}</button>
      </div>
    </div>

    <!-- 自定义槽位 -->
    <div class="eq-section">
      <div class="eq-section-label">
        自定义 ({{ eq.customSlots.length }}/5)
      </div>
      <div class="eq-presets">
        <div v-for="(slot, si) in eq.customSlots" :key="si" class="eq-custom-item">
          <button
            class="eq-preset-btn custom"
            :class="{ active: eq.activePresetName === slot.name }"
            @click="eq.selectPreset(slot.name)"
          >
            <span v-if="editingIndex !== si">{{ slot.name }}</span>
            <input
              v-else
              v-model="editingName"
              class="eq-rename-input"
              @keyup.enter="confirmRename(si)"
              @blur="confirmRename(si)"
              @click.stop
              ref="renameInput"
            />
          </button>
          <button class="eq-icon-btn" @click.stop="startRename(si)" title="重命名">
            <svg viewBox="0 0 24 24" width="14" height="14"><path fill="currentColor" d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/></svg>
          </button>
          <button class="eq-icon-btn danger" @click.stop="eq.removeCustomSlot(si)" title="删除">
            <svg viewBox="0 0 24 24" width="14" height="14"><path fill="currentColor" d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <button
          v-if="eq.customSlots.length < 5"
          class="eq-preset-btn add-btn"
          @click="addNewSlot"
        >+ 新建</button>
      </div>
    </div>

    <!-- 频段滑块 -->
    <div class="eq-sliders" :class="{ disabled: !eq.enabled }">
      <div v-for="(freq, i) in EQ_FREQUENCIES" :key="freq" class="eq-slider-col">
        <span class="eq-slider-value">{{ currentBands[i] > 0 ? '+' : '' }}{{ currentBands[i] }}</span>
        <div class="eq-slider-track">
          <input
            type="range"
            :min="EQ_MIN"
            :max="EQ_MAX"
            :value="currentBands[i]"
            @input="onBandChange(i, $event)"
            class="eq-slider"
            :disabled="!eq.enabled || !isCurrentCustom"
          />
          <div class="eq-slider-center"></div>
        </div>
        <span class="eq-slider-label">{{ EQ_LABELS[i] }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useEqualizerStore, BUILTIN_PRESETS, EQ_FREQUENCIES, EQ_LABELS, EQ_MIN, EQ_MAX } from '../stores/equalizer'

const eq = useEqualizerStore()

const editingIndex = ref(-1)
const editingName = ref('')
const renameInput = ref<HTMLInputElement[]>()

const currentBands = computed(() => eq.getCurrentBands())
const isCurrentCustom = computed(() => eq.isCustom(eq.activePresetName))

const onBandChange = (bandIndex: number, event: Event) => {
  const val = parseFloat((event.target as HTMLInputElement).value)
  // 找到当前自定义槽位的索引
  const slotIdx = eq.customSlots.findIndex(s => s.name === eq.activePresetName)
  if (slotIdx >= 0) {
    eq.updateCustomBand(slotIdx, bandIndex, val)
  }
}

const startRename = async (index: number) => {
  editingIndex.value = index
  editingName.value = eq.customSlots[index].name
  await nextTick()
  if (renameInput.value && renameInput.value[0]) {
    renameInput.value[0].focus()
    renameInput.value[0].select()
  }
}

const confirmRename = (index: number) => {
  if (editingName.value.trim() && editingName.value.trim() !== eq.customSlots[index].name) {
    eq.renameCustomSlot(index, editingName.value.trim())
  }
  editingIndex.value = -1
}

const addNewSlot = () => {
  const num = eq.customSlots.length + 1
  let name = `自定义${num}`
  // 避免重名
  while (eq.customSlots.some(s => s.name === name)) {
    name = `自定义${num}_${Math.random().toString(36).slice(2, 4)}`
  }
  eq.addCustomSlot(name)
  eq.selectPreset(name)
}
</script>

<style scoped>
.eq-panel {
  width: 340px;
  padding: 16px;
  user-select: none;
}

.eq-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.eq-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

/* 开关 */
.eq-toggle {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
  cursor: pointer;
}
.eq-toggle input { display: none; }
.eq-toggle-slider {
  position: absolute;
  inset: 0;
  background: var(--bg-secondary);
  border-radius: 11px;
  transition: background 0.2s;
}
.eq-toggle-slider::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 18px;
  height: 18px;
  background: #fff;
  border-radius: 50%;
  transition: transform 0.2s;
}
.eq-toggle input:checked + .eq-toggle-slider {
  background: var(--primary-color);
}
.eq-toggle input:checked + .eq-toggle-slider::after {
  transform: translateX(18px);
}

/* 区块 */
.eq-section {
  margin-bottom: 12px;
}
.eq-section-label {
  font-size: 11px;
  color: var(--text-secondary);
  margin-bottom: 6px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* 预设按钮组 */
.eq-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.eq-preset-btn {
  padding: 4px 10px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.eq-preset-btn:hover {
  border-color: var(--primary-color);
}
.eq-preset-btn.active {
  background: var(--primary-color);
  border-color: var(--primary-color);
  color: #000;
  font-weight: 600;
}
.eq-preset-btn.add-btn {
  border-style: dashed;
  color: var(--text-secondary);
}
.eq-preset-btn.add-btn:hover {
  color: var(--primary-color);
  border-color: var(--primary-color);
}

/* 自定义槽位 */
.eq-custom-item {
  display: flex;
  align-items: center;
  gap: 2px;
}

.eq-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  border-radius: 3px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}
.eq-icon-btn:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}
.eq-icon-btn.danger:hover {
  color: #e74c3c;
}

.eq-rename-input {
  width: 80px;
  padding: 2px 4px;
  border: 1px solid var(--primary-color);
  border-radius: 3px;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 12px;
  outline: none;
}

/* 频段滑块区域 */
.eq-sliders {
  display: flex;
  justify-content: space-between;
  gap: 6px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
  transition: opacity 0.2s;
}
.eq-sliders.disabled {
  opacity: 0.35;
  pointer-events: none;
}

.eq-slider-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  flex: 1;
}

.eq-slider-value {
  font-size: 10px;
  color: var(--text-secondary);
  font-variant-numeric: tabular-nums;
  min-width: 28px;
  text-align: center;
}

.eq-slider-track {
  position: relative;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.eq-slider {
  writing-mode: vertical-lr;
  direction: rtl;
  -webkit-appearance: none;
  appearance: none;
  width: 120px;
  height: 4px;
  background: var(--bg-secondary);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}

.eq-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 14px;
  height: 14px;
  background: var(--primary-color);
  border-radius: 50%;
  cursor: pointer;
}
.eq-slider::-moz-range-thumb {
  width: 14px;
  height: 14px;
  background: var(--primary-color);
  border-radius: 50%;
  border: none;
  cursor: pointer;
}

.eq-slider-center {
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--text-secondary);
  opacity: 0.3;
  pointer-events: none;
}

.eq-slider-label {
  font-size: 9px;
  color: var(--text-secondary);
}
</style>
