<template>
  <div class="cockpit-container">
    <!-- 最上面：标题栏 -->
    <div class="cockpit-header">
      <h1 class="title">大连斯频德 · 激光焊接智能驾驶舱</h1>
      <div class="work-shift">
        <span class="shift-label">工位: </span>
        <span class="shift-name">① 一号机 ② 二号机</span>
      </div>
      <button class="fullscreen-btn" @click="toggleFullscreen" :title="isInFullscreen ? '退出全屏 (ESC)' : '进入全屏'">
        <i :class="['fas', isInFullscreen ? 'fa-compress' : 'fa-expand']"></i>
      </button>
    </div>

    <!-- 顶部KPI卡片（6个） -->
    <div class="kpi-cards">
      <!-- OEE卡片 - 左右分栏显示两台设备 -->
      <div class="kpi-card kpi-card-dual">
        <div class="kpi-label">设备OEE</div>
        <div class="kpi-dual-content">
          <div class="kpi-dual-item">
            <div class="kpi-dual-label">一号机</div>
            <div class="kpi-dual-value">{{ device1OEE }}%</div>
          </div>
          <div class="kpi-dual-divider"></div>
          <div class="kpi-dual-item">
            <div class="kpi-dual-label">二号机</div>
            <div class="kpi-dual-value">{{ device2OEE }}%</div>
          </div>
        </div>
        <div class="kpi-trend">
          <span class="trend-arrow">{{ deviceTrend.arrow }}</span>
          <span class="trend-text">平均 {{ avgOEE }}%</span>
        </div>
        <div class="kpi-sparkline" ref="sparkline1"></div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">今日产量</div>
        <div class="kpi-value">{{ todayProduction }} <span class="unit">件</span></div>
        <div class="kpi-trend">
          <span class="trend-arrow">↑</span>
          <span class="trend-text">实时统计</span>
        </div>
        <div class="kpi-device-quality">
          <span class="kpi-device-quality-item">一号机 {{ device1TodayProduction }} 件</span>
          <span class="kpi-device-quality-item">二号机 {{ device2TodayProduction }} 件</span>
        </div>
        <div class="kpi-sparkline" ref="sparkline2"></div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">本月良品率</div>
        <div class="kpi-value">{{ qualityRate }}%</div>
        <div class="kpi-trend">
          <span class="trend-arrow">{{ qualityTrend.arrow }}</span>
          <span class="trend-text">{{ qualityTrend.text }}</span>
        </div>
        <div class="kpi-device-quality">
          <span v-for="item in activeDeviceQuality" :key="item.device_name" class="kpi-device-quality-item">
            {{ item.device_name }} {{ item.quality_rate !== null ? item.quality_rate.toFixed(1) + '%' : '暂无' }}
          </span>
        </div>
        <div class="kpi-sparkline" ref="sparkline3"></div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">当前在制工单</div>
        <div class="kpi-value">{{ activeOrders }}</div>
        <div class="kpi-subtitle">装配班盘管组</div>
        <div class="kpi-sparkline" ref="sparkline4"></div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">报警数</div>
        <div class="kpi-value alarm">{{ alarmCount }}</div>
        <div class="kpi-trend stable">{{ alarmCount > 0 ? '注意' : '稳定' }}</div>
        <div class="kpi-sparkline" ref="sparkline5"></div>
      </div>
      <!-- 性能稼动率卡片 - 左右分栏显示两台设备 -->
      <div class="kpi-card kpi-card-dual">
        <div class="kpi-label">性能稼动率</div>
        <div class="kpi-dual-content">
          <div class="kpi-dual-item">
            <div class="kpi-dual-label">一号机</div>
            <div class="kpi-dual-value">{{ device1Performance }}%</div>
          </div>
          <div class="kpi-dual-divider"></div>
          <div class="kpi-dual-item">
            <div class="kpi-dual-label">二号机</div>
            <div class="kpi-dual-value">{{ device2Performance }}%</div>
          </div>
        </div>
        <div class="kpi-trend">
          <span class="trend-arrow">{{ performanceTrend.arrow }}</span>
          <span class="trend-text">平均 {{ avgPerformanceEfficiency }}%</span>
        </div>
        <div class="kpi-sparkline" ref="sparkline6"></div>
      </div>
    </div>

    <!-- 主内容区域（三栏） -->
    <div class="main-content">
      <!-- 左栏 -->
      <div class="left-column">
        <!-- 计划总览 -->
        <div class="panel">
          <div class="panel-header">
            <span class="dot"></span>
            <span class="panel-title">计划总览</span>
          </div>
          <div class="plan-table-container">
            <table class="plan-table">
              <thead>
                <tr>
                  <th>计划编号</th>
                  <th>产品</th>
                  <th>工位</th>
                  <th>计划数</th>
                  <th>实绩</th>
                  <th>合格率</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in displayOrders" :key="item.id">
                  <td>{{ item.order_no }}</td>
                  <td>{{ item.product_code }}</td>
                  <td>{{ item.workstation }}</td>
                  <td>{{ item.plan_qty }}</td>
                  <td>{{ item.actual_qty }}</td>
                  <td :class="getQualityClass(item.quality_rate)">{{ item.quality_rate }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 今日产量（小时精度） -->
        <div class="panel chart-panel">
          <div class="panel-header">
            <span class="dot"></span>
            <span class="panel-title">今日良品</span>
          </div>
          <div class="chart-container" ref="productionChart"></div>
        </div>

        <!-- 质量分布 -->
        <div class="panel">
          <div class="panel-header">
            <span class="dot"></span>
            <span class="panel-title">当日质量分布</span>
          </div>
          <div class="quality-charts">
            <div class="quality-item">
              <div class="quality-chart" ref="qualityChart1"></div>
              <div class="quality-legend">
                <div class="legend-item">
                  <span class="legend-dot qualified"></span>
                  <span>合格</span>
                </div>
                <div class="legend-item">
                  <span class="legend-dot defect"></span>
                  <span>不合格</span>
                </div>
              </div>
            </div>
            <div class="quality-item">
              <div class="quality-chart" ref="qualityChart2"></div>
              <div class="quality-legend">
                <div class="legend-item">
                  <span class="legend-dot qualified"></span>
                  <span>合格</span>
                </div>
                <div class="legend-item">
                  <span class="legend-dot defect"></span>
                  <span>不合格</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 中栏：实时监控 -->
      <div class="center-column">
        <div class="panel monitor-panel">
          <div class="panel-header">
            <span class="dot"></span>
            <span class="panel-title">实时监控</span>
          </div>
          
          <!-- 监控视频 - 动态根据设备数量显示 -->
          <div 
            v-for="(device, index) in monitorDevices" 
            :key="device.device_id"
            class="monitor-item"
          >
            <div class="video-wrapper">
              <video 
                v-if="device.videoSrc"
                :ref="el => videoRefs[index] = el"
                :src="device.videoSrc" 
                autoplay 
                loop 
                muted 
                playsinline
                preload="auto"
                class="video-player"
                @loadedmetadata="onVideoLoaded(index)"
              ></video>
              <div v-else class="video-placeholder">
                <i class="fas fa-video"></i>
                <p>监控视频接入中...</p>
                <p class="video-hint">请将视频文件放置到: assets/videos/video{{ index + 1 }}.mp4</p>
              </div>
            </div>
            <div class="video-info">
              <div class="video-tags">
                <span class="status-tag" :class="getDeviceStatusClass(device)">
                  {{ getDeviceStatusText(device) }}
                </span>
                <span class="param-tag">{{ device.device_name }}</span>
                <span class="param-tag">{{ device.device_code }}</span>
                <span class="param-tag">{{ device.device_name }} · {{ device.device_code }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右栏 -->
      <div class="right-column">
        <!-- 人员稼动率 -->
        <div class="panel chart-panel">
          <div class="panel-header">
            <span class="dot"></span>
            <span class="panel-title">当日人员稼动率</span>
          </div>
          <div class="chart-container" ref="staffChart"></div>
        </div>

        <!-- 稼动率趋势 -->
        <div class="panel chart-panel trend-panel">
          <div class="panel-header">
            <span class="dot"></span>
            <span class="panel-title">OEE趋势</span>
          </div>
          <div class="chart-container" ref="trendChart"></div>
        </div>
      </div>
    </div>
  </div>

</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { useFullscreen } from '../composables/useFullscreen'

// 数据
const orders = ref([])
const devices = ref([])
const deviceStats = ref({})
const staffList = ref([])
const realtimeTags = ref([])
const hourlyProduction = ref([]) // 今日小时产量数据
const monthlyProduction = ref([]) // 月度产量汇总数据（按设备）
const monthlyQuality = ref([]) // 月度各设备良品率（来自工单表）
const dailyQuality = ref([]) // 今日各设备良品率（来自生产运行记录）
const activeOrderQuality = ref([]) // 在产工单各设备良品率（status 1/2）
const dailyQualityTrend = ref([]) // 本月每日良品率趋势
const staffEfficiency = ref([]) // 员工绩效数据
const utilizationTrend = ref([]) // 利用率趋势数据
const hourlyAlarmCount = ref([]) // 每小时报警数据
const productionCoefficient = ref(10) // 理论节拍系数（秒/件），默认10秒
const dailyWorkMinutes = ref(460) // 每日应工作分钟数（扣除休息后），默认460分钟
const breakTimes = ref([]) // 休息时间段配置


// 视频播放器引用
const videoRefs = []

// 用于监控显示的设备列表（固定显示2个）
const monitorDevices = computed(() => {
  // 固定返回2个监控位
  const monitor1 = devices.value[0] || { 
    device_id: 0, 
    device_name: '设备1', 
    device_code: '-',
    current_status: 0 
  }
  
  const monitor2 = devices.value[1] || { 
    device_id: 0, 
    device_name: '设备2', 
    device_code: '-',
    current_status: 0 
  }
  
  // 加载视频文件（同一段视频，第一个从头播放，第二个从中间播放）
  try {
    const videoPath = new URL('../assets/videos/video1.mp4', import.meta.url).href
    
    return [
      { ...monitor1, videoSrc: videoPath },
      { ...monitor2, videoSrc: videoPath }
    ]
  } catch (error) {
    console.error('视频加载失败:', error)
    return [
      { ...monitor1, videoSrc: '' },
      { ...monitor2, videoSrc: '' }
    ]
  }
})

// 历史数据用于迷你图
const historyData = ref({
  utilization: [],
  production: [],
  quality: [],
  orders: []
  // 注意：报警数和性能稼动率不使用历史数据点，直接使用 hourlyAlarmCount 和 hourlyOEE（每小时统计）
})

// ✅ OEE 数据（每小时）
const hourlyOEE = ref([])

let refreshTimer = null
let charts = []
let trendChartTooltipIndex = 0 // ✅ 记住OEE趋势图tooltip的当前位置
let trendChartTooltipTimer = null // ✅ 记住OEE趋势图的定时器

// 图表引用
const sparkline1 = ref(null)
const sparkline2 = ref(null)
const sparkline3 = ref(null)
const sparkline4 = ref(null)
const sparkline5 = ref(null)
const sparkline6 = ref(null)
const productionChart = ref(null)
const qualityChart1 = ref(null)
const qualityChart2 = ref(null)
const staffChart = ref(null)
const trendChart = ref(null)

// 计算属性
const displayOrders = computed(() => {
  return orders.value.slice(0, 5).map(order => {
    const qualityRate = order.actual_qty > 0 
      ? ((order.ok_qty / order.actual_qty) * 100).toFixed(1) + '%'
      : '-'
    
    // 根据设备ID推断工位
    let workstation = '-'
    if (order.target_device_id) {
      const device = devices.value.find(d => d.device_id === order.target_device_id)
      if (device) {
        workstation = device.device_name || `设备${order.target_device_id}`
      } else {
        // 如果找不到设备，使用默认命名
        workstation = `${order.target_device_id}号焊机`
      }
    }
    
    return {
      ...order,
      quality_rate: qualityRate,
      workstation: workstation
    }
  })
})

const activeOrders = computed(() => orders.value.filter(o => o.status === 1).length)

const todayProduction = computed(() => {
  // ✅ 使用真实的今日产量：所有设备所有小时的总产量（良品+不良品）
  if (hourlyProduction.value.length > 0) {
    return hourlyProduction.value.reduce((sum, h) => sum + (h.ok_qty || 0) + (h.ng_qty || 0), 0)
  }
  
  // 如果没有小时产量数据，返回 0
  return 0
})

const sumTodayProductionByDeviceNames = (deviceNames = []) => {
  if (hourlyProduction.value.length === 0) return 0

  return hourlyProduction.value
    .filter(item => deviceNames.includes(item.device_name))
    .reduce((sum, item) => sum + (item.ok_qty || 0) + (item.ng_qty || 0), 0)
}

const device1TodayProduction = computed(() => {
  return sumTodayProductionByDeviceNames(['设备#1', '设备1', '一号机'])
})

const device2TodayProduction = computed(() => {
  return sumTodayProductionByDeviceNames(['设备#2', '设备2', '二号机'])
})

// 本月各设备良品率（用于顶部卡片小字显示，以设备列表为框架，无数据显示"暂无"）
const activeDeviceQuality = computed(() => {
  const deviceList = devices.value.slice(0, 2)
  const fallback = [{ device_name: '一号机' }, { device_name: '二号机' }]
  const list = deviceList.length > 0 ? deviceList : fallback
  return list.map(d => {
    const matched = monthlyQuality.value.find(q => q.device_name === d.device_name)
    const total = matched ? (matched.ok_qty || 0) + (matched.ng_qty || 0) : 0
    return {
      device_name: d.device_name,
      quality_rate: total > 0 ? (matched.quality_rate || 100.0) : null
    }
  })
})

const qualityRate = computed(() => {
  // 从月度工单数据汇总良品率（含跨月工单）
  if (monthlyQuality.value.length === 0) return '100.0'
  const totalOk = monthlyQuality.value.reduce((sum, item) => sum + (item.ok_qty || 0), 0)
  const totalNg = monthlyQuality.value.reduce((sum, item) => sum + (item.ng_qty || 0), 0)
  const total = totalOk + totalNg
  if (total === 0) return '100.0'
  return ((totalOk / total) * 100).toFixed(1)
})

const qualityTrend = computed(() => {
  const rate = parseFloat(qualityRate.value)
  if (rate >= 99) return { arrow: '↑', text: '优秀' }
  if (rate >= 95) return { arrow: '→', text: '良好' }
  return { arrow: '↓', text: '需改善' }
})

// ✅ 计算平均OEE（优先使用汇总行数据）
const avgOEE = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  // 查找汇总行（time_period包含"合计"）
  const summaryRow = hourlyOEE.value.find(item => item.time_period && item.time_period.includes('合计'))
  if (summaryRow && summaryRow.oee_pct != null) {
    return summaryRow.oee_pct.toFixed(1)
  }
  
  // 否则计算平均值（排除汇总行）
  const hourlyData = hourlyOEE.value.filter(item => !item.time_period || !item.time_period.includes('合计'))
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.oee_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

// ✅ 计算平均时间稼动率
const avgAvailability = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  const summaryRow = hourlyOEE.value.find(item => item.time_period && item.time_period.includes('合计'))
  if (summaryRow && summaryRow.availability_pct != null) {
    return summaryRow.availability_pct.toFixed(1)
  }
  
  const hourlyData = hourlyOEE.value.filter(item => !item.time_period || !item.time_period.includes('合计'))
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.availability_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

// ✅ 计算平均性能稼动率（OEE中的P）
const avgOEEPerformance = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  const summaryRow = hourlyOEE.value.find(item => item.time_period && item.time_period.includes('合计'))
  if (summaryRow && summaryRow.performance_pct != null) {
    return summaryRow.performance_pct.toFixed(1)
  }
  
  const hourlyData = hourlyOEE.value.filter(item => !item.time_period || !item.time_period.includes('合计'))
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.performance_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

// ✅ 计算平均良品率（OEE中的Q）
const avgOEEQuality = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  const summaryRow = hourlyOEE.value.find(item => item.time_period && item.time_period.includes('合计'))
  if (summaryRow && summaryRow.quality_pct != null) {
    return summaryRow.quality_pct.toFixed(1)
  }
  
  const hourlyData = hourlyOEE.value.filter(item => !item.time_period || !item.time_period.includes('合计'))
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.quality_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

// ✅ 计算设备1的OEE
const device1OEE = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  const summaryRow = hourlyOEE.value.find(item => 
    item.device_name === '设备#1' && item.time_period && item.time_period.includes('合计')
  )
  if (summaryRow && summaryRow.oee_pct != null) {
    return summaryRow.oee_pct.toFixed(1)
  }
  
  const hourlyData = hourlyOEE.value.filter(item => 
    item.device_name === '设备#1' && (!item.time_period || !item.time_period.includes('合计'))
  )
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.oee_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

// ✅ 计算设备2的OEE
const device2OEE = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  const summaryRow = hourlyOEE.value.find(item => 
    item.device_name === '设备#2' && item.time_period && item.time_period.includes('合计')
  )
  if (summaryRow && summaryRow.oee_pct != null) {
    return summaryRow.oee_pct.toFixed(1)
  }
  
  const hourlyData = hourlyOEE.value.filter(item => 
    item.device_name === '设备#2' && (!item.time_period || !item.time_period.includes('合计'))
  )
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.oee_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

const deviceTrend = computed(() => {
  const oee = parseFloat(avgOEE.value)
  // ✅ 显示两台设备各自的OEE
  const text = `①${device1OEE.value}% ②${device2OEE.value}%`
  
  if (oee >= 85) return { arrow: '↑', text }
  if (oee >= 70) return { arrow: '→', text }
  return { arrow: '↓', text }
})

// 保留旧的设备稼动率（用于其他地方可能需要）
const deviceUtilization = computed(() => {
  return deviceStats.value.avg_utilization?.toFixed(1) || 0
})

// 使用数据库中的活动报警数（未恢复的报警）
const alarmCount = ref(0)

const avgStaffEfficiency = computed(() => {
  if (staffEfficiency.value.length === 0) return 0
  const totalEfficiency = staffEfficiency.value.reduce((sum, s) => sum + s.efficiency, 0)
  return Math.round(totalEfficiency / staffEfficiency.value.length)
})

const staffTrend = computed(() => {
  const rate = parseFloat(avgStaffEfficiency.value)
  if (rate >= 85) return { arrow: '↑', text: '优秀' }
  if (rate >= 70) return { arrow: '→', text: '正常' }
  return { arrow: '↓', text: '需改善' }
})

// ✅ 计算整体性能稼动率（直接使用汇总行数据）
const avgPerformanceEfficiency = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  // 优先使用汇总行数据
  const summaryRow = hourlyOEE.value.find(item => item.time_period && item.time_period.includes('合计'))
  if (summaryRow && summaryRow.performance_pct != null) {
    return summaryRow.performance_pct.toFixed(1)
  }
  
  // 否则计算平均值
  const hourlyData = hourlyOEE.value.filter(item => !item.time_period || !item.time_period.includes('合计'))
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.performance_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

// ✅ 计算设备1的性能稼动率
const device1Performance = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  const summaryRow = hourlyOEE.value.find(item => 
    item.device_name === '设备#1' && item.time_period && item.time_period.includes('合计')
  )
  if (summaryRow && summaryRow.performance_pct != null) {
    return summaryRow.performance_pct.toFixed(1)
  }
  
  const hourlyData = hourlyOEE.value.filter(item => 
    item.device_name === '设备#1' && (!item.time_period || !item.time_period.includes('合计'))
  )
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.performance_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

// ✅ 计算设备2的性能稼动率
const device2Performance = computed(() => {
  if (hourlyOEE.value.length === 0) return '0.0'
  
  const summaryRow = hourlyOEE.value.find(item => 
    item.device_name === '设备#2' && item.time_period && item.time_period.includes('合计')
  )
  if (summaryRow && summaryRow.performance_pct != null) {
    return summaryRow.performance_pct.toFixed(1)
  }
  
  const hourlyData = hourlyOEE.value.filter(item => 
    item.device_name === '设备#2' && (!item.time_period || !item.time_period.includes('合计'))
  )
  if (hourlyData.length === 0) return '0.0'
  const sum = hourlyData.reduce((acc, item) => acc + (item.performance_pct || 0), 0)
  return (sum / hourlyData.length).toFixed(1)
})

const performanceTrend = computed(() => {
  const rate = parseFloat(avgPerformanceEfficiency.value)
  // ✅ 显示两台设备各自的性能稼动率
  const text = `①${device1Performance.value}% ②${device2Performance.value}%`
  
  if (rate >= 90) return { arrow: '↑', text }
  if (rate >= 75) return { arrow: '→', text }
  return { arrow: '↓', text }
})

// ✅ 保留旧的基于工单的性能稼动率计算（可能其他地方需要）
const avgPerformanceEfficiencyByOrder = computed(() => {
  // 筛选有效工单：有开工时间且良品数>0
  const ordersWithTime = orders.value.filter(o => o.start_time && o.ok_qty > 0)
  
  if (ordersWithTime.length === 0) return 0
  
  // 计算每个工单的性能稼动率并求平均
  const totalEfficiency = ordersWithTime.reduce((sum, order) => {
    const efficiency = calculatePerformanceEfficiency(order)
    return sum + efficiency
  }, 0)
  
  return Math.round(totalEfficiency / ordersWithTime.length)
})

// ✅ 计算单个工单的性能稼动率
// 公式：性能稼动率 = 理论时间 / 实际时间 × 100%
// 理论时间 = 良品数 × 节拍（秒/件）
// 实际时间 = used_seconds + 当前段用时（如果正在生产中）
const calculatePerformanceEfficiency = (order) => {
  if (order.ok_qty === 0) return 0
  
  // 计算实际用时（秒）
  let actualSeconds = order.used_seconds || 0
  
  // 如果正在生产中且有当前开始时间，加上当前段用时
  if (order.status === 1 && order.current_start_time) {
    const currentStart = new Date(order.current_start_time)
    const elapsed = Math.floor((Date.now() - currentStart.getTime()) / 1000)
    actualSeconds += elapsed
  }
  
  if (actualSeconds <= 0) return 0
  
  // 计算理论时间（秒）
  const theoreticalSeconds = order.ok_qty * productionCoefficient.value
  
  // 性能稼动率 = 理论时间 / 实际时间 × 100%
  const efficiency = (theoreticalSeconds / actualSeconds) * 100
  
  return Math.round(efficiency)
}

// 方法
const getQualityClass = (rate) => {
  if (!rate || rate === '-') return 'normal'
  const value = parseFloat(rate)
  if (value >= 90) return 'excellent'  // 大于等于90%显示绿色
  return 'normal'  // 低于90%显示红色
}

const getDeviceStatusClass = (device) => {
  if (!device) return 'idle'
  if (device.current_status === 1) return 'running'
  if (device.current_status === 2) return 'fault'
  return 'idle'
}

const getDeviceStatusText = (device) => {
  if (!device) return '待机'
  if (device.current_status === 1) return '运行中'
  if (device.current_status === 2) return '故障'
  return '空闲'
}

// 使用全局全屏状态管理
const { 
  isFullscreen: isInFullscreen, 
  toggleFullscreen, 
  exitFullscreen,
  handleFullscreenChange 
} = useFullscreen()

// CN: 窗口切走再切回时，浏览器内核可能暂停视频，这里统一负责恢复播放。
// EN: When the window loses and regains focus, the browser engine may pause media; resume it here.
// JP: ウィンドウの切り替え後に動画が停止することがあるため、ここで再生を復帰する。
const playMonitorVideo = (video, index) => {
  if (!video) return

  if (index === 1 && video.duration && video.currentTime < 1) {
    video.currentTime = video.duration * 0.5
  }

  const playPromise = video.play()
  if (playPromise?.catch) {
    playPromise.catch(err => {
      console.log('视频自动播放失败:', err)
    })
  }
}

const onVideoLoaded = (index) => {
  const video = videoRefs[index]
  playMonitorVideo(video, index)
}

const resumeMonitorVideos = () => {
  if (document.hidden) return

  videoRefs.forEach((video, index) => {
    if (video?.paused) {
      playMonitorVideo(video, index)
    }
  })
}

const handleWindowResume = () => {
  setTimeout(resumeMonitorVideos, 0)
}

// 加载数据
const loadOrders = async () => {
  try {
    if (window.go?.main?.App?.GetAllOrders) {
      const result = await window.go.main.App.GetAllOrders()
      orders.value = result || []
    }
  } catch (e) {
    console.error('加载工单失败:', e)
  }
}

const loadDevices = async () => {
  try {
    if (window.go?.main?.App) {
      const result = await window.go.main.App.GetAllDevicesStatus()
      devices.value = result || []
      
      const stats = await window.go.main.App.GetDeviceStatusStats()
      deviceStats.value = stats || {}
    }
  } catch (e) {
    console.error('加载设备失败:', e)
  }
}

const loadStaff = async () => {
  try {
    if (window.go?.main?.App?.GetAllStaff) {
      const result = await window.go.main.App.GetAllStaff(null, null)
      staffList.value = result || []
    }
  } catch (e) {
    console.error('加载员工失败:', e)
  }
}

const loadRealtimeData = async () => {
  try {
    if (window.go?.main?.App?.GetRealtimeData) {
      const result = await window.go.main.App.GetRealtimeData()
      realtimeTags.value = result || []
    }
  } catch (e) {
    console.error('加载实时数据失败:', e)
  }
}

const loadHourlyProduction = async () => {
  try {
    if (window.go?.main?.App?.GetHourlyProductionAccurate) {
      const result = await window.go.main.App.GetHourlyProductionAccurate(null)
      hourlyProduction.value = result || []
    } else {
      console.warn('⚠️ GetHourlyProductionAccurate 方法不存在')
      hourlyProduction.value = []
    }
  } catch (e) {
    console.error('❌ 加载今日小时产量失败:', e)
    hourlyProduction.value = []
  }
}

const loadMonthlyProduction = async () => {
  try {
    if (window.go?.main?.App?.GetMonthlyProductionAccurate) {
      const result = await window.go.main.App.GetMonthlyProductionAccurate(null)
      monthlyProduction.value = result || []
    } else {
      monthlyProduction.value = []
    }
  } catch (e) {
    console.error('❌ 加载月度产量失败:', e)
    monthlyProduction.value = []
  }
}

const loadMonthlyQuality = async () => {
  try {
    if (window.go?.main?.App?.GetMonthlyQualityByOrder) {
      const result = await window.go.main.App.GetMonthlyQualityByOrder()
      monthlyQuality.value = result || []
    } else {
      monthlyQuality.value = []
    }
  } catch (e) {
    console.error('❌ 加载月度良品率失败:', e)
    monthlyQuality.value = []
  }
}

const loadDailyQuality = async () => {
  try {
    if (window.go?.main?.App?.GetDailyQualityByRun) {
      const result = await window.go.main.App.GetDailyQualityByRun()
      dailyQuality.value = result || []
    } else {
      dailyQuality.value = []
    }
  } catch (e) {
    console.error('❌ 加载今日良品率失败:', e)
    dailyQuality.value = []
  }
}

const loadActiveOrderQuality = async () => {
  try {
    if (window.go?.main?.App?.GetActiveOrderQuality) {
      const result = await window.go.main.App.GetActiveOrderQuality()
      activeOrderQuality.value = result || []
    } else {
      activeOrderQuality.value = []
    }
  } catch (e) {
    console.error('❌ 加载在产工单良品率失败:', e)
    activeOrderQuality.value = []
  }
}

const loadDailyQualityTrend = async () => {
  try {
    if (window.go?.main?.App?.GetMonthlyDailyQualityTrend) {
      const result = await window.go.main.App.GetMonthlyDailyQualityTrend()
      dailyQualityTrend.value = result || []
    } else {
      dailyQualityTrend.value = []
    }
  } catch (e) {
    console.error('❌ 加载每日良品率趋势失败:', e)
    dailyQualityTrend.value = []
  }
}

const loadStaffEfficiency = async () => {
  try {
    if (!window.go?.main?.App) return
    
    // 1. 获取所有员工
    const allStaff = await window.go.main.App.GetAllStaff(null, 1) // 只要在职员工
    if (!allStaff || allStaff.length === 0) {
      staffEfficiency.value = []
      return
    }
    
    // 2. ✅ 改为今日的班次记录
    const now = new Date()
    const todayStart = new Date(now)
    todayStart.setHours(0, 0, 0, 0) // 今日0点
    
    let allSessions = []
    try {
      const todayStr = now.toISOString().split('T')[0] // YYYY-MM-DD 格式
      
      allSessions = await window.go.main.App.GetSessionHistory(null, null, todayStr, todayStr) || []
      console.log(`✅ 获取今日所有班次 (${todayStr})，共`, allSessions.length, '个')
      
      // 补充活动班次（可能还未结束，不在历史记录中）
      const activeSessions = await window.go.main.App.GetAllActiveSessions() || []
      if (activeSessions.length > 0) {
        console.log('📝 获取到活动班次', activeSessions.length, '个')
        // 只添加今日的活动班次，并且去重（避免与历史记录重复）
        const todayActiveSessions = activeSessions.filter(s => {
          const loginTime = new Date(s.login_time)
          // 必须是今日的班次
          if (loginTime < todayStart) return false
          
          // ✅ 检查是否已经存在于历史记录中（根据session_id去重）
          const isDuplicate = allSessions.some(existing => existing.session_id === s.session_id)
          return !isDuplicate
        })
        
        if (todayActiveSessions.length > 0) {
          console.log('📝 补充活动班次（去重后）', todayActiveSessions.length, '个')
          allSessions = [...allSessions, ...todayActiveSessions]
        } else {
          console.log('📝 活动班次已在历史记录中，无需补充')
        }
      }
    } catch (e) {
      console.error('获取今日班次失败:', e)
      allSessions = []
    }
    
    // 3. ✅ 计算今日到当前时间的应工作时长（实时计算，扣除休息）
    // ✅ 使用配置的每日应工作分钟数
    const fullDayWorkMinutes = dailyWorkMinutes.value // 从配置加载
    
    // ✅ 使用配置的休息时间段（动态读取）
    const breaks = breakTimes.value.map(bt => ({
      start: { h: bt.start_hour, m: bt.start_min },
      end: { h: bt.end_hour, m: bt.end_min }
    }))
    
    // 工作时间：7:40 - 16:20
    const workStart = { h: 7, m: 40 }
    const workEnd = { h: 16, m: 20 }
    
    const currentHour = now.getHours()
    const currentMinute = now.getMinutes()
    
    // 计算从7:40到当前时间的总分钟数
    let totalMinutes = 0
    
    // 判断是否在工作时间内
    const currentTimeInMin = currentHour * 60 + currentMinute
    const workStartInMin = workStart.h * 60 + workStart.m
    const workEndInMin = workEnd.h * 60 + workEnd.m
    
    if (currentTimeInMin < workStartInMin) {
      // 还没到上班时间
      totalMinutes = 0
    } else if (currentTimeInMin > workEndInMin) {
      // 已经过了下班时间，使用配置的全天应工作时长
      totalMinutes = fullDayWorkMinutes
    } else {
      // 在工作时间内，计算从7:40到现在的时间
      totalMinutes = currentTimeInMin - workStartInMin
      
      // 扣除已经过去的休息时间
      for (const breakTime of breaks) {
        const breakStartInMin = breakTime.start.h * 60 + breakTime.start.m
        const breakEndInMin = breakTime.end.h * 60 + breakTime.end.m
        
        if (currentTimeInMin > breakEndInMin) {
          // 当前时间已过这个休息时段，全部扣除
          totalMinutes -= (breakEndInMin - breakStartInMin)
        } else if (currentTimeInMin > breakStartInMin && currentTimeInMin <= breakEndInMin) {
          // 当前时间正在这个休息时段内，扣除已经过去的部分
          totalMinutes -= (currentTimeInMin - breakStartInMin)
        }
        // 如果当前时间还没到这个休息时段，不扣除
      }
    }
    
    // 确保不为负数
    totalMinutes = Math.max(0, totalMinutes)
    
    console.log(`📊 今日稼动率计算基准 (实时):`)
    console.log(`   - 当前时间: ${currentHour}:${String(currentMinute).padStart(2, '0')}`)
    console.log(`   - 工作时间: 7:40 - 16:20`)
    console.log(`   - 休息时间: ${breakTimes.value.map(bt => `${bt.start_hour}:${String(bt.start_min).padStart(2,'0')}-${bt.end_hour}:${String(bt.end_min).padStart(2,'0')}`).join(', ')} [从配置加载]`)
    console.log(`   - 全天应工作: ${fullDayWorkMinutes}分钟 (${(fullDayWorkMinutes/60).toFixed(2)}小时) [从配置加载]`)
    console.log(`   - 到目前为止应工作: ${totalMinutes}分钟 (${(totalMinutes/60).toFixed(1)}小时)`)
    console.log(`📊 今日班次总数: ${allSessions.length} 个`)
    
    // 4. 为每个员工计算稼动率（基于今日的班次，扣除休息时间）
    const efficiencyData = []
    
    // 辅助函数：计算时间段内的休息时间（分钟）
    const calculateBreakMinutes = (startTime, endTime) => {
      const startInMin = startTime.getHours() * 60 + startTime.getMinutes()
      const endInMin = endTime.getHours() * 60 + endTime.getMinutes()
      
      let breakMinutes = 0
      
      for (const breakTime of breaks) {
        const breakStartInMin = breakTime.start.h * 60 + breakTime.start.m
        const breakEndInMin = breakTime.end.h * 60 + breakTime.end.m
        
        // 计算班次时间段与休息时间段的重叠部分
        const overlapStart = Math.max(startInMin, breakStartInMin)
        const overlapEnd = Math.min(endInMin, breakEndInMin)
        
        if (overlapStart < overlapEnd) {
          breakMinutes += (overlapEnd - overlapStart)
        }
      }
      
      return breakMinutes
    }
    
    for (const staff of allStaff) {
      let workingMinutes = 0
      
      // 查找包含该员工的今日班次
      for (const session of allSessions) {
        try {
          const staffIds = JSON.parse(session.staff_ids || '[]')
          if (staffIds.includes(staff.id)) {
            const loginTime = new Date(session.login_time)
            const logoutTime = session.logout_time ? new Date(session.logout_time) : now
            
            // 只统计今日的时间
            if (loginTime >= todayStart) {
              // 计算总时长（分钟）
              const totalMinutes = (logoutTime - loginTime) / 60000
              
              // ✅ 扣除休息时间
              const breakMinutes = calculateBreakMinutes(loginTime, logoutTime)
              const effectiveMinutes = totalMinutes - breakMinutes
              
              if (effectiveMinutes > 0) {
                workingMinutes += effectiveMinutes
                const sessionStatus = session.logout_time ? '已结束' : '进行中'
                console.log(`👷 员工 ${staff.name} 班次(${sessionStatus}):`)
                console.log(`   登录: ${loginTime.toLocaleString()}`)
                console.log(`   登出: ${logoutTime.toLocaleString()}`)
                console.log(`   总时长: ${Math.round(totalMinutes)}分钟`)
                console.log(`   休息时间: ${Math.round(breakMinutes)}分钟`)
                console.log(`   有效工作: ${Math.round(effectiveMinutes)}分钟`)
              }
            }
          }
        } catch (e) {
          console.error('解析班次数据失败:', e)
        }
      }
      
      // 计算稼动率（整数，允许超过100%）
      let efficiency = 0
      if (totalMinutes > 0) {
        efficiency = Math.round((workingMinutes / totalMinutes) * 100)
      }
      
      console.log(`📊 员工 ${staff.name}: 今日工作 ${Math.round(workingMinutes)} 分钟 / 应工作 ${Math.round(totalMinutes)} 分钟 = ${efficiency}%`)
      
      efficiencyData.push({
        staff_id: staff.id,
        staff_name: staff.name,
        working_min: Math.round(workingMinutes),
        efficiency: efficiency // ✅ 允许超过100%（加班情况）
      })
    }
    
    // 按稼动率降序排序
    staffEfficiency.value = efficiencyData.sort((a, b) => b.efficiency - a.efficiency)
    
  } catch (e) {
    console.error('加载人员稼动率失败:', e)
    staffEfficiency.value = []
  }
}

const loadUtilizationTrend = async () => {
  try {
    if (window.go?.main?.App?.GetDeviceUtilizationTrend) {
      const result = await window.go.main.App.GetDeviceUtilizationTrend(null)
      utilizationTrend.value = result || []
    }
  } catch (e) {
    console.error('加载利用率趋势失败:', e)
  }
}

const loadAlarmCount = async () => {
  try {
    // ✅ 使用新接口：获取今日未确认的报警数
    if (window.go?.main?.App?.GetTodayUnacknowledgedAlarmCount) {
      const count = await window.go.main.App.GetTodayUnacknowledgedAlarmCount()
      alarmCount.value = count || 0
      console.log('📊 今日未确认报警数:', alarmCount.value)
    }
  } catch (e) {
    console.error('加载报警数失败:', e)
    alarmCount.value = 0
  }
}

const loadHourlyAlarmCount = async () => {
  try {
    // ✅ 加载今日每小时的报警数（用于迷你图）
    if (window.go?.main?.App?.GetTodayHourlyAlarmCount) {
      const result = await window.go.main.App.GetTodayHourlyAlarmCount()
      hourlyAlarmCount.value = result || []
      console.log('📊 今日每小时报警数:', hourlyAlarmCount.value)
    }
  } catch (e) {
    console.error('加载每小时报警数失败:', e)
    hourlyAlarmCount.value = []
  }
}

// ✅ 加载理论节拍系数（用于计算性能稼动率）
const loadProductionCoefficient = async () => {
  try {
    if (window.go?.main?.App?.GetProductionCoefficient) {
      const coefficient = await window.go.main.App.GetProductionCoefficient()
      productionCoefficient.value = coefficient
      console.log('📊 理论节拍系数:', coefficient, '秒/件')
    }
  } catch (e) {
    console.error('加载理论节拍系数失败:', e)
    // 使用默认值 10 秒/件
    productionCoefficient.value = 10
  }
}

// ✅ 加载每日应工作分钟数（用于计算人员稼动率）
const loadDailyWorkMinutes = async () => {
  try {
    if (window.go?.main?.App?.GetDailyWorkMinutes) {
      const minutes = await window.go.main.App.GetDailyWorkMinutes()
      dailyWorkMinutes.value = minutes
      console.log('📊 每日应工作分钟数:', minutes, '分钟 (', (minutes/60).toFixed(2), '小时)')
    }
  } catch (e) {
    console.error('加载每日应工作分钟数失败:', e)
    // 使用默认值 460 分钟（7小时40分钟）
    dailyWorkMinutes.value = 460
  }
}

// ✅ 加载休息时间段配置（用于计算人员稼动率和OEE）
const loadBreakTimes = async () => {
  try {
    if (window.go?.main?.App?.GetBreakTimes) {
      const times = await window.go.main.App.GetBreakTimes()
      breakTimes.value = times || []
      console.log('📊 休息时间段配置:', breakTimes.value.length, '个')
      breakTimes.value.forEach(bt => {
        console.log(`   - ${bt.name}: ${String(bt.start_hour).padStart(2,'0')}:${String(bt.start_min).padStart(2,'0')} - ${String(bt.end_hour).padStart(2,'0')}:${String(bt.end_min).padStart(2,'0')}`)
      })
    }
  } catch (e) {
    console.error('加载休息时间段配置失败:', e)
    // 使用默认值
    breakTimes.value = [
      { id: 1, name: '上午休息', start_hour: 9, start_min: 40, end_hour: 9, end_min: 50 },
      { id: 2, name: '午餐休息', start_hour: 11, start_min: 40, end_hour: 12, end_min: 20 },
      { id: 3, name: '下午休息', start_hour: 14, start_min: 20, end_hour: 14, end_min: 30 }
    ]
  }
}

// ✅ 加载每小时OEE数据
const loadHourlyOEE = async () => {
  try {
    if (window.go?.main?.App?.GetHourlyOEE) {
      console.log('🔍 开始加载每小时OEE数据...')
      
      // 构建设备配置（使用理论节拍系数）
      const configs = [
        {
          device_id: 1,
          device_name: "设备#1",
          var_ok: 1,
          var_ng_add: 72,
          var_ng_sub: 71,
          cycle_time: productionCoefficient.value || 100
        },
        {
          device_id: 2,
          device_name: "设备#2",
          var_ok: 95,
          var_ng_add: 97,
          var_ng_sub: 96,
          cycle_time: productionCoefficient.value || 100
        }
      ]
      
      const result = await window.go.main.App.GetHourlyOEE(configs)
      hourlyOEE.value = result || []
      
      console.log('✅ 加载每小时OEE:', hourlyOEE.value.length, '条记录')
      
      if (hourlyOEE.value.length > 0) {
        console.log('📊 OEE原始数据:', hourlyOEE.value)
        
        // 查找汇总行
        const summaryRow = hourlyOEE.value.find(item => item.time_period && item.time_period.includes('合计'))
        console.log('📊 汇总行数据:', summaryRow)
        
        // 过滤出每小时数据
        const hourlyData = hourlyOEE.value.filter(item => !item.time_period || !item.time_period.includes('合计'))
        console.log('📊 每小时数据:', hourlyData)
        
        console.log('📊 OEE统计:')
        console.log('  - 平均OEE:', avgOEE.value + '%')
        console.log('  - 平均时间稼动率(A):', avgAvailability.value + '%')
        console.log('  - 平均性能稼动率(P):', avgOEEPerformance.value + '%')
        console.log('  - 平均良品率(Q):', avgOEEQuality.value + '%')
      } else {
        console.log('📊 今日暂无OEE数据')
      }
    } else {
      console.warn('⚠️ GetHourlyOEE 方法不存在')
      hourlyOEE.value = []
    }
  } catch (e) {
    console.error('❌ 加载每小时OEE失败:', e)
    console.error('错误详情:', e.message)
    hourlyOEE.value = []
  }
}

const refreshAll = async () => {
  await Promise.all([
    loadOrders(),
    loadDevices(),
    loadStaff(),
    loadRealtimeData(),
    loadHourlyProduction(),
    loadMonthlyProduction(),
    loadMonthlyQuality(),
    loadDailyQuality(),
    loadActiveOrderQuality(),
    loadDailyQualityTrend(),
    loadStaffEfficiency(),
    loadUtilizationTrend(),
    loadAlarmCount(),
    loadHourlyAlarmCount(),  // ✅ 加载每小时报警数
    loadHourlyOEE()  // ✅ 加载每小时OEE
  ])
  
  // 更新历史数据用于迷你图
  updateHistoryData()
  updateCharts()
}

const updateHistoryData = () => {
  // ✅ 删除缓慢加载逻辑，直接使用完整数据
  // 稼动率：保留最近13个数据点（对应7-19点）
  const util = parseFloat(deviceUtilization.value)
  historyData.value.utilization.push(util)
  if (historyData.value.utilization.length > 13) historyData.value.utilization.shift()
  
  historyData.value.production.push(todayProduction.value)
  if (historyData.value.production.length > 13) historyData.value.production.shift()
  
  const qRate = parseFloat(qualityRate.value)
  historyData.value.quality.push(qRate)
  if (historyData.value.quality.length > 13) historyData.value.quality.shift()
  
  historyData.value.orders.push(activeOrders.value)
  if (historyData.value.orders.length > 13) historyData.value.orders.shift()
  
  // 📊 调试：打印迷你图数据
  console.log('📈 迷你图数据:', {
    utilization: historyData.value.utilization,
    production: historyData.value.production,
    quality: historyData.value.quality,
    orders: historyData.value.orders,
    hourlyAlarms: hourlyAlarmCount.value, // 每小时报警数
    hourlyOEE: hourlyOEE.value, // OEE数据（包含性能稼动率）
    avgPerformance: avgPerformanceEfficiency.value // 平均性能稼动率
  })
}

// 初始化迷你图
const initSparklines = () => {
  // ✅ 报警数迷你图：使用今日每小时的报警数统计
  // 构建完整的时间轴数据（7:00-19:00），没有数据的小时用null，让线条连接
  // ✅ 处理OEE迷你图：只显示有数据的点（删除null点）
  let oeeData = []
  const hourlyData = hourlyOEE.value.filter(item => !item.time_period || !item.time_period.includes('合计'))
  if (hourlyData.length > 0) {
    // 只取有OEE数据的点（过滤掉null和0）
    oeeData = hourlyData
      .filter(h => h.oee_pct != null && h.oee_pct > 0)
      .map(h => h.oee_pct)
  }
  
  // 如果完全没有数据，显示一个0点
  if (oeeData.length === 0) {
    oeeData = [0]
  }
  
  let alarmData = []
  if (hourlyAlarmCount.value.length > 0) {
    // 有数据：构建7-19点的完整数组
    for (let hour = 7; hour <= 19; hour++) {
      const found = hourlyAlarmCount.value.find(h => h.hour === hour)
      alarmData.push(found ? found.alarm_count : null)  // 没有数据用null
    }
  } else {
    // 完全没有数据：显示一条为0的直线
    alarmData = Array(13).fill(0)  // 7-19点共13个点
  }
  
  // ✅ 性能稼动率迷你图：只显示有数据的点（删除null点）
  let performanceData = []
  if (hourlyData.length > 0) {
    // 只取有性能稼动率数据的点（过滤掉null和0）
    performanceData = hourlyData
      .filter(h => h.performance_pct != null && h.performance_pct > 0)
      .map(h => h.performance_pct)
  }
  
  // 如果完全没有数据，显示一个0点
  if (performanceData.length === 0) {
    performanceData = [0]
  }
  
  const sparklines = [
    { ref: sparkline1, color: '#00d4ff', data: oeeData },  // ✅ 使用OEE数据
    { ref: sparkline2, color: '#00ff88', data: historyData.value.production },
    { ref: sparkline3, color: '#a855f7', data: dailyQualityTrend.value.map(d => d.quality_rate) },
    { ref: sparkline4, color: '#fbbf24', data: historyData.value.orders },
    { ref: sparkline5, color: '#ef4444', data: alarmData },  // 使用每小时报警数
    { ref: sparkline6, color: '#3b82f6', data: performanceData }  // ✅ 使用OEE中的性能稼动率
  ]

  sparklines.forEach(({ ref: element, color, data }, index) => {
    if (!element.value) return
    
    const chart = echarts.init(element.value)
    charts.push(chart)
    
    // ✅ 直接使用数据，不做缓慢加载
    let displayData = data
    
    // 如果数据为空，显示一条为0的直线
    if (!displayData || displayData.length === 0) {
      displayData = [0]
    }
    
    console.log(`📊 Sparkline ${index + 1} 数据:`, displayData)
    
    chart.setOption({
      animation: false, // 禁用动画
      grid: { top: 0, right: 0, bottom: 0, left: 0 },
      xAxis: { 
        type: 'category', 
        show: false, 
        data: displayData.map((_, i) => i),
        boundaryGap: false  // 让线条填满整个宽度
      },
      yAxis: { type: 'value', show: false },
      series: [{
        type: 'line',
        data: displayData,
        smooth: true,
        symbol: 'none',
        connectNulls: true,  // ✅ 关键：连接null值，让线条连续
        lineStyle: { color: color, width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: color + '80' },
            { offset: 1, color: color + '00' }
          ])
        }
      }]
    })
  })
}

