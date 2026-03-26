<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-cogs"></i>
          设备状态监控
        </div>
        <div class="page-subtitle">Device Status Monitoring System</div>
      </div>
      <div class="header-actions">
        <button class="action-btn refresh" @click="loadData" title="刷新数据">
          <i class="fas fa-sync-alt"></i> 刷新数据
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card card-running">
        <div class="stat-icon">
          <i class="fas fa-play-circle"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.running_count || 0 }}</div>
          <div class="stat-label">运行中</div>
        </div>
      </div>
      <div class="stat-card card-idle">
        <div class="stat-icon">
          <i class="fas fa-pause-circle"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.idle_count || 0 }}</div>
          <div class="stat-label">停机中</div>
        </div>
      </div>
      <div class="stat-card card-total">
        <div class="stat-icon">
          <i class="fas fa-cogs"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.total_count || 0 }}</div>
          <div class="stat-label">设备总数</div>
        </div>
      </div>
      <div class="stat-card card-utilization">
        <div class="stat-icon">
          <i class="fas fa-chart-line"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.avg_utilization?.toFixed(1) || 0 }}%</div>
          <div class="stat-label">平均利用率</div>
        </div>
      </div>
    </div>

    <!-- 今日设备运行甘特图 -->
    <div class="card">
      <div class="card-header">
        <h3><i class="fas fa-chart-gantt"></i> 设备运行甘特图（今日24小时）</h3>
        <div class="legend">
          <span class="legend-item">
            <span class="legend-color running"></span> 运行
          </span>
          <span class="legend-item">
            <span class="legend-color idle"></span> 停机
          </span>
        </div>
      </div>
      <div class="gantt-container">
        <!-- 时间轴 -->
        <div class="gantt-timeline">
          <div class="timeline-header">
            <div class="device-label">设备</div>
            <div class="timeline-scale">
              <div 
                v-for="hour in 24" 
                :key="hour" 
                class="timeline-hour"
              >
                {{ String(hour - 1).padStart(2, '0') }}:00
              </div>
            </div>
          </div>
        </div>
        <!-- 甘特图数据行 -->
        <div class="gantt-rows">
          <div 
            v-for="device in currentDevices" 
            :key="device.device_id" 
            class="gantt-row"
          >
            <div class="device-name-cell">
              <i class="fas fa-cog"></i>
              {{ device.device_name }}
            </div>
            <div class="gantt-bars">
              <!-- 今日24小时的状态条 -->
              <div 
                v-for="(bar, index) in device.ganttBars" 
                :key="index"
                :class="['gantt-bar', getGanttBarClass(bar.status)]"
                :style="{ 
                  left: bar.left + '%', 
                  width: bar.width + '%' 
                }"
                :title="getBarTooltip(bar)"
              >
              </div>
            </div>
          </div>
        </div>
        <div v-if="currentDevices.length === 0" class="gantt-empty">
          <i class="fas fa-inbox"></i>
          <p>暂无在线设备</p>
        </div>
      </div>
    </div>

    <!-- 历史记录查询 -->
    <div class="card">
      <div class="card-header">
        <h3><i class="fas fa-history"></i> 设备状态历史记录</h3>
        <span class="badge">{{ filteredHistory.length }} 条记录</span>
      </div>
      
      <!-- 查询条件 -->
      <n-space :size="12" style="margin-bottom: 20px;">
        <n-select
          v-model:value="historyFilter.deviceId"
          :options="deviceOptions"
          placeholder="选择设备"
          clearable
          style="width: 200px;"
        >
          <template #prefix>
            <i class="fas fa-cog" style="margin-right: 8px; color: #78909c;"></i>
          </template>
        </n-select>
        
        <n-select
          v-model:value="historyFilter.status"
          :options="statusOptions"
          placeholder="选择状态"
          clearable
          style="width: 150px;"
        >
          <template #prefix>
            <i class="fas fa-filter" style="margin-right: 8px; color: #78909c;"></i>
          </template>
        </n-select>
        
        <n-date-picker
          v-model:value="historyFilter.startTimeValue"
          type="datetime"
          placeholder="开始时间"
          clearable
          style="width: 200px;"
        />
        
        <n-date-picker
          v-model:value="historyFilter.endTimeValue"
          type="datetime"
          placeholder="结束时间"
          clearable
          style="width: 200px;"
        />
        
        <n-button type="primary" @click="loadHistory">
          <template #icon>
            <i class="fas fa-search"></i>
          </template>
          查询
        </n-button>
      </n-space>

      <!-- 历史记录表格 -->
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th width="50">#</th>
              <th>设备名称</th>
              <th>设备编码</th>
              <th>开始时间</th>
              <th>结束时间</th>
              <th>设备状态</th>
              <th>持续时长</th>
              <th>班组</th>
              <th>操作人员</th>
              <th>备注</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredHistory.length === 0">
              <td colspan="10" class="empty-row">
                <i class="fas fa-inbox"></i>
                <p>暂无历史记录</p>
              </td>
            </tr>
            <tr v-for="(record, index) in filteredHistory" :key="record.id">
              <td>{{ index + 1 }}</td>
              <td>
                <div class="device-name">
                  <i class="fas fa-cog"></i>
                  <strong>{{ record.device_name }}</strong>
                </div>
              </td>
              <td>
                <span class="device-code">{{ record.device_code }}</span>
              </td>
              <td>{{ formatDateTime(record.start_time) }}</td>
              <td>
                <span :class="record.end_time ? '' : 'running-badge'">
                  {{ formatDateTime(record.end_time) || '进行中' }}
                </span>
              </td>
              <td>
                <span :class="['status-badge', getStatusClass(record.status)]">
                  <i :class="getStatusIcon(record.status)"></i>
                  {{ getStatusName(record.status) }}
                </span>
              </td>
              <td>
                <span class="duration-text">
                  {{ calculateDuration(record) }}
                </span>
              </td>
              <td>
                <span v-if="record.team_name" class="team-badge">
                  <i class="fas fa-users"></i>
                  {{ record.team_name }}
                </span>
                <span v-else class="text-muted">-</span>
              </td>
              <td>
                <span class="operator-badge">
                  <i class="fas fa-user"></i>
                  {{ record.operators || '-' }}
                </span>
              </td>
              <td>
                <span class="remark-text" :title="record.remark">
                  {{ record.remark }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'

const currentDevices = ref([])
const historyRecords = ref([])
const allDevices = ref([])
const stats = ref({
  running_count: 0,
  idle_count: 0,
  fault_count: 0,
  total_count: 0,
  avg_utilization: 0
})
const loading = ref(false)
const historyFilter = ref({
  deviceId: null,
  status: null,
  startTime: '',
  endTime: '',
  startTimeValue: null,
  endTimeValue: null
})
let refreshTimer = null

// 设备选项
const deviceOptions = computed(() => {
  return allDevices.value.map(d => ({
    label: d.device_name,
    value: d.id
  }))
})

// 状态选项
const statusOptions = [
  { label: '停机', value: 0 },
  { label: '运行', value: 1 }
]

// 加载当前在线设备状态和甘特图
const loadCurrentDevices = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      // 1. 获取在线设备状态
      const result = await window.go.main.App.GetAllDevicesStatus()
      currentDevices.value = result || []
      
      // 2. 获取所有设备列表（用于甘特图，包括不在线的设备）
      const allDevs = await window.go.main.App.GetAllDevices()
      
      // 3. 为所有设备加载今日甘特图数据（不仅仅是在线的）
      const devicesWithGantt = []
      for (const device of allDevs) {
        const ganttBars = await loadTodayGantt(device.id)
        devicesWithGantt.push({
          device_id: device.id,
          device_name: device.device_name,
          device_code: device.device_code,
          ganttBars
        })
      }
      
      // 用甘特图数据覆盖（这样包含所有设备）
      currentDevices.value = devicesWithGantt
      
      // 4. 加载统计数据
      const statsResult = await window.go.main.App.GetDeviceStatusStats()
      stats.value = statsResult || {}
    }
  } catch (e) {
    console.error('加载设备状态失败:', e)
  }
}

