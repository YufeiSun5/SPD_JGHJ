<!-- 成功或者失败弹窗 -->
<template>
  <Teleport to="body">
    <Transition name="toast-fade">
      <div v-if="show" class="toast-overlay">
        <div :class="['toast-container', type, { 'toast-hiding': isHiding }]">
          <i :class="getIcon()"></i>
          <span>{{ message }}</span>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  message: {
    type: String,
    default: ''
  },
  type: {
    type: String,
    default: 'success', // success, error, warning, info
    validator: (value) => ['success', 'error', 'warning', 'info'].includes(value)
  },
  duration: {
    type: Number,
    default: 1000  // 改为1秒显示时间
  },
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:show', 'close'])

const isHiding = ref(false)
let timer = null
let hideTimer = null

// 监听 show 变化，自动隐藏
watch(() => props.show, (newVal) => {
  if (newVal) {
    isHiding.value = false
    
    // 清除之前的定时器
    if (timer) {
      clearTimeout(timer)
    }
    if (hideTimer) {
      clearTimeout(hideTimer)
    }
    
    // 设置隐藏动画定时器（显示时间后开始淡出）
    timer = setTimeout(() => {
      isHiding.value = true
      
      // 淡出动画完成后真正隐藏
      hideTimer = setTimeout(() => {
        emit('update:show', false)
        emit('close')
        isHiding.value = false
      }, 500) // 500ms 淡出动画时间
    }, props.duration)
  }
})

// 获取图标
const getIcon = () => {
  const icons = {
    success: 'fas fa-check-circle',
    error: 'fas fa-times-circle',
    warning: 'fas fa-exclamation-triangle',
    info: 'fas fa-info-circle'
  }
  return icons[props.type] || icons.info
}
</script>

<style scoped>
/* 自定义提示框 */
.toast-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  pointer-events: none;
}

.toast-container {
  padding: 16px 24px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  font-weight: 500;
  color: #fff;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
  backdrop-filter: blur(10px);
  animation: toastSlideIn 0.3s ease-out;
  pointer-events: auto;
  min-width: 300px;
  max-width: 500px;
  opacity: 1;
  transform: translateY(0) scale(1);
  transition: opacity 0.5s ease-out, transform 0.5s ease-out;
}

/* 隐藏时的淡出动画 */
.toast-container.toast-hiding {
  opacity: 0;
  transform: translateY(-20px) scale(0.95);
}

@keyframes toastSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* Vue Transition 动画 */
.toast-fade-enter-active {
  animation: toastSlideIn 0.3s ease-out;
}

.toast-fade-leave-active {
  transition: opacity 0.5s ease-out;
}

.toast-fade-leave-to {
  opacity: 0;
}

.toast-container i {
  font-size: 20px;
  flex-shrink: 0;
}

.toast-container.success {
  background: linear-gradient(135deg, rgba(46, 204, 113, 0.95) 0%, rgba(39, 174, 96, 0.95) 100%);
  border: 1px solid rgba(46, 204, 113, 0.5);
}

.toast-container.error {
  background: linear-gradient(135deg, rgba(231, 76, 60, 0.95) 0%, rgba(192, 57, 43, 0.95) 100%);
  border: 1px solid rgba(231, 76, 60, 0.5);
}

.toast-container.warning {
  background: linear-gradient(135deg, rgba(243, 156, 18, 0.95) 0%, rgba(230, 126, 34, 0.95) 100%);
  border: 1px solid rgba(243, 156, 18, 0.5);
}

.toast-container.info {
  background: linear-gradient(135deg, rgba(52, 152, 219, 0.95) 0%, rgba(41, 128, 185, 0.95) 100%);
  border: 1px solid rgba(52, 152, 219, 0.5);
}
</style>

