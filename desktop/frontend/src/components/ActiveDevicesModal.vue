<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3>
          <i class="fas fa-desktop"></i>
          在线设备班次
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <div v-if="sessions.length === 0" class="empty-state">
          <i class="fas fa-clock"></i>
          <p>当前没有设备运行</p>
          <p class="hint">点击"班次登记"开始工作</p>
        </div>
        
        <div v-else class="device-sessions">
          <div 
            v-for="session in sessionsWithDetails" 
            :key="session.id"
            class="session-card"
          >
            <div class="session-header-card">
              <div class="device-info">
                <i class="fas fa-cog"></i>
                <span class="device-name">{{ session.device_name }}</span>
                <span class="device-code">{{ session.device_code }}</span>
              </div>
              <button 
                class="logout-btn-mini" 
                @click="handleLogout(session.device_id)"
                title="结束班次"
              >
                <i class="fas fa-stop-circle"></i> 结束班次
              </button>
            </div>
            
            <div class="session-details">
              <div class="detail-item">
                <i class="fas fa-users"></i>
                <span class="label">班组：</span>
                <span class="value">{{ session.team?.team_name || '-' }}</span>
              </div>
              <div class="detail-item">
                <i class="fas fa-clock"></i>
                <span class="label">开始时间：</span>
                <span class="value">{{ formatTime(session.login_time) }}</span>
              </div>
              <div class="detail-item">
                <i class="fas fa-hourglass-half"></i>
                <span class="label">运行时长：</span>
                <span class="value highlight">{{ calculateDuration(session.login_time) }}</span>
              </div>
            </div>
            
            <div class="staff-badges">
              <span 
                v-for="staff in session.staff_list" 
                :key="staff.id"
                class="staff-badge"
              >
                <i class="fas fa-user"></i>
                {{ staff.name }}
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="modal-btn cancel" @click="$emit('close')">
          <i class="fas fa-times"></i> 关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const props = defineProps({
  sessions: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'logout'])

const devices = ref([])
const allStaff = ref([])

// 带设备详情的班次列表
const sessionsWithDetails = computed(() => {
  return props.sessions.map(session => {
    const device = devices.value.find(d => d.id === session.device_id)
    
    // 解析员工列表并获取真实姓名
    let staff_list = []
    try {
      const staffIds = JSON.parse(session.staff_ids)
      staff_list = staffIds.map(id => {
        const staff = allStaff.value.find(s => s.id === id)
        return {
          id,
          name: staff?.name || `员工${id}`,
          staff_code: staff?.staff_code || '-'
        }
      })
    } catch (e) {
      console.error('解析员工列表失败:', e)
    }
    
    return {
      ...session,
      device_name: device?.device_name || `设备${session.device_id}`,
      device_code: device?.device_code || '-',
      staff_list
    }
  })
})

// 处理下班
const handleLogout = (deviceId) => {
  emit('logout', deviceId)
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 计算时长
const calculateDuration = (startTime) => {
  if (!startTime) return '-'
  const start = new Date(startTime)
  const now = new Date()
  const diff = now - start
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  return `${hours}小时${minutes}分钟`
}

// 加载设备列表
const loadDevices = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllDevices()
      devices.value = result || []
    }
  } catch (e) {
    console.error('加载设备失败:', e)
  }
}

// 加载员工列表
const loadStaff = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllStaff(null, 1)
      allStaff.value = result || []
    }
  } catch (e) {
    console.error('加载员工失败:', e)
  }
}

onMounted(() => {
  loadDevices()
  loadStaff()
})
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

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-container {
  background: rgba(20, 30, 48, 0.98);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 90%;
  max-width: 700px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  animation: slideUp 0.3s;
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
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
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
  gap: 10px;
  background: rgba(10, 14, 39, 0.5);
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

.modal-btn.cancel {
  background: rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.8);
}

.modal-btn.cancel:hover {
  background: rgba(255,255,255,0.15);
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: rgba(255,255,255,0.5);
}

.empty-state i {
  font-size: 64px;
  margin-bottom: 16px;
  opacity: 0.3;
}

.empty-state p {
  margin: 8px 0;
  font-size: 16px;
}

.empty-state .hint {
  font-size: 14px;
  color: rgba(255,255,255,0.4);
}

/* 班次卡片列表 */
.device-sessions {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.session-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  padding: 16px;
  transition: all 0.3s;
}

.session-card:hover {
  background: rgba(255,255,255,0.05);
  border-color: rgba(102, 126, 234, 0.3);
  transform: translateX(4px);
}

.session-header-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.device-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.device-info i {
  color: #667eea;
  font-size: 18px;
}

.device-name {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
}

.device-code {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-family: 'Courier New', monospace;
}

.logout-btn-mini {
  padding: 6px 14px;
  background: rgba(231, 76, 60, 0.2);
  border: 1px solid #e74c3c;
  border-radius: 6px;
  color: #e74c3c;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.logout-btn-mini:hover {
  background: #e74c3c;
  color: #fff;
}

.session-details {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.detail-item i {
  color: rgba(255,255,255,0.5);
  font-size: 14px;
}

.detail-item .label {
  color: rgba(255,255,255,0.6);
}

.detail-item .value {
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

.detail-item .value.highlight {
  color: #2ecc71;
  font-weight: 600;
}

.staff-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.staff-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: rgba(102, 126, 234, 0.15);
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 16px;
  color: #667eea;
  font-size: 12px;
  font-weight: 500;
}

.staff-badge i {
  font-size: 11px;
}
</style>