// 加载今日甘特图数据
const loadTodayGantt = async (deviceId) => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      // 获取今日00:00到现在的历史记录
      const now = new Date()
      const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      
      const startTimeStr = formatDateTimeLocal(todayStart)
      const endTimeStr = formatDateTimeLocal(now)
      
      console.log(`加载设备 ${deviceId} 今日甘特图: ${startTimeStr} 至 ${endTimeStr}`)
      
      const history = await window.go.main.App.GetDeviceStatusHistoryAll(
        deviceId,
        startTimeStr,
        endTimeStr
      )
      
      console.log(`设备 ${deviceId} 今日历史记录数:`, history?.length || 0)
      if (history && history.length > 0) {
        console.log(`设备 ${deviceId} 记录详情:`, history.map(h => ({
          id: h.id,
          status: h.status,
          start: h.start_time,
          end: h.end_time
        })))
      }
      
      // 转换为甘特图数据
      const bars = convertToGanttBars(history || [], todayStart, now)
      console.log(`设备 ${deviceId} 甘特图条数:`, bars.length, bars)
      
      return bars
    }
  } catch (e) {
    console.error(`加载设备 ${deviceId} 甘特图数据失败:`, e)
    return []
  }
  return []
}

// 转换历史记录为甘特图条
const convertToGanttBars = (history, dayStart, dayEnd) => {
  const bars = []
  const totalMinutes = 24 * 60
  
  // 按开始时间排序
  const sortedHistory = [...history].sort((a, b) => {
    return new Date(a.start_time) - new Date(b.start_time)
  })
  
  for (const record of sortedHistory) {
    const startTime = new Date(record.start_time)
    const endTime = record.end_time ? new Date(record.end_time) : dayEnd
    
    // 确保时间在今日范围内
    let effectiveStart = startTime < dayStart ? dayStart : startTime
    let effectiveEnd = endTime > dayEnd ? dayEnd : endTime
    
    // 如果记录完全不在今日范围内，跳过
    if (effectiveEnd <= dayStart || effectiveStart >= dayEnd) {
      continue
    }
    
    // 计算相对于今日00:00的分钟偏移
    const startOffset = (effectiveStart - dayStart) / (1000 * 60)
    const endOffset = (effectiveEnd - dayStart) / (1000 * 60)
    
    // 转换为百分比
    const left = (startOffset / totalMinutes) * 100
    const width = ((endOffset - startOffset) / totalMinutes) * 100
    
    // 确保有最小宽度
    const finalWidth = Math.max(0.8, width)
    
    bars.push({
      status: record.status,
      left: Math.max(0, Math.min(100, left)),
      width: finalWidth,
      start: effectiveStart,
      end: effectiveEnd,
      remark: record.remark || ''
    })
  }
  
  return bars
}

