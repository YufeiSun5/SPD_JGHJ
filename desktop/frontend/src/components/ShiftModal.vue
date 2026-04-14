<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container large">
      <div class="modal-header">
        <h3>
          <i class="fas fa-clipboard-check"></i>
          设备班次登记
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <!-- 当前在线设备提示 -->
        <div v-if="activeSessions.length > 0" class="active-devices-hint">
          <div class="hint-header">
            <i class="fas fa-info-circle"></i>
            <span>当前有 {{ activeSessions.length }} 台设备正在运行</span>
          </div>
          <div class="devices-list">
            <span 
              v-for="session in sessionsWithDeviceName" 
              :key="session.id"
              class="device-badge"
            >
              {{ session.device_name }} - {{ session.team?.team_name }}
            </span>
          </div>
        </div>

        <!-- 设备选择 -->
        <div class="form-section">
          <h4><i class="fas fa-desktop"></i> 选择设备</h4>
          <div class="device-selector">
            <div class="custom-select">
              <select v-model="form.device_id">
                <option v-for="device in devices" :key="device.id" :value="device.id">
                  {{ device.device_name }} ({{ device.device_code }})
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- 班组选择 -->
        <div class="form-section">
          <h4><i class="fas fa-users"></i> 选择班组</h4>
          <div class="team-grid">
            <div 
              v-for="team in teams" 
              :key="team.id"
              :class="['team-card', { active: form.team_id === team.id }]"
              @click="selectTeam(team)"
            >
              <div class="team-icon">
                <i class="fas fa-users"></i>
              </div>
              <div class="team-info">
                <div class="team-name">{{ team.team_name }}</div>
                <div class="team-leader">
                  <i class="fas fa-user"></i>
                  {{ team.leader_name || '未指定班长' }}
                </div>
                <div class="team-count">
                  <i class="fas fa-user-friends"></i>
                  {{ getTeamStaffCount(team.id) }} 人
                </div>
              </div>
              <div v-if="form.team_id === team.id" class="check-mark">
                <i class="fas fa-check-circle"></i>
              </div>
            </div>
          </div>
        </div>

        <!-- 人员选择 - 双栏拖拽模式 -->
        <div v-if="form.team_id" class="form-section">
          <h4>
            <i class="fas fa-user-check"></i> 
            选择操作人员 
            <span class="count-badge">{{ form.staff_ids.length }} 人</span>
          </h4>
          
          <div class="staff-drag-container">
            <!-- 左侧：所有员工 -->
            <div class="staff-pool">
              <div class="pool-header">
                <i class="fas fa-users"></i>
                <span>所有员工</span>
                <span class="pool-count">{{ availableStaff.length }}人</span>
              </div>
              <div class="pool-search">
                <i class="fas fa-search"></i>
                <input 
                  v-model="searchQuery" 
                  type="text" 
                  placeholder="搜索员工..."
                />
              </div>
              <div class="staff-list">
                <div 
                  v-for="staff in filteredAvailableStaff" 
                  :key="'available-' + staff.id"
                  class="staff-item"
                  draggable="true"
                  @dragstart="handleDragStart(staff, $event)"
                  @click="addStaff(staff.id)"
                >
                  <div class="staff-avatar">
                    <i class="fas fa-user-circle"></i>
                  </div>
                  <div class="staff-details">
                    <div class="staff-name">{{ staff.name }}</div>
                    <div class="staff-code">{{ staff.staff_code }}</div>
                    <div class="staff-team" v-if="staff.current_team">
                      <i class="fas fa-users"></i>
                      {{ staff.current_team.team_name }}
                    </div>
                  </div>
                  <div class="staff-action">
                    <i class="fas fa-plus-circle"></i>
                  </div>
                </div>
                <div v-if="filteredAvailableStaff.length === 0" class="empty-staff">
                  <i class="fas fa-search"></i>
                  <p>{{ searchQuery ? '无匹配结果' : '暂无可选员工' }}</p>
                </div>
              </div>
            </div>

            <!-- 中间：拖拽提示 -->
            <div class="drag-hint">
              <i class="fas fa-arrows-alt-h"></i>
              <p>点击或拖拽</p>
            </div>

            <!-- 右侧：已选员工 -->
            <div 
              class="staff-selected"
              @dragover.prevent
              @drop="handleDrop"
            >
              <div class="pool-header selected">
                <i class="fas fa-user-check"></i>
                <span>操作人员</span>
                <span class="pool-count">{{ form.staff_ids.length }}人</span>
              </div>
              <div class="quick-actions-top">
                <button class="quick-btn-mini" @click="selectTeamMembers" title="选择本组成员">
                  <i class="fas fa-users"></i> 本组
                </button>
                <button class="quick-btn-mini" @click="selectNone" title="清空">
                  <i class="fas fa-times"></i> 清空
                </button>
              </div>
              <div class="staff-list">
                <div 
                  v-for="staff in selectedStaff" 
                  :key="'selected-' + staff.id"
                  class="staff-item selected"
                  @click="removeStaff(staff.id)"
                >
                  <div class="staff-avatar">
                    <i class="fas fa-user-circle"></i>
                  </div>
                  <div class="staff-details">
                    <div class="staff-name">{{ staff.name }}</div>
                    <div class="staff-code">{{ staff.staff_code }}</div>
                  </div>
                  <div class="staff-action remove">
                    <i class="fas fa-times-circle"></i>
                  </div>
                </div>
                <div v-if="form.staff_ids.length === 0" class="empty-staff">
                  <i class="fas fa-user-plus"></i>
                  <p>请选择操作人员</p>
                  <p class="hint-text">从左侧点击或拖拽添加</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="modal-btn cancel" @click="$emit('close')">
          <i class="fas fa-times"></i> 取消
        </button>
        <button 
          class="modal-btn confirm" 
          @click="handleShift"
          :disabled="!canShift"
        >
          <i class="fas fa-check"></i> 
          确认登记
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'