// 初始化今日产量图表 - 使用新的数据结构（time_slot + device_name）
const initProductionChart = () => {
  if (!productionChart.value) return
  const chart = echarts.init(productionChart.value)
  charts.push(chart)

  // ✅ 固定时间轴：7:00 到 17:00（与OEE趋势图一致，对应工作时间 7:40-16:20）
  const startHour = 7
  const endHour = 17
  
  const hours = []
  for (let i = startHour; i <= endHour; i++) {
    hours.push(`${i}:00`)
  }
  
  console.log(`📊 产量图时间轴: ${startHour}点-${endHour}点, 共${hours.length}小时`)
  
  // ✅ 从新的数据结构中获取设备列表
  const deviceNames = [...new Set(hourlyProduction.value.map(h => h.device_name))]
  console.log('📊 有产量数据的设备:', deviceNames)
  
  if (deviceNames.length === 0) {
    console.warn('⚠️ 没有产量数据，显示空图表')
    chart.setOption({
      animation: false,
      grid: { top: 35, right: 25, bottom: 35, left: 45 },
      xAxis: { type: 'category', data: hours, axisLine: { lineStyle: { color: '#1e3a5f' } }, axisLabel: { color: '#6b7f9e', fontSize: 10 } },
      yAxis: { type: 'value', splitLine: { lineStyle: { color: '#1e3a5f' } }, axisLabel: { color: '#6b7f9e', fontSize: 10 } },
      series: []
    })
    return
  }
  
  // ✅ 为每个设备生成一个系列
  const seriesData = []
  const legendData = []
  const colors = [
    ['#00d4ff', '#0088ff'],  // 蓝色渐变
    ['#00ff88', '#00aa55'],  // 绿色渐变
    ['#ff9800', '#ff5722'],  // 橙色渐变
    ['#a855f7', '#7c3aed'],  // 紫色渐变
  ]
  
  const today = new Date().toISOString().split('T')[0] // YYYY-MM-DD
  
  deviceNames.forEach((deviceName, index) => {
    // 提取该设备每小时的产量
    const deviceHourlyData = hours.map((hourLabel) => {
      const hour = parseInt(hourLabel.split(':')[0])
      const timeSlot = `${today} ${String(hour).padStart(2, '0')}:00:00`
      
      const hourData = hourlyProduction.value.find(
        h => h.device_name === deviceName && h.time_slot === timeSlot
      )
      
      // 如果没有数据，返回null（断开图表，但保留X轴刻度）
      if (!hourData || hourData.total_qty === 0) {
        return null
      }
      
      return hourData.ok_qty || 0
    })
    
    legendData.push(deviceName)
    
    const colorIndex = index % colors.length
    seriesData.push({
      name: deviceName,
      type: 'bar',
      data: deviceHourlyData,
      barWidth: deviceNames.length === 1 ? '50%' : '35%',
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: colors[colorIndex][0] },
          { offset: 1, color: colors[colorIndex][1] }
        ])
      },
      // ✅ 在柱子上方显示数字
      label: {
        show: true,
        position: 'top',
        formatter: (params) => {
          // 只显示有数据的柱子的数字
          return params.value != null ? params.value : ''
        },
        color: '#fff',
        fontSize: 10,
        fontWeight: 'bold'
      },
      connectNulls: false
    })
    
    console.log(`📊 ${deviceName} 产量:`, deviceHourlyData)
  })
  
  chart.setOption({
    animation: false, // 禁用动画
    grid: { top: 35, right: 25, bottom: 35, left: 45 },
    legend: {
      data: legendData,
      top: 5,
      right: 20,
      textStyle: { color: '#8b9eb5', fontSize: 11 }
    },
    xAxis: {
      type: 'category',
      data: hours,
      axisLine: { lineStyle: { color: '#1e3a5f' } },
      axisLabel: { color: '#6b7f9e', fontSize: 10 }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: '#1e3a5f' } },
      axisLabel: { color: '#6b7f9e', fontSize: 10 }
    },
    series: seriesData
  })
}