// 加载设备列表
const loadDevices = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllDevices()
      allDevices.value = result || []
    }
  } catch (e) {
    console.error('加载设备列表失败:', e)
  }
}

// 加载历史记录（默认今天）
const loadHistory = async () => {
  try {
    loading.value = true
    
    // 转换时间戳为字符串
    let startTime = historyFilter.value.startTime
    let endTime = historyFilter.value.endTime
    
    if (historyFilter.value.startTimeValue) {
      const date = new Date(historyFilter.value.startTimeValue)
      startTime = formatDateTimeLocal(date)
    }
    
    if (historyFilter.value.endTimeValue) {
      const date = new Date(historyFilter.value.endTimeValue)
      endTime = formatDateTimeLocal(date)
    }
    
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetDeviceStatusHistoryAll(
        historyFilter.value.deviceId,
        startTime,
        endTime
      )
      
      historyRecords.value = (result || []).map(record => {
        // 解析 extra_data
        let temperature = '-'
        let humidity = '-'
        
        if (record.extra_data) {
          try {
            const extraData = JSON.parse(record.extra_data)
            if (extraData.temperature) {
              temperature = `${extraData.temperature.toFixed(1)}°C`
            }
            if (extraData.humidity) {
              humidity = `${extraData.humidity.toFixed(1)}%`
            }
          } catch (e) {
            console.error('解析extra_data失败:', e)
          }
        }
        
        // 获取设备信息（如果Device已预加载）
        const deviceName = record.device?.device_name || `设备${record.device_id}`
        const deviceCode = record.device?.device_code || '-'
        
        return {
          id: record.id,
          device_id: record.device_id,
          device_name: deviceName,
          device_code: deviceCode,
          status: record.status,
          start_time: record.start_time,
          end_time: record.end_time,
          duration_min: record.duration_min,
          temperature,
          humidity,
          team_name: record.team_name || '',
          operators: record.operators || '',
          remark: record.remark || '-'
        }
      })
    }
  } catch (e) {
    console.error('加载历史记录失败:', e)
  } finally {
    loading.value = false
  }
}

