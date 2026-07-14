<template>
  <div class="page-transition-container">
    <!-- 底层：页面内容 -->
    <div class="content-layer">
      <slot></slot>
    </div>

    <!-- 中层1：黑色遮罩 -->
    <div class="mask-black" ref="blackMask"></div>

    <!-- 中层2：黄色遮罩 -->
    <div class="mask-yellow" ref="yellowMask"></div>

    <!-- 顶层：右侧常驻标题栏 -->
    <div class="side-title-bar">
      <div class="title-text">集成音乐系统</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, provide } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// DOM 引用
const blackMask = ref<HTMLDivElement | null>(null)
const yellowMask = ref<HTMLDivElement | null>(null)

// 动画状态
const isAnimating = ref(false)

// 执行页面切换动画
const navigateWithAnimation = async (targetPath: string, currentIndex: number, targetIndex: number) => {
  if (!blackMask.value || !yellowMask.value || isAnimating.value) {
    // 如果动画进行中或 DOM 未就绪，直接跳转
    await router.push(targetPath)
    return
  }

  isAnimating.value = true
  const isForward = targetIndex > currentIndex

  // 阶段1：黑色遮罩入场 (0-0.2s)
  blackMask.value.style.animation = isForward
    ? 'slideInFromRight 0.2s cubic-bezier(0.2, 0.8, 0.2, 1) forwards'
    : 'slideInFromLeft 0.2s cubic-bezier(0.2, 0.8, 0.2, 1) forwards'

  await sleep(150) // 等待到0.15s，黑色遮罩即将铺满

  // 阶段2：黄色遮罩递进覆盖 (0.15-0.35s，与阶段1重叠0.05s)
  yellowMask.value.style.animation = isForward
    ? 'slideInFromRight 0.2s cubic-bezier(0.2, 0.8, 0.2, 1) forwards'
    : 'slideInFromLeft 0.2s cubic-bezier(0.2, 0.8, 0.2, 1) forwards'

  await sleep(200) // 等待到0.35s，黄色遮罩完全覆盖

  // 切换路由（在遮罩完全覆盖时）
  await router.push(targetPath)

  // 阶段3：遮罩归位 (0.35-0.6s)
  // 黄色遮罩先收回右侧标题栏
  yellowMask.value.style.animation = isForward
    ? 'slideOutToRight 0.25s cubic-bezier(0.2, 0.8, 0.2, 1) forwards'
    : 'slideOutToLeft 0.25s cubic-bezier(0.2, 0.8, 0.2, 1) forwards'

  await sleep(50)

  // 黑色遮罩紧跟收回
  blackMask.value.style.animation = isForward
    ? 'slideOutToRight 0.25s cubic-bezier(0.2, 0.8, 0.2, 1) forwards'
    : 'slideOutToLeft 0.25s cubic-bezier(0.2, 0.8, 0.2, 1) forwards'

  await sleep(250) // 等待到0.6s，动画完成

  // 重置遮罩层状态
  blackMask.value.style.animation = ''
  yellowMask.value.style.animation = ''
  blackMask.value.style.transform = isForward ? 'translateX(100vw)' : 'translateX(-100vw)'
  yellowMask.value.style.transform = isForward ? 'translateX(100vw)' : 'translateX(-100vw)'

  // 动画结束，重置状态
  isAnimating.value = false
}

// 暴露方法给父组件使用
defineExpose({
  navigateWithAnimation
})

// 提供给子组件使用
provide('navigateWithAnimation', navigateWithAnimation)

const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))
</script>

<style scoped>
.page-transition-container {
  position: relative;
  width: 100%;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

/* 底层：内容层 */
.content-layer {
  position: relative;
  width: 100%;
  height: 100%;
  z-index: 1;
  overflow-y: auto;
}

/* 中层1：黑色遮罩 */
.mask-black {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: #000000;
  z-index: 10;
  transform: translateX(100vw);
  will-change: transform;
}

/* 中层2：黄色遮罩 */
.mask-yellow {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: #FFD700;
  z-index: 20;
  transform: translateX(100vw);
  will-change: transform;
}

/* 关键帧动画 */
@keyframes slideInFromRight {
  from {
    transform: translateX(100vw);
  }
  to {
    transform: translateX(0);
  }
}

@keyframes slideOutToRight {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(calc(100vw - 80px)); /* 收进右侧标题栏（80px宽） */
  }
}

@keyframes slideInFromLeft {
  from {
    transform: translateX(-100vw);
  }
  to {
    transform: translateX(0);
  }
}

@keyframes slideOutToLeft {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(-100vw);
  }
}

/* 顶层：右侧常驻标题栏 */
.side-title-bar {
  position: fixed;
  top: 0;
  right: 0;
  width: 80px;
  height: 100vh;
  background-color: #FFD700;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1);
}

.title-text {
  writing-mode: vertical-rl;
  text-orientation: mixed;
  font-size: 1.2rem;
  font-weight: 700;
  color: #1A1A1A;
  letter-spacing: 0.3em;
}

/* 顶层：左下角切换按钮 */
.nav-buttons {
  position: fixed;
  bottom: 40px;
  left: 40px;
  display: flex;
  gap: 20px;
  z-index: 100;
}

.nav-btn {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  border: 2px solid #FFD700;
  background-color: transparent;
  color: #FFD700;
  font-size: 1.5rem;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-btn:hover:not(:disabled) {
  background-color: #FFD700;
  color: #1A1A1A;
  transform: scale(1.1);
}

.nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.btn-icon {
  font-weight: 700;
}

/* 日间/夜间模式适配 */
[data-theme="dark"] .side-title-bar {
  background-color: #FFD700;
}

[data-theme="dark"] .title-text {
  color: #0A0A0A;
}

[data-theme="dark"] .nav-btn {
  border-color: #FFD700;
  color: #FFD700;
}

[data-theme="dark"] .nav-btn:hover:not(:disabled) {
  background-color: #FFD700;
  color: #0A0A0A;
}
</style>