const props = defineProps({
  teams: {
    type: Array,
    default: () => []
  },
  activeSessions: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'shift'])

const form = ref({
  device_id: 1,
  team_id: null,
  staff_ids: []
})

const allStaff = ref([])
const devices = ref([])
const searchQuery = ref('')
const draggedStaff = ref(null)

// 带设备名称的会话列表
const sessionsWithDeviceName = computed(() => {
  return props.activeSessions.map(session => {
    const device = devices.value.find(d => d.id === session.device_id)
    return {
      ...session,
      device_name: device?.device_name || `设备${session.device_id}`
    }
  })
})

// 可选员工（未被选中的）
const availableStaff = computed(() => {
  return allStaff.value.filter(s => 
    s.is_active === 1 && !form.value.staff_ids.includes(s.id)
  )
})

// 搜索过滤后的可选员工
const filteredAvailableStaff = computed(() => {
  if (!searchQuery.value) return availableStaff.value
  const query = searchQuery.value.toLowerCase()
  return availableStaff.value.filter(s => 
    s.name.toLowerCase().includes(query) || 
    s.staff_code.toLowerCase().includes(query)
  )
})

// 已选员工
const selectedStaff = computed(() => {
  return form.value.staff_ids.map(id => 
    allStaff.value.find(s => s.id === id)
  ).filter(Boolean)
})

// 是否可以换班
const canShift = computed(() => {
  return form.value.device_id && form.value.team_id && form.value.staff_ids.length > 0
})

// 获取班组人员数量
const getTeamStaffCount = (teamId) => {
  return allStaff.value.filter(s => s.current_team_id === teamId && s.is_active === 1).length
}

// 选择班组
const selectTeam = async (team) => {
  form.value.team_id = team.id
  
  // 每次切换班组时，重新加载最新的员工数据
  console.log('🔄 [换班] 切换到班组:', team.team_name, '重新加载员工数据...')
  await loadStaff()
  
  // 自动选中该班组的所有在职人员
  const teamMembers = allStaff.value.filter(s => 
    s.current_team_id === team.id && s.is_active === 1
  )
  form.value.staff_ids = teamMembers.map(s => s.id)
  
  console.log('📋 [换班] 班组成员:', teamMembers.map(s => `${s.name}(id:${s.id})`))
  console.log('📋 [换班] 选中的员工ID:', form.value.staff_ids)
}

