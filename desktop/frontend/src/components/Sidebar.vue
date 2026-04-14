<template>
  <div class="sidebar show">
    <div class="sidebar-header">
      <div class="sidebar-logo" @click="handleLogoClick">
        <img src="/assets/images/logo_sp.png" alt="斯频德" class="logo-img">
      </div>
      <button class="icon-btn" @click="$emit('toggleMenu')" title="切换菜单模式">
        <i class="fas fa-bars"></i>
      </button>
    </div>
    
    <!-- 主菜单项 -->
    <div
      v-for="item in mainMenuItems"
      :key="item.key"
      :class="['menu-item', { active: currentPage === item.key }]"
      @click="$emit('changePage', item.key); $router.push({ name: item.key })"
    >
      <i :class="item.icon"></i>
      <span>{{ item.label }}</span>
    </div>
    
    <!-- 系统维护分组 -->
    <div class="menu-group" ref="menuGroupRef">
      <div class="menu-group-header" @click="toggleSystemMenu">
        <div class="menu-group-title">
          <i class="fas fa-cog"></i>
          <span>系统维护</span>
        </div>
        <i :class="['fas', showSystemMenu ? 'fa-chevron-up' : 'fa-chevron-down']"></i>
      </div>
      <transition name="slide">
        <div v-show="showSystemMenu" class="menu-group-items" @click.stop>
          <div
            v-for="item in visibleSystemMenuItems"
            :key="item.key"
            :class="['menu-item sub-item', { active: currentPage === item.key }]"
            @click="handleMenuClick(item.key)"
          >
            <i :class="item.icon"></i>
            <span>{{ item.label }}</span>
          </div>
        </div>
      </transition>
    </div>

    <!-- 密码验证对话框 -->
    <div v-if="showPasswordDialog" class="password-overlay" @click.self="closePasswordDialog">
      <div class="password-dialog">
        <div class="password-header">
          <h3>系统验证</h3>
          <button class="close-btn" @click="closePasswordDialog">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="password-body">
          <p>输入调试员密码以访问调试功能</p
          <input
            ref="passwordInput"
            v-model="password"
            type="password"
            class="password-input"
            placeholder="请输入密码"
            @keyup.enter="verifyPassword"
          />
          <div v-if="passwordError" class="password-error">
            <i class="fas fa-exclamation-circle"></i>
            {{ passwordError }}
          </div>
        </div>
        <div class="password-footer">
          <button class="btn-cancel" @click="closePasswordDialog">取消</button>
          <button class="btn-confirm" @click="verifyPassword">确认</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'

defineProps({
  currentPage: String
})
const emit = defineEmits(['changePage', 'toggleMenu'])
const router = useRouter()

const showSystemMenu = ref(false)
const showPasswordDialog = ref(false)
const password = ref('')
const passwordError = ref('')
const isAuthenticated = ref(false)
const logoClickCount = ref(0)
const logoClickTimer = ref(null)
const passwordInput = ref(null)
const menuGroupRef = ref(null)

// 管理员密码
const ADMIN_PASSWORD = '512114'

const mainMenuItems = [
  { key: 'cockpit', label: '驾驶舱', icon: 'fas fa-tachometer-alt' },
  { key: 'production', label: '生产计划', icon: 'fas fa-clipboard-list' },
  { key: 'staff', label: '人员管理', icon: 'fas fa-users' },
  { key: 'device', label: '设备状态', icon: 'fas fa-cogs' },
  { key: 'alarm', label: '报警管理', icon: 'fas fa-bell' },
  { key: 'quality', label: '不良品管理', icon: 'fas fa-exclamation-triangle' },
  { key: 'history', label: '历史查询', icon: 'fas fa-history' },
  { key: 'shiftReport', label: '班次追溯', icon: 'fas fa-clipboard-check' },
  { key: 'assistant', label: '智能助手', icon: 'fas fa-robot' },
  { key: 'dashboard', label: '数据概览', icon: 'fas fa-chart-line' },
  { key: 'oeeDebug', label: 'OEE调试', icon: 'fas fa-calculator' },
  { key: 'settings', label: '参数配置', icon: 'fas fa-sliders-h' }
]

const systemMenuItems = [
  { key: 'config', label: '采集配置', icon: 'fas fa-wrench', requireAuth: true },
  { key: 'tasks', label: '任务管理', icon: 'fas fa-tasks', requireAuth: true }
]

// 根据认证状态过滤菜单项
const visibleSystemMenuItems = computed(() => {
  return systemMenuItems.filter(item => !item.requireAuth || isAuthenticated.value)
})

const toggleSystemMenu = () => {
  showSystemMenu.value = !showSystemMenu.value
}

// 处理菜单项点击
const handleMenuClick = (key) => {
  emit('changePage', key)
  router.push({ name: key })
  showSystemMenu.value = false
}

// 处理 Logo 点击（连续点击5次触发密码输入）
const handleLogoClick = () => {
  logoClickCount.value++
  
  // 清除之前的计时器
  if (logoClickTimer.value) {
    clearTimeout(logoClickTimer.value)
  }
  
  // 2秒内点击5次触发
  if (logoClickCount.value >= 5) {
    logoClickCount.value = 0
    openPasswordDialog()
  } else {
    // 2秒后重置计数
    logoClickTimer.value = setTimeout(() => {
      logoClickCount.value = 0
    }, 2000)
  }
}

// 打开密码对话框
const openPasswordDialog = () => {
  showPasswordDialog.value = true
  password.value = ''
  passwordError.value = ''
  
  // 自动聚焦输入框
  nextTick(() => {
    if (passwordInput.value) {
      passwordInput.value.focus()
    }
  })
}