// 过滤后的历史记录
const filteredHistory = computed(() => {
  let result = historyRecords.value
  
  if (historyFilter.value.status !== null) {
    result = result.filter(r => r.status === historyFilter.value.status)
  }
  
  return result
})

// 加载所有数据
const loadData = async () => {
  await loadDevices()
  await loadCurrentDevices()
  
  // 默认加载今天的历史记录
  if (!historyFilter.value.startTime) {
    const now = new Date()
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    historyFilter.value.startTime = formatDateTimeLocal(todayStart)
    historyFilter.value.endTime = formatDateTimeLocal(now)
    historyFilter.value.startTimeValue = todayStart.getTime()
    historyFilter.value.endTimeValue = now.getTime()
  }
  
  await loadHistory()
}

// 格式化日期时间为 datetime-local 格式
const formatDateTimeLocal = (date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}`
}

// 格式化日期时间显示
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 获取甘特图条颜色类
const getGanttBarClass = (status) => {
  const classMap = {
    0: 'bar-idle',
    1: 'bar-running'
  }
  return classMap[status] || 'bar-idle'
}

// 获取甘特图条提示信息
const getBarTooltip = (bar) => {
  const statusName = getStatusName(bar.status)
  const startTime = bar.start.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  
  // 判断是否为进行中的状态（结束时间是当前时间）
  const now = new Date()
  const isOngoing = Math.abs(bar.end - now) < 60000 // 1分钟内认为是当前时间
  
  let endTimeStr, duration
  if (isOngoing) {
    endTimeStr = '进行中'
    duration = formatDuration(Math.floor((now - bar.start) / (1000 * 60)))
  } else {
    endTimeStr = bar.end.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    duration = formatDuration(Math.floor((bar.end - bar.start) / (1000 * 60)))
  }
  
  return `${statusName} | ${startTime} - ${endTimeStr} | ${duration}${bar.remark ? ' | ' + bar.remark : ''}`
}

// 获取温度样式类
const getTempClass = (temperature) => {
  if (!temperature || temperature === '-') return ''
  
  const temp = parseFloat(temperature)
  if (temp >= 60) return 'temp-danger'
  if (temp >= 45) return 'temp-warning'
  return 'temp-normal'
}

// 获取状态样式类
const getStatusClass = (status) => {
  const statusMap = {
    0: 'status-idle',
    1: 'status-running'
  }
  return statusMap[status] || 'status-idle'
}

// 获取状态图标
const getStatusIcon = (status) => {
  const iconMap = {
    0: 'fas fa-pause-circle',
    1: 'fas fa-play-circle'
  }
  return iconMap[status] || 'fas fa-pause-circle'
}

// 获取状态名称
const getStatusName = (status) => {
  const names = { 0: '停机', 1: '运行' }
  return names[status] || '停机'
}

// 计算持续时间（考虑进行中的状态）
const calculateDuration = (record) => {
  if (!record.start_time) return '0分'
  
  const startTime = new Date(record.start_time)
  let endTime
  
  if (record.end_time) {
    // 已结束，使用结束时间
    endTime = new Date(record.end_time)
  } else {
    // 进行中，使用当前时间
    endTime = new Date()
  }
  
  const durationMs = endTime - startTime
  const minutes = Math.floor(durationMs / (1000 * 60))
  
  return formatDuration(minutes)
}

// 格式化时长
const formatDuration = (minutes) => {
  if (!minutes || minutes === 0) return '0分'
  
  const hours = Math.floor(minutes / 60)
  const mins = Math.floor(minutes % 60)
  
  if (hours > 0) {
    return `${hours}h${mins}m`
  }
  return `${mins}分`
}

// 页面挂载时加载数据
onMounted(() => {
  loadData()
  
  // 设置自动刷新（每30秒刷新当前状态和甘特图）
  refreshTimer = setInterval(() => {
    loadCurrentDevices()
  }, 30000)
})

// 页面卸载时清除定时器
onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
/* 页面容器使用全局样式，padding: 40px */

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 28px;
  font-weight: bold;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title i {
  color: #667eea;
}

.page-subtitle {
  font-size: 14px;
  color: rgba(255,255,255,0.5);
  margin-top: 4px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

/* 统计卡片 - 与人员管理页面完全一致 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  padding: 24px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 20px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.3);
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0,0,0,0.4);
}

/* 工业低调配色 - 与人员管理页面统一 */
.stat-card.card-running {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

.stat-card.card-idle {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
}

.stat-card.card-total {
  background: linear-gradient(135deg, #556070 0%, #6e7e8e 100%);
}

.stat-card.card-utilization {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
}

.stat-icon {
  width: 64px;
  height: 64px;
  background: rgba(255,255,255,0.2);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 36px;
  font-weight: bold;
  color: #fff;
  line-height: 1;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

/* 卡片 */
.card {
  background: rgba(30, 40, 60, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  border: 1px solid rgba(255,255,255,0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.card-header h3 {
  font-size: 18px;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
}

.card-header h3 i {
  color: #78909c;
}

.badge {
  background: rgba(120, 144, 156, 0.2);
  color: #b0bec5;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}

.legend {
  display: flex;
  gap: 16px;
  align-items: center;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: rgba(255,255,255,0.7);
}

.legend-color {
  width: 20px;
  height: 12px;
  border-radius: 3px;
}

.legend-color.running {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

.legend-color.idle {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
}


/* 甘特图 */
.gantt-container {
  background: rgba(20, 30, 48, 0.4);
  border-radius: 8px;
  overflow: hidden;
}

.gantt-timeline {
  background: rgba(120, 144, 156, 0.1);
  border-bottom: 2px solid rgba(120, 144, 156, 0.3);
}

.timeline-header {
  display: flex;
  height: 50px;
}

.device-label {
  width: 180px;
  padding: 12px;
  font-weight: bold;
  color: rgba(255,255,255,0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(120, 144, 156, 0.15);
  border-right: 1px solid rgba(255,255,255,0.1);
}

.timeline-scale {
  flex: 1;
  display: flex;
  position: relative;
}

.timeline-hour {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  color: rgba(255,255,255,0.5);
  border-right: 1px dashed rgba(255,255,255,0.15);
  position: relative;
}

.gantt-rows {
  min-height: 200px;
  max-height: 400px;
  overflow-y: auto;
}

.gantt-row {
  display: flex;
  height: 50px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  transition: background 0.2s;
}

.gantt-row:hover {
  background: rgba(120, 144, 156, 0.08);
}

.device-name-cell {
  width: 180px;
  padding: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: rgba(255,255,255,0.8);
  border-right: 1px solid rgba(255,255,255,0.1);
  background: rgba(20, 30, 48, 0.3);
}

.device-name-cell i {
  color: #78909c;
}

.gantt-bars {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
  background-image: repeating-linear-gradient(
    to right,
    transparent,
    transparent calc(100% / 24 - 1px),
    rgba(255,255,255,0.08) calc(100% / 24 - 1px),
    rgba(255,255,255,0.08) calc(100% / 24)
  );
}

.gantt-bar {
  position: absolute;
  height: 30px;
  border-radius: 4px;
  transition: all 0.3s;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
}

.gantt-bar:hover {
  transform: scaleY(1.15);
  z-index: 10;
  box-shadow: 0 4px 12px rgba(0,0,0,0.5);
}

.gantt-bar.bar-running {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

.gantt-bar.bar-idle {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
}


.gantt-empty {
  padding: 60px 20px;
  text-align: center;
  color: rgba(255,255,255,0.3);
}

.gantt-empty i {
  font-size: 48px;
  margin-bottom: 12px;
}

.gantt-empty p {
  margin: 0;
  font-size: 14px;
}


/* 表格 */
.table-container {
  overflow-x: auto;
  max-height: 500px;
  overflow-y: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table thead {
  background: rgba(120, 144, 156, 0.1);
  position: sticky;
  top: 0;
  z-index: 10;
}

.data-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  font-size: 13px;
  border-bottom: 2px solid rgba(120, 144, 156, 0.3);
}

.data-table tbody tr {
  border-bottom: 1px solid rgba(255,255,255,0.05);
  transition: all 0.2s;
}

.data-table tbody tr:hover {
  background: rgba(120, 144, 156, 0.08);
}

.data-table td {
  padding: 14px 12px;
  color: rgba(255,255,255,0.8);
  font-size: 13px;
}

.empty-row {
  text-align: center;
  padding: 60px 20px !important;
  color: rgba(255,255,255,0.4);
}

.empty-row i {
  font-size: 48px;
  display: block;
  margin-bottom: 12px;
  opacity: 0.3;
}

.empty-row p {
  margin: 0;
  font-size: 14px;
}

/* 设备名称 */
.device-name {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255,255,255,0.9);
}

.device-name i {
  color: #78909c;
}

/* 设备编码 */
.device-code {
  background: rgba(120, 144, 156, 0.2);
  color: #b0bec5;
  padding: 4px 8px;
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  font-weight: 500;
}

/* 状态徽章 - 工业色系 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.status-running {
  background: rgba(94, 139, 126, 0.2);
  color: #5e8b7e;
}

.status-badge.status-idle {
  background: rgba(158, 126, 94, 0.2);
  color: #9e7e5e;
}


/* 温度徽章 */
.temp-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.temp-badge.temp-normal {
  background: rgba(94, 139, 126, 0.15);
  color: #7ea896;
}

.temp-badge.temp-warning {
  background: rgba(158, 126, 94, 0.15);
  color: #b8967e;
}

.temp-badge.temp-danger {
  background: rgba(142, 110, 110, 0.15);
  color: #a88e8e;
}

/* 湿度徽章 */
.humidity-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: rgba(110, 126, 142, 0.15);
  color: #8ea0b4;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

/* 时间徽章 */
.time-badge {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  display: inline-block;
}

.time-badge.green {
  background: rgba(94, 139, 126, 0.15);
  color: #7ea896;
}

.time-badge.orange {
  background: rgba(158, 126, 94, 0.15);
  color: #b8967e;
}

.time-badge.red {
  background: rgba(142, 110, 110, 0.15);
  color: #a88e8e;
}

.duration-text {
  color: rgba(255,255,255,0.7);
  font-weight: 500;
}

.running-badge {
  color: #5e8b7e;
  font-weight: 500;
}

/* 利用率进度条 */
.utilization-bar {
  position: relative;
  width: 100%;
  height: 24px;
  background: rgba(0,0,0,0.3);
  border-radius: 12px;
  overflow: hidden;
}

.utilization-fill {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  background: linear-gradient(90deg, #4a6b60 0%, #5e8b7e 100%);
  transition: width 0.5s ease;
}

.utilization-text {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  text-shadow: 0 1px 3px rgba(0,0,0,0.5);
  z-index: 1;
}

/* 操作人员徽章 */
.operator-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: rgba(110, 126, 142, 0.15);
  color: #8ea0b4;
  border-radius: 6px;
  font-size: 12px;
}

/* 班组徽章 */
.team-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: rgba(126, 126, 110, 0.15);
  color: #9e9e8e;
  border-radius: 6px;
  font-size: 12px;
}

.text-muted {
  color: rgba(255,255,255,0.4);
}

/* 备注文本 */
.remark-text {
  color: rgba(255,255,255,0.6);
  font-size: 12px;
  max-width: 200px;
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 响应式 */
@media (max-width: 1400px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