// 初始化质量分布图表 - 使用在产工单数据
const initQualityCharts = () => {
  const qualityData = activeOrderQuality.value
  // 取前两台设备（按 device_id 去重后取前两条，防止脏数据导致同一设备重复出现）
  const seen = new Set()
  const deviceList = devices.value.filter(d => {
    if (seen.has(d.device_id)) return false
    seen.add(d.device_id)
    return true
  }).slice(0, 2)
  const chartRefs = [qualityChart1.value, qualityChart2.value]

  chartRefs.forEach((chartEl, idx) => {
    if (!chartEl) return
    const chart = echarts.init(chartEl)
    charts.push(chart)

    // 优先用在产工单数据匹配设备，找不到则用设备列表里的名称
    const deviceName = deviceList[idx]?.device_name || (idx === 0 ? '一号机' : '二号机')
    const matched = qualityData.find(q => q.device_name === deviceName)
    const total = matched ? (matched.ok_qty || 0) + (matched.ng_qty || 0) : 0

    if (matched && total > 0) {
      const qualified = parseFloat(((matched.ok_qty / total) * 100).toFixed(1))
      const defect = parseFloat(((matched.ng_qty / total) * 100).toFixed(1))
      chart.setOption({
        animation: false,
        series: [{
          type: 'pie',
          radius: ['40%', '55%'],
          center: ['50%', '42%'],
          data: [
            { value: qualified, name: '合格', itemStyle: { color: '#00ffaa' } },
            { value: defect, name: '不合格', itemStyle: { color: '#ff4466' } }
          ],
          label: { show: false },
          emphasis: { scale: false },
          labelLine: { show: false }
        }],
        graphic: [
          {
            type: 'text',
            left: 'center',
            top: '38%',
            style: { text: qualified.toFixed(1) + '%', fontSize: 18, fontWeight: 'bold', fill: '#fff', textAlign: 'center' }
          },
          {
            type: 'text',
            left: 'center',
            top: '82%',
            style: { text: deviceName, fontSize: 11, fill: '#7b9cc5', textAlign: 'center' }
          }
        ]
      })
    } else {
      // 无在产工单，显示暂无工单
      chart.setOption({
        animation: false,
        graphic: [
          {
            type: 'text',
            left: 'center',
            top: '38%',
            style: { text: '暂无工单', fontSize: 14, fill: '#4a5f8a', textAlign: 'center' }
          },
          {
            type: 'text',
            left: 'center',
            top: '62%',
            style: { text: deviceName, fontSize: 11, fill: '#7b9cc5', textAlign: 'center' }
          }
        ]
      })
    }
  })
}

