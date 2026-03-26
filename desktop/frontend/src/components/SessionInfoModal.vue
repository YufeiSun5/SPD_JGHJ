<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3>
          <i class="fas fa-info-circle"></i>
          班次详情
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <!-- 班次基本信息 -->
        <div class="info-card">
          <div class="card-header">
            <i class="fas fa-clipboard"></i>
            <span>班次信息</span>
          </div>
          <div class="info-grid">
            <div class="info-item">
              <label>班组名称</label>
              <div class="value">{{ session.team?.team_name || '-' }}</div>
            </div>
            <div class="info-item">
              <label>班组长</label>
              <div class="value">{{ session.team?.leader_name || '-' }}</div>
            </div>
            <div class="info-item">
              <label>上班时间</label>
              <div class="value">{{ formatDateTime(session.login_time) }}</div>
            </div>
            <div class="info-item">
              <label>状态</label>
              <div class="value">
                <span class="status-badge active">
                  <i class="fas fa-clock"></i> 进行中
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 当班人员 -->
        <div class="info-card">
          <div class="card-header">
            <i class="fas fa-users"></i>
            <span>当班人员（{{ session.staff_list?.length || 0 }}人）</span>
          </div>
          <div class="staff-list">
            <div 
              v-for="staff in session.staff_list" 
              :key="staff.id"
              class="staff-tag"
            >
              <i class="fas fa-user"></i>
              <span>{{ staff.name }}</span>
              <small>{{ staff.staff_code }}</small>
            </div>
          </div>
        </div>

        <!-- 工作统计 -->
        <div class="stats-grid">
          <div class="stat-card blue">
            <div class="stat-icon">
              <i class="fas fa-clock"></i>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatMinutes(currentDuration) }}</div>
              <div class="stat-label">累计时长</div>
            </div>
          </div>
          <div class="stat-card green">
            <div class="stat-icon">
              <i class="fas fa-play"></i>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatMinutes(session.worked_min) }}</div>
              <div class="stat-label">工作时长</div>
            </div>
          </div>
          <div class="stat-card orange">
            <div class="stat-icon">
              <i class="fas fa-pause"></i>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatMinutes(session.idle_min) }}</div>
              <div class="stat-label">空闲时长</div>
            </div>
          </div>
          <div class="stat-card purple">
            <div class="stat-icon">
              <i class="fas fa-chart-line"></i>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ session.efficiency?.toFixed(1) || 0 }}%</div>
              <div class="stat-label">工时利用率</div>
            </div>
          </div>
        </div>

        <!-- 效率进度条 -->
        <div class="efficiency-bar">
          <div class="bar-header">
            <span>工时利用率</span>
            <span class="percentage">{{ session.efficiency?.toFixed(1) || 0 }}%</span>
          </div>
          <div class="progress-bar">
            <div 
              class="progress-fill" 
              :style="{ width: (session.efficiency || 0) + '%' }"
              :class="getEfficiencyClass(session.efficiency)"
            ></div>
          </div>
          <div class="bar-legend">
            <div class="legend-item">
              <div class="legend-color green"></div>
              <span>工作中</span>
            </div>
            <div class="legend-item">
              <div class="legend-color orange"></div>
              <span>空闲</span>
            </div>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="modal-btn danger" @click="$emit('logout')">
          <i class="fas fa-sign-out-alt"></i> 下班签退
        </button>
        <button class="modal-btn confirm" @click="$emit('close')">
          <i class="fas fa-check"></i> 确定
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  session: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'logout'])

// 计算当前总时长
const currentDuration = computed(() => {
  if (!props.session.login_time) return 0
  const start = new Date(props.session.login_time)
  const now = new Date()
  return Math.floor((now - start) / (1000 * 60))
})

// 格式化分钟数
const formatMinutes = (minutes) => {
  if (!minutes) return '0小时'
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return hours > 0 ? `${hours}小时${mins}分钟` : `${mins}分钟`
}

// 格式化日期时间
const formatDateTime = (datetime) => {
  if (!datetime) return '-'
  const date = new Date(datetime)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 获取效率等级样式
const getEfficiencyClass = (efficiency) => {
  if (!efficiency) return 'low'
  if (efficiency >= 80) return 'high'
  if (efficiency >= 60) return 'medium'
  return 'low'
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
  max-height: 90vh;
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.modal-btn.confirm:hover {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  transform: translateY(-2px);
}

.modal-btn.danger {
  background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
  color: #fff;
}

.modal-btn.danger:hover {
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.4);
}

.modal-body {
  padding: 24px;
  max-height: 70vh;
  overflow-y: auto;
}

/* 信息卡片 */
.info-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.info-item label {
  display: block;
  font-size: 12px;
  color: rgba(255,255,255,0.6);
  margin-bottom: 6px;
}

.info-item .value {
  font-size: 15px;
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.active {
  background: rgba(46, 204, 113, 0.2);
  color: #2ecc71;
  border: 1px solid #2ecc71;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

/* 人员列表 */
.staff-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.staff-tag {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(102, 126, 234, 0.1);
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 20px;
  font-size: 13px;
  color: rgba(255,255,255,0.9);
}

.staff-tag i {
  color: #667eea;
}

.staff-tag small {
  color: rgba(255,255,255,0.6);
  font-size: 11px;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}

.stat-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 16px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.stat-card.green { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
.stat-card.orange { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.stat-card.blue { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
.stat-card.purple { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }

.stat-icon {
  width: 48px;
  height: 48px;
  background: rgba(255,255,255,0.2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #fff;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 20px;
  font-weight: bold;
  color: #fff;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: rgba(255,255,255,0.9);
}

/* 效率进度条 */
.efficiency-bar {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  padding: 16px;
}

.bar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  color: rgba(255,255,255,0.9);
}

.percentage {
  font-weight: 600;
  color: #667eea;
}

.progress-bar {
  height: 24px;
  background: rgba(255,255,255,0.05);
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 12px;
}

.progress-fill {
  height: 100%;
  border-radius: 12px;
  transition: width 0.3s, background 0.3s;
  position: relative;
}

.progress-fill.high {
  background: linear-gradient(90deg, #2ecc71 0%, #27ae60 100%);
}

.progress-fill.medium {
  background: linear-gradient(90deg, #f39c12 0%, #e67e22 100%);
}

.progress-fill.low {
  background: linear-gradient(90deg, #e74c3c 0%, #c0392b 100%);
}

.bar-legend {
  display: flex;
  gap: 20px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: rgba(255,255,255,0.7);
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 3px;
}

.legend-color.green {
  background: #2ecc71;
}

.legend-color.orange {
  background: #f39c12;
}

/* 底部按钮 */
.modal-btn.danger {
  background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
  color: #fff;
}

.modal-btn.danger:hover {
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.4);
}
</style>

