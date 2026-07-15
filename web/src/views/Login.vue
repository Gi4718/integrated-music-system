<template>
  <div class="login-page">
    <!-- 系统未登录状态 -->
    <div v-if="!systemAuth.isSystemLoggedIn" class="system-login-section">
      <div class="login-card">
        <h2>系统登录</h2>
        <form @submit.prevent="handleSystemLogin">
          <div class="form-group">
            <label>用户名</label>
            <input
              v-model="systemForm.username"
              type="text"
              placeholder="请输入用户名"
              required
            />
          </div>
          <div class="form-group">
            <label>密码</label>
            <input
              v-model="systemForm.password"
              type="password"
              placeholder="请输入密码"
              required
            />
          </div>
          <button type="submit" class="login-btn" :disabled="systemLoading">
            {{ systemLoading ? '登录中...' : '登录' }}
          </button>
          <div v-if="systemError" class="error-message">{{ systemError }}</div>
        </form>
        <div class="register-link" v-if="!hasSystemUser">
          还没有账号？<router-link to="/register">立即注册</router-link>
        </div>
      </div>
    </div>

    <!-- 系统已登录，显示网易云登录 -->
    <div v-else class="netease-login-section">
      <!-- 网易云未登录 -->
      <div v-if="!authStore.isLoggedIn" class="netease-login-card">
        <h2>网易云音乐登录</h2>
        <p class="subtitle">登录后享受个性化推荐和歌单同步</p>

        <div class="login-tabs">
          <button
            :class="{ active: loginMethod === 'qr' }"
            @click="loginMethod = 'qr'"
          >
            扫码登录
          </button>
          <button
            :class="{ active: loginMethod === 'phone' }"
            @click="loginMethod = 'phone'"
          >
            手机登录
          </button>
          <button
            :class="{ active: loginMethod === 'email' }"
            @click="loginMethod = 'email'"
          >
            邮箱登录
          </button>
        </div>

        <!-- 扫码登录 -->
        <div v-if="loginMethod === 'qr'" class="qr-login">
          <div v-if="qrCodeUrl" class="qr-code-container">
            <img :src="qrCodeUrl" alt="二维码" class="qr-code" />
            <p class="qr-hint">请使用网易云音乐 APP 扫码登录</p>
            <p v-if="qrStatus" class="qr-status">{{ qrStatus }}</p>
          </div>
          <div v-else class="loading">正在生成二维码...</div>
        </div>

        <!-- 手机登录 -->
        <div v-if="loginMethod === 'phone'" class="phone-login">
          <div class="phone-login-tabs">
            <button
              :class="{ active: phoneLoginType === 'password' }"
              @click="phoneLoginType = 'password'"
            >
              密码登录
            </button>
            <button
              :class="{ active: phoneLoginType === 'sms' }"
              @click="phoneLoginType = 'sms'"
            >
              验证码登录
            </button>
          </div>

          <!-- 密码登录 -->
          <form v-if="phoneLoginType === 'password'" @submit.prevent="handlePhonePasswordLogin">
            <div class="form-group">
              <label>手机号</label>
              <input
                v-model="phonePasswordForm.phone"
                type="tel"
                placeholder="请输入手机号"
                pattern="[0-9]{11}"
                required
              />
            </div>
            <div class="form-group">
              <label>密码</label>
              <input
                v-model="phonePasswordForm.password"
                type="password"
                placeholder="请输入密码"
                required
              />
            </div>
            <button type="submit" class="login-btn" :disabled="phonePasswordLoading">
              {{ phonePasswordLoading ? '登录中...' : '登录' }}
            </button>
            <div v-if="phonePasswordError" class="error-message">{{ phonePasswordError }}</div>
          </form>

          <!-- 验证码登录 -->
          <form v-if="phoneLoginType === 'sms'" @submit.prevent="handlePhoneSMSLogin">
            <div class="form-group">
              <label>手机号</label>
              <input
                v-model="phoneSMSForm.phone"
                type="tel"
                placeholder="请输入手机号"
                pattern="[0-9]{11}"
                required
              />
            </div>
            <div class="form-group">
              <label>验证码</label>
              <div class="sms-input-group">
                <input
                  v-model="phoneSMSForm.captcha"
                  type="text"
                  placeholder="请输入验证码"
                  maxlength="6"
                  required
                />
                <button
                  type="button"
                  class="sms-btn"
                  :disabled="smsCooldown > 0 || phoneSMSLoading"
                  @click="sendSMSCode"
                >
                  {{ smsCooldown > 0 ? `${smsCooldown}s` : '获取验证码' }}
                </button>
              </div>
            </div>
            <button type="submit" class="login-btn" :disabled="phoneSMSLoading">
              {{ phoneSMSLoading ? '登录中...' : '登录' }}
            </button>
            <div v-if="phoneSMSError" class="error-message">{{ phoneSMSError }}</div>
          </form>
        </div>

        <!-- 邮箱登录 -->
        <div v-if="loginMethod === 'email'" class="email-login">
          <form @submit.prevent="handleEmailLogin">
            <div class="form-group">
              <label>邮箱</label>
              <input
                v-model="emailForm.email"
                type="email"
                placeholder="请输入邮箱"
                required
              />
            </div>
            <div class="form-group">
              <label>密码</label>
              <input
                v-model="emailForm.password"
                type="password"
                placeholder="请输入密码"
                required
              />
            </div>
            <button type="submit" class="login-btn" :disabled="emailLoading">
              {{ emailLoading ? '登录中...' : '登录' }}
            </button>
            <div v-if="emailError" class="error-message">{{ emailError }}</div>
          </form>
        </div>
      </div>

      <!-- 网易云已登录 -->
      <div v-else class="netease-logged-in">
        <div class="user-info-card">
          <img :src="authStore.user?.avatar" alt="头像" class="user-avatar" />
          <div class="user-details">
            <h3>{{ authStore.user?.nickname }}</h3>
            <div class="user-badges">
              <span v-if="authStore.user?.vipType" class="vip-badge">
                {{ authStore.user?.vipType }}
              </span>
              <span class="status-badge">
                {{ authStore.user?.cookieValid ? 'Cookie 有效' : 'Cookie 已过期' }}
              </span>
            </div>
          </div>
          <button @click="handleNeteaseLogout" class="logout-btn">
            退出网易云登录
          </button>
        </div>

        <div class="system-logout-section">
          <button @click="handleSystemLogout" class="system-logout-btn">
            退出系统登录
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useSystemAuthStore } from '../stores/systemAuth'
import { authAPI } from '../api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const systemAuth = useSystemAuthStore()