// 初始化人员稼动率图表
const initStaffChart = () => {
  if (!staffChart.value) return
  const chart = echarts.init(staffChart.value)
  charts.push(chart)

  // 使用真实的人员稼动率数据
  if (staffEfficiency.value.length === 0) {
    chart.setOption({
      animation: false, // 禁用动画
      grid: { top: 30, right: 25, bottom: 30, left: 45 },
      xAxis: { type: 'category', data: [], axisLine: { lineStyle: { color: '#1e3a5f' } }, axisLabel: { color: '#6b7f9e', fontSize: 10 } },
      yAxis: { type: 'value', max: 100, splitLine: { lineStyle: { color: '#1e3a5f' } }, axisLabel: { color: '#6b7f9e', fontSize: 10, formatter: '{value}%' } },
      series: [{ type: 'bar', data: [] }]
    })
    return
  }
  
  // 使用真实稼动率数据，取前8名
  const staffData = staffEfficiency.value.slice(0, 8).map((staff) => ({
    name: staff.staff_name,
    value: staff.efficiency // 稼动率 = 工作时长/总时长 (%)
  }))

  chart.setOption({
    animation: false, // 禁用动画
    grid: { top: 30, right: 25, bottom: 30, left: 45 },
    xAxis: {
      type: 'category',
      data: staffData.map(s => s.name),
      axisLine: { lineStyle: { color: '#1e3a5f' } },
      axisLabel: { color: '#6b7f9e', fontSize: 10 }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: '#1e3a5f' } },
      axisLabel: { 
        color: '#6b7f9e', 
        fontSize: 10,
        formatter: '{value}%'
      }
    },
    series: [{
      type: 'bar',
      data: staffData.map(s => s.value),
      barWidth: '50%',
      label: {
        show: true,
        position: 'top',
        formatter: (params) => Math.round(params.value) + '%',
        color: '#fff',
        fontSize: 10
      },
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#3b82f6' },
          { offset: 1, color: '#1e40af' }
        ])
      }
    }]
  })
}

