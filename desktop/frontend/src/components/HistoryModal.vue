<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3>
          <i class="fas fa-history"></i>
          调动历史
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="staff-info">
          <div class="avatar">
            <i class="fas fa-user-circle"></i>
          </div>
          <div>
            <div class="staff-name">{{ staff.name }}</div>
            <div class="staff-code">工号：{{ staff.staff_code }}</div>
          </div>
        </div>

        <div class="timeline">
          <div v-if="history.length === 0" class="empty-state">
            <i class="fas fa-clipboard-list"></i>
            <p>暂无调动记录</p>
          </div>
          <div 
            v-for="(item, index) in sortedHistory" 
            :key="item.id"
            class="timeline-item"
          >
            <div class="timeline-dot" :class="getActionClass(item.action_type)">
              <i :class="getActionIcon(item.action_type)"></i>
            </div>
            <div class="timeline-content">
              <div class="timeline-header">
                <span class="action-type" :class="getActionClass(item.action_type)">
                  {{ getActionText(item.action_type) }}
                </span>
                <span class="time">{{ formatDateTime(item.happened_at) }}</span>
              </div>
              <div class="timeline-body">
                <div class="team-info">
                  <i class="fas fa-users"></i>
                  <span>{{ item.team?.team_name || '未知班组' }}</span>
                </div>
                <div v-if="item.operator_name" class="operator-info">
                  <i class="fas fa-user-edit"></i>
                  <span>操作人：{{ item.operator_name }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="modal-btn confirm" @click="$emit('close')">
          <i class="fas fa-check"></i> 关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  staff: {
    type: Object,
    required: true
  },
  history: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close'])

// 按时间倒序排列
const sortedHistory = computed(() => {
  return [...props.history].sort((a, b) => {
    return new Date(b.happened_at) - new Date(a.happened_at)
  })
})

const getActionText = (actionType) => {
  const map = {
    1: '加入班组',
    2: '离开班组'
  }
  return map[actionType] || '未知操作'
}

const getActionIcon = (actionType) => {
  const map = {
    1: 'fas fa-sign-in-alt',
    2: 'fas fa-sign-out-alt'
  }
  return map[actionType] || 'fas fa-question'
}

const getActionClass = (actionType) => {
  const map = {
    1: 'join',
    2: 'leave'
  }
  return map[actionType] || ''
}

const formatDateTime = (datetime) => {
  if (!datetime) return '-'
  const date = new Date(datetime)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<style scoped>
.modal-overlay {
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
  z-index: 1000;
  animation: fadeIn 0.3s;
}

.modal-container {
  background: rgba(20, 30, 48, 0.98);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  animation: slideUp 0.3s;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, rgba(52, 152, 219, 0.1) 0%, rgba(41, 128, 185, 0.1) 100%);
}

.modal-header h3 {
  font-size: 18px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
  margin: 0;
}

.modal-close {
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

.modal-close:hover {
  background: rgba(255,255,255,0.2);
  color: #fff;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: flex-end;
  background: rgba(10, 14, 39, 0.5);
}

.staff-info {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: rgba(255,255,255,0.05);
  border-radius: 12px;
  margin-bottom: 24px;
}

.avatar {
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: #fff;
}

.staff-name {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 4px;
}

.staff-code {
  font-size: 13px;
  color: rgba(255,255,255,0.6);
}

.timeline {
  position: relative;
  padding-left: 40px;
}

.timeline::before {
  content: '';
  position: absolute;
  left: 16px;
  top: 0;
  bottom: 0;
  width: 2px;
  background: linear-gradient(180deg, rgba(102, 126, 234, 0.5) 0%, rgba(102, 126, 234, 0.1) 100%);
}

.timeline-item {
  position: relative;
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.timeline-item:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.timeline-dot {
  position: absolute;
  left: -32px;
  top: 0;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  color: #fff;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}

.timeline-dot.join {
  background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
}

.timeline-dot.leave {
  background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
}

.timeline-content {
  background: rgba(255,255,255,0.03);
  border-radius: 12px;
  padding: 16px;
  margin-left: 8px;
  border: 1px solid rgba(255,255,255,0.05);
  transition: all 0.2s;
}

.timeline-content:hover {
  background: rgba(255,255,255,0.05);
  border-color: rgba(102, 126, 234, 0.3);
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.action-type {
  font-size: 14px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 10px;
}

.action-type.join {
  background: rgba(46, 204, 113, 0.2);
  color: #2ecc71;
  border: 1px solid #2ecc71;
}

.action-type.leave {
  background: rgba(231, 76, 60, 0.2);
  color: #e74c3c;
  border: 1px solid #e74c3c;
}

.time {
  font-size: 12px;
  color: rgba(255,255,255,0.5);
}

.timeline-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.team-info,
.operator-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: rgba(255,255,255,0.7);
}

.team-info i {
  color: #667eea;
}

.operator-info i {
  color: #95a5a6;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: rgba(255,255,255,0.5);
}

.empty-state i {
  font-size: 48px;
  display: block;
  margin-bottom: 12px;
  opacity: 0.5;
}

.modal-btn {
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

.modal-btn.confirm {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.modal-btn.confirm:hover {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  transform: translateY(-2px);
}
</style>

