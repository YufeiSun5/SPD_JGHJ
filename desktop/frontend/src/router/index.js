import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  // 🔧 关键: 使用Hash模式，避免Wails路径问题
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'cockpit',
      component: () => import('../views/Cockpit.vue')
    },
    {
      path: '/production',
      name: 'production',
      component: () => import('../views/Production.vue')
    },
    {
      path: '/quality',
      name: 'quality',
      component: () => import('../views/Quality.vue')
    },
    {
      path: '/alarm',
      name: 'alarm',
      component: () => import('../views/Alarm.vue')
    },
    {
      path: '/staff',
      name: 'staff',
      component: () => import('../views/Staff.vue')
    },
    {
      path: '/device',
      name: 'device',
      component: () => import('../views/DeviceStatus.vue')
    },
    {
      path: '/assistant',
      name: 'assistant',
      component: () => import('../views/Assistant.vue')
    },
    {
      path: '/history',
      name: 'history',
      component: () => import('../views/History.vue')
    },
    {
      path: '/config',
      name: 'config',
      component: () => import('../views/DeviceConfig.vue')
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/Dashboard.vue')
    },
    {
      path: '/tasks',
      name: 'tasks',
      component: () => import('../views/TaskManagement.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SystemSettings.vue')
    },
    {
      path: '/oee-debug',
      name: 'oeeDebug',
      component: () => import('../views/OEEDebug.vue')
    }
  ]
})

export default router