// 初始化OEE趋势图表
const initTrendChart = () => {
  if (!trendChart.value) return
  const chart = echarts.init(trendChart.value)
  charts.push(chart)

  // 获取当前时间
  const now = new Date()
  const currentHour = now.getHours()
  
  // ✅ 修改时间轴：7:00 到 17:00（对应实际工作时间 7:40-16:20）
  // 7:00-8:00 包含 7:40 开始的数据
  // 16:00-17:00 包含到 16:20 结束的数据
  const times = []
  const startHour = 7
  const endHour = 17
  
  for (let i = startHour; i <= endHour; i++) {
    times.push(`${i}:00`)
  }
  
  // 排除汇总行
  const hourlyData = hourlyOEE.value.filter(item => !item.time_period || !item.time_period.includes('合计'))
  
  console.log('📊 OEE趋势图 - 可用OEE数据:', hourlyData.length, '条')
  console.log('📊 OEE趋势图 - 设备列表:', [...new Set(hourlyData.map(d => d.device_name))])
  console.log('📊 OEE趋势图 - 原始数据:', hourlyData)
  
  // ✅ 从OEE数据中提取设备#1和设备#2的数据
  // 区分：未发生的时间（null，断开）vs 已发生但为0（显示0）
  const machine1 = times.map((timeLabel) => {
    const hour = parseInt(timeLabel.split(':')[0])
    
    // 判断该时间段是否已经开始（例如10:00-11:00的时段，只要当前时间>=10:00就算已发生）
    const hasHappened = currentHour >= hour
    
    const oeeData = hourlyData.find(
      t => t.device_name === '设备#1' && t.hour === hour
    )
    
    // ✅ 如果时间未发生，返回null（折线断开）
    if (!hasHappened) {
      return null
    }
    
    // ✅ 如果时间已发生但没有数据记录，返回null（可能设备未运行）
    if (!oeeData) {
      return null
    }
    
    // ✅ 如果有数据记录，返回OEE值（即使是0也显示，因为0是真实值）
    return oeeData.oee_pct || 0
  })
  
  const machine2 = times.map((timeLabel) => {
    const hour = parseInt(timeLabel.split(':')[0])
    
    // 判断该时间段是否已经开始
    const hasHappened = currentHour >= hour
    
    const oeeData = hourlyData.find(
      t => t.device_name === '设备#2' && t.hour === hour
    )
    
    // ✅ 如果时间未发生，返回null（折线断开）
    if (!hasHappened) {
      return null
    }
    
    // ✅ 如果时间已发生但没有数据记录，返回null
    if (!oeeData) {
      return null
    }
    
    // ✅ 如果有数据记录，返回OEE值（即使是0也显示）
    return oeeData.oee_pct || 0
  })
  
  console.log('📊 OEE趋势图 - 设备#1数据:', machine1)
  console.log('📊 OEE趋势图 - 设备#2数据:', machine2)

  // 使用固定的设备名称（与OEE数据一致）
  const legend1 = '设备#1'
  const legend2 = '设备#2'
  
  chart.setOption({
    animation: false, // 禁用动画
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(10, 20, 45, 0.95)',
      borderColor: 'rgba(0, 170, 255, 0.5)',
      borderWidth: 1,
      textStyle: {
        color: '#fff',
        fontSize: 12
      },
      formatter: function(params) {
        if (!params || params.length === 0) return ''
        
        let result = `<div style="font-weight: bold; margin-bottom: 5px;">${params[0].axisValue}</div>`
        
        params.forEach(param => {
          if (param.value != null) {
            result += `
              <div style="margin: 3px 0;">
                <span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${param.color};margin-right:5px;"></span>
                <span>${param.seriesName}: ${param.value.toFixed(1)}%</span>
              </div>
            `
          }
        })
        
        return result
      }
    },
    grid: { top: 35, right: 35, bottom: 35, left: 45 },
    legend: {
      data: [legend1, legend2],
      top: 5,
      right: 20,
      textStyle: { color: '#8b9eb5', fontSize: 11 }
    },
    xAxis: {
      type: 'category',
      data: times,
      boundaryGap: false,
      axisLine: { lineStyle: { color: '#1e3a5f' } },
      axisLabel: { color: '#6b7f9e', fontSize: 10 }
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: 100,
      splitLine: { lineStyle: { color: '#1e3a5f', type: 'dashed' } },
      axisLabel: { color: '#6b7f9e', fontSize: 10 }
    },
    series: [
      {
        name: legend1,
        type: 'line',
        data: machine1,
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: { color: '#00ffaa', width: 2 },
        itemStyle: { color: '#00ffaa' },
        connectNulls: false,  // 不连接null值，让折线断开
        markLine: {
          silent: true,
          symbol: 'none',
          label: {
            show: true,
            position: 'end',
            formatter: 'OEE标准线 75%',
            color: '#ff9800',
            fontSize: 10
          },
          lineStyle: {
            color: '#ff9800',
            type: 'dashed',
            width: 1
          },
          data: [{
            yAxis: 75,
            name: 'OEE标准线'
          }]
        }
      },
      {
        name: legend2,
        type: 'line',
        data: machine2,
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: { color: '#3b82f6', width: 2 },
        itemStyle: { color: '#3b82f6' },
        connectNulls: false  // 不连接null值，让折线断开
      }
    ]
  })
  
  // ✅ 启动tooltip自动循环播放（记住上次位置）
  const autoShowTooltip = () => {
    // 清除之前的定时器
    if (trendChartTooltipTimer) {
      clearInterval(trendChartTooltipTimer)
    }
    
    trendChartTooltipTimer = setInterval(() => {
      // 只在有数据的点上显示tooltip
      const validIndices = []
      times.forEach((time, index) => {
        if (machine1[index] != null || machine2[index] != null) {
          validIndices.push(index)
        }
      })
      
      if (validIndices.length === 0) return
      
      // ✅ 使用全局变量记住位置，循环显示
      if (trendChartTooltipIndex >= validIndices.length) {
        trendChartTooltipIndex = 0
      }
      
      chart.dispatchAction({
        type: 'showTip',
        seriesIndex: 0,
        dataIndex: validIndices[trendChartTooltipIndex]
      })
      
      trendChartTooltipIndex++
    }, 3000) // 每3秒切换一次
  }
  
  // 启动自动播放
  autoShowTooltip()
  
  // 鼠标悬停时暂停自动播放
  chart.on('mouseover', () => {
    if (trendChartTooltipTimer) {
      clearInterval(trendChartTooltipTimer)
      trendChartTooltipTimer = null
    }
  })
  
  // 鼠标离开时恢复自动播放
  chart.on('mouseout', () => {
    autoShowTooltip()
  })
}

