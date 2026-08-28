<template>
  <div class="eq-panel">
    <!-- 顶部：标题 + 开关 -->
    <div class="eq-top">
      <span class="eq-title">均衡器</span>
      <label class="eq-toggle">
        <input type="checkbox" :checked="eq.enabled" @change="eq.setEnabled(!eq.enabled)" />
        <span class="eq-toggle-slider"></span>
      </label>
    </div>

    <!-- 预设下拉 + 保存按钮 -->
    <div class="eq-select-row">
      <div class="eq-select-wrap">
        <select class="eq-select" :value="eq.activePresetName" @change="onPresetChange">
          <optgroup label="内置预设">
            <option v-for="p in BUILTIN_PRESETS" :key="p.name" :value="p.name">{{ p.name }}</option>
          </optgroup>
          <optgroup v-if="eq.customSlots.length" label="自定义">
            <option v-for="s in eq.customSlots" :key="s.name" :value="s.name">{{ s.name }}</option>
          </optgroup>
        </select>
      </div>
      <button v-if="isCurrentCustom" class="eq-save-btn" @click="showSaveDialog = true">保存</button>
    </div>

    <!-- 自定义管理区（仅当前是自定义时显示） -->
    <div v-if="isCurrentCustom" class="eq-custom-bar">
      <button class="eq-icon-btn" @click="startRename(currentSlotIndex)" title="重命名">
        <svg viewBox="0 0 24 24" width="14" height="14"><path fill="currentColor" d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/></svg>
      </button>
      <button class="eq-icon-btn danger" @click="eq.removeCustomSlot(currentSlotIndex)" title="删除此自定义">
        <svg viewBox="0 0 24 24" width="14" height="14"><path fill="currentColor" d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
      </button>
      <span class="eq-custom-label">{{ eq.activePresetName }}</span>
    </div>

    <!-- 频段滑块 -->
    <div class="eq-sliders" :class="{ disabled: !eq.enabled }">
      <div v-for="(freq, i) in EQ_FREQUENCIES" :key="freq" class="eq-slider-col">
        <span class="eq-db-label">+{{ EQ_MAX }}dB</span>
        <div class="eq-slider-wrap">
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
        <span class="eq-db-label">0dB</span>
        <span class="eq-db-label">-{{ EQ_MAX }}dB</span>
        <span class="eq-freq-label">{{ EQ_LABELS[i] }}</span>
      </div>
    </div>

    <!-- 新建自定义按钮 -->
    <div class="eq-footer">
      <button
        v-if="eq.customSlots.length < 5"
        class="eq-add-btn"
        @click="addNewSlot"
      >+ 新建自定义</button>
      <span v-else class="eq-full-tip">自定义槽位已满 (5/5)</span>
    </div>

    <!-- 重命名弹窗 -->
    <div v-if="editingIndex >= 0" class="eq-dialog-mask" @click="cancelRename">
      <div class="eq-dialog" @click.stop>
        <div class="eq-dialog-title">重命名</div>
        <input
          v-model="editingName"
          class="eq-dialog-input"
          @keyup.enter="confirmRename"
          @keyup.escape="cancelRename"
          ref="renameInput"
          placeholder="输入新名称"
        />
        <div class="eq-dialog-btns">
          <button class="eq-dialog-btn cancel" @click="cancelRename">取消</button>
          <button class="eq-dialog-btn confirm" @click="confirmRename">确定</button>
        </div>
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
const renameInput = ref<HTMLInputElement>()
const showSaveDialog = ref(false)

const currentBands = computed(() => eq.getCurrentBands())
const isCurrentCustom = computed(() => eq.isCustom(eq.activePresetName))
const currentSlotIndex = computed(() => eq.customSlots.findIndex(s => s.name === eq.activePresetName))

const onPresetChange = (e: Event) => {
  eq.selectPreset((e.target as HTMLSelectElement).value)
}