// 选择本班成员
const selectTeamMembers = () => {
  const teamMembers = allStaff.value.filter(s => 
    s.current_team_id === form.value.team_id && s.is_active === 1
  )
  form.value.staff_ids = [...new Set([...form.value.staff_ids, ...teamMembers.map(s => s.id)])]
}

// 添加员工
const addStaff = (staffId) => {
  if (!form.value.staff_ids.includes(staffId)) {
    form.value.staff_ids.push(staffId)
  }
}

// 移除员工
const removeStaff = (staffId) => {
  const index = form.value.staff_ids.indexOf(staffId)
  if (index > -1) {
    form.value.staff_ids.splice(index, 1)
  }
}

// 清空
const selectNone = () => {
  form.value.staff_ids = []
}

// 拖拽开始
const handleDragStart = (staff, event) => {
  draggedStaff.value = staff
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/html', event.target.innerHTML)
}

// 拖拽放下
const handleDrop = (event) => {
  event.preventDefault()
  if (draggedStaff.value) {
    addStaff(draggedStaff.value.id)
    draggedStaff.value = null
  }
}

// 处理换班
const handleShift = () => {
  console.log('🔄 [换班] handleShift 被调用', form.value)
  console.log('canShift:', canShift.value)
  
  if (!form.value.device_id) {
    alert('请选择设备')
    return
  }
  
  if (!form.value.team_id) {
    alert('请选择班组')
    return
  }
  
  if (form.value.staff_ids.length === 0) {
    alert('请至少选择一名员工')
    return
  }
  
  console.log('✅ [换班] 准备 emit shift 事件:', form.value)
  
  // emit 换班事件，父组件会处理所有的员工班组更新逻辑和数据刷新
  emit('shift', { ...form.value })
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

// 加载所有员工
const loadStaff = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllStaff(null, 1) // 只加载在职员工
      allStaff.value = result || []
    }
  } catch (e) {
    console.error('加载员工失败:', e)
  }
}

// 加载设备列表
const loadDevices = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllDevices()
      devices.value = result || []
      
      // 如果有设备且没有选中设备，默认选第一个
      if (devices.value.length > 0 && !form.value.device_id) {
        form.value.device_id = devices.value[0].id
      }
    }
  } catch (e) {
    console.error('加载设备失败:', e)
  }
}

onMounted(async () => {
  await loadDevices()
  await loadStaff()
  
  // 检查选择的设备是否已有活动班次
  if (props.activeSessions && props.activeSessions.length > 0) {
    // 如果当前选择的设备已有班次，预填数据
    const currentDeviceSession = props.activeSessions.find(s => s.device_id === form.value.device_id)
    if (currentDeviceSession) {
      form.value.team_id = currentDeviceSession.team_id
      
      // 解析当前班次的人员ID
      try {
        const staffIds = JSON.parse(currentDeviceSession.staff_ids)
        form.value.staff_ids = staffIds
      } catch (e) {
        console.error('解析员工ID失败:', e)
      }
    } else {
      // 该设备没有活动班次，加载历史记录
      await loadLastSession()
    }
  } else {
    // 没有任何活动班次，加载历史记录
    await loadLastSession()
  }
})