const updateCharts = () => {
  // ✅ 清除OEE趋势图的定时器（但保留位置索引）
  if (trendChartTooltipTimer) {
    clearInterval(trendChartTooltipTimer)
    trendChartTooltipTimer = null
  }
  
  // 先销毁所有现有图表实例
  charts.forEach(chart => {
    if (chart && !chart.isDisposed()) {
      chart.dispose()
    }
  })
  charts.length = 0  // 清空数组
  
  // 重新初始化所有图表（OEE趋势图会从上次的位置继续播放）
  initSparklines()
  initProductionChart()
  initQualityCharts()
  initStaffChart()
  initTrendChart()
}

// 窗口大小变化时重新调整图表
const handleResize = () => {
  charts.forEach(chart => chart.resize())
}

// 监听ESC键退出全屏
const handleKeydown = (e) => {
  if (e.key === 'Escape' && isInFullscreen.value) {
    exitFullscreen()
  }
}

onMounted(async () => {
  // ✅ 初始化历史数据，填充初始值以确保图表有足够宽度
  historyData.value.utilization = Array(13).fill(0)
  historyData.value.production = Array(13).fill(0)
  historyData.value.quality = Array(13).fill(100)
  historyData.value.orders = Array(13).fill(0)
  // performance 不再需要，直接使用 hourlyOEE 中的数据
  
  console.log('🚀 驾驶舱启动，直接加载完整数据（无缓慢加载）')
  console.log('📊 报警数和性能稼动率迷你图使用OEE接口数据')
  
  // ✅ 先加载配置参数
  await loadProductionCoefficient() // 理论节拍系数（用于计算性能稼动率）
  await loadDailyWorkMinutes()      // 每日应工作分钟数（用于计算人员稼动率）
  await loadBreakTimes()            // 休息时间段配置（用于计算人员稼动率和OEE）
  
  // ✅ 直接加载数据并初始化图表，refreshAll 内已调用 updateCharts() 完成初始化
  await refreshAll()
  updateHistoryData()

  // 每5秒刷新一次数据
  refreshTimer = setInterval(refreshAll, 5000)
  
  window.addEventListener('resize', handleResize)
  window.addEventListener('focus', handleWindowResume)
  window.addEventListener('pageshow', handleWindowResume)
  document.addEventListener('keydown', handleKeydown)
  document.addEventListener('visibilitychange', handleWindowResume)
  
  // 监听全屏状态变化
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.addEventListener('msfullscreenchange', handleFullscreenChange)

  handleWindowResume()
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
  if (trendChartTooltipTimer) clearInterval(trendChartTooltipTimer) // ✅ 清除tooltip定时器
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('focus', handleWindowResume)
  window.removeEventListener('pageshow', handleWindowResume)
  document.removeEventListener('keydown', handleKeydown)
  document.removeEventListener('visibilitychange', handleWindowResume)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  document.removeEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.removeEventListener('msfullscreenchange', handleFullscreenChange)
  charts.forEach(chart => chart.dispose())
})
</script>