// 系统登录状态
const hasSystemUser = ref(false)
const systemLoading = ref(false)
const systemError = ref('')
const systemForm = ref({
  username: '',
  password: ''
})

// 网易云登录方式
const loginMethod = ref<'qr' | 'phone' | 'email'>('qr')

// 扫码登录
const qrCodeUrl = ref('')
const qrStatus = ref('')
let qrCheckInterval: number | null = null

// 手机登录
const phoneLoginType = ref<'password' | 'sms'>('password')
const phonePasswordLoading = ref(false)
const phonePasswordError = ref('')
const phonePasswordForm = ref({ phone: '', password: '' })
const phoneSMSLoading = ref(false)
const phoneSMSError = ref('')
const phoneSMSForm = ref({ phone: '', captcha: '' })
const smsCooldown = ref(0)
let smsTimer: number | null = null

// 邮箱登录
const emailLoading = ref(false)
const emailError = ref('')
const emailForm = ref({
  email: '',
  password: ''
})

// 系统登录
const handleSystemLogin = async () => {
  systemLoading.value = true
  systemError.value = ''

  try {
    const success = await systemAuth.login(systemForm.value.username, systemForm.value.password)
    if (success) {
      ElMessage.success('系统登录成功')
      // 系统登录成功后立即检查网易云登录状态
      await authStore.checkLoginStatus()
      // 如果是扫码模式且网易云未登录，生成二维码
      if (loginMethod.value === 'qr' && !authStore.isLoggedIn) {
        await getQRCode()
      }
    } else {
      systemError.value = '登录失败，请检查用户名和密码'
    }
  } catch (error: any) {
    systemError.value = error.response?.data?.error || '登录失败'
  } finally {
    systemLoading.value = false
  }
}