const onBandChange = (bandIndex: number, event: Event) => {
  const val = parseFloat((event.target as HTMLInputElement).value)
  const slotIdx = eq.customSlots.findIndex(s => s.name === eq.activePresetName)
  if (slotIdx >= 0) {
    eq.updateCustomBand(slotIdx, bandIndex, val)
  }
}

const startRename = async (index: number) => {
  if (index < 0) return
  editingIndex.value = index
  editingName.value = eq.customSlots[index].name
  await nextTick()
  renameInput.value?.focus()
  renameInput.value?.select()
}

const confirmRename = () => {
  if (editingIndex.value >= 0 && editingName.value.trim()) {
    eq.renameCustomSlot(editingIndex.value, editingName.value.trim())
  }
  editingIndex.value = -1
}

const cancelRename = () => {
  editingIndex.value = -1
}

const addNewSlot = () => {
  const num = eq.customSlots.length + 1
  let name = `自定义${num}`
  while (eq.customSlots.some(s => s.name === name)) {
    name = `自定义${num}_${Math.random().toString(36).slice(2, 4)}`
  }
  eq.addCustomSlot(name)
  eq.selectPreset(name)
}
</script>

<style scoped>
.eq-panel {
  width: 100%;
  padding: 12px 16px;
  user-select: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 顶部 */
.eq-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.eq-title {
  font-size: 16px;
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

/* 下拉选择行 */
.eq-select-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.eq-select-wrap {
  flex: 1;
}
.eq-select {
  width: 100%;
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  cursor: pointer;
  appearance: auto;
}
.eq-select:focus {
  border-color: var(--primary-color);
}
.eq-save-btn {
  padding: 6px 16px;
  border: 1px solid var(--primary-color);
  border-radius: 6px;
  background: transparent;
  color: var(--primary-color);
  font-size: 13px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}
.eq-save-btn:hover {
  background: var(--primary-color);
  color: #000;
}

/* 自定义管理条 */
.eq-custom-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 0;
}
.eq-custom-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-left: 4px;
}

.eq-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 4px;
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

/* 频段滑块 */
.eq-sliders {
  display: flex;
  justify-content: space-between;
  gap: 0;
  padding: 4px 16px;
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
  gap: 2px;
  flex: 1;
  max-width: 50px;
}

.eq-db-label {
  font-size: 7px;
  color: var(--text-secondary);
  opacity: 0.5;
  line-height: 1;
}

.eq-slider-wrap {
  position: relative;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.eq-slider {
  writing-mode: vertical-lr;
  direction: rtl;
  -webkit-appearance: none;
  appearance: none;
  width: 60px;
  height: 3px;
  background: var(--bg-secondary);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}

.eq-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 10px;
  height: 10px;
  background: var(--primary-color);
  border-radius: 50%;
  cursor: pointer;
}
.eq-slider::-moz-range-thumb {
  width: 10px;
  height: 10px;
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
  opacity: 0.2;
  pointer-events: none;
}

.eq-freq-label {
  font-size: 9px;
  color: var(--text-secondary);
  margin-top: 1px;
}

/* 底部 */
.eq-footer {
  display: flex;
  justify-content: center;
  padding-top: 4px;
}
.eq-add-btn {
  padding: 4px 14px;
  border: 1px dashed var(--border-color);
  border-radius: 4px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.eq-add-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}
.eq-full-tip {
  font-size: 11px;
  color: var(--text-secondary);
}

/* 重命名弹窗 */
.eq-dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.eq-dialog {
  background: var(--card-bg);
  border-radius: 10px;
  padding: 20px;
  width: 280px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.3);
}
.eq-dialog-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}
.eq-dialog-input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  margin-bottom: 14px;
}
.eq-dialog-input:focus {
  border-color: var(--primary-color);
}
.eq-dialog-btns {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.eq-dialog-btn {
  padding: 6px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}
.eq-dialog-btn.cancel {
  background: var(--bg-secondary);
  color: var(--text-secondary);
}
.eq-dialog-btn.confirm {
  background: var(--primary-color);
  color: #000;
  font-weight: 600;
}
</style>