// 加载上一次的班次记录
const loadLastSession = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      // 获取最近的历史记录
      const history = await window.go.main.App.GetSessionHistory(
        form.value.device_id,
        null,
        null,
        null
      )
      
      if (history && history.length > 0) {
        const lastSession = history[0]
        form.value.team_id = lastSession.team_id
        
        // 解析上次的人员
        try {
          const staffIds = JSON.parse(lastSession.staff_ids)
          form.value.staff_ids = staffIds
        } catch (e) {
          console.error('解析历史员工ID失败:', e)
        }
      }
    }
  } catch (e) {
    console.error('加载历史班次失败:', e)
  }
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
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background: rgba(30, 40, 60, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-container.large {
  max-width: 900px;
  max-height: 90vh;
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
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
  border: none;
  background: none;
  color: rgba(255,255,255,0.6);
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
  transition: all 0.2s;
}

.modal-close:hover {
  color: #fff;
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

.modal-btn.confirm {
  background: linear-gradient(135deg, #f39c12 0%, #e67e22 100%);
  color: #fff;
}

.modal-btn.confirm:hover {
  box-shadow: 0 4px 12px rgba(243, 156, 18, 0.4);
  transform: translateY(-2px);
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  max-height: calc(90vh - 160px);
}

/* 在班设备提示 */
.active-devices-hint {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 24px;
}

.hint-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #667eea;
  margin-bottom: 12px;
  font-weight: 500;
}

.devices-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.device-badge {
  padding: 4px 12px;
  background: rgba(102, 126, 234, 0.2);
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 16px;
  color: #667eea;
  font-size: 12px;
  font-weight: 500;
}

/* 当前班次信息 */
.current-session {
  background: linear-gradient(135deg, rgba(46, 204, 113, 0.1) 0%, rgba(39, 174, 96, 0.1) 100%);
  border: 1px solid rgba(46, 204, 113, 0.3);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
}

.session-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 600;
  color: #2ecc71;
  margin-bottom: 16px;
}

.session-info {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.info-item label {
  font-size: 12px;
  color: rgba(255,255,255,0.6);
}

.info-item .value {
  font-size: 15px;
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

.info-item .value.highlight {
  color: #2ecc71;
  font-weight: 600;
}

.logout-btn {
  width: 100%;
  padding: 10px;
  background: rgba(231, 76, 60, 0.2);
  border: 1px solid #e74c3c;
  border-radius: 8px;
  color: #e74c3c;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.logout-btn:hover {
  background: #e74c3c;
  color: #fff;
}

/* 表单区块 */
.form-section {
  margin-bottom: 24px;
}

.form-section h4 {
  font-size: 14px;
  color: rgba(255,255,255,0.9);
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.count-badge {
  padding: 2px 10px;
  background: rgba(102, 126, 234, 0.2);
  border: 1px solid #667eea;
  border-radius: 10px;
  font-size: 12px;
  color: #667eea;
  font-weight: 600;
  margin-left: 8px;
}

/* 设备选择器 */
.device-selector .custom-select {
  width: 100%;
}

.custom-select {
  position: relative;
}

.custom-select::after {
  content: '\f078';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255,255,255,0.5);
  pointer-events: none;
  font-size: 11px;
}

.custom-select select {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  width: 100%;
  padding: 14px 36px 14px 16px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.2s;
}

.custom-select select option {
  background: rgba(20, 30, 48, 0.98);
  color: #fff;
  padding: 12px 16px;
  font-size: 15px;
  line-height: 1.6;
}

.custom-select select:hover {
  border-color: rgba(102, 126, 234, 0.5);
}

.custom-select select:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}

/* 班组网格 */
.team-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.team-card {
  background: rgba(255,255,255,0.03);
  border: 2px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s;
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.team-card:hover {
  background: rgba(255,255,255,0.05);
  border-color: rgba(102, 126, 234, 0.3);
  transform: translateY(-2px);
}

.team-card.active {
  background: rgba(102, 126, 234, 0.1);
  border-color: #667eea;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.team-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #fff;
}

.team-card.active .team-icon {
  animation: pulse-scale 1s infinite;
}

@keyframes pulse-scale {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}

.team-info {
  flex: 1;
}

.team-name {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 6px;
}

.team-leader, .team-count {
  font-size: 12px;
  color: rgba(255,255,255,0.6);
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
}

.check-mark {
  position: absolute;
  top: 8px;
  right: 8px;
  color: #2ecc71;
  font-size: 20px;
  animation: check-bounce 0.3s;
}

@keyframes check-bounce {
  0% { transform: scale(0); }
  50% { transform: scale(1.2); }
  100% { transform: scale(1); }
}

/* 人员拖拽选择 - 双栏布局 */
.staff-drag-container {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 16px;
  height: 400px;
}

.staff-pool,
.staff-selected {
  display: flex;
  flex-direction: column;
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  overflow: hidden;
}

.pool-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: rgba(102, 126, 234, 0.1);
  border-bottom: 1px solid rgba(255,255,255,0.1);
  font-size: 14px;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
}

.pool-header.selected {
  background: rgba(46, 204, 113, 0.1);
}

.pool-count {
  margin-left: auto;
  padding: 2px 8px;
  background: rgba(255,255,255,0.1);
  border-radius: 10px;
  font-size: 12px;
}

.pool-search {
  position: relative;
  padding: 12px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.pool-search i {
  position: absolute;
  left: 24px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255,255,255,0.5);
  font-size: 12px;
}

.pool-search input {
  width: 100%;
  padding: 8px 12px 8px 32px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 6px;
  color: #fff;
  font-size: 13px;
}

.pool-search input:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}

.quick-actions-top {
  display: flex;
  gap: 6px;
  padding: 8px 12px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.quick-btn-mini {
  flex: 1;
  padding: 6px 10px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 6px;
  color: rgba(255,255,255,0.8);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.quick-btn-mini:hover {
  background: rgba(255,255,255,0.08);
  border-color: rgba(102, 126, 234, 0.3);
}

.staff-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.staff-list::-webkit-scrollbar {
  width: 6px;
}

.staff-list::-webkit-scrollbar-track {
  background: rgba(255,255,255,0.03);
}

.staff-list::-webkit-scrollbar-thumb {
  background: rgba(102, 126, 234, 0.3);
  border-radius: 3px;
}

.staff-list::-webkit-scrollbar-thumb:hover {
  background: rgba(102, 126, 234, 0.5);
}

.staff-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.staff-item:hover {
  background: rgba(255,255,255,0.05);
  border-color: rgba(102, 126, 234, 0.3);
  transform: translateX(2px);
}

.staff-item.selected {
  background: rgba(46, 204, 113, 0.1);
  border-color: rgba(46, 204, 113, 0.3);
}

.staff-item.selected:hover {
  background: rgba(231, 76, 60, 0.1);
  border-color: rgba(231, 76, 60, 0.3);
  transform: translateX(-2px);
}

.staff-avatar {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #fff;
  flex-shrink: 0;
}

.staff-item.selected .staff-avatar {
  background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
}

.staff-details {
  flex: 1;
  min-width: 0;
}

.staff-name {
  font-size: 13px;
  font-weight: 500;
  color: #fff;
  margin-bottom: 3px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.staff-code {
  font-size: 11px;
  color: rgba(255,255,255,0.5);
}

.staff-team {
  font-size: 11px;
  color: rgba(102, 126, 234, 0.8);
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.staff-action {
  font-size: 18px;
  color: rgba(46, 204, 113, 0.6);
  transition: all 0.2s;
}

.staff-item:hover .staff-action {
  color: #2ecc71;
  transform: scale(1.2);
}

.staff-action.remove {
  color: rgba(231, 76, 60, 0.6);
}

.staff-item.selected:hover .staff-action.remove {
  color: #e74c3c;
  transform: scale(1.2);
}

.drag-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: rgba(255,255,255,0.3);
  font-size: 24px;
}

.drag-hint p {
  font-size: 12px;
  margin: 0;
}

.empty-staff {
  text-align: center;
  padding: 40px;
  color: rgba(255,255,255,0.5);
}

.empty-staff i {
  font-size: 40px;
  display: block;
  margin-bottom: 12px;
  opacity: 0.5;
}


/* 底部按钮 */
.modal-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-btn:disabled:hover {
  transform: none;
  box-shadow: none;
}
</style>