// 系统退出
const handleSystemLogout = () => {
  systemAuth.logout()
  ElMessage.success('已退出系统登录')
}

// 网易云退出
const handleNeteaseLogout = async () => {
  try {
    await authAPI.logout()
  } catch (error) {
    console.error('退出登录API调用失败:', error)
  } finally {
    // 无论API是否成功，都要清空前端状态
    authStore.clearUser()
    ElMessage.success('已退出网易云登录')
    // 退出后自动生成二维码
    if (loginMethod.value === 'qr') {
      // 先清空旧状态，确保视图立即切换到二维码区域
      qrCodeUrl.value = ''
      qrStatus.value = ''
      stopQRCheck()
      // 延迟一帧确保视图更新完成
      await nextTick()
      await getQRCode()
    }
  }
}

// 获取二维码
const getQRCode = async () => {
  try {
    // 先清空旧状态
    qrCodeUrl.value = ''
    qrStatus.value = '正在生成二维码...'
    stopQRCheck()

    const keyRes = await authAPI.getQRKey()
    const qrKey = keyRes.data.key

    const qrRes = await authAPI.getQRCode(qrKey)
    qrCodeUrl.value = qrRes.data.qr_img
    qrStatus.value = ''

    // 开始轮询检查扫码状态
    startQRCheck(qrKey)
  } catch (error) {
    console.error('获取二维码失败:', error)
    qrStatus.value = '获取二维码失败，请重试'
    qrCodeUrl.value = ''
  }
}

// 轮询二维码状态
const startQRCheck = (key: string) => {
  stopQRCheck()
  qrCheckInterval = window.setInterval(async () => {
    try {
      const res = await authAPI.checkQRStatus(key)
      const { code, message } = res.data

      qrStatus.value = message

      if (code === 803) {
        // 授权成功
        stopQRCheck()
        qrStatus.value = '登录成功，正在获取用户信息...'

        try {
          // 后端已经保存了 cookie 和用户信息到数据库
          // 只需要更新前端的登录状态
          await authStore.checkLoginStatus()

          if (authStore.isLoggedIn) {
            ElMessage.success('网易云登录成功')
            // 登录成功后跳转到首页，让首页重新加载推荐内容
            setTimeout(() => {
              router.push('/')
            }, 500)
          } else {
            ElMessage.error('获取用户信息失败，请刷新页面重试')
            qrStatus.value = '登录失败，请刷新页面重试'
          }
        } catch (statusError) {
          console.error('获取登录状态失败:', statusError)
          ElMessage.error('获取用户信息失败，请刷新页面重试')
          qrStatus.value = '登录失败，请刷新页面重试'
        }
      } else if (code === 800) {
        // 二维码过期
        stopQRCheck()
        qrStatus.value = '二维码已过期，请刷新页面'
        qrCodeUrl.value = ''
      }
    } catch (error) {
      console.error('检查二维码状态失败', error)
    }
  }, 3000)
}

const stopQRCheck = () => {
  if (qrCheckInterval) {
    clearInterval(qrCheckInterval)
    qrCheckInterval = null
  }
}