<style scoped>
.cockpit-container {
  width: 100%;
  height: 100vh;
  background: #0a0e27;
  color: #fff;
  padding: 1.2vh 1.2vw;
  box-sizing: border-box;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 标题栏 */
.cockpit-header {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 1vw 1vh;
  flex-shrink: 0;
  position: relative;
}

.title {
  font-size: clamp(20px, 2vw, 28px);
  font-weight: bold;
  background: linear-gradient(90deg, #00d4ff, #0088ff);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0;
  text-align: center;
}

.work-shift {
  margin-top: 0.2vw;
  margin-left: 1.5vw;
  font-size: clamp(11px, 1vw, 14px);
  color: #7b9cc5;
  padding: 0.6vh 1.2vw;
  border: 1.5px solid rgba(0, 170, 255, 0.5);
  border-radius: 2vw;
  background: linear-gradient(135deg, rgba(15, 35, 70, 0.4), rgba(10, 20, 45, 0.6));
  backdrop-filter: blur(8px);
  box-shadow: 0 0 15px rgba(0, 170, 255, 0.2);
  transition: all 0.3s ease;
  white-space: nowrap;
}

.work-shift:hover {
  border-color: rgba(0, 212, 255, 0.8);
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.4);
  transform: translateY(-1px);
}

.shift-label {
  color: #7b9cc5;
}

.shift-name {
  color: #00ffaa;
  font-weight: 500;
}

.fullscreen-btn {
  position: absolute;
  right: 1vw;
  background: rgba(0, 170, 255, 0.2);
  border: 1px solid rgba(0, 170, 255, 0.4);
  color: #00aaff;
  font-size: clamp(16px, 1.4vw, 20px);
  padding: 0.6vh 0.8vw;
  border-radius: 0.5vw;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fullscreen-btn:hover {
  background: rgba(0, 170, 255, 0.3);
  border-color: #00aaff;
  transform: scale(1.05);
}

/* KPI卡片 */
.kpi-cards {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 1vw;
  margin-bottom: 1.5vh;
  flex-shrink: 0;
}

.kpi-card {
  background: linear-gradient(135deg, rgba(15, 35, 70, 0.8), rgba(10, 20, 45, 0.9));
  border: 1px solid rgba(0, 170, 255, 0.25);
  border-radius: 0.8vw;
  padding: 1.2vh 1vw;
  position: relative;
  overflow: hidden;
  padding-bottom: 3.5vh;
}

.kpi-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #00d4ff, #0088ff, #00ffaa);
}

.kpi-label {
  font-size: clamp(10px, 0.75vw, 12px);
  color: #7b9cc5;
  margin-bottom: 0.8vh;
  position: relative;
  z-index: 1;
}

.kpi-value {
  font-size: clamp(24px, 2.2vw, 32px);
  font-weight: bold;
  color: #fff;
  margin-bottom: 0.5vh;
  position: relative;
  z-index: 1;
}

.kpi-value.alarm {
  color: #00ff88;
}

.kpi-value .unit {
  font-size: clamp(14px, 1.2vw, 18px);
  margin-left: 0.3vw;
}

/* 双栏显示卡片样式 */
.kpi-card-dual .kpi-dual-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8vw;
  margin-bottom: 0.5vh;
  position: relative;
  z-index: 1;
}

