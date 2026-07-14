<template>
  <div class="message-panel" v-if="visible">
    <div class="panel-header">
      <h3>任务日志</h3>
      <button class="close-btn" @click="$emit('close')">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
        </svg>
      </button>
    </div>
    <div class="panel-content" ref="panelContent">
      <div v-if="tasks.length === 0" class="empty-state">
        <p>暂无任务</p>
      </div>
      <div v-else class="task-list">
        <div v-for="task in tasks" :key="task.id" class="task-item">
          <div class="task-header">
            <span class="task-type-badge" :class="task.type">
              {{ getTypeLabel(task.type) }}
            </span>
            <span class="task-status" :class="task.status">
              {{ getStatusLabel(task.status) }}
            </span>
            <button
              v-if="task.status === 'running' || task.status === 'pending'"
              class="cancel-btn"
              @click="cancelTask(task.id)"
              title="终止任务"
            >
              <svg viewBox="0 0 24 24" width="16" height="16">
                <path fill="currentColor" d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
              </svg>
            </button>
          </div>
          <div class="task-title">{{ task.title }}</div>
          <div v-if="task.description" class="task-desc">{{ task.description }}</div>
          <div v-if="task.status === 'running' || task.status === 'completed'" class="task-progress">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: task.progress + '%' }"></div>
            </div>
            <span class="progress-text">{{ task.current }}/{{ task.total }}</span>
          </div>
          <div v-if="task.current_bytes > 0" class="task-bytes">
            {{ formatBytes(task.current_bytes) }} / {{ formatBytes(task.total_bytes) }}
          </div>
          <div v-if="task.current_file" class="task-current-file">{{ task.current_file }}</div>
          <div v-if="task.error" class="task-error">{{ task.error }}</div>
          <div class="task-time">开始时间：{{ formatTime(task.created_at) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { taskAPI } from '../api'

defineProps<{
  visible: boolean
}>()

defineEmits<{
  (e: 'close'): void
}>()

interface Task {
  id: string
  type: string
  title: string
  description: string
  status: string
  progress: number
  total: number
  current: number
  current_bytes: number
  total_bytes: number
  current_file: string
  created_at: string
  error?: string
}

const tasks = ref<Task[]>([])
let pollTimer: number | null = null

const getTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    download: '下载',
    data_complete: '补全',
    sync: '同步',
    scan: '扫描'
  }
  return labels[type] || type
}

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    pending: '等待中',
    running: '进行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已终止'
  }
  return labels[status] || status
}

const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

const loadTasks = async () => {
  try {
    const res = await taskAPI.getTasks()
    const newTasks = res.data.tasks || []
    // 按创建时间升序排序，保证顺序稳定
    newTasks.sort((a: Task, b: Task) => {
      return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    })
    tasks.value = newTasks
  } catch (e) {
    console.error('加载任务失败', e)
  }
}

const cancelTask = async (taskId: string) => {
  try {
    await taskAPI.cancelTask(taskId)
    // 立即刷新列表
    await loadTasks()
  } catch (e) {
    console.error('终止任务失败', e)
  }
}

onMounted(() => {
  loadTasks()
  pollTimer = window.setInterval(loadTasks, 3000)
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})
</script>

<style scoped>
.message-panel {
  position: fixed;
  top: 64px;
  right: 20px;
  width: 380px;
  max-height: calc(100vh - 100px);
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-secondary);
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.2s;
}

.close-btn:hover {
  background: var(--bg-secondary);
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
  color: var(--text-secondary);
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-item {
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.task-type-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.task-type-badge.download {
  background: #FFD700;
  color: #000;
}

.task-type-badge.data_complete {
  background: #4CAF50;
  color: #fff;
}

.task-type-badge.sync {
  background: #2196F3;
  color: #fff;
}

.task-type-badge.scan {
  background: #9C27B0;
  color: #fff;
}

.task-status {
  font-size: 12px;
  font-weight: 500;
}

.task-status.pending {
  color: var(--text-secondary);
}

.task-status.running {
  color: #FFD700;
}

.task-status.completed {
  color: #4CAF50;
}

.task-status.failed {
  color: #f44336;
}

.task-status.cancelled {
  color: #ff9800;
}

.cancel-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #f44336;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.2s;
  opacity: 0.6;
}

.cancel-btn:hover {
  background: rgba(244, 67, 54, 0.1);
  opacity: 1;
}

.task-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.task-desc {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.task-progress {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: var(--border-color);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #FFD700;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 12px;
  color: var(--text-secondary);
  min-width: 60px;
  text-align: right;
}

.task-error {
  font-size: 12px;
  color: #f44336;
  margin-bottom: 8px;
}

.task-time {
  font-size: 11px;
  color: var(--text-secondary);
}

.task-bytes {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.task-current-file {
  font-size: 11px;
  color: var(--text-secondary);
  font-family: 'Courier New', monospace;
  word-break: break-all;
  margin-bottom: 4px;
}

@media (max-width: 768px) {
  .message-panel {
    top: 56px;
    right: 8px;
    left: 8px;
    width: auto;
    max-height: calc(100vh - 80px);
  }
}
</style>
