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
      <!-- OEE卡片 - 当班 OEE，左右分栏显示两台设备 -->
      <div class="kpi-card kpi-card-dual">
        <div class="kpi-label">当班OEE</div>
        <div class="kpi-dual-content">
          <div class="kpi-dual-item">
            <div class="kpi-dual-value">{{ shiftDevice1OEE }}%</div>
            <div class="kpi-dual-label">一号机</div>
          </div>
          <div class="kpi-dual-divider"></div>
          <div class="kpi-dual-item">
            <div class="kpi-dual-value">{{ shiftDevice2OEE }}%</div>
            <div class="kpi-dual-label">二号机</div>
          </div>
        </div>
        <div class="kpi-trend">
          <span class="trend-arrow">{{ deviceTrend.arrow }}</span>
          <span class="trend-text">平均 {{ shiftAvgOEE }}%</span>
        </div>
        <!-- 历史班次小字 -->
        <div class="kpi-shift-history" v-if="pastShiftsOEE.length">
          <div v-for="s in pastShiftsOEE" :key="s.shift_name" class="kpi-shift-row">
            {{ s.shift_name }}: ①{{ getDeviceOEE(s, '设备#1') }}% ②{{ getDeviceOEE(s, '设备#2') }}%
          </div>
        </div>
        <div class="kpi-sparkline" ref="sparkline1"></div>
      </div>
      <!-- 产量卡片 - 当班产量 -->
      <div class="kpi-card">
        <div class="kpi-label">当班产量</div>
        <div class="kpi-value">{{ shiftTotalProduction }} <span class="unit">件</span></div>
        <div class="kpi-trend">
          <span class="trend-arrow">↑</span>
          <span class="trend-text">当班统计</span>
        </div>
        <div class="kpi-device-quality">
          <span class="kpi-device-quality-item">一号机 {{ shiftDevice1Production }} 件</span>
          <span class="kpi-device-quality-item">二号机 {{ shiftDevice2Production }} 件</span>
        </div>
        <!-- 历史班次小字 -->
        <div class="kpi-shift-history" v-if="pastShiftsProdSummary.length">
          <div v-for="s in pastShiftsProdSummary" :key="s.shift_name" class="kpi-shift-row">
            {{ s.shift_name }}: ①{{ s.d1 }} ②{{ s.d2 }} 件
          </div>
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
      <!-- 性能稼动率卡片 - 当班稼动率，左右分栏显示两台设备 -->
      <div class="kpi-card kpi-card-dual">
        <div class="kpi-label">当班稼动率</div>
        <div class="kpi-dual-content">
          <div class="kpi-dual-item">
            <div class="kpi-dual-value">{{ shiftDevice1Performance }}%</div>
            <div class="kpi-dual-label">一号机</div>
          </div>
          <div class="kpi-dual-divider"></div>
          <div class="kpi-dual-item">
            <div class="kpi-dual-value">{{ shiftDevice2Performance }}%</div>
            <div class="kpi-dual-label">二号机</div>
          </div>
        </div>
        <div class="kpi-trend">
          <span class="trend-arrow">{{ performanceTrend.arrow }}</span>
          <span class="trend-text">平均 {{ shiftAvgPerformance }}%</span>
        </div>
        <!-- 历史班次小字 -->
        <div class="kpi-shift-history" v-if="pastShiftsOEE.length">
          <div v-for="s in pastShiftsOEE" :key="s.shift_name" class="kpi-shift-row">
            {{ s.shift_name }}: ①{{ getDevicePerf(s, '设备#1') }}% ②{{ getDevicePerf(s, '设备#2') }}%
          </div>
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
            <span class="panel-title">当班质量分布</span>
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
          <!-- 历史班次良品率小字 -->
          <div class="kpi-shift-history" v-if="pastShiftsQuality.length" style="padding: 4px 6px 0;">
            <div v-for="s in pastShiftsQuality" :key="s.shift_name" class="kpi-shift-row">
              {{ s.shift_name }}: ①{{ getDeviceQuality(s, '设备#1') }}% ②{{ getDeviceQuality(s, '设备#2') }}%
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

// ✅ OEE 数据（每小时，全天，供 sparkline 折线图使用）
const hourlyOEE = ref([])

// CN: 逻辑日活动班次列表（含 has_arrived / is_current 标记，用于 OEE 趋势图 x 轴构建）
// EN: Active shifts for the current logical day (with arrival flags for OEE chart x-axis).
// JP: 現在の論理日のアクティブシフト一覧（OEEチャートのx軸構築用、到達フラグ付き）。
const shiftsForDay = ref([])

// CN: 当前逻辑日所有已到达班次的 OEE 汇总（含当班+历史班，用于顶部卡片当班大字+历史小字）
// EN: OEE summaries for all arrived shifts of the logical day (current shift large text + history small text).
// JP: 到達済シフト全体のOEE集計（当班大字表示＋過去班小字履歴）。
const allShiftsOEE = ref([])

// CN: 当前逻辑日所有已到达班次的良品率汇总（用于左下角饼图当班数据+历史小字）
// EN: Quality summaries for all arrived shifts of the logical day.
// JP: 到達済シフト全体の良品率集計（左下角饼图用）。
const allShiftsQuality = ref([])

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

