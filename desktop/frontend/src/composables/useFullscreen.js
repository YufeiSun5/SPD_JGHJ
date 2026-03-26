import { ref } from 'vue'

// 全局共享的全屏状态
const isFullscreen = ref(false)
const showFullscreenHint = ref(false)

export function useFullscreen() {
  const enterFullscreen = () => {
    isFullscreen.value = true
    showFullscreenHint.value = true
    setTimeout(() => {
      showFullscreenHint.value = false
    }, 3000)
    
    if (window.runtime && window.runtime.WindowFullscreen) {
      window.runtime.WindowFullscreen()
    } else {
      // 浏览器全屏API
      const elem = document.documentElement
      if (elem.requestFullscreen) {
        elem.requestFullscreen()
      } else if (elem.webkitRequestFullscreen) {
        elem.webkitRequestFullscreen()
      } else if (elem.msRequestFullscreen) {
        elem.msRequestFullscreen()
      }
    }
  }

  const exitFullscreen = () => {
    isFullscreen.value = false
    showFullscreenHint.value = false
    
    if (window.runtime && window.runtime.WindowUnfullscreen) {
      window.runtime.WindowUnfullscreen()
    } else {
      // 浏览器全屏API
      if (document.fullscreenElement) {
        document.exitFullscreen()
      } else if (document.webkitFullscreenElement) {
        document.webkitExitFullscreen()
      } else if (document.msFullscreenElement) {
        document.msExitFullscreen()
      }
    }
  }

  const toggleFullscreen = () => {
    if (isFullscreen.value) {
      exitFullscreen()
    } else {
      enterFullscreen()
    }
  }

  // 监听浏览器全屏状态变化
  const handleFullscreenChange = () => {
    const browserIsFullscreen = !!(
      document.fullscreenElement || 
      document.webkitFullscreenElement || 
      document.msFullscreenElement
    )
    isFullscreen.value = browserIsFullscreen
  }

  return {
    isFullscreen,
    showFullscreenHint,
    enterFullscreen,
    exitFullscreen,
    toggleFullscreen,
    handleFullscreenChange
  }
}

























