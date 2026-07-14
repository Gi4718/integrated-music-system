<template>
  <el-config-provider :locale="zhCn">
    <div class="app-container">
      <TopNav />
      <main class="main-content">
        <div class="content-wrapper">
          <router-view />
        </div>
      </main>

      <!-- 底部播放控件 -->
      <MusicPlayer />

      <!-- 页面切换动画遮罩层 -->
      <div class="mask-black" ref="blackMask"></div>
      <div class="mask-yellow" ref="yellowMask"></div>
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
// @ts-ignore
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { useThemeStore } from './stores/theme'
import { onMounted, ref, provide } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import TopNav from './components/TopNav.vue'
import MusicPlayer from './components/MusicPlayer.vue'

const theme = useThemeStore()
const router = useRouter()
const route = useRoute()

// DOM 引用
const blackMask = ref<HTMLDivElement | null>(null)
const yellowMask = ref<HTMLDivElement | null>(null)
const isAnimating = ref(false)

// 页面顺序（用于判断前进/后退）
const pageOrder = ['/', '/search', '/playlist', '/downloads', '/settings', '/login']

const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// 执行页面切换动画
const navigateWithAnimation = async (targetPath: string) => {
  // 检查是否禁用页面动画
  const disabled = localStorage.getItem('disablePageAnimation') === 'true'
  if (disabled) {
    await router.push(targetPath)
    return
  }

  if (!blackMask.value || !yellowMask.value || isAnimating.value) {
    await router.push(targetPath)
    return
  }

  const currentIndex = pageOrder.indexOf(route.path)
  const targetIndex = pageOrder.indexOf(targetPath)

  if (currentIndex === -1 || targetIndex === -1) {
    await router.push(targetPath)
    return
  }

  isAnimating.value = true
  const isForward = targetIndex > currentIndex

  const enterFrom = isForward ? 'translateX(100vw)' : 'translateX(-100vw)'
  const exitTo = isForward ? 'translateX(-100vw)' : 'translateX(100vw)'

  // 将遮罩重置到入场起点（禁用过渡，立即归位）
  blackMask.value.style.transition = 'none'
  yellowMask.value.style.transition = 'none'
  blackMask.value.style.transform = enterFrom
  yellowMask.value.style.transform = enterFrom

  // 强制重绘，确保浏览器应用无过渡的初始位置
  void blackMask.value.offsetHeight

  // 恢复过渡效果
  blackMask.value.style.transition = ''
  yellowMask.value.style.transition = ''

  // 阶段1：黑色遮罩入场 (0-0.2s)
  blackMask.value.style.transform = 'translateX(0)'

  await sleep(150)

  // 阶段2：黄色遮罩递进覆盖 (0.15-0.35s)
  yellowMask.value.style.transform = 'translateX(0)'

  await sleep(200)

  // 切换路由
  await router.push(targetPath)

  // 阶段3：遮罩归位 (0.35-0.6s)
  blackMask.value.style.transform = exitTo

  await sleep(50)

  yellowMask.value.style.transform = exitTo

  await sleep(250)

  isAnimating.value = false
}

// 提供给子组件使用
provide('navigateWithAnimation', navigateWithAnimation)

onMounted(() => {
  theme.initTheme()
})
</script>

<style>
:root {
  --primary-color: #FFD700;
  --primary-hover: #FFC107;
  --bg-color: #FFFFFF;
  --bg-secondary: #F5F5F5;
  --text-primary: #1A1A1A;
  --text-secondary: #666666;
  --border-color: #E0E0E0;
  --shadow-color: rgba(0, 0, 0, 0.1);
  --card-bg: #FFFFFF;
  --nav-bg: #FFD700;
  --accent-gradient: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
}

[data-theme="dark"] {
  --primary-color: #FFD700;
  --primary-hover: #FFC107;
  --bg-color: #0A0A0A;
  --bg-secondary: #1A1A1A;
  --text-primary: #FFFFFF;
  --text-secondary: #B0B0B0;
  --border-color: #2A2A2A;
  --shadow-color: rgba(255, 215, 0, 0.1);
  --card-bg: #1A1A1A;
  --nav-bg: #1A1A1A;
  --accent-gradient: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  width: 100%;
  height: 100%;
}

body {
  font-family: 'Helvetica Neue', Arial, sans-serif;
  background-color: var(--bg-color);
  color: var(--text-primary);
  transition: background-color 0.3s ease, color 0.3s ease;
}

#app {
  width: 100%;
  height: 100%;
}

.app-container {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
  background-color: var(--bg-color);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.content-wrapper {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

/* 页面切换动画遮罩层 */
.mask-black {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: #000000;
  z-index: 9998;
  transform: translateX(100vw);
  transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  will-change: transform;
  pointer-events: none;
}

.mask-yellow {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: #FFD700;
  z-index: 9999;
  transform: translateX(100vw);
  transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  will-change: transform;
  pointer-events: none;
}

/* Element Plus 主题覆盖 */
.el-button--primary {
  --el-button-bg-color: var(--primary-color) !important;
  --el-button-border-color: var(--primary-color) !important;
  --el-button-hover-bg-color: var(--primary-hover) !important;
  --el-button-hover-border-color: var(--primary-hover) !important;
  --el-button-text-color: #000 !important;
}

.el-input__wrapper {
  background-color: var(--card-bg) !important;
  box-shadow: 0 0 0 1px var(--border-color) inset !important;
}

.el-input__inner {
  color: var(--text-primary) !important;
}

.el-table {
  --el-table-bg-color: var(--card-bg) !important;
  --el-table-tr-bg-color: var(--card-bg) !important;
  --el-table-header-bg-color: var(--bg-secondary) !important;
  --el-table-row-hover-bg-color: var(--bg-secondary) !important;
  --el-table-text-color: var(--text-primary) !important;
  --el-table-header-text-color: var(--text-primary) !important;
  --el-table-border-color: var(--border-color) !important;
}

.el-card {
  --el-card-bg-color: var(--card-bg) !important;
  border-color: var(--border-color) !important;
}

.el-form-item__label {
  color: var(--text-primary) !important;
}

.el-divider__text {
  background-color: var(--bg-color) !important;
  color: var(--primary-color) !important;
  font-weight: 600;
}

.el-tabs__item {
  color: var(--text-secondary) !important;
}

.el-tabs__item.is-active {
  color: var(--primary-color) !important;
}

.el-tabs__active-bar {
  background-color: var(--primary-color) !important;
}

.el-tag {
  --el-tag-bg-color: var(--bg-secondary) !important;
  --el-tag-text-color: var(--text-primary) !important;
  --el-tag-border-color: var(--border-color) !important;
}

.el-empty__description p {
  color: var(--text-secondary) !important;
}
</style>
