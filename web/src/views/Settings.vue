<template>
  <div class="settings-page">
    <!-- 本地下载配置 -->
    <div class="settings-card">
      <h3 class="card-title">本地下载配置</h3>

      <div class="setting-row">
        <label class="setting-label">下载路径</label>
        <div class="setting-input-wrap">
          <input v-model="settings.downloadPath" class="setting-input" placeholder="/music" />
          <button class="help-btn" @mouseenter="showTip('path')" @mouseleave="hideTip()">?</button>
          <div v-if="activeTip === 'path'" class="tip-popup tip-left">
            1. 下载路径格式。Mac/Linux: /path/to/... |Windows: C:\Users\YourUserName\Downloads<br/>
            2. 请注意，如果本服务部署在 Docker 中，下载路径应当为 Docker 容器内的路径。你需要将容器内的下载路径映射到宿主机的相应目录。
          </div>
        </div>
      </div>

      <div class="setting-row">
        <label class="setting-label">单曲下载的文件名格式</label>
        <div class="setting-input-wrap">
          <input v-model="settings.songFormat" class="setting-input" placeholder="{songName} - {artist}" />
          <button class="help-btn" @mouseenter="showTip('songFormat')" @mouseleave="hideTip()">?</button>
          <div v-if="activeTip === 'songFormat'" class="tip-popup tip-left">
            支持的变量：{songName}，{artist}，{album}<br/>
            示例：{album}-{artist}-{songName}<br/>
            支持目录结构：{artist}/{album}/{songName}
          </div>
        </div>
      </div>

      <div class="section-save-bar">
        <button class="save-btn" @click="saveDownloadSettings" :disabled="savingDownload">
          {{ savingDownload ? '保存中...' : '保存' }}
        </button>
        <span v-if="downloadSavedTip" class="save-tip">保存成功</span>
      </div>
    </div>

    <!-- 歌单同步到本地 -->
    <div class="settings-card">
      <h3 class="card-title">歌单同步到本地</h3>

      <div class="setting-row">
        <div class="setting-label-col">
          <span class="setting-name">自动同步</span>
          <span class="setting-desc">开启后将按设定规则自动同步歌单到本地</span>
        </div>
        <div class="setting-control-row">
          <label class="switch">
            <input type="checkbox" v-model="settings.autoSync" />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div v-if="settings.autoSync" class="setting-row">
        <div class="setting-label-col">
          <span class="setting-name">同步模式</span>
          <span class="setting-desc">选择同步触发方式</span>
        </div>
        <div class="setting-control-row">
          <div class="mode-btns">
            <button :class="{ active: settings.syncMode === 'interval' }" @click="settings.syncMode = 'interval'">间隔模式</button>
            <button :class="{ active: settings.syncMode === 'schedule' }" @click="settings.syncMode = 'schedule'">定时模式</button>
          </div>
        </div>
      </div>

      <div v-if="settings.autoSync && settings.syncMode === 'interval'" class="setting-row">
        <div class="setting-label-col">
          <span class="setting-name">同步间隔</span>
          <span class="setting-desc">每隔指定时间同步一次</span>
        </div>
        <div class="setting-control-row">
          <span class="control-text">每</span>
          <input type="number" v-model.number="settings.syncInterval" class="number-input" min="1" max="72" />
          <div class="unit-btns">
            <button :class="{ active: settings.syncUnit === 'hour' }" @click="settings.syncUnit = 'hour'">小时</button>
            <button :class="{ active: settings.syncUnit === 'day' }" @click="settings.syncUnit = 'day'">天</button>
          </div>
        </div>
      </div>

      <div v-if="settings.autoSync && settings.syncMode === 'schedule'" class="setting-row">
        <div class="setting-label-col">
          <span class="setting-name">同步时间</span>
          <span class="setting-desc">选择星期和时间进行同步</span>
        </div>
        <div class="setting-control-row">
          <div class="weekday-btns">
            <button v-for="day in weekdays" :key="day.value"
                    :class="{ active: settings.syncWeekdays.includes(day.value) }"
                    @click="toggleWeekday(day.value)">
              {{ day.label }}
            </button>
          </div>
          <input type="time" v-model="settings.syncTime" class="time-input" />
        </div>
      </div>

      <div class="setting-row">
        <div class="setting-label-col">
          <span class="setting-name">断点续传</span>
          <span class="setting-desc">继续之前未完成的下载任务，避免重复下载</span>
        </div>
        <div class="setting-control-row">
          <label class="switch">
            <input type="checkbox" v-model="settings.resumeDownloads" />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="setting-row checkbox-row">
        <label class="checkbox-label">
          <input type="checkbox" v-model="settings.deleteRemoved" />
          <span>当歌单里的歌曲移除时，同时删除本地对应的歌曲文件</span>
        </label>
      </div>

      <div class="setting-row">
        <label class="setting-label">歌单歌曲的文件名格式</label>
        <div class="setting-input-wrap">
          <input v-model="settings.playlistFormat" class="setting-input" placeholder="{playlistName}/{songName} - {artist}" />
          <button class="help-btn" @mouseenter="showTip('playlistFormat')" @mouseleave="hideTip()">?</button>
          <div v-if="activeTip === 'playlistFormat'" class="tip-popup tip-left">
            支持的变量：{playlistName}，{songName}，{artist}，{album}<br/>
            示例：{playlistName}/{album}-{artist}-{songName}
          </div>
        </div>
      </div>

      <div class="setting-row">
        <label class="setting-label">音质偏好</label>
        <div class="quality-btns">
          <button :class="{ active: settings.quality === 'high' }" @click="settings.quality = 'high'">高质量</button>
          <button :class="{ active: settings.quality === 'lossless' }" @click="settings.quality = 'lossless'">无损</button>
        </div>
      </div>

      <!-- 数据补全配置 -->
      <div class="setting-row">
        <div class="setting-label-col">
          <span class="setting-name">自动数据补全</span>
          <span class="setting-desc">数据补全将在同步完成后自动执行</span>
        </div>
        <div class="setting-control-row">
          <label class="switch">
            <input type="checkbox" v-model="settings.autoDataComplete" />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="setting-row checkbox-row">
        <label class="checkbox-label">
          <input type="checkbox" v-model="settings.dataCompleteCover" />
          <span>补全专辑封面</span>
        </label>
      </div>

      <div class="setting-row checkbox-row">
        <label class="checkbox-label">
          <input type="checkbox" v-model="settings.dataCompleteLyrics" />
          <span>补全歌词</span>
        </label>
      </div>

      <div class="setting-row checkbox-row">
        <label class="checkbox-label">
          <input type="checkbox" v-model="settings.dataCompleteArtist" />
          <span>补全艺人信息</span>
        </label>
      </div>

      <div class="sync-status">
        <div class="sync-status-header">
          <svg viewBox="0 0 24 24" width="16" height="16"><path fill="currentColor" d="M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm.5-13H11v6l5.25 3.15.75-1.23-4.5-2.67z"/></svg>
          <span>同步状态</span>
        </div>
        <div class="sync-status-info">
          <div class="sync-info-item">
            <span class="sync-info-label">上次同步时间</span>
            <span class="sync-info-value">{{ lastSyncTime }}</span>
          </div>
          <div class="sync-info-item">
            <span class="sync-info-label">下次同步时间</span>
            <span class="sync-info-value">{{ nextSyncTime }}</span>
          </div>
          <div class="sync-info-item">
            <span class="sync-info-label">距离下次同步</span>
            <span class="sync-info-value">{{ countdownText }}</span>
          </div>
        </div>
      </div>

      <div class="section-save-bar">
        <button class="save-btn" @click="saveSyncSettings" :disabled="savingSync">
          {{ savingSync ? '保存中...' : '保存' }}
        </button>
        <span v-if="syncSavedTip" class="save-tip">保存成功</span>
      </div>
    </div>

    <!-- SSL 设置 -->
    <div class="settings-card">
      <h3 class="card-title">SSL设置</h3>

      <!-- SSL 错误信息 -->
      <div v-if="sslError" class="ssl-error-banner">
        <svg viewBox="0 0 24 24" width="20" height="20"><path fill="#f56c6c" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
        <span>{{ sslError }}</span>
      </div>

      <div class="setting-row ssl-mode-row">
        <label class="setting-label">证书配置方式</label>
        <select v-model="settings.sslMode" class="ssl-mode-select">
          <option value="none">不使用 SSL</option>
          <option value="cert">本地证书（路径映射或上传）</option>
          <option value="acme">DNS-ACME 自动申请</option>
        </select>
      </div>

      <!-- 证书模式：路径映射或上传 -->
      <div v-if="settings.sslMode === 'cert'">
        <div class="setting-row">
          <label class="setting-label">证书路径<span class="required">*</span></label>
          <div class="path-input-wrap">
            <input v-model="settings.sslCertPath" class="setting-input" placeholder="证书路径" />
            <button class="upload-btn" title="上传证书文件" @click="triggerUpload('cert')">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M9 16h6v-6h4l-7-7-7 7h4v6zm-4 2h14v2H5v-2z"/></svg>
            </button>
            <button class="help-btn" @mouseenter="showTip('certPath')" @mouseleave="hideTip()">?</button>
            <div v-if="activeTip === 'certPath'" class="tip-popup tip-left">填写证书映射到容器内的绝对路径（包括证书文件名），或点击右侧按钮上传</div>
          </div>
        </div>

        <div class="setting-row">
          <label class="setting-label">私钥路径<span class="required">*</span></label>
          <div class="path-input-wrap">
            <input v-model="settings.sslKeyPath" class="setting-input" placeholder="私钥路径" />
            <button class="upload-btn" title="上传私钥文件" @click="triggerUpload('key')">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M9 16h6v-6h4l-7-7-7 7h4v6zm-4 2h14v2H5v-2z"/></svg>
            </button>
            <button class="help-btn" @mouseenter="showTip('keyPath')" @mouseleave="hideTip()">?</button>
            <div v-if="activeTip === 'keyPath'" class="tip-popup tip-left">填写证书密钥映射到容器内的绝对路径（包括密钥文件名），或点击右侧按钮上传</div>
          </div>
        </div>

        <div class="setting-row">
          <label class="setting-label">中间证书</label>
          <div class="path-input-wrap">
            <input v-model="settings.sslChainPath" class="setting-input" placeholder="中间证书路径" />
            <button class="upload-btn" title="上传中间证书文件" @click="triggerUpload('chain')">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M9 16h6v-6h4l-7-7-7 7h4v6zm-4 2h14v2H5v-2z"/></svg>
            </button>
            <button class="help-btn" @mouseenter="showTip('chainPath')" @mouseleave="hideTip()">?</button>
            <div v-if="activeTip === 'chainPath'" class="tip-popup tip-left">填写中间证书映射到容器内的绝对路径（包括证书文件名），或点击右侧按钮上传</div>
          </div>
        </div>

        <!-- 隐藏的文件选择器 -->
        <input ref="certFileInput" type="file" accept=".pem,.crt,.cert" style="display:none" @change="e => handleFileUpload('cert', e)" />
        <input ref="keyFileInput" type="file" accept=".key,.pem" style="display:none" @change="e => handleFileUpload('key', e)" />
        <input ref="chainFileInput" type="file" accept=".pem,.crt,.cert" style="display:none" @change="e => handleFileUpload('chain', e)" />

        <!-- 证书验证按钮和信息显示 -->
        <div class="setting-row" v-if="settings.sslCertPath && settings.sslKeyPath">
          <label class="setting-label">证书验证</label>
          <div class="cert-validation-section">
            <button class="validate-btn" @click="validateSSLCert" :disabled="certValidating">
              {{ certValidating ? '验证中...' : '验证证书' }}
            </button>

            <div v-if="certInfo" class="cert-info" :class="{ 'valid': certValid, 'invalid': !certValid }">
              <div class="cert-status">
                <span class="status-icon">{{ certValid ? '✓' : '✗' }}</span>
                <span class="status-text">{{ certInfo.message }}</span>
              </div>

              <div v-if="certValid && certInfo.subject" class="cert-details">
                <div class="detail-row">
                  <span class="detail-label">颁发给:</span>
                  <span class="detail-value">{{ certInfo.subject }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">颁发者:</span>
                  <span class="detail-value">{{ certInfo.issuer }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">有效期:</span>
                  <span class="detail-value">{{ certInfo.not_before }} 至 {{ certInfo.not_after }}</span>
                </div>
                <div class="detail-row" v-if="certInfo.domains && certInfo.domains.length > 0">
                  <span class="detail-label">域名:</span>
                  <span class="detail-value">{{ certInfo.domains.join(', ') }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- DNS-ACME 模式 -->
      <div v-if="settings.sslMode === 'acme'">
        <div class="setting-row">
          <label class="setting-label">DNS API 插件</label>
          <select v-model="selectedPluginId" class="ssl-mode-select" @change="onPluginChange">
            <option v-for="p in acmePlugins" :key="p.id" :value="p.id">{{ p.label }}</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label">验证延迟（秒）</label>
          <input type="number" v-model.number="pluginDelay" class="setting-input" style="width:120px" />
        </div>
        <div class="setting-row">
          <label class="setting-label">邮箱</label>
          <input v-model="settings.acmeEmail" class="setting-input" placeholder="your@email.com" />
        </div>
        <div class="setting-row">
          <label class="setting-label">域名</label>
          <input v-model="settings.acmeDomain" class="setting-input" placeholder="music.example.com" />
        </div>

        <!-- 动态插件字段 -->
        <template v-for="field in currentPluginFields" :key="field.key">
          <div class="setting-row">
            <label class="setting-label">
              {{ field.label }}<span v-if="field.required" class="required">*</span>
            </label>
            <div class="path-input-wrap">
              <input
                v-model="pluginFieldValues[field.key]"
                class="setting-input"
                :type="getFieldInputType(field.key)"
                :placeholder="field.placeholder || field.label"
                @focus="onFieldFocus(field.key)"
                @blur="onFieldBlur(field.key)"
              />
              <button v-if="field.hint" class="help-btn" @mouseenter="showTip(field.key)" @mouseleave="hideTip()">?</button>
              <div v-if="activeTip === field.key" class="tip-popup tip-left">{{ field.hint }}</div>
            </div>
          </div>
        </template>

        <div class="acme-actions">
          <button class="apply-acme-btn" @click="applyACME" :disabled="acmeApplying">
            {{ acmeApplying ? '申请中...' : '申请证书' }}
          </button>
          <span v-if="acmeMessage" class="acme-message" :class="acmeSuccess ? 'success' : 'error'">{{ acmeMessage }}</span>
        </div>

        <!-- ACME 证书验证详情 -->
        <div class="setting-row" v-if="settings.sslCertPath && settings.sslKeyPath">
          <label class="setting-label">证书验证</label>
          <div class="cert-validation-section">
            <button class="validate-btn" @click="validateSSLCert" :disabled="certValidating">
              {{ certValidating ? '验证中...' : '验证证书' }}
            </button>

            <div v-if="certInfo" class="cert-info" :class="{ 'valid': certValid, 'invalid': !certValid }">
              <div class="cert-status">
                <span class="status-icon">{{ certValid ? '✓' : '✗' }}</span>
                <span class="status-text">{{ certInfo.message }}</span>
              </div>

              <div v-if="certValid && certInfo.subject" class="cert-details">
                <div class="detail-row">
                  <span class="detail-label">颁发给:</span>
                  <span class="detail-value">{{ certInfo.subject }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">颁发者:</span>
                  <span class="detail-value">{{ certInfo.issuer }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">有效期:</span>
                  <span class="detail-value">{{ certInfo.not_before }} 至 {{ certInfo.not_after }}</span>
                </div>
                <div class="detail-row" v-if="certInfo.domains && certInfo.domains.length > 0">
                  <span class="detail-label">域名:</span>
                  <span class="detail-value">{{ certInfo.domains.join(', ') }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 端口配置 -->
      <div class="setting-row port-row">
        <div class="port-group">
          <label class="port-label">HTTP端口</label>
          <input type="number" v-model.number="settings.httpPort" class="port-input" />
        </div>
        <div class="port-group">
          <label class="port-label">HTTPS端口</label>
          <input type="number" v-model.number="settings.httpsPort" class="port-input" />
        </div>
      </div>

      <div class="setting-row">
        <label class="setting-label">HTTP自动跳转HTTPS</label>
        <label class="switch" :class="{ disabled: !sslRedirectEnabled }">
          <input type="checkbox" v-model="settings.sslRedirect" :disabled="!sslRedirectEnabled" />
          <span class="slider"></span>
        </label>
        <span v-if="!sslRedirectEnabled" class="redirect-hint">请先验证证书</span>
      </div>

      <div class="section-save-bar">
        <button class="save-btn" @click="saveSSLSettings" :disabled="savingSSL">
          {{ savingSSL ? '保存中...' : '保存' }}
        </button>
        <span v-if="sslSavedTip" class="save-tip">保存成功</span>
      </div>
    </div>

    <!-- 个性化 -->
    <div class="settings-card">
      <h3 class="card-title">个性化</h3>

      <div class="setting-row">
        <label class="setting-label">关闭页面切换动画</label>
        <label class="switch">
          <input type="checkbox" v-model="disablePageAnimation" />
          <span class="slider"></span>
        </label>
      </div>

      <div class="section-save-bar">
        <button class="save-btn" @click="savePersonalization" :disabled="savingPersonal">
          {{ savingPersonal ? '保存中...' : '保存' }}
        </button>
        <span v-if="personalSavedTip" class="save-tip">保存成功</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { settingsAPI } from '../api'
import { ElMessage } from 'element-plus'

const activeTip = ref('')
const certFileInput = ref<HTMLInputElement>()
const keyFileInput = ref<HTMLInputElement>()
const chainFileInput = ref<HTMLInputElement>()

// 分区块保存状态
const savingDownload = ref(false)
const downloadSavedTip = ref(false)
const savingSync = ref(false)
const syncSavedTip = ref(false)
const savingSSL = ref(false)
const sslSavedTip = ref(false)
const sslError = ref('')
const savingPersonal = ref(false)
const personalSavedTip = ref(false)
const disablePageAnimation = ref(false)

const settings = ref({
  downloadPath: '',
  songFormat: '{songName} - {artist}',
  autoSync: false,
  syncInterval: 12,
  syncUnit: 'hour' as 'hour' | 'day',
  deleteRemoved: false,
  playlistFormat: '{playlistName}/{songName} - {artist}',
  quality: 'high' as 'high' | 'lossless',
  resumeDownloads: true,
  autoDataComplete: false,
  dataCompleteInterval: 24,
  dataCompleteUnit: 'hour' as 'hour' | 'day',
  dataCompleteCover: true,
  dataCompleteLyrics: true,
  dataCompleteArtist: true,
  sslMode: 'none' as 'none' | 'cert' | 'acme',
  sslCertPath: '',
  sslKeyPath: '',
  sslChainPath: '',
  httpPort: 33550,
  httpsPort: 33551,
  sslRedirect: false,
  acmeProvider: '',
  acmeEmail: '',
  acmeDomain: '',
  acmeFields: {} as Record<string, string>,
  lastSyncTime: '',
  nextSyncTime: '',
  syncMode: 'interval' as 'interval' | 'schedule',
  syncWeekdays: [] as number[],
  syncTime: '08:00'
})

const weekdays = [
  { value: 1, label: '周一' },
  { value: 2, label: '周二' },
  { value: 3, label: '周三' },
  { value: 4, label: '周四' },
  { value: 5, label: '周五' },
  { value: 6, label: '周六' },
  { value: 0, label: '周日' }
]

const toggleWeekday = (day: number) => {
  const index = settings.value.syncWeekdays.indexOf(day)
  if (index > -1) {
    settings.value.syncWeekdays.splice(index, 1)
  } else {
    settings.value.syncWeekdays.push(day)
  }
}

const acmeApplying = ref(false)
const acmeMessage = ref('')
const acmeSuccess = ref(false)

// SSL 证书验证状态
const certValidating = ref(false)
const certValid = ref(false)
const certInfo = ref<{
  valid: boolean
  message: string
  subject?: string
  issuer?: string
  not_before?: string
  not_after?: string
  domains?: string[]
} | null>(null)

interface PluginField {
  key: string
  label: string
  type: string
  required: boolean
  placeholder: string
  hint: string
}

interface ACMEPlugin {
  id: string
  label: string
  delay: number
  fields: PluginField[]
}

const acmePlugins = ref<ACMEPlugin[]>([])
const selectedPluginId = ref('')
const pluginDelay = ref(30)
const pluginFieldValues = ref<Record<string, string>>({})
const maskedFieldValues = ref<Record<string, string>>({}) // 存储敏感字段的真实值
const focusedField = ref<string | null>(null) // 当前聚焦的字段

const currentPlugin = computed(() => {
  return acmePlugins.value.find(p => p.id === selectedPluginId.value) || null
})

const currentPluginFields = computed(() => {
  return currentPlugin.value?.fields || []
})

const sslRedirectEnabled = computed(() => {
  return certValid.value && settings.value.sslCertPath !== '' && settings.value.sslKeyPath !== ''
})

const onPluginChange = () => {
  const plugin = currentPlugin.value
  if (plugin) {
    pluginDelay.value = plugin.delay
    pluginFieldValues.value = {}
    maskedFieldValues.value = {}
    plugin.fields.forEach(f => {
      if (settings.value.acmeFields[f.key]) {
        const realValue = settings.value.acmeFields[f.key]
        maskedFieldValues.value[f.key] = realValue // 存储真实值
        pluginFieldValues.value[f.key] = maskValue(realValue) // 显示打码值
      }
    })
  }
}

const nextSyncTime = computed(() => {
  if (!settings.value.autoSync) return '--'
  // 使用后端返回的真实下次同步时间
  if (settings.value.nextSyncTime) {
    const next = new Date(settings.value.nextSyncTime)
    return next.toLocaleString('zh-CN', { year: 'numeric', month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit' }).replace(/\//g, '/')
  }
  return '--'
})

const lastSyncTime = computed(() => {
  if (!settings.value.lastSyncTime) return '--'
  const last = new Date(settings.value.lastSyncTime)
  return last.toLocaleString('zh-CN', { year: 'numeric', month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit' }).replace(/\//g, '/')
})

const countdownText = computed(() => {
  if (!settings.value.autoSync || !settings.value.nextSyncTime) return '--'
  const next = new Date(settings.value.nextSyncTime)
  const now = new Date()
  const diff = next.getTime() - now.getTime()
  if (diff <= 0) return '即将同步'
  const minutes = Math.floor(diff / 60000)
  if (minutes < 60) return `${minutes} 分钟`
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return `${hours} 小时 ${mins} 分钟`
})

const showTip = (name: string) => { activeTip.value = name }
const hideTip = () => { activeTip.value = '' }

// 敏感字段打码：前后保留3位，中间用 * 替代
const maskValue = (val: string) => {
  if (!val) return ''
  if (val.length <= 6) return '***'
  return val.slice(0, 3) + '***' + val.slice(-3)
}

// 敏感字段列表（这些字段需要打码显示）
const sensitiveFieldKeys = ['token', 'secret', 'key', 'password', 'apikey', 'api_key', 'global_api_key', 'email', 'account', 'zone', 'client_id', 'tenant', 'subscription', 'consumer']

const isSensitiveField = (fieldKey: string) => {
  const lower = fieldKey.toLowerCase()
  return sensitiveFieldKeys.some(k => lower.includes(k))
}

// 获取输入框类型：敏感字段聚焦时为 text，失焦时为 text（用打码值）
const getFieldInputType = (_fieldKey: string) => {
  return 'text'
}

// 聚焦时显示真实值
const onFieldFocus = (fieldKey: string) => {
  focusedField.value = fieldKey
  if (maskedFieldValues.value[fieldKey]) {
    pluginFieldValues.value[fieldKey] = maskedFieldValues.value[fieldKey]
  }
}

// 失焦时恢复打码显示
const onFieldBlur = (fieldKey: string) => {
  focusedField.value = null
  if (maskedFieldValues.value[fieldKey]) {
    // 用户编辑后更新真实值
    maskedFieldValues.value[fieldKey] = pluginFieldValues.value[fieldKey]
    pluginFieldValues.value[fieldKey] = maskValue(pluginFieldValues.value[fieldKey])
  }
}

const triggerUpload = (type: 'cert' | 'key' | 'chain') => {
  const input = type === 'cert' ? certFileInput.value : type === 'key' ? keyFileInput.value : chainFileInput.value
  input?.click()
}

const handleFileUpload = async (type: 'cert' | 'key' | 'chain', e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('type', type)
    const res = await settingsAPI.uploadSSLCertFile(formData)
    if (res.data.container_path) {
      if (type === 'cert') settings.value.sslCertPath = res.data.container_path
      else if (type === 'key') settings.value.sslKeyPath = res.data.container_path
      else settings.value.sslChainPath = res.data.container_path
      ElMessage.success(`${type === 'cert' ? '证书' : type === 'key' ? '私钥' : '中间证书'}已上传`)

      // 上传证书/私钥后默认关闭跳转，需重新验证才能开启
      settings.value.sslRedirect = false
      certValid.value = false
      certInfo.value = null

      // 如果证书和私钥都已上传，自动验证
      if (settings.value.sslCertPath && settings.value.sslKeyPath) {
        await validateSSLCert()
      }
    }
  } catch {
    ElMessage.error('上传失败')
  } finally {
    input.value = ''
  }
}

const validateSSLCert = async () => {
  if (!settings.value.sslCertPath || !settings.value.sslKeyPath) {
    ElMessage.warning('请先上传证书和私钥')
    return
  }

  certValidating.value = true
  certInfo.value = null
  certValid.value = false

  try {
    const res = await settingsAPI.validateSSLCert({
      cert_path: settings.value.sslCertPath,
      key_path: settings.value.sslKeyPath
    })

    certInfo.value = res.data
    certValid.value = res.data.valid

    if (res.data.valid) {
      ElMessage.success('证书验证通过')
    } else {
      ElMessage.error(res.data.message || '证书验证失败')
    }
  } catch (error: any) {
    ElMessage.error('验证失败: ' + (error.response?.data?.error || error.message))
    certInfo.value = {
      valid: false,
      message: '验证请求失败'
    }
  } finally {
    certValidating.value = false
  }
}

const loadPlugins = async () => {
  try {
    const res = await settingsAPI.getACMEPlugins()
    acmePlugins.value = res.data.plugins || []
  } catch {}
}

const loadSettings = async () => {
  try {
    const res = await settingsAPI.getSettings()
    const s = res.data.settings
    if (s) {
      settings.value.downloadPath = s.download_path || ''
      settings.value.songFormat = s.song_format || '{songName} - {artist}'
      settings.value.autoSync = s.auto_sync === 'true'
      settings.value.syncInterval = parseInt(s.sync_interval) || 12
      settings.value.syncUnit = s.sync_unit === 'day' ? 'day' : 'hour'
      settings.value.deleteRemoved = s.delete_removed === 'true'
      settings.value.playlistFormat = s.playlist_format || '{playlistName}/{songName} - {artist}'
      settings.value.quality = s.quality || 'high'
      settings.value.resumeDownloads = s.resume_downloads !== 'false'
      settings.value.autoDataComplete = s.auto_data_complete === 'true'
      settings.value.dataCompleteInterval = parseInt(s.data_complete_interval) || 24
      settings.value.dataCompleteUnit = s.data_complete_unit === 'day' ? 'day' : 'hour'
      settings.value.dataCompleteCover = s.data_complete_cover !== 'false'
      settings.value.dataCompleteLyrics = s.data_complete_lyrics !== 'false'
      settings.value.dataCompleteArtist = s.data_complete_artist !== 'false'
      settings.value.sslMode = s.ssl_mode || 'none'
      settings.value.sslCertPath = s.ssl_cert_path || ''
      settings.value.sslKeyPath = s.ssl_key_path || ''
      settings.value.sslChainPath = s.ssl_chain_path || ''
      settings.value.httpPort = parseInt(s.http_port) || 33550
      settings.value.httpsPort = parseInt(s.https_port) || 33551
      settings.value.sslRedirect = s.ssl_redirect === 'true'
      sslError.value = s.ssl_error || ''
      settings.value.acmeProvider = s.acme_provider || ''
      settings.value.acmeEmail = s.acme_email || ''
      settings.value.acmeDomain = s.acme_domain || ''
      // 同步已保存的插件选择到 UI
      if (s.acme_provider) {
        selectedPluginId.value = s.acme_provider
      }
      if (s.acme_fields) {
        settings.value.acmeFields = typeof s.acme_fields === 'string'
          ? JSON.parse(s.acme_fields)
          : s.acme_fields
        // 同步已保存的字段值到 UI，并对敏感字段进行打码
        if (typeof s.acme_fields === 'string') {
          try {
            const parsed = JSON.parse(s.acme_fields)
            maskedFieldValues.value = { ...parsed } // 存储真实值
            pluginFieldValues.value = { ...parsed }
            // 对敏感字段进行打码显示
            Object.keys(pluginFieldValues.value).forEach(key => {
              if (isSensitiveField(key) && pluginFieldValues.value[key]) {
                pluginFieldValues.value[key] = maskValue(pluginFieldValues.value[key])
              }
            })
          } catch {}
        } else {
          const fields = s.acme_fields as Record<string, string>
          maskedFieldValues.value = { ...fields } // 存储真实值
          pluginFieldValues.value = { ...fields }
          // 对敏感字段进行打码显示
          Object.keys(pluginFieldValues.value).forEach(key => {
            if (isSensitiveField(key) && pluginFieldValues.value[key]) {
              pluginFieldValues.value[key] = maskValue(pluginFieldValues.value[key])
            }
          })
        }
      }
      settings.value.lastSyncTime = s.last_sync_time || ''
      settings.value.nextSyncTime = s.next_sync_time || ''
      settings.value.syncMode = s.sync_mode === 'schedule' ? 'schedule' : 'interval'
      if (s.sync_weekdays) {
        try {
          settings.value.syncWeekdays = typeof s.sync_weekdays === 'string'
            ? JSON.parse(s.sync_weekdays)
            : s.sync_weekdays
        } catch {
          settings.value.syncWeekdays = []
        }
      }
      settings.value.syncTime = s.sync_time || '08:00'
      const savedAnim = localStorage.getItem('disablePageAnimation')
      if (savedAnim !== null) {
        disablePageAnimation.value = savedAnim === 'true'
      } else if (s.disable_page_animation !== undefined) {
        disablePageAnimation.value = String(s.disable_page_animation) === 'true' || s.disable_page_animation === true
      }
    }

    if (settings.value.sslCertPath && settings.value.sslKeyPath) {
      await validateSSLCert()
    }
  } catch {}
}

// 保存本地下载配置
const saveDownloadSettings = async () => {
  savingDownload.value = true
  downloadSavedTip.value = false
  try {
    const data = {
      download_path: settings.value.downloadPath,
      song_format: settings.value.songFormat
    }
    await settingsAPI.updateSettings(data)
    downloadSavedTip.value = true
    await loadSettings()
    setTimeout(() => { downloadSavedTip.value = false }, 2000)
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingDownload.value = false
  }
}

// 保存歌单同步配置
const saveSyncSettings = async () => {
  savingSync.value = true
  syncSavedTip.value = false
  try {
    const data = {
      auto_sync: settings.value.autoSync.toString(),
      sync_interval: settings.value.syncInterval.toString(),
      sync_unit: settings.value.syncUnit,
      sync_mode: settings.value.syncMode,
      sync_weekdays: JSON.stringify(settings.value.syncWeekdays),
      sync_time: settings.value.syncTime,
      delete_removed: settings.value.deleteRemoved.toString(),
      playlist_format: settings.value.playlistFormat,
      quality: settings.value.quality,
      resume_downloads: settings.value.resumeDownloads.toString(),
      auto_data_complete: settings.value.autoDataComplete.toString(),
      data_complete_cover: settings.value.dataCompleteCover.toString(),
      data_complete_lyrics: settings.value.dataCompleteLyrics.toString(),
      data_complete_artist: settings.value.dataCompleteArtist.toString()
    }
    await settingsAPI.updateSettings(data)
    syncSavedTip.value = true
    await loadSettings()
    setTimeout(() => { syncSavedTip.value = false }, 2000)
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingSync.value = false
  }
}

// 保存 SSL 配置
const saveSSLSettings = async () => {
  // 如果启用 SSL，必须先验证证书
  if (settings.value.sslMode === 'cert' && settings.value.sslCertPath && settings.value.sslKeyPath) {
    if (!certValid.value) {
      ElMessage.error('请先验证证书有效性')
      return
    }
  }

  savingSSL.value = true
  sslSavedTip.value = false
  try {
    const data: any = {
      ssl_mode: settings.value.sslMode,
      http_port: settings.value.httpPort.toString(),
      https_port: settings.value.httpsPort.toString(),
      ssl_redirect: settings.value.sslRedirect.toString()
    }
    if (settings.value.sslMode === 'cert') {
      data.ssl_cert_path = settings.value.sslCertPath
      data.ssl_key_path = settings.value.sslKeyPath
      data.ssl_chain_path = settings.value.sslChainPath
    } else if (settings.value.sslMode === 'acme') {
      data.acme_provider = selectedPluginId.value
      data.acme_email = settings.value.acmeEmail
      data.acme_domain = settings.value.acmeDomain
      // 保存真实值，而不是打码后的值
      const realFieldValues = { ...pluginFieldValues.value }
      Object.keys(realFieldValues).forEach(key => {
        if (maskedFieldValues.value[key]) {
          realFieldValues[key] = maskedFieldValues.value[key]
        }
      })
      data.acme_fields = JSON.stringify(realFieldValues)
    }
    await settingsAPI.updateSettings(data)
    sslSavedTip.value = true
    await loadSettings()
    setTimeout(() => { sslSavedTip.value = false }, 2000)
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingSSL.value = false
  }
}

// 保存个性化配置
const savePersonalization = async () => {
  savingPersonal.value = true
  personalSavedTip.value = false
  try {
    await settingsAPI.updateSettings({
      disable_page_animation: disablePageAnimation.value.toString()
    })
    // 同步到 localStorage 供 App.vue 读取
    localStorage.setItem('disablePageAnimation', disablePageAnimation.value.toString())
    personalSavedTip.value = true
    await loadSettings()
    setTimeout(() => { personalSavedTip.value = false }, 2000)
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingPersonal.value = false
  }
}

const applyACME = async () => {
  if (!selectedPluginId.value || !settings.value.acmeEmail || !settings.value.acmeDomain) {
    ElMessage.warning('请选择 DNS 插件、填写邮箱和域名')
    return
  }
  acmeApplying.value = true
  acmeMessage.value = ''
  acmeSuccess.value = false
  try {
    // 构建真实值，而不是打码后的值
    const realFieldValues = { ...pluginFieldValues.value }
    Object.keys(realFieldValues).forEach(key => {
      if (maskedFieldValues.value[key]) {
        realFieldValues[key] = maskedFieldValues.value[key]
      }
    })

    const res = await settingsAPI.applyACME({
      provider: selectedPluginId.value,
      email: settings.value.acmeEmail,
      domain: settings.value.acmeDomain,
      fields: realFieldValues
    })

    // 后端已经保存了 ssl_mode=acme、ssl_cert_path、ssl_key_path，并触发了热加载
    const sslReloadOk = res.data?.ssl_reload_success === true

    if (sslReloadOk) {
      acmeMessage.value = res.data.message || '证书申请成功，HTTPS服务已启用'
      acmeSuccess.value = true
    } else {
      acmeMessage.value = (res.data.message || '证书申请成功') + '，但 HTTPS 热加载失败，请检查日志'
      acmeSuccess.value = false
    }

    // 重新加载设置，确保 UI 显示正确的 ssl_mode 和证书路径
    await loadSettings()
  } catch (e: any) {
    acmeMessage.value = e?.response?.data?.error || '证书申请失败'
    acmeSuccess.value = false
  } finally {
    acmeApplying.value = false
  }
}

onMounted(() => {
  // 先从 localStorage 恢复个性化设置，避免闪烁
  const savedAnimationDisabled = localStorage.getItem('disablePageAnimation')
  if (savedAnimationDisabled !== null) {
    disablePageAnimation.value = savedAnimationDisabled === 'true'
  }
  loadPlugins()
  loadSettings()
})

// 同步插件字段值到 settings.acmeFields
const syncPluginFields = () => {
  settings.value.acmeFields = { ...pluginFieldValues.value }
}

watch(pluginFieldValues, syncPluginFields, { deep: true })
</script>

<style scoped>
.settings-page {
  padding: 24px 32px;
  max-width: 900px;
  min-height: 100%;
}

.settings-card {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 24px 32px;
  margin-bottom: 24px;
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

.setting-row:last-child {
  border-bottom: none;
}

.setting-label {
  width: 200px;
  min-width: 200px;
  text-align: right;
  padding-right: 24px;
  font-size: 14px;
  color: var(--text-primary);
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

.setting-input-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}

.setting-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
}

.setting-input:focus {
  border-color: #FFFA00;
}

.help-btn {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tip-popup {
  position: absolute;
  bottom: calc(100% + 8px);
  right: 0;
  background: #1a1a2e;
  color: #eee;
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.6;
  white-space: nowrap;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  max-width: 500px;
  white-space: normal;
}

.tip-popup::after {
  content: '';
  position: absolute;
  top: 100%;
  right: 12px;
  border: 6px solid transparent;
  border-top-color: #1a1a2e;
}

.setting-control-row {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.control-text {
  font-size: 14px;
  color: var(--text-primary);
}

.number-input {
  width: 70px;
  padding: 4px 8px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 14px;
  text-align: center;
}

.unit-btns {
  display: flex;
}

.unit-btns button {
  padding: 4px 14px;
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
}

.unit-btns button:first-child {
  border-radius: 4px 0 0 4px;
}

.unit-btns button:last-child {
  border-radius: 0 4px 4px 0;
  border-left: none;
}

.unit-btns button.active {
  background: #FFFA00;
  border-color: #FFFA00;
  color: #000;
}

.mode-btns {
  display: flex;
}

.mode-btns button {
  padding: 4px 14px;
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
}

.mode-btns button:first-child {
  border-radius: 4px 0 0 4px;
}

.mode-btns button:last-child {
  border-radius: 0 4px 4px 0;
  border-left: none;
}

.mode-btns button.active {
  background: #FFFA00;
  border-color: #FFFA00;
  color: #000;
}

.weekday-btns {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.weekday-btns button {
  padding: 4px 10px;
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  border-radius: 4px;
}

.weekday-btns button.active {
  background: #FFFA00;
  border-color: #FFFA00;
  color: #000;
}

.time-input {
  padding: 4px 8px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 14px;
}

.checkbox-row {
  justify-content: flex-end;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-primary);
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  accent-color: #FFFA00;
  width: 16px;
  height: 16px;
}

.quality-btns {
  display: flex;
}

.quality-btns button {
  padding: 6px 18px;
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
}

.quality-btns button:first-child {
  border-radius: 4px 0 0 4px;
}

.quality-btns button:last-child {
  border-radius: 0 4px 4px 0;
  border-left: none;
}

.quality-btns button.active {
  background: #FFFA00;
  border-color: #FFFA00;
  color: #000;
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
  top: 0; left: 0; right: 0; bottom: 0;
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

.switch.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.switch.disabled input {
  cursor: not-allowed;
}

.redirect-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-left: 8px;
}

/* Sync status */
.sync-status {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.sync-status-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.sync-status-info {
  display: flex;
  justify-content: space-between;
}

.sync-info-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.sync-info-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.sync-info-value {
  font-size: 14px;
  color: var(--text-primary);
}

/* Section save bar */
.section-save-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 20px 0 0 0;
  margin-top: 16px;
  border-top: 1px solid var(--border-color);
}

.save-btn {
  padding: 8px 32px;
  background: #FFFA00;
  color: #000;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.save-btn:hover {
  opacity: 0.85;
}

.save-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.save-tip {
  color: #4caf50;
  font-size: 13px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* SSL specific styles */
.ssl-mode-row {
  justify-content: flex-start;
}

.ssl-mode-select {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
}

.ssl-mode-select:focus {
  border-color: #FFFA00;
}

.path-input-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}

.upload-btn {
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.upload-btn:hover {
  background: var(--bg-color);
  border-color: #FFFA00;
}

.required {
  color: #f44336;
  margin-left: 4px;
}

.port-row {
  justify-content: flex-start;
  gap: 40px;
}

.port-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.port-label {
  font-size: 14px;
  color: var(--text-primary);
}

.port-input {
  width: 100px;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
}

.port-input:focus {
  border-color: #FFFA00;
}

.acme-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--border-color);
}

.apply-acme-btn {
  padding: 8px 24px;
  background: #FFFA00;
  color: #000;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.apply-acme-btn:hover {
  opacity: 0.85;
}

.apply-acme-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.acme-message {
  font-size: 13px;
}

.acme-message.success {
  color: #4caf50;
}

.acme-message.error {
  color: #f44336;
}

/* 证书验证样式 */
.cert-validation-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.validate-btn {
  padding: 8px 24px;
  background: #FFFA00;
  color: #000;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
  align-self: flex-start;
}

.validate-btn:hover {
  opacity: 0.85;
}

.validate-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cert-info {
  padding: 12px 16px;
  border-radius: 6px;
  border: 1px solid;
}

.cert-info.valid {
  background: rgba(76, 175, 80, 0.1);
  border-color: #4caf50;
}

.cert-info.invalid {
  background: rgba(244, 67, 54, 0.1);
  border-color: #f44336;
}

.cert-status {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.status-icon {
  font-size: 18px;
  font-weight: bold;
}

.cert-info.valid .status-icon {
  color: #4caf50;
}

.cert-info.invalid .status-icon {
  color: #f44336;
}

.status-text {
  font-size: 14px;
  font-weight: 600;
}

.cert-info.valid .status-text {
  color: #4caf50;
}

.cert-info.invalid .status-text {
  color: #f44336;
}

.cert-details {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--border-color);
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

/* 移动端适配 */
@media (max-width: 768px) {
  .settings-page {
    padding: 16px;
  }

  .settings-card {
    padding: 16px;
  }

  .card-title {
    font-size: 18px;
    margin-bottom: 16px;
  }

  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .setting-label {
    width: 100%;
    text-align: left;
  }

  .setting-control {
    width: 100%;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .quality-buttons,
  .unit-buttons {
    width: 100%;
  }

  .quality-buttons button,
  .unit-buttons button {
    flex: 1;
  }

  .sync-status {
    flex-direction: column;
    gap: 12px;
  }

  .sync-status-label {
    min-width: auto;
  }
}
</style>
