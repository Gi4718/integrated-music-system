<template>
  <div class="register-page">
    <div class="register-container">
      <h2 class="page-title">系统注册</h2>
      <p class="page-desc">创建管理员账号以保护您的服务</p>

      <div class="register-form">
        <div class="form-group">
          <label>用户名</label>
          <input v-model="form.username" type="text" placeholder="请输入用户名" class="form-input" />
        </div>

        <div class="form-group">
          <label>密码</label>
          <input v-model="form.password" type="password" placeholder="请输入密码" class="form-input" />
        </div>

        <div class="form-group">
          <label>确认密码</label>
          <input v-model="form.confirmPassword" type="password" placeholder="再次输入密码" class="form-input" />
        </div>

        <button class="register-btn" @click="handleRegister" :disabled="loading">
          {{ loading ? '注册中...' : '注册' }}
        </button>

        <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { systemAPI } from '../api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const errorMsg = ref('')

const form = ref({
  username: '',
  password: '',
  confirmPassword: ''
})

const handleRegister = async () => {
  errorMsg.value = ''

  if (!form.value.username || !form.value.password || !form.value.confirmPassword) {
    errorMsg.value = '请填写所有字段'
    return
  }

  if (form.value.password !== form.value.confirmPassword) {
    errorMsg.value = '两次输入的密码不一致'
    return
  }

  if (form.value.password.length < 6) {
    errorMsg.value = '密码长度至少6位'
    return
  }

  loading.value = true
  try {
    await systemAPI.register({
      username: form.value.username,
      password: form.value.password
    })

    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (error: any) {
    errorMsg.value = error.response?.data?.error || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: var(--bg-color);
}

.register-container {
  width: 100%;
  max-width: 400px;
  background: var(--card-bg);
  border: 2px solid var(--primary-color);
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 8px 32px var(--shadow-color);
}

.page-title {
  font-size: 24px;
  font-weight: bold;
  color: var(--text-primary);
  margin: 0 0 8px 0;
  text-align: center;
}

.page-desc {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0 0 32px 0;
  text-align: center;
}

.register-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.form-input {
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.register-btn {
  padding: 12px;
  background: var(--primary-color);
  color: #000;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
  margin-top: 8px;
}

.register-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.register-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-msg {
  padding: 12px;
  background: rgba(255, 0, 0, 0.1);
  border: 1px solid rgba(255, 0, 0, 0.3);
  border-radius: 8px;
  color: #ff4444;
  font-size: 14px;
  text-align: center;
}

@media (max-width: 768px) {
  .register-container {
    padding: 32px 24px;
  }

  .page-title {
    font-size: 20px;
  }
}
</style>