// 关闭密码对话框
const closePasswordDialog = () => {
  showPasswordDialog.value = false
  password.value = ''
  passwordError.value = ''
}

// 验证密码
const verifyPassword = () => {
  if (!password.value) {
    passwordError.value = '请输入密码'
    return
  }
  
  if (password.value === ADMIN_PASSWORD) {
    isAuthenticated.value = true
    passwordError.value = ''
    closePasswordDialog()
    
    // 显示成功提示
    if (window.$message) {
      window.$message.success('验证成功，高级功能已解锁')
    }
  } else {
    passwordError.value = '密码错误，请重试'
    password.value = ''
    
    // 重新聚焦
    nextTick(() => {
      if (passwordInput.value) {
        passwordInput.value.focus()
      }
    })
  }
}

// 点击外部关闭下拉菜单
const handleClickOutside = (event) => {
  if (showSystemMenu.value && menuGroupRef.value && !menuGroupRef.value.contains(event.target)) {
    showSystemMenu.value = false
  }
}

onMounted(() => {
  // 使用 capture 阶段避免冲突
  document.addEventListener('click', handleClickOutside, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside, true)
  if (logoClickTimer.value) {
    clearTimeout(logoClickTimer.value)
  }
})
</script>

<style scoped>
.sidebar {
  width: 240px;
  background: rgba(20, 30, 48, 0.95);
  backdrop-filter: blur(10px);
  color: #ecf0f1;
  overflow-y: auto;
  transition: all 0.3s;
  border-right: 1px solid rgba(255,255,255,0.1);
}

.sidebar-header {
  padding: 15px 20px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.sidebar-logo {
  flex: 1;
  display: flex;
  align-items: center;
}

.logo-img {
  height: 35px;
  width: auto;
  object-fit: contain;
}

.icon-btn {
  width: 32px;
  height: 32px;
  background: rgba(255,255,255,0.1);
  border: none;
  border-radius: 0;
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: rgba(84, 110, 122, 0.5);
}

.menu-item {
  padding: 15px 20px;
  cursor: pointer;
  border-left: 3px solid transparent;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 12px;
  color: rgba(255,255,255,0.7);
  font-weight: 500;
}

.menu-item i {
  font-size: 16px;
  width: 24px;
  text-align: center;
}

.menu-item:hover {
  background: rgba(84, 110, 122, 0.3);
  border-left-color: #546e7a;
  color: #fff;
}

.menu-item.active {
  background: linear-gradient(90deg, rgba(84, 110, 122, 0.3) 0%, transparent 100%);
  border-left-color: #607d8b;
  color: #fff;
  font-weight: 500;
}

/* 分组样式 */
.menu-group {
  margin-top: 10px;
  border-top: 1px solid rgba(255,255,255,0.05);
}

.menu-group-header {
  padding: 15px 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: rgba(255,255,255,0.7);
  transition: all 0.2s;
  font-weight: 500;
  background: rgba(0,0,0,0.1);
}

.menu-group-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.menu-group-title i {
  font-size: 16px;
  width: 24px;
  text-align: center;
}

.menu-group-header:hover {
  background: rgba(84, 110, 122, 0.2);
  color: #fff;
}

.menu-group-items {
  overflow: hidden;
}

.menu-item.sub-item {
  padding-left: 56px;
  font-size: 13px;
}

.menu-item.sub-item i {
  font-size: 14px;
}

/* 折叠动画 */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
  max-height: 200px;
}

.slide-enter-from,
.slide-leave-to {
  max-height: 0;
  opacity: 0;
}

/* Logo 可点击提示 */
.sidebar-logo {
  cursor: pointer;
  user-select: none;
}

/* 密码对话框 */
.password-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.password-dialog {
  background: rgba(30, 40, 60, 0.95);
  border: 1px solid rgba(255,255,255,0.2);
  border-radius: 12px;
  width: 400px;
  max-width: 90%;
  box-shadow: 0 8px 32px rgba(0,0,0,0.5);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.password-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.password-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 8px;
}

.close-btn {
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: rgba(255,255,255,0.6);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255,255,255,0.1);
  color: #fff;
}

.password-body {
  padding: 24px;
}

.password-body p {
  margin: 0 0 16px 0;
  color: rgba(255,255,255,0.7);
  font-size: 14px;
}

.password-input {
  width: 100%;
  padding: 12px 16px;
  background: rgba(20, 30, 50, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 15px;
  outline: none;
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.password-input:focus {
  border-color: #546e7a;
  background: rgba(20, 30, 50, 0.8);
  box-shadow: 0 0 0 3px rgba(84, 110, 122, 0.1);
}

.password-input::placeholder {
  color: rgba(255,255,255,0.3);
}

.password-error {
  margin-top: 12px;
  padding: 10px 12px;
  background: rgba(142, 110, 110, 0.2);
  border: 1px solid rgba(142, 110, 110, 0.3);
  border-radius: 6px;
  color: #d4a8a8;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.password-footer {
  padding: 16px 24px;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-cancel,
.btn-confirm {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-cancel {
  background: rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.7);
}

.btn-cancel:hover {
  background: rgba(255,255,255,0.15);
  color: #fff;
}

.btn-confirm {
  background: rgba(84, 110, 122, 0.3);
  color: #546e7a;
  border: 1px solid rgba(84, 110, 122, 0.3);
}

.btn-confirm:hover {
  background: rgba(84, 110, 122, 0.4);
  transform: translateY(-1px);
}
</style>



