import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/Home.vue')
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/Login.vue')
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/Register.vue')
    },
    {
      path: '/search',
      name: 'search',
      component: () => import('../views/Search.vue')
    },
    {
      path: '/playlist',
      name: 'playlist',
      component: () => import('../views/Playlist.vue')
    },
    {
      path: '/playlist/:id',
      name: 'playlist-detail',
      component: () => import('../views/Playlist.vue')
    },
    {
      path: '/playlist/:id/:page',
      name: 'playlist-page',
      component: () => import('../views/Playlist.vue')
    },
    {
      path: '/downloads',
      name: 'downloads',
      component: () => import('../views/Downloads.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/Settings.vue')
    },
    {
      path: '/accounts',
      name: 'accounts',
      component: () => import('../views/Accounts.vue')
    },
  ]
})

// 路由守卫：检查系统认证
router.beforeEach(async (to, _from, next) => {
  // 登录页面：允许访问
  if (to.name === 'login') {
    next()
    return
  }

  // 注册页面：如果系统已有用户，直接遣返到登录页
  if (to.name === 'register') {
    try {
      const response = await fetch('/api/system/check')
      const data = await response.json()
      if (data.has_user) {
        ElMessage.warning('系统已有管理员账号，请直接登录')
        next({ name: 'login' })
        return
      }
    } catch {
      // 出错时允许访问注册页
    }
    next()
    return
  }

  // 检查是否已登录系统
  const token = localStorage.getItem('system_token')
  if (!token) {
    // 检查系统是否已有用户
    try {
      const response = await fetch('/api/system/check')
      const data = await response.json()

      if (!data.has_user) {
        // 系统没有用户，跳转到注册页面
        ElMessage.warning('请先注册系统账号')
        next({ name: 'register' })
      } else {
        // 系统有用户但未登录，跳转到登录页面
        ElMessage.warning('无有效登录状态，请前往登录')
        next({ name: 'login' })
      }
    } catch (error) {
      // 出错时跳转到登录页面
      ElMessage.warning('无有效登录状态，请前往登录')
      next({ name: 'login' })
    }
  } else {
    next()
  }
})

export default router
