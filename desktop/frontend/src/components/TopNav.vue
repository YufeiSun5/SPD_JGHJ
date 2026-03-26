<template>
  <div class="top-nav">
    <div class="logo" @click="handleLogoClick">
      <img src="/assets/images/logo_sp.png" alt="斯频德" class="logo-img">
    </div>
    <div class="top-menu">
      <div
        v-for="item in mainMenuItems"
        :key="item.key"
        :class="['top-menu-item', { active: currentPage === item.key }]"
        @click="$emit('changePage', item.key); $router.push({ name: item.key })"
      >
        <i :class="item.icon"></i>
        <span>{{ item.label }}</span>
      </div>
      
      <!-- 系统维护下拉菜单 -->
      <div class="top-menu-dropdown" ref="dropdownRef">
        <div 
          class="top-menu-item dropdown-trigger"
          @click="toggleSystemMenu"
        >
          <i class="fas fa-cog"></i>
          <span>系统维护</span>
          <i :class="['fas', showSystemMenu ? 'fa-chevron-up' : 'fa-chevron-down']" style="font-size: 12px;"></i>
        </div>
        <div v-if="showSystemMenu" class="dropdown-menu" @click.stop>
          <div
            v-for="item in visibleSystemMenuItems"
            :key="item.key"
            :class="['dropdown-item', { active: currentPage === item.key }]"
            @click="handleMenuClick(item.key)"
          >
            <i :class="item.icon"></i>
            <span>{{ item.label }}</span>
          </div>
        </div>
      </div>
    </div>
    <div class="top-controls">
      <button class="icon-btn" @click="$emit('toggleMenu')" title="切换菜单模式">
        <i class="fas fa-bars"></i>
      </button>
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
          <p>请输入调试员密码以访问调试功能</p>
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
const dropdownRef = ref(null)

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
  { key: 'assistant', label: '智能助手', icon: 'fas fa-robot' }
]

const systemMenuItems = [
  { key: 'dashboard', label: '数据概览', icon: 'fas fa-chart-line', requireAuth: true },
  { key: 'config', label: '采集配置', icon: 'fas fa-wrench', requireAuth: true },
  { key: 'tasks', label: '任务管理', icon: 'fas fa-tasks', requireAuth: true },
  { key: 'oeeDebug', label: 'OEE调试', icon: 'fas fa-calculator', requireAuth: true },
  { key: 'settings', label: '参数配置', icon: 'fas fa-sliders-h', requireAuth: false }
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
  
  if (logoClickTimer.value) {
    clearTimeout(logoClickTimer.value)
  }
  
  if (logoClickCount.value >= 5) {
    logoClickCount.value = 0
    openPasswordDialog()
  } else {
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
    
    if (window.$message) {
      window.$message.success('验证成功，高级功能已解锁')
    }
  } else {
    passwordError.value = '密码错误，请重试'
    password.value = ''
    
    nextTick(() => {
      if (passwordInput.value) {
        passwordInput.value.focus()
      }
    })
  }
}

// 点击外部关闭下拉菜单
const handleClickOutside = (event) => {
  if (showSystemMenu.value && dropdownRef.value && !dropdownRef.value.contains(event.target)) {
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
.top-nav {
  height: 60px;
  background: rgba(20, 30, 48, 0.95);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  padding: 0 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  position: relative;
  z-index: 100;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.logo {
  display: flex;
  align-items: center;
  min-width: 200px;
  padding: 5px 0;
  transition: all 0.3s;
  cursor: pointer;
  user-select: none;
}

.logo-img {
  height: 40px;
  width: auto;
  object-fit: contain;
  transition: all 0.3s;
}

.top-menu {
  display: flex;
  gap: 5px;
  margin-left: 40px;
  flex: 1;
  position: relative;
  overflow: visible;
  transition: all 0.3s;
}


.top-menu-item {
  padding: 10px 20px;
  color: rgba(255,255,255,0.7);
  cursor: pointer;
  border-radius: 0;
  transition: all 0.2s;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;
}

.top-menu-item i {
  font-size: 16px;
  flex-shrink: 0;
}

.top-menu-item span {
  transition: all 0.3s;
}

.top-menu-item:hover {
  background: rgba(84, 110, 122, 0.3);
  color: #fff;
}

.top-menu-item.active {
  background: linear-gradient(135deg, #546e7a 0%, #607d8b 100%);
  color: #fff;
  font-weight: 500;
}

/* 响应式设计 - 中等屏幕：隐藏文字，只显示图标 */
@media (max-width: 1400px) {
  .logo {
    min-width: 150px;
  }
  
  .logo-img {
    height: 35px;
  }
  
  .top-menu {
    margin-left: 20px;
    gap: 3px;
  }
  
  .top-menu-item {
    padding: 10px 15px;
  }
}

@media (max-width: 1200px) {
  .top-menu {
    margin-left: 15px;
    gap: 2px;
  }
  
  .top-menu-item {
    padding: 10px 12px;
  }
  
  .top-menu-item span {
    max-width: 60px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

/* 小屏幕：只显示图标 */
@media (max-width: 1000px) {
  .logo {
    min-width: 120px;
  }
  
  .logo-img {
    height: 30px;
  }
  
  .top-menu {
    margin-left: 10px;
    gap: 0;
  }
  
  .top-menu-item {
    padding: 10px;
    min-width: 40px;
    justify-content: center;
  }
  
  .top-menu-item span {
    display: none;
  }
  
  .top-menu-item i {
    font-size: 18px;
  }
}

/* 下拉菜单 */
.top-menu-dropdown {
  position: relative;
}

.dropdown-trigger {
  position: relative;
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  min-width: 160px;
  background: rgba(30, 40, 58, 0.98);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  z-index: 9999;
  margin-top: 5px;
}

.dropdown-item {
  padding: 12px 20px;
  color: rgba(255,255,255,0.7);
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 500;
  border-left: 3px solid transparent;
}

.dropdown-item i {
  font-size: 14px;
  width: 18px;
}

.dropdown-item:hover {
  background: rgba(84, 110, 122, 0.3);
  color: #fff;
  border-left-color: #546e7a;
}

.dropdown-item.active {
  background: rgba(84, 110, 122, 0.5);
  color: #fff;
  border-left-color: #607d8b;
  font-weight: 500;
}

.top-controls {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.icon-btn {
  width: 36px;
  height: 36px;
  background: rgba(255,255,255,0.1);
  border: none;
  border-radius: 0;
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  transition: all 0.2s;
  flex-shrink: 0;
}

.icon-btn:hover {
  background: rgba(84, 110, 122, 0.5);
}

/* 响应式 - 下拉菜单在小屏幕的处理 */
@media (max-width: 1000px) {
  .dropdown-menu {
    right: 0;
    left: auto;
  }
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
  z-index: 99999;
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