// ── 当班数据辅助 computed ──────────────────────────────────────────────────────
// CN: 从 allShiftsOEE 中找出当前班次（is_current 优先，否则最后一个已到达的）
// EN: Resolve the current shift summary from allShiftsOEE (is_current preferred, else last arrived).
// JP: allShiftsOEEから当班サマリを特定（is_current優先、なければ最後の到達済）。
const currentShiftOEESummary = computed(() =>
  allShiftsOEE.value.find(s => s.is_current) ||
  [...allShiftsOEE.value].reverse().find(s => s.has_arrived) ||
  allShiftsOEE.value[0] || null
)

// CN: 过去2班（排除当前班，最近的排前）
// EN: Past 2 shifts excluding current, most recent first.
// JP: 当番を除く過去2シフト（最新が先頭）。
const pastShiftsOEE = computed(() => {
  const currName = currentShiftOEESummary.value?.shift_name
  return allShiftsOEE.value
    .filter(s => s.has_arrived && s.shift_name !== currName)
    .slice(-2).reverse()
})

const currentShiftQuality = computed(() =>
  allShiftsQuality.value.find(s => s.is_current) ||
  [...allShiftsQuality.value].reverse().find(s => s.has_arrived) ||
  allShiftsQuality.value[0] || null
)

const pastShiftsQuality = computed(() => {
  const currName = currentShiftQuality.value?.shift_name
  return allShiftsQuality.value
    .filter(s => s.has_arrived && s.shift_name !== currName)
    .slice(-2).reverse()
})

// CN: 辅助函数：从班次OEE汇总中提取指定设备的数值
// EN: Helper: extract per-device values from a ShiftOEESummary/ShiftQualitySummary object.
// JP: ヘルパー：ShiftOEESummary/ShiftQualitySummaryから設備別値を取得。
const getDeviceOEE = (summary, deviceName) => {
  const d = summary?.devices?.find(d => d.device_name === deviceName)
  return d?.oee_pct != null ? d.oee_pct.toFixed(1) : '0.0'
}
const getDevicePerf = (summary, deviceName) => {
  const d = summary?.devices?.find(d => d.device_name === deviceName)
  return d?.performance_pct != null ? d.performance_pct.toFixed(1) : '0.0'
}
const getDeviceQuality = (summary, deviceName) => {
  const d = summary?.devices?.find(d => d.device_name === deviceName)
  return d?.quality_rate != null ? d.quality_rate.toFixed(1) : '100.0'
}

// 当班 OEE ──
const shiftDevice1OEE = computed(() => getDeviceOEE(currentShiftOEESummary.value, '设备#1'))
const shiftDevice2OEE = computed(() => getDeviceOEE(currentShiftOEESummary.value, '设备#2'))
const shiftAvgOEE = computed(() => {
  const s = currentShiftOEESummary.value
  if (!s?.devices?.length) return '0.0'
  const sum = s.devices.reduce((acc, d) => acc + (d.oee_pct || 0), 0)
  return (sum / s.devices.length).toFixed(1)
})

// 当班性能稼动率 ──
const shiftDevice1Performance = computed(() => getDevicePerf(currentShiftOEESummary.value, '设备#1'))
const shiftDevice2Performance = computed(() => getDevicePerf(currentShiftOEESummary.value, '设备#2'))
const shiftAvgPerformance = computed(() => {
  const s = currentShiftOEESummary.value
  if (!s?.devices?.length) return '0.0'
  const sum = s.devices.reduce((acc, d) => acc + (d.performance_pct || 0), 0)
  return (sum / s.devices.length).toFixed(1)
})

// 当班产量（前端过滤 hourlyProduction 按班次时间窗口）──
// CN: 从 shiftsForDay 中找当前班（与 OEE 保持一致的逻辑）
// EN: Resolve current shift from shiftsForDay (same fallback logic as OEE).
// JP: shiftsForDayから当班を特定（OEEと同じフォールバックロジック）。
const currentShiftFromLogicalDay = computed(() =>
  shiftsForDay.value.find(s => s.is_current) ||
  [...shiftsForDay.value].reverse().find(s => s.has_arrived) ||
  shiftsForDay.value[0] || null
)

const filterProdByShift = (shift, deviceNames = null) => {
  if (!shift) return 0
  const pad = n => String(n).padStart(2, '0')
  const base = shift.logical_date
  const startMin = shift.start_hour * 60 + shift.start_min
  const endMin = shift.end_hour * 60 + shift.end_min
  const startDt = new Date(`${base}T${pad(shift.start_hour)}:${pad(shift.start_min)}:00`)
  let endDt
  if (endMin > startMin) {
    endDt = new Date(`${base}T${pad(shift.end_hour)}:${pad(shift.end_min)}:00`)
  } else {
    const next = new Date(base)
    next.setDate(next.getDate() + 1)
    const nd = next.toISOString().slice(0, 10)
    endDt = new Date(`${nd}T${pad(shift.end_hour)}:${pad(shift.end_min)}:00`)
  }
  return hourlyProduction.value
    .filter(item => {
      if (deviceNames && !deviceNames.includes(item.device_name)) return false
      const t = new Date(item.time_slot)
      return t >= startDt && t < endDt
    })
    .reduce((sum, item) => sum + (item.ok_qty || 0) + (item.ng_qty || 0), 0)
}

const shiftTotalProduction = computed(() => filterProdByShift(currentShiftFromLogicalDay.value))
const shiftDevice1Production = computed(() => filterProdByShift(currentShiftFromLogicalDay.value, ['设备#1', '设备1', '一号机']))
const shiftDevice2Production = computed(() => filterProdByShift(currentShiftFromLogicalDay.value, ['设备#2', '设备2', '二号机']))