// 手机密码登录（支持二次验证）
const handlePhonePasswordLogin = async () => {
  phonePasswordLoading.value = true
  phonePasswordError.value = ''

  try {
    const res = await authAPI.loginByPhonePassword(phonePasswordForm.value.phone, phonePasswordForm.value.password)
    if (res.data.code === 200) {
      // 后端已经保存了 cookie 和用户信息，无需再次调用 saveCookie
      await authStore.checkLoginStatus()
      ElMessage.success('网易云登录成功')
    } else if (res.data.needSecondVerify) {
      const verifyCode = prompt('需要二次验证，请输入验证码：')
      if (verifyCode) {
        const verifyRes = await authAPI.secondVerify(phonePasswordForm.value.phone, phonePasswordForm.value.password, verifyCode)
        if (verifyRes.data.code === 200) {
          await authStore.checkLoginStatus()
          ElMessage.success('网易云登录成功')
        } else {
          phonePasswordError.value = verifyRes.data.error || '二次验证失败'
        }
      }
    } else {
      phonePasswordError.value = res.data.msg || res.data.error || '登录失败'
    }
  } catch (error: any) {
    phonePasswordError.value = error.response?.data?.msg || error.response?.data?.error || '登录失败'
  } finally {
    phonePasswordLoading.value = false
  }
}

// 发送短信验证码
const sendSMSCode = async () => {
  if (!phoneSMSForm.value.phone || phoneSMSForm.value.phone.length !== 11) {
    phoneSMSError.value = '请输入正确的手机号'
    return
  }

  try {
    await authAPI.sendSMSCode(phoneSMSForm.value.phone)
    ElMessage.success('验证码已发送')
    smsCooldown.value = 60
    smsTimer = window.setInterval(() => {
      smsCooldown.value--
      if (smsCooldown.value <= 0) {
        if (smsTimer) {
          clearInterval(smsTimer)
          smsTimer = null
        }
      }
    }, 1000)
  } catch (error: any) {
    phoneSMSError.value = error.response?.data?.error || '发送验证码失败'
  }
}

// 手机验证码登录
const handlePhoneSMSLogin = async () => {
  phoneSMSLoading.value = true
  phoneSMSError.value = ''

  try {
    const res = await authAPI.loginByPhone(phoneSMSForm.value.phone, phoneSMSForm.value.captcha)
    if (res.data.code === 200) {
      // 后端已经保存了 cookie 和用户信息，无需再次调用 saveCookie
      await authStore.checkLoginStatus()
      ElMessage.success('网易云登录成功')
    } else {
      phoneSMSError.value = res.data.msg || res.data.error || '登录失败'
    }
  } catch (error: any) {
    phoneSMSError.value = error.response?.data?.msg || error.response?.data?.error || '登录失败'
  } finally {
    phoneSMSLoading.value = false
  }
}

// 邮箱登录
const handleEmailLogin = async () => {
  emailLoading.value = true
  emailError.value = ''

  try {
    const res = await authAPI.loginByEmail(emailForm.value.email, emailForm.value.password)
    if (res.data.code === 200) {
      // 后端已经保存了 cookie 和用户信息，无需再次调用 saveCookie
      await authStore.checkLoginStatus()
      ElMessage.success('网易云登录成功')
    } else {
      emailError.value = res.data.msg || res.data.error || '登录失败'
    }
  } catch (error: any) {
    emailError.value = error.response?.data?.msg || error.response?.data?.error || '登录失败'
  } finally {
    emailLoading.value = false
  }
}

onMounted(async () => {
  // 防御性恢复：确保 store 的 token 与 localStorage 同步
  const savedToken = localStorage.getItem('system_token')
  if (savedToken && !systemAuth.token) {
    systemAuth.token = savedToken
    systemAuth.username = localStorage.getItem('system_username')
  }

  // 检查是否有系统用户
  hasSystemUser.value = await systemAuth.checkSystemUser()

  // 如果系统已登录，检查网易云登录状态
  if (systemAuth.isSystemLoggedIn) {
    await authStore.checkLoginStatus()

    // 如果是扫码登录，生成二维码
    if (loginMethod.value === 'qr' && !authStore.isLoggedIn) {
      await getQRCode()
    }
  }
})

onUnmounted(() => {
  stopQRCheck()
})
</script>

