<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider>
      <MessageSetup>
        <div id="app">
        <!-- 全屏提示 -->
        <div v-if="showFullscreenHint" class="fullscreen-hint">
          <i class="fas fa-info-circle"></i> 按 ESC 键退出全屏
        </div>

        <!-- 自定义标题栏 -->
        <TitleBar 
          v-if="!isFullscreen"
          @minimize="minimizeWindow"
          @maximize="toggleMaximize"
          @fullscreen="toggleFullscreen"
          @close="closeWindow"
        />

        <!-- 顶部导航栏 -->
        <TopNav
          v-if="menuMode === 'top' && !isFullscreen"
          :current-page="currentPage"
          @change-page="currentPage = $event"
          @toggle-menu="toggleMenuMode"
        />

        <!-- 主容器 -->
        <div class="main-container" :class="containerClass">
          <!-- 左侧菜单 -->
          <Sidebar
            v-if="menuMode === 'sidebar' && !isFullscreen"
            :current-page="currentPage"
            @change-page="currentPage = $event"
            @toggle-menu="toggleMenuMode"
          />

          <!-- 内容区域 -->
          <div class="content">
            <router-view />
          </div>
        </div>
      </div>
    </MessageSetup>
  </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { zhCN, dateZhCN } from 'naive-ui'
import TitleBar from './components/TitleBar.vue'
import TopNav from './components/TopNav.vue'
import Sidebar from './components/Sidebar.vue'
import MessageSetup from './components/MessageSetup.vue'
import { useFullscreen } from './composables/useFullscreen'

const router = useRouter()
const currentPage = ref('dashboard')
const menuMode = ref('top')

// 使用全局全屏状态管理
const { 
  isFullscreen, 
  showFullscreenHint, 
  toggleFullscreen, 
  exitFullscreen,
  handleFullscreenChange 
} = useFullscreen()

const containerClass = computed(() => ({
  'sidebar-mode': menuMode.value === 'sidebar',
  'fullscreen-mode': isFullscreen.value
}))

const toggleMenuMode = () => {
  menuMode.value = menuMode.value === 'top' ? 'sidebar' : 'top'
}

const minimizeWindow = () => {
  if (window.runtime && window.runtime.WindowMinimise) {
    window.runtime.WindowMinimise()
  }
}

const toggleMaximize = () => {
  if (window.runtime && window.runtime.WindowToggleMaximise) {
    window.runtime.WindowToggleMaximise()
  }
}

const closeWindow = () => {
  if (window.runtime && window.runtime.Quit) {
    window.runtime.Quit()
  }
}

onMounted(() => {
  // 监听ESC键退出全屏
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && isFullscreen.value) {
      exitFullscreen()
    }
  })
  
  // 监听浏览器全屏状态变化
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.addEventListener('msfullscreenchange', handleFullscreenChange)
  
  // 同步路由到currentPage
  router.afterEach((to) => {
    currentPage.value = to.name || 'dashboard'
  })
})
</script>

<style scoped>
/* App容器基础样式在global.css中 */
.fullscreen-hint {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0,0,0,0.8);
  padding: 10px 20px;
  border-radius: 20px;
  font-size: 13px;
  color: #fff;
  z-index: 9999;
  animation: fadeOut 3s forwards;
}

@keyframes fadeOut {
  0%, 70% { opacity: 1; }
  100% { opacity: 0; pointer-events: none; }
}

.main-container {
  display: flex;
  height: calc(100vh - 92px);
}

.main-container.sidebar-mode {
  height: calc(100vh - 32px);
}

.main-container.fullscreen-mode {
  height: 100vh;
}

.content {
  flex: 1;
  overflow-y: auto;
  background: #0a0e27;
}
</style>