// 过去2班产量摘要（用于卡片底部小字）
const pastShiftsProdSummary = computed(() => {
  const currName = currentShiftFromLogicalDay.value?.name
  return shiftsForDay.value
    .filter(s => s.has_arrived && s.name !== currName)
    .slice(-2).reverse()
    .map(s => ({
      shift_name: s.name,
      d1: filterProdByShift(s, ['设备#1', '设备1', '一号机']),
      d2: filterProdByShift(s, ['设备#2', '设备2', '二号机'])
    }))
})

// ── 当班数据辅助 end ───────────────────────────────────────────────────────────

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

// CN: 加载人员稼动率 — 基于逻辑日班次（shiftsForDay）计算，与 OEE/良品图保持一致。
//     应工作时长 = 所有已到达班次的时间窗口之和（扣除班内休息），
//     实际工作时长 = 员工 session 在班次窗口内的有效时长（同样扣除重叠休息段）。
// EN: Load staff efficiency — computed from logical-day shifts (consistent with OEE/production).
//     Available work = sum of arrived shift windows minus breaks;
//     Actual work = session overlap with shift windows minus break overlaps.
// JP: 人員稼働率読込 — OEE/良品と同じ論理日シフトに基づき算出。
//     応工作 = 到達済シフト窓の合計−休憩、実作業 = セッションとシフト窓の重畳−休憩重畳。
const loadStaffEfficiency = async () => {
  try {
    if (!window.go?.main?.App) return

    // 1. 获取所有员工
    const allStaff = await window.go.main.App.GetAllStaff(null, 1)
    if (!allStaff || allStaff.length === 0) {
      staffEfficiency.value = []
      return
    }

    // 2. ✅ 使用逻辑日班次（与 OEE 一致）
    const shifts = shiftsForDay.value || []
    // 只取已到达的班次（顺序前缀，与 OEE 图逻辑相同）
    let lastArrivedIdx = -1
    for (let i = 0; i < shifts.length; i++) {
      if (shifts[i].has_arrived) lastArrivedIdx = i
      else break
    }
    const arrivedShifts = lastArrivedIdx >= 0 ? shifts.slice(0, lastArrivedIdx + 1) : []
    if (arrivedShifts.length === 0) {
      staffEfficiency.value = []
      return
    }

    const logicalDate = shifts[0]?.logical_date || new Date().toISOString().split('T')[0]
    const logicalBase = new Date(logicalDate + 'T00:00:00')
    const now = new Date()

    // 3. 构建班次窗口列表（绝对 Date 对象，便于后续计算）
    const shiftWindows = arrivedShifts.map(s => {
      const startMin = s.start_hour * 60 + s.start_min
      const endMin   = s.end_hour * 60 + s.end_min
      const isCross  = endMin <= startMin
      const shiftStart = new Date(logicalBase.getTime() + startMin * 60000)
      const shiftEnd   = new Date(logicalBase.getTime() + (isCross ? endMin + 24 * 60 : endMin) * 60000)
      // 如果当前班次还在进行中，截止到"现在"
      const effectiveEnd = s.is_current && now < shiftEnd ? now : shiftEnd
      return { start: shiftStart, end: effectiveEnd, shift: s }
    })

    // 4. 休息段转为绝对 Date 辅助函数（休息段属于逻辑日内，可能跨日）
    const breaks = breakTimes.value.map(bt => ({
      start: bt.start_hour * 60 + bt.start_min,
      end:   bt.end_hour * 60 + bt.end_min
    }))

    // 辅助：计算「时间段 [a, b) 内被休息段占用的分钟数」
    const overlapBreakMinutes = (aStart, aEnd) => {
      const aStartMin = (aStart - logicalBase) / 60000
      const aEndMin   = (aEnd - logicalBase) / 60000
      let total = 0
      for (const brk of breaks) {
        const os = Math.max(aStartMin, brk.start)
        const oe = Math.min(aEndMin, brk.end)
        if (os < oe) total += (oe - os)
      }
      return total
    }

    // 5. 计算应工作总分钟数（所有已到达班次窗口扣除休息）
    let totalAvailableMinutes = 0
    for (const w of shiftWindows) {
      const raw = (w.end - w.start) / 60000
      const brk = overlapBreakMinutes(w.start, w.end)
      totalAvailableMinutes += Math.max(0, raw - brk)
    }

    console.log(`📊 人员稼动率（逻辑日 ${logicalDate}，已到达 ${arrivedShifts.length} 个班次）`)
    console.log(`   应工作总时长: ${Math.round(totalAvailableMinutes)} 分钟 (${(totalAvailableMinutes/60).toFixed(1)} 小时)`)

    // 6. 获取逻辑日的班次记录（按日期范围查询）
    // 跨日班次：结束日期可能是 logicalDate + 1
    const endDate = (() => {
      const last = shiftWindows[shiftWindows.length - 1]
      return last.end.toISOString().split('T')[0]
    })()

    let allSessions = []
    try {
      allSessions = await window.go.main.App.GetSessionHistory(null, null, logicalDate, endDate) || []
      // 补充当前活动 session
      const activeSessions = await window.go.main.App.GetAllActiveSessions() || []
      if (activeSessions.length > 0) {
        const existIds = new Set(allSessions.map(s => s.session_id))
        const extra = activeSessions.filter(s => !existIds.has(s.session_id))
        if (extra.length > 0) allSessions = [...allSessions, ...extra]
      }
      console.log(`   获取到 ${allSessions.length} 条 session 记录`)
    } catch (e) {
      console.error('获取逻辑日 session 失败:', e)
    }

    // 7. 为每个员工计算稼动率
    const efficiencyData = []
    for (const staff of allStaff) {
      let workingMinutes = 0

      for (const session of allSessions) {
        try {
          const staffIds = JSON.parse(session.staff_ids || '[]')
          if (!staffIds.includes(staff.id)) continue

          const loginTime  = new Date(session.login_time)
          const logoutTime = session.logout_time ? new Date(session.logout_time) : now

          // 计算 session 与每个班次窗口的有效重叠（扣除休息）
          for (const w of shiftWindows) {
            const os = Math.max(loginTime.getTime(), w.start.getTime())
            const oe = Math.min(logoutTime.getTime(), w.end.getTime())
            if (os >= oe) continue
            const rawOverlap = (oe - os) / 60000
            const brkOverlap = overlapBreakMinutes(new Date(os), new Date(oe))
            workingMinutes += Math.max(0, rawOverlap - brkOverlap)
          }
        } catch (e) {
          // 跳过解析失败的 session
        }
      }

      const efficiency = totalAvailableMinutes > 0
        ? Math.round((workingMinutes / totalAvailableMinutes) * 100)
        : 0

      efficiencyData.push({
        staff_id: staff.id,
        staff_name: staff.name,
        working_min: Math.round(workingMinutes),
        efficiency
      })
    }

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

// CN: 加载逻辑日班次列表，用于 OEE 趋势图 x 轴分段和背景色渲染
// EN: Load logical-day shift list for OEE chart x-axis segmentation and background coloring.
// JP: OEEチャートのx軸分割と背景色描画用に論理日シフト一覧を読み込む。
const loadShiftsForDay = async () => {
  try {
    if (window.go?.main?.App?.GetShiftsForLogicalDay) {
      const result = await window.go.main.App.GetShiftsForLogicalDay()
      shiftsForDay.value = result || []
    }
  } catch (e) {
    console.error('加载逻辑日班次失败:', e)
    shiftsForDay.value = []
  }
}

// ✅ 加载每小时OEE数据
const loadHourlyOEE = async () => {
  try {
    if (window.go?.main?.App?.GetHourlyOEE) {
      console.log('🔍 开始加载每小时OEE数据...')
      const result = await window.go.main.App.GetHourlyOEE()
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

// CN: 加载当前逻辑日所有已到达班次的 OEE 汇总（用于卡片当班大字+历史班次小字）
// EN: Load OEE summaries for all arrived shifts of the current logical day.
// JP: 現在の論理日の到達済シフト全体のOEEサマリを読み込む。
const loadAllShiftsOEE = async () => {
  try {
    if (window.go?.main?.App?.GetAllShiftsOEESummary) {
      allShiftsOEE.value = await window.go.main.App.GetAllShiftsOEESummary() || []
    }
  } catch (e) {
    console.error('❌ 加载班次OEE汇总失败:', e)
    allShiftsOEE.value = []
  }
}

// CN: 加载当前逻辑日所有已到达班次的良品率汇总（用于左下角饼图当班数据+历史小字）
// EN: Load quality summaries for all arrived shifts of the current logical day.
// JP: 現在の論理日の到達済シフト全体の良品率サマリを読み込む。
const loadAllShiftsQuality = async () => {
  try {
    if (window.go?.main?.App?.GetAllShiftsQualitySummary) {
      allShiftsQuality.value = await window.go.main.App.GetAllShiftsQualitySummary() || []
    }
  } catch (e) {
    console.error('❌ 加载班次良品率汇总失败:', e)
    allShiftsQuality.value = []
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
    loadHourlyOEE(),         // ✅ 加载每小时OEE（全天，用于sparkline）
    loadShiftsForDay(),      // ✅ 加载逻辑日班次（OEE x轴）
    loadAllShiftsOEE(),      // ✅ 加载所有班次OEE汇总（当班大字+历史小字）
    loadAllShiftsQuality()   // ✅ 加载所有班次良品率汇总（左下角饼图）
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
// CN: 初始化今日良品图（柱状图），x 轴与 OEE 趋势图保持相同的班次分段逻辑
// EN: Init today's good-parts bar chart; x-axis uses the same shift-based segment logic as OEE.
// JP: 今日の良品棒グラフを初期化。x軸はOEEトレンドチャートと同じシフト分割ロジックを使用。
const initProductionChart = () => {
  if (!productionChart.value) return
  const chart = echarts.init(productionChart.value)
  charts.push(chart)

  // ── 与 OEE 图相同：顺序前缀 → 只取已到达班次 ──────────────
  let lastArrivedIdx = -1
  for (let i = 0; i < shiftsForDay.value.length; i++) {
    if (shiftsForDay.value[i].has_arrived) lastArrivedIdx = i
    else break
  }
  const arrivedShifts = lastArrivedIdx >= 0 ? shiftsForDay.value.slice(0, lastArrivedIdx + 1) : []

  // ── 逻辑日期（用于 time_slot 匹配）─────────────────────────
  const logicalDate = shiftsForDay.value[0]?.logical_date || new Date().toISOString().split('T')[0]

  // ── 班次配色（与 OEE 保持一致）──────────────────────────────
  const shiftBgColors = [
    'rgba(0, 210, 130, 0.11)',
    'rgba(50, 140, 255, 0.11)',
    'rgba(255, 165, 30, 0.11)',
    'rgba(200, 80, 220, 0.11)',
    'rgba(0, 210, 220, 0.11)',
  ]

  // ── 构建 x 轴 & markArea（value 轴 + 小时小数值 = 真正分钟级精度）────
  // CN: x 轴使用 value 类型。ECharts category 轴的 OrdinalScale.parse() 会对小数
  //     做 Math.round() 取整，导致 markArea 分钟精度被抹杀。改用 value 轴后，
  //     markArea 的 xAxis 直接用小时小数（如 7.667 = 07:40），真正精确对齐班次分钟边界。
  // EN: Uses value-type x-axis. Category axis OrdinalScale.parse() rounds fractional values
  //     via Math.round(), destroying minute precision. With value axis, markArea xAxis uses
  //     hour decimals (e.g. 7.667 = 07:40) for true minute-level shift boundary alignment.
  // JP: value型x軸を使用。category軸のOrdinalScale.parse()は小数をMath.round()で丸めるため
  //     分精度が消える。value軸なら markArea の xAxis に時刻小数(例: 7.667=07:40)を直接使用し、
  //     シフト境界を分単位で正確に揃えられる。
  const hourValues = []
  const hourSet = new Set()
  const markAreaData = []

  if (arrivedShifts.length === 0) {
    chart.setOption({
      graphic: [{ type: 'text', left: 'center', top: 'middle',
        style: { text: '等待班次开始...', fill: '#6b7f9e', fontSize: 13 } }]
    })
    return
  }

  // 构建 hourValues 数组（整数小时值，供数据映射用）
  const shiftMeta = arrivedShifts.map(shift => {
    const startH  = shift.start_hour
    const endHRaw = shift.end_hour
    const endM    = shift.end_min
    const isCrossMidnight = endHRaw < startH || (endHRaw === startH && endM < shift.start_min)
    const effectiveEndH   = isCrossMidnight ? endHRaw + 24 : endHRaw
    const lastBucketH     = (endM === 0) ? effectiveEndH - 1 : effectiveEndH

    for (let h = startH; h <= lastBucketH; h++) {
      if (!hourSet.has(h)) {
        hourValues.push(h)
        hourSet.add(h)
      }
    }
    return { shift, startH, effectiveEndH }
  })

  // markArea 直接使用小时小数值：hour + min/60（如 7 + 40/60 ≈ 7.667）
  for (let i = 0; i < shiftMeta.length; i++) {
    const { shift, startH, effectiveEndH } = shiftMeta[i]
    markAreaData.push([
      {
        xAxis: startH + shift.start_min / 60,
        itemStyle: { color: shiftBgColors[i % shiftBgColors.length] },
        label: { show: true, position: 'insideTopLeft', formatter: shift.name,
          color: 'rgba(180,200,220,0.7)', fontSize: 9 }
      },
      { xAxis: effectiveEndH + shift.end_min / 60 }
    ])
  }

  // ── 获取有产量数据的设备名称 ──────────────────────────────────
  const deviceNames = [...new Set(hourlyProduction.value.map(h => h.device_name))]

  const barColors = [
    ['#00d4ff', '#0088ff'],
    ['#00ff88', '#00aa55'],
    ['#ff9800', '#ff5722'],
    ['#a855f7', '#7c3aed'],
  ]

  // ── 为每个设备生成柱状系列（value 轴需 [x, y] 格式）─────────
  const seriesData = []
  const legendData = []

  deviceNames.forEach((deviceName, index) => {
    const deviceHourlyData = hourValues.map(h => {
      // 用实际 hour_idx（跨日时 > 23）拼接 time_slot，对应后端 FLOOR(TIMESTAMPDIFF/3600) 的绝对小时
      const displayH = h % 24
      const nextDay = h >= 24
      const dateStr = nextDay
        ? (() => { const d = new Date(logicalDate); d.setDate(d.getDate() + 1); return d.toISOString().split('T')[0] })()
        : logicalDate
      const timeSlot = `${dateStr} ${String(displayH).padStart(2, '0')}:00:00`
      const found = hourlyProduction.value.find(
        hp => hp.device_name === deviceName && hp.time_slot === timeSlot
      )
      if (!found || found.total_qty === 0) return [h, null]
      return [h, found.ok_qty || 0]
    })

    legendData.push(deviceName)
    const ci = index % barColors.length
    seriesData.push({
      name: deviceName,
      type: 'bar',
      data: deviceHourlyData,
      barWidth: deviceNames.length === 1 ? '50%' : '35%',
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: barColors[ci][0] },
          { offset: 1, color: barColors[ci][1] }
        ])
      },
      label: {
        show: true, position: 'top',
        formatter: p => {
          const v = Array.isArray(p.value) ? p.value[1] : p.value
          return v != null ? v : ''
        },
        color: '#fff', fontSize: 10, fontWeight: 'bold'
      },
      connectNulls: false,
      // 仅第一个设备系列挂 markArea（避免重叠）
      ...(index === 0 ? { markArea: { silent: true, data: markAreaData } } : {})
    })
  })

  // CN: 若无数据，追加一条隐藏 dummy 系列来承载班次背景色。
  // EN: If no data, add an invisible dummy series to keep shift backgrounds.
  // JP: データなしの場合は非表示の dummy シリーズで背景色を維持する。
  const finalSeries = seriesData.length > 0
    ? seriesData
    : [{
        type: 'bar',
        data: hourValues.map(h => [h, null]),
        barWidth: 0,
        silent: true,
        itemStyle: { opacity: 0 },
        markArea: { silent: true, data: markAreaData }
      }]

  chart.setOption({
    animation: false,
    grid: { top: 19, right: 12, bottom: 18, left: 22 },
    legend: {
      data: legendData, top: 5, right: 20,
      textStyle: { color: '#8b9eb5', fontSize: 11 }
    },
    xAxis: {
      type: 'value',
      min: hourValues[0],
      max: hourValues[hourValues.length - 1] + 1,
      interval: (() => { const span = hourValues.length; return span > 16 ? 3 : span > 10 ? 2 : 1 })(),
      splitLine: { show: false },
      axisTick: { show: false },
      axisLine: { lineStyle: { color: '#1e3a5f' } },
      axisLabel: {
        color: '#6b7f9e', fontSize: 10,
        showMaxLabel: false,
        formatter: v => {
          if (v !== Math.round(v)) return ''
          return `${String(Math.round(v) % 24).padStart(2, '0')}:00`
        }
      }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: '#1e3a5f' } },
      axisLabel: { color: '#6b7f9e', fontSize: 10 }
    },
    series: finalSeries
  })
}

// 初始化质量分布图表 - 使用在产工单数据
const initQualityCharts = () => {
  // CN: 使用当班良品率数据（allShiftsQuality → currentShiftQuality），无数据则显示100%满绿环
  // EN: Use current-shift quality data; show 100% full green ring when no data.
  // JP: 当班良品率データを使用。データなしの場合は100%フルグリーン表示。
  const qualityData = currentShiftQuality.value?.devices || []
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

    // 优先用当班质量数据匹配设备，找不到则用设备列表里的名称
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
      // CN: 当班暂无质量数据，显示100%满绿环（符合需求：无数据=100%良品）
      // EN: No shift quality data yet — show 100% full green ring per spec.
      // JP: 当班データ未着時は100%フルグリーン（仕様：データなし=100%良品）。
      chart.setOption({
        animation: false,
        series: [{
          type: 'pie',
          radius: ['40%', '55%'],
          center: ['50%', '42%'],
          data: [{ value: 100, name: '合格', itemStyle: { color: '#00ffaa' } }],
          label: { show: false },
          emphasis: { scale: false },
          labelLine: { show: false }
        }],
        graphic: [
          {
            type: 'text',
            left: 'center',
            top: '38%',
            style: { text: '100.0%', fontSize: 18, fontWeight: 'bold', fill: '#fff', textAlign: 'center' }
          },
          {
            type: 'text',
            left: 'center',
            top: '82%',
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

// CN: 初始化 OEE 趋势图（按逻辑日班次动态构建 x 轴）
// EN: Init OEE trend chart with shift-aware x-axis for the current logical day.
// JP: 論理日シフトに基づく動的x軸でOEEトレンドグラフを初期化する。
//
// 核心规则 / Core rules:
//   1. 只在 x 轴展示已到达（has_arrived=true）的班次时间段
//   2. 不同班次用不同半透明背景色（markArea）区分
//   3. 未来时段数据为 null，折线自然断开
//   4. 最小加载粒度为完整班次（一旦班次开始，该班次全部小时都进 x 轴）
const initTrendChart = () => {
  if (!trendChart.value) return
  const chart = echarts.init(trendChart.value)
  charts.push(chart)

  const now = new Date()
  // CN: 跨日修正 currentHour：若逻辑日属于昨天（凌晨时刻还在昨天的班次里），
  //     当前时刻对应的 hour_idx 应为 24 + now.getHours()，否则历史班次数据会被误判为"未来"。
  // EN: Cross-midnight correction: if logical date is yesterday, add 24 to current hour so
  //     past data buckets (h=14..23 from yesterday) are not mistakenly treated as future.
  // JP: 論理日が昨日の場合、currentHour に 24 を加算することで、昨日のバケット（h=14..23）
  //     が未来として誤判定されるのを防ぐ。
  const logicalDate = shiftsForDay.value[0]?.logical_date
  const todayStr = now.toISOString().split('T')[0]
  const isNextDay = logicalDate && logicalDate < todayStr  // 逻辑日是昨天 → 处于跨日凌晨段
  const currentHour = isNextDay ? 24 + now.getHours() : now.getHours()

  // 排除汇总行
  const hourlyData = hourlyOEE.value.filter(
    item => !item.time_period || !item.time_period.includes('合计')
  )

  // CN: 顺序前缀：从头扫描，遇到第一个未到达的班次就停止
  //     避免跳过未到达的中间班次（如班1已结束、班2未开始、班3已开始 → 只显示班1）
  // EN: Sequential prefix: stop at the first non-arrived shift to prevent skipping gaps.
  // JP: 順序プレフィックス：最初の未到達シフトで打ち切り、間のシフトを飛ばさない。
  let lastArrivedIdx = -1
  for (let i = 0; i < shiftsForDay.value.length; i++) {
    if (shiftsForDay.value[i].has_arrived) {
      lastArrivedIdx = i
    } else {
      break
    }
  }
  const arrivedShifts = lastArrivedIdx >= 0 ? shiftsForDay.value.slice(0, lastArrivedIdx + 1) : []

  // ─── 今天还没到第一班：显示等待提示 ───────────────────────
  if (arrivedShifts.length === 0) {
    chart.setOption({
      graphic: [{
        type: 'text',
        left: 'center',
        top: 'middle',
        style: { text: '等待班次开始...', fill: '#6b7f9e', fontSize: 13 }
      }]
    })
    return
  }

  // ─── 班次配色（科技深色风格）──────────────────────────────
  // CN: 每个班次使用不同的半透明背景色；颜色数组循环使用
  // EN: Each shift uses a distinct translucent background; colors cycle if more shifts than colors.
  const shiftBgColors = [
    'rgba(0, 210, 130, 0.11)',   // 一班 - 绿
    'rgba(50, 140, 255, 0.11)',  // 二班 - 蓝
    'rgba(255, 165, 30, 0.11)',  // 三班 - 橙
    'rgba(200, 80, 220, 0.11)',  // 四班 - 紫
    'rgba(0, 210, 220, 0.11)',   // 五班 - 青
  ]

  // ─── 构建 x 轴和班次 markArea（value 轴 + 小时小数值 = 真正分钟级精度）─
  // CN: 与今日良品图相同，改用 value 轴。ECharts category 轴的 OrdinalScale.parse()
  //     对小数做 Math.round() 取整，导致 markArea 分钟精度被抹杀。value 轴的 markArea
  //     xAxis 直接用小时小数（如 7.667=07:40），真正精确对齐班次分钟边界。
  // EN: Same as production chart – switched to value axis. Category OrdinalScale.parse()
  //     rounds fractional values, destroying minute precision. Value axis markArea uses
  //     hour decimals directly for true minute-level alignment.
  // JP: 良品グラフと同様にvalue軸に変更。category軸のOrdinalScale.parse()は小数を丸めるため
  //     分精度が消える。value軸ならmarkAreaのxAxisに時刻小数を直接使用し正確に揃えられる。
  const hourValues = []
  const hourSet = new Set()
  const markAreaData = []

  // 构建 hourValues 数组（整数小时值，供数据映射用）
  const oeeShiftMeta = arrivedShifts.map(shift => {
    const startH  = shift.start_hour
    const endHRaw = shift.end_hour
    const endM    = shift.end_min
    const isCrossMidnight = endHRaw < startH || (endHRaw === startH && endM < shift.start_min)
    const effectiveEndH   = isCrossMidnight ? endHRaw + 24 : endHRaw
    const lastBucketH     = (endM === 0) ? effectiveEndH - 1 : effectiveEndH

    for (let h = startH; h <= lastBucketH; h++) {
      if (!hourSet.has(h)) {
        hourValues.push(h)
        hourSet.add(h)
      }
    }
    return { shift, startH, effectiveEndH }
  })

  // markArea 直接使用小时小数值：hour + min/60（如 7 + 40/60 ≈ 7.667）
  for (let i = 0; i < oeeShiftMeta.length; i++) {
    const { shift, startH, effectiveEndH } = oeeShiftMeta[i]
    markAreaData.push([
      {
        xAxis: startH + shift.start_min / 60,
        itemStyle: { color: shiftBgColors[i % shiftBgColors.length] },
        label: {
          show: true,
          position: 'insideTopLeft',
          formatter: shift.name,
          color: 'rgba(180,200,220,0.7)',
          fontSize: 9
        }
      },
      { xAxis: effectiveEndH + shift.end_min / 60 }
    ])
  }

  // ─── 映射 OEE 数据到 x 轴 ────────────────────────────────
  // CN: 直接用整数小时值匹配后端 t.hour（跨日时可 > 23）。
  //     "未来"判断：h > currentHour。数据格式为 [hour, oee_pct]，供 value 轴使用。
  // EN: Match hourly data using integer hour values. Future hours appear as null.
  //     Data format: [hour, oee_pct] for value axis.
  // JP: 整数時間で後端 t.hour と直接照合。未来のhourはnull。データ形式: [hour, oee_pct]。
  const mapOEE = (deviceName) => hourValues.map(h => {
    if (h > currentHour) return null  // 未来时段：折线断开
    const d = hourlyData.find(t => t.device_name === deviceName && t.hour === h)
    return d ? (d.oee_pct ?? null) : null
  })

  const machine1 = mapOEE('设备#1')
  const machine2 = mapOEE('设备#2')

  // 转换为 [x, y] 格式供 value 轴使用
  const machine1XY = hourValues.map((h, i) => [h, machine1[i]])
  const machine2XY = hourValues.map((h, i) => [h, machine2[i]])

  chart.setOption({
    animation: false,
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(10, 20, 45, 0.95)',
      borderColor: 'rgba(0, 170, 255, 0.5)',
      borderWidth: 1,
      textStyle: { color: '#fff', fontSize: 12 },
      formatter(params) {
        if (!params?.length) return ''
        const hourVal = params[0].axisValue
        const label = `${String(Math.round(hourVal) % 24).padStart(2, '0')}:00`
        let html = `<div style="font-weight:bold;margin-bottom:5px;">${label}</div>`
        params.forEach(p => {
          const val = Array.isArray(p.value) ? p.value[1] : p.value
          if (val != null) {
            html += `<div style="margin:3px 0;">
              <span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${p.color};margin-right:5px;"></span>
              <span>${p.seriesName}: ${val.toFixed(1)}%</span>
            </div>`
          }
        })
        return html
      }
    },
    grid: { top: 38, right: 18, bottom: 18, left: 22 },
    legend: {
      data: ['设备#1', '设备#2'],
      top: 5, right: 20,
      textStyle: { color: '#8b9eb5', fontSize: 11 }
    },
    xAxis: {
      type: 'value',
      min: hourValues[0],
      max: hourValues[hourValues.length - 1] + 1,
      interval: (() => { const span = hourValues.length; return span > 16 ? 3 : span > 10 ? 2 : 1 })(),
      splitLine: { show: false },
      axisTick: { show: false },
      axisLine: { lineStyle: { color: '#1e3a5f' } },
      axisLabel: {
        color: '#6b7f9e', fontSize: 10,
        showMaxLabel: false,
        formatter: v => {
          if (v !== Math.round(v)) return ''
          return `${String(Math.round(v) % 24).padStart(2, '0')}:00`
        }
      }
    },
    yAxis: {
      type: 'value',
      min: 0, max: 100,
      splitLine: { lineStyle: { color: '#1e3a5f', type: 'dashed' } },
      axisLabel: { color: '#6b7f9e', fontSize: 10 }
    },
    series: [
      {
        name: '设备#1',
        type: 'line',
        data: machine1XY,
        smooth: true,
        symbol: 'circle', symbolSize: 6,
        lineStyle: { color: '#00ffaa', width: 2 },
        itemStyle: { color: '#00ffaa' },
        connectNulls: false,
        markLine: {
          silent: true,
          symbol: 'none',
          label: { show: true, position: 'end', formatter: 'OEE 75%', color: '#ff9800', fontSize: 10 },
          lineStyle: { color: '#ff9800', type: 'dashed', width: 1 },
          data: [{ yAxis: 75 }]
        },
        markArea: { silent: true, data: markAreaData }
      },
      {
        name: '设备#2',
        type: 'line',
        data: machine2XY,
        smooth: true,
        symbol: 'circle', symbolSize: 6,
        lineStyle: { color: '#3b82f6', width: 2 },
        itemStyle: { color: '#3b82f6' },
        connectNulls: false
      }
    ]
  })

  // ─── tooltip 自动轮播（保留上次位置）────────────────────────
  const autoShowTooltip = () => {
    if (trendChartTooltipTimer) clearInterval(trendChartTooltipTimer)
    trendChartTooltipTimer = setInterval(() => {
      const validIndices = []
      hourValues.forEach((_, i) => {
        if (machine1[i] != null || machine2[i] != null) validIndices.push(i)
      })
      if (!validIndices.length) return
      if (trendChartTooltipIndex >= validIndices.length) trendChartTooltipIndex = 0
      chart.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: validIndices[trendChartTooltipIndex] })
      trendChartTooltipIndex++
    }, 3000)
  }

  autoShowTooltip()
  chart.on('mouseover', () => { if (trendChartTooltipTimer) { clearInterval(trendChartTooltipTimer); trendChartTooltipTimer = null } })
  chart.on('mouseout', () => { autoShowTooltip() })
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
  /* CN: 锁定卡片行高度占视口14%，防止班次历史行将整行撑高
     EN: Cap the KPI row at 14 vh so shift-history rows cannot push the row taller.
     JP: シフト履歴行による行拡張を防ぐため、KPIカード行高を14vhに固定。 */
  height: 14vh;
  align-items: stretch;
}

.kpi-card {
  background: linear-gradient(135deg, rgba(15, 35, 70, 0.8), rgba(10, 20, 45, 0.9));
  border: 1px solid rgba(0, 170, 255, 0.25);
  border-radius: 0.8vw;
  padding: 1.2vh 1vw;
  /* CN: flex 列布局，历史班次块通过 margin-top:auto 固定在底部，消除卡片间视觉高低不齐
     EN: Flex column keeps main content top-aligned; shift-history is pushed to the bottom via margin-top:auto.
     JP: flex列でメインコンテンツを上揃え、シフト履歴はmargin-top:autoで底部に固定しカード間の高さ不揃いを解消。 */
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
  padding-bottom: 2.8vh;
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
  /* CN: 数值下方的设备标签，移到值后确保大数值与其他卡片顶部对齐
     EN: Device sub-label now below the value so big numbers align to the same row across all cards.
     JP: 数値の下に移動したデバイスサブラベル。全カードの大数値が同じ高さに揃う。 */
  margin-top: 2px;
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
  /* CN: 高度缩至 24%，为底部历史班次文字留出不遮挡的空间
     EN: Reduced to 24% height so the decorative sparkline does not obscure shift-history text.
     JP: 高さ24%に縮小し、シフト履歴テキストへの重なりを軽減。 */
  height: 24%;
  opacity: 0.3;
  pointer-events: none;
}

.kpi-device-quality {
  display: flex;
  /* CN: 设备行改为横排，两台设备并列一行，减少卡片内容行数
     EN: Device rows laid out in a single horizontal line to reduce row count and visual crowding.
     JP: デバイス行を横並びにしてカードの行数を減らし、視覚的な詰め込みを解消。 */
  flex-direction: row;
  flex-wrap: wrap;
  gap: 0 0.8vw;
  margin-top: 0.2vh;
  position: relative;
  z-index: 1;
}

.kpi-device-quality-item {
  font-size: clamp(9px, 0.65vw, 11px);
  color: #7b9cc5;
  line-height: 1.3;
}

/* CN: 当班历史班次小字区块，用 margin-top:auto 在 flex 列中始终贴卡片底部
   EN: Past-shift history block; margin-top:auto in flex column pins it to the card bottom.
   JP: シフト履歴ブロック。flex列内でmargin-top:autoにより常にカード底部へ配置。 */
.kpi-shift-history {
  padding: 2px 8px 0;
  border-top: 1px solid rgba(100, 140, 200, 0.15);
  margin-top: auto;
  position: relative;
  z-index: 2;
}

.kpi-shift-row {
  font-size: 10px;
  color: #5b7a9d;
  line-height: 1.35;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
