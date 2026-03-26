<!-- 弹出信息确认窗 -->
<template>
  <Teleport to="body">
    <div v-if="show" class="confirm-overlay" @click.self="handleCancel">
      <div class="confirm-container">
        <div class="confirm-header">
          <div class="header-icon" :class="type">
            <i :class="getIcon()"></i>
          </div>
          <h3>{{ title }}</h3>
          <button class="confirm-close" @click="handleCancel">
            <i class="fas fa-times"></i>
          </button>
        </div>
        
        <div class="confirm-body">
          <div class="confirm-message" v-html="message"></div>
          
          <div v-if="details && details.length > 0" class="confirm-details">
            <div v-for="(detail, index) in details" :key="index" class="detail-item">
              <span class="detail-icon">{{ detail.icon }}</span>
              <span class="detail-label">{{ detail.label }}:</span>
              <span class="detail-value">{{ detail.value }}</span>
            </div>
          </div>
          
          <div v-if="warnings && warnings.length > 0" class="confirm-warnings">
            <div class="warnings-title">
              <i class="fas fa-exclamation-triangle"></i>
              {{ warningTitle || '注意' }}：
            </div>
            <ul class="warnings-list">
              <li v-for="(warning, index) in warnings" :key="index">{{ warning }}</li>
            </ul>
          </div>
        </div>
        
        <div class="confirm-footer">
          <button v-if="cancelText" class="confirm-btn cancel" @click="handleCancel">
            <i class="fas fa-times"></i> {{ cancelText }}
          </button>
          <button class="confirm-btn confirm" :class="type" @click="handleConfirm">
            <i :class="confirmIcon"></i> {{ confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  type: {
    type: String,
    default: 'info', // info, success, warning, danger
    validator: (value) => ['info', 'success', 'warning', 'danger'].includes(value)
  },
  title: {
    type: String,
    default: '确认操作'
  },
  message: {
    type: String,
    required: true
  },
  details: {
    type: Array,
    default: () => []
    // 格式: [{ icon: '📋', label: '工单号', value: 'xxx' }]
  },
  warnings: {
    type: Array,
    default: () => []
    // 格式: ['警告1', '警告2']
  },
  warningTitle: {
    type: String,
    default: '注意'
  },
  confirmText: {
    type: String,
    default: '确认'
  },
  cancelText: {
    type: String,
    default: '取消'
  },
  confirmIcon: {
    type: String,
    default: 'fas fa-check'
  }
})

const emit = defineEmits(['update:show', 'confirm', 'cancel'])

const getIcon = () => {
  const icons = {
    info: 'fas fa-info-circle',
    success: 'fas fa-check-circle',
    warning: 'fas fa-exclamation-triangle',
    danger: 'fas fa-exclamation-circle'
  }
  return icons[props.type] || icons.info
}

const handleConfirm = () => {
  emit('confirm')
  emit('update:show', false)
}

const handleCancel = () => {
  emit('cancel')
  emit('update:show', false)
}
</script>

<style scoped>
.confirm-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.7);
  backdrop-filter: blur(5px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  animation: fadeIn 0.2s;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.confirm-container {
  background: rgba(20, 30, 48, 0.98);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  animation: slideUp 0.3s;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.confirm-header {
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.header-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}

.header-icon.info {
  background: rgba(52, 152, 219, 0.2);
  color: #3498db;
}

.header-icon.success {
  background: rgba(46, 204, 113, 0.2);
  color: #2ecc71;
}

.header-icon.warning {
  background: rgba(243, 156, 18, 0.2);
  color: #f39c12;
}

.header-icon.danger {
  background: rgba(231, 76, 60, 0.2);
  color: #e74c3c;
}

.confirm-header h3 {
  flex: 1;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  margin: 0;
}

.confirm-close {
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255,255,255,0.1);
  border-radius: 8px;
  color: rgba(255,255,255,0.7);
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.confirm-close:hover {
  background: rgba(255,255,255,0.2);
  color: #fff;
}

.confirm-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.confirm-message {
  font-size: 15px;
  line-height: 1.6;
  color: rgba(255,255,255,0.9);
  margin-bottom: 20px;
  white-space: pre-line;
}

.confirm-details {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-icon {
  font-size: 18px;
  width: 24px;
  text-align: center;
}

.detail-label {
  font-size: 13px;
  color: rgba(255,255,255,0.6);
  min-width: 80px;
}

.detail-value {
  font-size: 14px;
  color: #fff;
  font-weight: 500;
  flex: 1;
}

.confirm-warnings {
  background: rgba(243, 156, 18, 0.1);
  border: 1px solid rgba(243, 156, 18, 0.3);
  border-radius: 12px;
  padding: 16px;
}

.warnings-title {
  font-size: 14px;
  font-weight: 600;
  color: #f39c12;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.warnings-list {
  margin: 0;
  padding-left: 24px;
  color: rgba(255,255,255,0.8);
}

.warnings-list li {
  font-size: 13px;
  line-height: 1.6;
  margin-bottom: 6px;
}

.warnings-list li:last-child {
  margin-bottom: 0;
}

.confirm-footer {
  padding: 16px 24px;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: center;
  gap: 10px;
  background: rgba(10, 14, 39, 0.5);
}

.confirm-footer:has(.cancel) {
  justify-content: flex-end;
}

.confirm-btn {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.confirm-btn.cancel {
  background: rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.8);
}

.confirm-btn.cancel:hover {
  background: rgba(255,255,255,0.15);
}

.confirm-btn.confirm {
  color: #fff;
}

.confirm-btn.confirm.info {
  background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
}

.confirm-btn.confirm.info:hover {
  box-shadow: 0 4px 12px rgba(52, 152, 219, 0.4);
}

.confirm-btn.confirm.success {
  background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
}

.confirm-btn.confirm.success:hover {
  box-shadow: 0 4px 12px rgba(46, 204, 113, 0.4);
}

.confirm-btn.confirm.warning {
  background: linear-gradient(135deg, #f39c12 0%, #e67e22 100%);
}

.confirm-btn.confirm.warning:hover {
  box-shadow: 0 4px 12px rgba(243, 156, 18, 0.4);
}

.confirm-btn.confirm.danger {
  background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
}

.confirm-btn.confirm.danger:hover {
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.4);
}
</style>