.kpi-dual-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.kpi-dual-label {
  font-size: clamp(9px, 0.7vw, 11px);
  color: #7b9cc5;
  margin-bottom: 0.5vh;
}

.kpi-dual-value {
  font-size: clamp(20px, 1.8vw, 26px);
  font-weight: bold;
  color: #00ffaa;
}

.kpi-dual-divider {
  width: 1px;
  height: 4vh;
  background: linear-gradient(to bottom, 
    rgba(0, 170, 255, 0),
    rgba(0, 170, 255, 0.5),
    rgba(0, 170, 255, 0)
  );
}

.kpi-subtitle {
  font-size: clamp(9px, 0.7vw, 11px);
  color: #ff9800;
  position: relative;
  z-index: 1;
}

.kpi-trend {
  font-size: clamp(9px, 0.7vw, 11px);
  color: #00ff88;
  display: flex;
  align-items: center;
  gap: 0.3vw;
  position: relative;
  z-index: 1;
}

.kpi-trend.stable {
  color: #7b9cc5;
}

.trend-arrow {
  font-weight: bold;
}

.kpi-sparkline {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 30%;
  opacity: 0.3;
  pointer-events: none;
}

.kpi-device-quality {
  display: flex;
  flex-direction: column;
  gap: 0.15vw;
  margin-top: 0.2vw;
  position: relative;
  z-index: 1;
}

.kpi-device-quality-item {
  font-size: clamp(9px, 0.65vw, 11px);
  color: #7b9cc5;
  line-height: 1.3;
}

/* 主内容区域 */
.main-content {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1.3fr 1fr;
  gap: 1vw;
  min-height: 0;
}

.left-column,
.center-column,
.right-column {
  display: flex;
  flex-direction: column;
  gap: 1vh;
  min-height: 0;
}

/* 面板样式 */
.panel {
  background: linear-gradient(135deg, rgba(15, 35, 70, 0.6), rgba(10, 20, 45, 0.8));
  border: 1px solid rgba(0, 170, 255, 0.2);
  border-radius: 0.8vw;
  padding: 1.2vh 1vw;
  backdrop-filter: blur(10px);
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 0.6vw;
  margin-bottom: 1vh;
  padding-bottom: 0.8vh;
  border-bottom: 1px solid rgba(0, 170, 255, 0.2);
}

.dot {
  width: 0.5vw;
  height: 0.5vw;
  background: #00ffaa;
  border-radius: 50%;
  box-shadow: 0 0 0.6vw #00ffaa;
  flex-shrink: 0;
}

.panel-title {
  font-size: clamp(11px, 0.9vw, 14px);
  font-weight: bold;
  color: #00aaff;
}

/* 计划总览表格 */
.left-column > .panel:nth-child(1) {
  flex: 0 0 auto;
}

.plan-table-container {
  max-height: 18vh;
  overflow-y: auto;
}

.plan-table {
  width: 100%;
  border-collapse: collapse;
  font-size: clamp(9px, 0.7vw, 11px);
}

.plan-table th {
  background: rgba(0, 170, 255, 0.1);
  color: #7b9cc5;
  padding: 0.8vh 0.5vw;
  text-align: left;
  font-weight: normal;
  border-bottom: 1px solid rgba(0, 170, 255, 0.3);
  white-space: nowrap;
}

.plan-table td {
  padding: 0.8vh 0.5vw;
  border-bottom: 1px solid rgba(0, 170, 255, 0.1);
  color: #fff;
}

.plan-table td.excellent {
  color: #00ff88;  /* 绿色：合格率 >= 90% */
  font-weight: bold;
}

.plan-table td.normal {
  color: #ff6666;  /* 红色：合格率 < 90% */
}

/* 今日产量 */
.left-column > .panel:nth-child(2) {
  flex: 1;
  min-height: 0;
}

.chart-panel {
  display: flex;
  flex-direction: column;
}

.chart-container {
  flex: 1;
  min-height: 0;
}

/* 质量分布 */
.left-column > .panel:nth-child(3) {
  flex: 0 0 auto;
}

.quality-charts {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5vw; /* 增加间距，避免图表重叠 */
  padding: 0.5vh 0; /* 添加上下内边距 */
}

.quality-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 0; /* 防止flex项目溢出 */
}

.quality-chart {
  width: 100%;
  height: 18vh;
  min-height: 150px; /* 增加最小高度，确保设备名有足够空间显示 */
}

.quality-legend {
  display: flex;
  gap: 1vw;
  margin-top: 0.5vh;
  font-size: clamp(9px, 0.7vw, 11px);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.3vw;
  color: #7b9cc5;
}

.legend-dot {
  width: 0.6vw;
  height: 0.6vw;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-dot.qualified {
  background: #00ffaa;
}

.legend-dot.defect {
  background: #ff4466;
}

/* 实时监控 */
.monitor-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.monitor-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  margin-bottom: 1vh;
  min-height: 0;
}

.monitor-item:last-child {
  margin-bottom: 0;
}

.video-wrapper {
  flex: 1;
  background: #000;
  border-radius: 0.5vw;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 0;
}

.video-player {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #4a5f8a;
  gap: 1vh;
}

.video-placeholder i {
  font-size: clamp(32px, 3vw, 48px);
  opacity: 0.5;
}

.video-placeholder p {
  font-size: clamp(11px, 0.9vw, 14px);
  margin: 0;
}

.video-placeholder .video-hint {
  font-size: clamp(9px, 0.7vw, 11px);
  color: #7b9cc5;
  margin-top: 0.5vh;
}

.video-info {
  margin-top: 0.8vh;
  padding: 0.5vh 0;
}

.video-tags {
  display: flex;
  gap: 0.6vw;
  margin-bottom: 0.6vh;
  flex-wrap: nowrap; /* 不换行，保持在一行 */
  line-height: 1.4;
  align-items: center;
}

.status-tag {
  padding: 0.3vh 0.6vw;
  border-radius: 0.3vw;
  font-size: clamp(9px, 0.7vw, 11px);
  font-weight: bold;
  white-space: nowrap;
}

.status-tag.running {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
  border: 1px solid #00ff88;
}

.status-tag.idle {
  background: rgba(255, 152, 0, 0.2);
  color: #ff9800;
  border: 1px solid #ff9800;
}

.status-tag.fault {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
  border: 1px solid #ff4444;
}

.param-tag {
  font-size: clamp(9px, 0.7vw, 11px);
  color: #7b9cc5;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.video-label {
  font-size: clamp(10px, 0.8vw, 12px);
  color: #7b9cc5;
  margin-top: 0.4vh;
  line-height: 1.3;
}

/* 员工绩效 */
.right-column > .panel:nth-child(1) {
  flex: 0 0 32%;
}

/* 稼动率趋势 */
.right-column > .panel:nth-child(2) {
  flex: 1;
  min-height: 0;
}

/* 滚动条 */
::-webkit-scrollbar {
  width: 0.4vw;
  height: 0.4vw;
}

::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 0.2vw;
}

::-webkit-scrollbar-thumb {
  background: rgba(0, 170, 255, 0.3);
  border-radius: 0.2vw;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 170, 255, 0.5);
}

</style>