<style scoped>
.login-page {
  min-height: calc(100vh - 64px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  background: var(--bg-color);
}

/* 系统登录区域 */
.system-login-section {
  width: 100%;
  max-width: 400px;
}

.login-card {
  background: var(--card-bg);
  border: 2px solid var(--primary-color);
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 8px 32px var(--shadow-color);
}

.login-card h2 {
  font-size: 24px;
  font-weight: bold;
  color: var(--text-primary);
  margin: 0 0 32px 0;
  text-align: center;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.form-group input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.login-btn {
  width: 100%;
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

.login-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.login-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-message {
  margin-top: 16px;
  padding: 12px;
  background: rgba(255, 0, 0, 0.1);
  border: 1px solid rgba(255, 0, 0, 0.3);
  border-radius: 8px;
  color: #ff4444;
  font-size: 14px;
  text-align: center;
}

.register-link {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
  color: var(--text-secondary);
}

.register-link a {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 600;
}

.register-link a:hover {
  text-decoration: underline;
}

/* 网易云登录区域 */
.netease-login-section {
  width: 100%;
  max-width: 500px;
}

.netease-login-card {
  background: var(--card-bg);
  border: 2px solid var(--primary-color);
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 8px 32px var(--shadow-color);
}

.netease-login-card h2 {
  font-size: 24px;
  font-weight: bold;
  color: var(--text-primary);
  margin: 0 0 8px 0;
  text-align: center;
}

.subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  text-align: center;
  margin: 0 0 32px 0;
}

.login-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 32px;
  border-bottom: 1px solid var(--border-color);
}

.login-tabs button {
  flex: 1;
  padding: 12px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.login-tabs button.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
}

.login-tabs button:hover {
  color: var(--primary-color);
}

/* 二维码登录 */
.qr-login {
  text-align: center;
}

.qr-code-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.qr-code {
  width: 200px;
  height: 200px;
  border: 2px solid var(--border-color);
  border-radius: 8px;
}

.qr-hint {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

.qr-status {
  font-size: 14px;
  color: var(--primary-color);
  font-weight: 500;
  margin: 0;
}

.loading {
  font-size: 14px;
  color: var(--text-secondary);
  padding: 40px;
}

/* 手机登录子tab */
.phone-login-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.phone-login-tabs button {
  flex: 1;
  padding: 8px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.phone-login-tabs button.active {
  background: var(--primary-color);
  color: #000;
  border-color: var(--primary-color);
}

.sms-input-group {
  display: flex;
  gap: 8px;
}

.sms-input-group input {
  flex: 1;
}

.sms-btn {
  padding: 12px 16px;
  background: var(--primary-color);
  color: #000;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.2s;
}

.sms-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.sms-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 网易云已登录 */
.netease-logged-in {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.user-info-card {
  background: var(--card-bg);
  border: 2px solid var(--primary-color);
  border-radius: 16px;
  padding: 32px;
  box-shadow: 0 8px 32px var(--shadow-color);
  display: flex;
  align-items: center;
  gap: 24px;
}

.user-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--primary-color);
}

.user-details {
  flex: 1;
}

.user-details h3 {
  font-size: 20px;
  font-weight: bold;
  color: var(--text-primary);
  margin: 0 0 12px 0;
}

.user-badges {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.vip-badge {
  padding: 4px 12px;
  background: linear-gradient(135deg, #ffd700 0%, #ffa500 100%);
  color: #000;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge {
  padding: 4px 12px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.logout-btn {
  padding: 10px 24px;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.system-logout-section {
  text-align: center;
}

.system-logout-btn {
  padding: 10px 24px;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.system-logout-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

/* 移动端适配 */
@media (max-width: 768px) {
  .login-card,
  .netease-login-card {
    padding: 32px 24px;
  }

  .user-info-card {
    flex-direction: column;
    text-align: center;
  }

  .user-details {
    text-align: center;
  }

  .user-badges {
    justify-content: center;
  }
}
</style>
