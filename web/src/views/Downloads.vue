<template>
  <div class="downloads-page">
    <div class="downloads-header">
      <h2>下载记录</h2>
      <span class="subtitle">DOWNLOAD HISTORY</span>
    </div>

    <!-- 实时任务进度 -->
    <div v-if="activeTasks.length" class="active-tasks-section">
      <h3 class="section-title">正在进行的任务</h3>
      <div v-for="task in activeTasks" :key="task.id" class="task-card">
        <div class="task-header">
          <span class="task-title">{{ task.title }}</span>
          <span class="task-status" :class="task.status">{{ getStatusText(task.status) }}</span>
        </div>
        <div class="task-progress-bar">
          <div class="progress-fill" :style="{ width: task.progress + '%' }"></div>
        </div>
        <div class="task-details">
          <div class="detail-row">
            <span class="detail-label">进度:</span>
            <span class="detail-value">{{ task.current }} / {{ task.total }} ({{ task.progress }}%)</span>
          </div>
          <div v-if="task.current_file" class="detail-row">
            <span class="detail-label">当前文件:</span>
            <span class="detail-value filename">{{ task.current_file }}</span>
          </div>
          <div v-if="task.current_bytes > 0" class="detail-row">
            <span class="detail-label">已下载:</span>
            <span class="detail-value">{{ formatBytes(task.current_bytes) }} / {{ formatBytes(task.total_bytes) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 下载历史 -->
    <div class="history-section">
      <h3 class="section-title">下载历史</h3>
      <table class="download-table" v-if="downloads.length">
        <thead>
          <tr>
            <th style="width:40px">#</th>
            <th>歌曲</th>
            <th>歌手</th>
            <th>音质</th>
            <th style="width:100px">状态</th>
            <th style="width:100px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in downloads" :key="item.id">
            <td>{{ index + 1 }}</td>
            <td>{{ item.song_name || '未知歌曲' }}</td>
            <td>{{ item.artist }}</td>
            <td>{{ getQualityText(item.quality) }}</td>
            <td>
              <span class="status-badge" :class="item.status">
                {{ getStatusText(item.status) }}
              </span>
            </td>
            <td>
              <button
                v-if="item.status === 'failed'"
                class="retry-btn"
                @click="retryDownload(item)"
              >
                重试
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">暂无下载记录</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { downloadAPI, taskAPI } from '../api'
import { ElMessage } from 'element-plus'

const downloads = ref<any[]>([])
const tasks = ref<any[]>([])
let pollTimer: number | null = null

const activeTasks = computed(() => {
  return tasks.value.filter(t => t.status === 'running' || t.status === 'pending')
})

const loadHistory = async () => {
  try {
    const res = await downloadAPI.getHistory()
    downloads.value = res.data.history || []
  } catch {}
}

const loadTasks = async () => {
  try {
    const res = await taskAPI.getTasks()
    tasks.value = res.data.tasks || []
  } catch {}
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '等待中',
    running: '进行中',
    completed: '已完成',
    failed: '失败'
  }
  return texts[status] || status
}

const getQualityText = (quality: string) => {
  const texts: Record<string, string> = {
    standard: '标准',
    high: '高质量',
    lossless: '无损'
  }
  return texts[quality] || quality
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

const retryDownload = async (item: any) => {
  try {
    await downloadAPI.downloadSong(item.song_id, item.quality)
    ElMessage.success('已重新加入下载队列')
    await loadHistory()
  } catch {
    ElMessage.error('重试失败')
  }
}

const startPolling = () => {
  pollTimer = window.setInterval(() => {
    loadTasks()
    loadHistory()
  }, 2000)
}

onMounted(() => {
  loadHistory()
  loadTasks()
  startPolling()
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})
</script>

<style scoped>
.downloads-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px 32px;
  min-height: 100%;
}

.downloads-header {
  display: flex;
  align-items: baseline;
  gap: 1rem;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid var(--primary-color);
}

.downloads-header h2 {
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--text-primary);
}

.downloads-header .subtitle {
  font-size: 0.9rem;
  color: var(--text-secondary);
  letter-spacing: 0.2em;
}

.section-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 1rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border-color);
}

.active-tasks-section {
  margin-bottom: 2rem;
}

.task-card {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.task-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.task-status {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.task-status.pending {
  background: var(--bg-secondary);
  color: var(--text-secondary);
}

.task-status.running {
  background: #FFF3CD;
  color: #856404;
}

.task-status.completed {
  background: #D4EDDA;
  color: #155724;
}

.task-status.failed {
  background: #F8D7DA;
  color: #721C24;
}

.task-progress-bar {
  width: 100%;
  height: 8px;
  background: var(--bg-secondary);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 12px;
}

.progress-fill {
  height: 100%;
  background: #FFFA00;
  transition: width 0.3s ease;
}

.task-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  display: flex;
  gap: 8px;
  font-size: 13px;
}

.detail-label {
  color: var(--text-secondary);
  min-width: 80px;
}

.detail-value {
  color: var(--text-primary);
  flex: 1;
}

.detail-value.filename {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  word-break: break-all;
}

.history-section {
  margin-top: 2rem;
}

.download-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.download-table thead th {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 2px solid var(--border-color);
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 13px;
}

.download-table tbody td {
  padding: 10px 8px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
}

.download-table tbody tr:hover {
  background: var(--bg-secondary);
}

.status-badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
}

.status-badge.pending {
  background: var(--bg-secondary);
  color: var(--text-secondary);
}

.status-badge.downloading {
  background: #FFF3CD;
  color: #856404;
}

.status-badge.completed {
  background: #D4EDDA;
  color: #155724;
}

.status-badge.failed {
  background: #F8D7DA;
  color: #721C24;
}

.retry-btn {
  padding: 4px 12px;
  background: #FFFA00;
  color: #000;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}

.retry-btn:hover {
  opacity: 0.85;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
  font-size: 16px;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .downloads-page {
    padding: 16px;
  }

  .downloads-header h2 {
    font-size: 1.5rem;
  }

  .download-table {
    font-size: 12px;
  }

  .download-table thead th,
  .download-table tbody td {
    padding: 8px 4px;
  }

  .download-table th:nth-child(3),
  .download-table td:nth-child(3) {
    display: none;
  }
}
</style>
