<template>
  <div class="accounts-page">
    <div class="accounts-card">
      <h3 class="card-title">账户管理</h3>

      <!-- 多用户注册开关 -->
      <div class="setting-row">
        <div class="setting-label-col">
          <span class="setting-name">允许多用户注册</span>
          <span class="setting-desc">开启后允许其他用户注册系统账号</span>
        </div>
        <div class="setting-control-row">
          <label class="switch">
            <input type="checkbox" v-model="multiUserEnabled" @change="toggleMultiUser" />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <!-- 用户列表 -->
      <div class="user-list-section">
        <h4 class="section-title">账号列表</h4>

        <div class="user-list">
          <div v-for="user in users" :key="user.id" class="user-item">
            <div class="user-info">
              <div class="user-name">{{ user.username }}</div>
              <div class="user-meta">
                <span class="role-tag" :class="user.role">{{ user.role === 'admin' ? '管理员' : '普通用户' }}</span>
                <span class="meta-text">注册时间: {{ formatDate(user.created_at) }}</span>
                <span v-if="user.last_login_at" class="meta-text">最后登录: {{ formatDate(user.last_login_at) }}</span>
              </div>
            </div>

            <div class="user-actions">
              <button
                v-if="user.id !== currentUserId"
                class="action-btn role-btn"
                @click="toggleRole(user)"
                :title="user.role === 'admin' ? '降级为普通用户' : '提升为管理员'"
              >
                {{ user.role === 'admin' ? '降级' : '提升' }}
              </button>

              <button
                v-if="user.id !== currentUserId"
                class="action-btn pwd-btn"
                @click="showPasswordDialog(user)"
              >
                改密
              </button>

              <button
                v-if="user.id !== currentUserId"
                class="action-btn delete-btn"
                @click="confirmDelete(user)"
              >
                删除
              </button>

              <span v-if="user.id === currentUserId" class="current-tag">当前账号</span>
            </div>
          </div>

          <div v-if="users.length === 0" class="empty-state">
            <p>暂无用户</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 修改密码对话框 -->
    <div v-if="passwordDialog.visible" class="dialog-overlay" @click.self="closePasswordDialog">
      <div class="dialog">
        <h4 class="dialog-title">修改密码 - {{ passwordDialog.username }}</h4>
        <div class="dialog-content">
          <div class="form-group">
            <label>新密码</label>
            <input
              v-model="passwordDialog.newPassword"
              type="password"
              placeholder="请输入新密码（至少6位）"
              class="dialog-input"
            />
          </div>
        </div>
        <div class="dialog-actions">
          <button class="dialog-btn cancel-btn" @click="closePasswordDialog">取消</button>
          <button class="dialog-btn confirm-btn" @click="submitPasswordChange" :disabled="passwordDialog.submitting">
            {{ passwordDialog.submitting ? '提交中...' : '确认修改' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <div v-if="deleteDialog.visible" class="dialog-overlay" @click.self="closeDeleteDialog">
      <div class="dialog">
        <h4 class="dialog-title">确认删除</h4>
        <div class="dialog-content">
          <p>确定要删除用户 <strong>{{ deleteDialog.username }}</strong> 吗？</p>
          <p class="warning-text">此操作不可恢复，该用户的设置和网易云账号信息将被清除。</p>
        </div>
        <div class="dialog-actions">
          <button class="dialog-btn cancel-btn" @click="closeDeleteDialog">取消</button>
          <button class="dialog-btn confirm-btn danger" @click="submitDelete" :disabled="deleteDialog.submitting">
            {{ deleteDialog.submitting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useSystemAuthStore } from '../stores/systemAuth'
import { systemAPI, settingsAPI } from '../api'
import { ElMessage } from 'element-plus'

const systemAuth = useSystemAuthStore()

interface User {
  id: number
  username: string
  role: string
  created_at: string
  last_login_at?: string
  netease_user_id?: number
  netease_nick?: string
  cookie_valid?: boolean
}

const users = ref<User[]>([])
const multiUserEnabled = ref(false)

const currentUserId = computed(() => {
  const token = systemAuth.token
  if (!token) return 0
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.user_id || 0
  } catch {
    return 0
  }
})

const passwordDialog = ref({
  visible: false,
  userId: 0,
  username: '',
  newPassword: '',
  submitting: false
})

const deleteDialog = ref({
  visible: false,
  userId: 0,
  username: '',
  submitting: false
})

const loadUsers = async () => {
  try {
    const res = await systemAPI.listUsers()
    if (res.data.users) {
      users.value = res.data.users
    }
  } catch (e) {
    console.error('加载用户列表失败', e)
    ElMessage.error('加载用户列表失败')
  }
}

const loadMultiUserSetting = async () => {
  try {
    const res = await settingsAPI.getSettings()
    multiUserEnabled.value = res.data.settings?.multi_user_enabled === 'true'
  } catch (e) {
    console.error('加载设置失败', e)
  }
}

const toggleMultiUser = async () => {
  try {
    await settingsAPI.updateSettings({ multi_user_enabled: multiUserEnabled.value.toString() })
    ElMessage.success(multiUserEnabled.value ? '已开启多用户注册' : '已关闭多用户注册')
  } catch {
    ElMessage.error('保存失败')
    multiUserEnabled.value = !multiUserEnabled.value
  }
}

const toggleRole = async (user: User) => {
  const newRole = user.role === 'admin' ? 'user' : 'admin'
  try {
    const res = await systemAPI.updateUserRole(user.id, newRole)
    if (res.data.message) {
      ElMessage.success(`已将 ${user.username} ${newRole === 'admin' ? '提升为管理员' : '降级为普通用户'}`)
      await loadUsers()
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '操作失败')
  }
}

const showPasswordDialog = (user: User) => {
  passwordDialog.value = {
    visible: true,
    userId: user.id,
    username: user.username,
    newPassword: '',
    submitting: false
  }
}

const closePasswordDialog = () => {
  passwordDialog.value.visible = false
  passwordDialog.value.newPassword = ''
}

const submitPasswordChange = async () => {
  if (passwordDialog.value.newPassword.length < 6) {
    ElMessage.warning('密码至少6位')
    return
  }
  passwordDialog.value.submitting = true
  try {
    const res = await systemAPI.updateUserPassword(passwordDialog.value.userId, passwordDialog.value.newPassword)
    if (res.data.message) {
      ElMessage.success('密码已修改')
      closePasswordDialog()
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '修改失败')
  } finally {
    passwordDialog.value.submitting = false
  }
}

const confirmDelete = (user: User) => {
  deleteDialog.value = {
    visible: true,
    userId: user.id,
    username: user.username,
    submitting: false
  }
}

const closeDeleteDialog = () => {
  deleteDialog.value.visible = false
}

const submitDelete = async () => {
  deleteDialog.value.submitting = true
  try {
    const res = await systemAPI.deleteUser(deleteDialog.value.userId)
    if (res.data.message) {
      ElMessage.success('用户已删除')
      closeDeleteDialog()
      await loadUsers()
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '删除失败')
  } finally {
    deleteDialog.value.submitting = false
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '--'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  loadUsers()
  loadMultiUserSetting()
})
</script>

<style scoped>
.accounts-page {
  padding: 24px 32px;
  max-width: 900px;
  min-height: 100%;
}

.accounts-card {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 24px 32px;
  background: var(--card-bg);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 20px 0;
}

.setting-row {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.setting-label-col {
  width: 200px;
  min-width: 200px;
  text-align: right;
  padding-right: 24px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.setting-name {
  font-size: 14px;
  color: var(--text-primary);
}

.setting-desc {
  font-size: 12px;
  color: var(--text-secondary);
}

.setting-control-row {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

/* Switch toggle */
.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #ccc;
  border-radius: 24px;
  transition: 0.3s;
}

.slider::before {
  content: '';
  position: absolute;
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background: #fff;
  border-radius: 50%;
  transition: 0.3s;
}

.switch input:checked + .slider {
  background: #FFFA00;
}

.switch input:checked + .slider::before {
  transform: translateX(20px);
}

/* User list */
.user-list-section {
  margin-top: 24px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 16px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-secondary);
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.user-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.role-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.role-tag.admin {
  background: rgba(255, 215, 0, 0.2);
  color: #FFD700;
  border: 1px solid #FFD700;
}

.role-tag.user {
  background: rgba(128, 128, 128, 0.2);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}

.meta-text {
  font-size: 12px;
  color: var(--text-secondary);
}

.user-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover {
  background: var(--border-color);
}

.role-btn:hover {
  background: rgba(255, 215, 0, 0.2);
  border-color: #FFD700;
}

.pwd-btn:hover {
  background: rgba(66, 133, 244, 0.2);
  border-color: #4285f4;
}

.delete-btn:hover {
  background: rgba(244, 67, 54, 0.2);
  border-color: #f44336;
  color: #f44336;
}

.current-tag {
  padding: 4px 10px;
  background: rgba(76, 175, 80, 0.2);
  color: #4caf50;
  border: 1px solid #4caf50;
  border-radius: 4px;
  font-size: 12px;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

/* Dialog */
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
  z-index: 1000;
}

.dialog {
  background: var(--card-bg);
  border-radius: 8px;
  padding: 24px;
  min-width: 400px;
  max-width: 90vw;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.dialog-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 20px 0;
}

.dialog-content {
  margin-bottom: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  color: var(--text-primary);
}

.dialog-input {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
}

.dialog-input:focus {
  border-color: #FFFA00;
}

.warning-text {
  color: #f44336;
  font-size: 13px;
  margin-top: 8px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.dialog-btn {
  padding: 8px 20px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.cancel-btn:hover {
  background: var(--border-color);
}

.confirm-btn {
  background: #FFFA00;
  color: #000;
  border-color: #FFFA00;
  font-weight: 600;
}

.confirm-btn:hover {
  opacity: 0.85;
}

.confirm-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.confirm-btn.danger {
  background: #f44336;
  color: #fff;
  border-color: #f44336;
}

.confirm-btn.danger:hover {
  opacity: 0.85;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .accounts-page {
    padding: 16px;
  }

  .accounts-card {
    padding: 16px;
  }

  .setting-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .setting-label-col {
    width: 100%;
    text-align: left;
  }

  .user-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .user-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .dialog {
    min-width: auto;
    width: 90vw;
  }
}
</style>
