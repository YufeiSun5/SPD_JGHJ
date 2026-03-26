<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-clipboard-check"></i>
          工单及不良品管理
        </div>
        <div class="page-subtitle">Order & Quality Management</div>
      </div>
      <div class="header-actions">
        <button 
          class="action-btn primary" 
          @click="showOrderModal('add')" 
          title="新增工单"
        >
          <i class="fas fa-plus"></i> 新增工单
        </button>
        <button 
          class="action-btn refresh" 
          @click="loadOrders" 
          title="刷新数据"
        >
          <i class="fas fa-sync-alt"></i> 刷新
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card card-total">
        <div class="stat-icon">
          <i class="fas fa-clipboard-list"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.totalOrders }}</div>
          <div class="stat-label">工单总数</div>
        </div>
      </div>
      <div class="stat-card card-quality">
        <div class="stat-icon">
          <i class="fas fa-check-circle"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.avgQualityRate }}%</div>
          <div class="stat-label">平均良品率</div>
        </div>
      </div>
      <div class="stat-card card-efficiency">
        <div class="stat-icon">
          <i class="fas fa-tachometer-alt"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.avgEfficiency }}%</div>
          <div class="stat-label">平均性能稼动率</div>
        </div>
      </div>
      <div class="stat-card card-production">
        <div class="stat-icon">
          <i class="fas fa-boxes"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.totalProduction }}</div>
          <div class="stat-label">累计产量</div>
        </div>
      </div>
    </div>

    <!-- 筛选与设置栏 -->
    <div class="card">
      <div class="card-header">
        <h3><i class="fas fa-filter"></i> 筛选条件与设置</h3>
        <button 
          class="action-btn-mini secondary" 
          @click="showSettings = !showSettings"
          title="显示/隐藏高级设置"
        >
          <i :class="showSettings ? 'fas fa-eye-slash' : 'fas fa-cog'"></i>
          {{ showSettings ? '隐藏设置' : '高级设置' }}
        </button>
      </div>
      <div class="card-content">
        <!-- 筛选条件 -->
        <div class="filter-row">
          <div class="filter-group">
            <label><i class="fas fa-calendar-alt"></i> 时间范围：</label>
            <div class="custom-select">
              <select v-model="filters.timeRange" @change="applyFilters">
                <option value="all">全部时间</option>
                <option value="today">今天</option>
                <option value="yesterday">昨天</option>
                <option value="week">最近7天</option>
                <option value="month">最近30天</option>
              </select>
            </div>
          </div>
          
          <div class="filter-group">
            <label><i class="fas fa-desktop"></i> 设备筛选：</label>
            <div class="custom-select">
              <select v-model="filters.deviceId" @change="applyFilters">
                <option :value="null">全部设备</option>
                <option v-for="device in devices" :key="device.id" :value="device.id">
                  {{ device.device_name }}
                </option>
              </select>
            </div>
          </div>
          
          <div class="filter-group">
            <label><i class="fas fa-tasks"></i> 工单状态：</label>
            <div class="custom-select">
              <select v-model="filters.status" @change="applyFilters">
                <option :value="null">全部状态</option>
                <option :value="0">待产</option>
                <option :value="1">进行中</option>
                <option :value="2">暂停</option>
                <option :value="3">完工</option>
                <option :value="4">关闭</option>
              </select>
            </div>
          </div>
          
          <button class="action-btn secondary" @click="resetFilters">
            <i class="fas fa-undo"></i> 重置
          </button>
        </div>
        
        <!-- 高级设置（默认隐藏） -->
        <div v-if="showSettings" class="advanced-settings">
          <div class="settings-divider"></div>
          <div class="setting-row">
            <div class="setting-label">
              <i class="fas fa-clock"></i>
              <span>理论节拍（单件加工时间）</span>
            </div>
            <div class="setting-input">
              <input 
                v-model.number="productionCoefficient" 
                type="number" 
                min="0" 
                step="0.1"
                placeholder="例如：10"
              />
              <span class="unit">秒/件</span>
            </div>
            <button class="action-btn-mini secondary" @click="saveCoefficient">
              <i class="fas fa-save"></i> 保存
            </button>
          </div>
          <div class="setting-hint">
            <i class="fas fa-info-circle"></i>
            <span>性能稼动率 = 理论时间 / 实际时间 × 100%，理论时间 = 节拍 × 良品数</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 数据表格卡片 -->
    <div class="card">
      <div class="card-header">
        <h3>
          <i class="fas fa-table"></i> 工单列表
          <span class="badge-count">{{ orders.length }} 条</span>
        </h3>
        <button 
          class="action-btn success"
          @click="exportData" 
          :disabled="orders.length === 0"
        >
          <i class="fas fa-file-excel"></i> 导出CSV
        </button>
      </div>
      
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th width="60">序号</th>
              <th width="140">工单号</th>
              <th width="120">产品编码</th>
              <th width="100">设备</th>
              <th width="80">计划数</th>
              <th width="80">实际数</th>
              <th width="80">良品数</th>
              <th width="80">不良数</th>
              <th width="100">良品率</th>
              <th width="120">性能稼动率</th>
              <th width="100">状态</th>
              <th width="120">实际用时</th>
              <th width="160">开工时间</th>
              <th width="160">完工时间</th>
              <th width="180">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="orders.length === 0">
              <td colspan="15" class="empty-row">
                <i class="fas fa-inbox"></i>
                <p>暂无工单数据</p>
              </td>
            </tr>
            <tr v-for="(order, index) in orders" :key="order.id">
              <td>{{ index + 1 }}</td>
              <td>
                <span class="order-code">{{ order.order_no }}</span>
              </td>
              <td>
                <span class="product-code">{{ order.product_code }}</span>
              </td>
              <td>
                <span class="device-name">{{ getDeviceName(order.target_device_id) }}</span>
              </td>
              <td>{{ order.plan_qty }}</td>
              <td>{{ order.actual_qty }}</td>
              <td class="text-success">{{ order.ok_qty }}</td>
              <td class="text-danger">{{ order.ng_qty }}</td>
              <td>
                <span :class="['quality-badge', getQualityClass(order.quality_rate)]">
                  {{ order.quality_rate }}%
                </span>
              </td>
              <td>
                <span :class="['efficiency-badge', getEfficiencyClass(order.efficiency)]">
                  {{ order.efficiency }}%
                </span>
              </td>
              <td>
                <span :class="['status-badge', getStatusClass(order.status)]">
                  <i :class="getStatusIcon(order.status)"></i>
                  {{ getStatusText(order.status) }}
                </span>
              </td>
              <td>
                <span class="time-badge">{{ formatUsedTime(order) }}</span>
              </td>
              <td>
                <span class="time-text">{{ formatDateTime(order.start_time) }}</span>
              </td>
              <td>
                <span class="time-text">{{ formatDateTime(order.end_time) }}</span>
              </td>
              <td>
                <div class="action-buttons">
                  <button 
                    class="table-btn info" 
                    @click="viewOrder(order)" 
                    title="查看详情"
                  >
                    <i class="fas fa-eye"></i>
                  </button>
                  <button 
                    class="table-btn edit" 
                    @click="showOrderModal('edit', order)" 
                    title="编辑"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button 
                    class="table-btn delete" 
                    @click="deleteOrder(order)" 
                    title="删除"
                  >
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Toast 提示组件 -->
    <Toast
      :show="toast.show"
      :message="toast.message"
      :type="toast.type"
      :duration="toast.duration"
      @update:show="toast.show = $event"
      @close="toast.show = false"
    />

    <!-- 确认对话框组件 -->
    <ConfirmDialog
      v-if="confirmDialog.show"
      :show="true"
      :type="confirmDialog.type"
      :title="confirmDialog.title"
      :message="confirmDialog.message"
      :details="confirmDialog.details"
      :warnings="confirmDialog.warnings"
      :warning-title="confirmDialog.warningTitle"
      :confirm-text="confirmDialog.confirmText"
      :cancel-text="confirmDialog.cancelText"
      :confirm-icon="confirmDialog.confirmIcon"
      @confirm="confirmDialog.onConfirm"
      @cancel="confirmDialog.onCancel"
    />

    <!-- 工单新增/编辑弹窗 -->
    <PlanModal
      v-if="planModal.show"
      :mode="planModal.mode"
      :plan="planModal.form"
      @close="closePlanModal"
      @save="savePlan"
      @delete="confirmDeleteOrder"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Toast from '../components/Toast.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import PlanModal from '../components/PlanModal.vue'

const orders = ref([])
const allOrders = ref([]) // 存储所有工单（未筛选）
const devices = ref([])
const productionCoefficient = ref(10) // 默认10秒/件
const loading = ref(false)
const showSettings = ref(false) // 控制高级设置显示/隐藏

// 筛选条件
const filters = ref({
  timeRange: 'all',
  deviceId: null,
  status: null
})

const toast = ref({
  show: false,
  message: '',
  type: 'success',
  duration: 2000
})

const confirmDialog = ref({
  show: false,
  type: 'info',
  title: '确认操作',
  message: '',
  details: [],
  warnings: [],
  warningTitle: '注意',
  confirmText: '确认',
  cancelText: '取消',
  confirmIcon: 'fas fa-check',
  onConfirm: () => {},
  onCancel: () => {}
})

const planModal = ref({
  show: false,
  mode: 'add',
  form: {}
})

// 统计数据
const stats = computed(() => {
  const totalOrders = orders.value.length
  const totalProduction = orders.value.reduce((sum, o) => sum + (o.ok_qty || 0), 0)
  
  // 计算平均良品率
  const ordersWithProduction = orders.value.filter(o => o.actual_qty > 0)
  const totalQualityRate = ordersWithProduction.reduce((sum, o) => {
    const rate = o.actual_qty > 0 ? (o.ok_qty / o.actual_qty * 100) : 0
    return sum + rate
  }, 0)
  const avgQualityRate = ordersWithProduction.length > 0 
    ? Math.round(totalQualityRate / ordersWithProduction.length) 
    : 0
  
  // 计算平均性能稼动率
  const ordersWithTime = orders.value.filter(o => o.start_time && o.ok_qty > 0)
  const totalEfficiency = ordersWithTime.reduce((sum, o) => {
    const efficiency = calculateEfficiency(o)
    return sum + efficiency
  }, 0)
  const avgEfficiency = ordersWithTime.length > 0 
    ? Math.round(totalEfficiency / ordersWithTime.length) 
    : 0
  
  return {
    totalOrders,
    totalProduction,
    avgQualityRate,
    avgEfficiency
  }
})

// 计算良品率
const calculateQualityRate = (order) => {
  if (order.actual_qty === 0) return 0
  return Math.round((order.ok_qty / order.actual_qty) * 100)
}

// 计算性能稼动率
const calculateEfficiency = (order) => {
  if (order.ok_qty === 0) return 0
  
  // 使用累计用时 + 当前段用时
  let actualSeconds = order.used_seconds || 0
  
  // 如果正在生产中且有当前开始时间，加上当前段用时
  if (order.status === 1 && order.current_start_time) {
    const currentStart = new Date(order.current_start_time)
    const elapsed = Math.floor((Date.now() - currentStart.getTime()) / 1000)
    actualSeconds += elapsed
  }
  
  if (actualSeconds <= 0) return 0
  
  const theoreticalSeconds = order.ok_qty * productionCoefficient.value // 理论时间（秒）
  const efficiency = (theoreticalSeconds / actualSeconds) * 100
  
  return Math.round(efficiency)
}

// 加载工单列表
const loadOrders = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllOrders()
      allOrders.value = (result || []).map(order => ({
        ...order,
        quality_rate: calculateQualityRate(order),
        efficiency: calculateEfficiency(order)
      }))
      console.log('加载工单列表:', allOrders.value.length, '条')
      applyFilters() // 应用筛选
    }
  } catch (e) {
    console.error('加载工单失败:', e)
    showToast('加载工单失败: ' + e, 'error')
  }
}

// 应用筛选
const applyFilters = () => {
  let filtered = [...allOrders.value]
  
  // 时间范围筛选
  if (filters.value.timeRange !== 'all') {
    const now = new Date()
    let startTime = null
    
    switch (filters.value.timeRange) {
      case 'today':
        startTime = new Date(now.getFullYear(), now.getMonth(), now.getDate())
        break
      case 'yesterday':
        startTime = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1)
        const endTime = new Date(now.getFullYear(), now.getMonth(), now.getDate())
        filtered = filtered.filter(order => {
          const createdAt = new Date(order.created_at)
          return createdAt >= startTime && createdAt < endTime
        })
        startTime = null // 已处理完，不需要后续的 >= 判断
        break
      case 'week':
        startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      case 'month':
        startTime = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
        break
    }
    
    if (startTime) {
      filtered = filtered.filter(order => {
        const createdAt = new Date(order.created_at)
        return createdAt >= startTime
      })
    }
  }
  
  // 设备筛选
  if (filters.value.deviceId !== null) {
    filtered = filtered.filter(order => order.target_device_id === filters.value.deviceId)
  }
  
  // 状态筛选
  if (filters.value.status !== null) {
    filtered = filtered.filter(order => order.status === filters.value.status)
  }
  
  orders.value = filtered
  console.log('筛选后工单数:', orders.value.length)
}

// 重置筛选
const resetFilters = () => {
  filters.value = {
    timeRange: 'all',
    deviceId: null,
    status: null
  }
  applyFilters()
  showToast('筛选条件已重置', 'success')
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

// 获取设备名称
const getDeviceName = (deviceId) => {
  if (!deviceId) return '-'
  const device = devices.value.find(d => d.id === deviceId)
  return device ? device.device_name : `设备${deviceId}`
}

// 获取状态文本
const getStatusText = (status) => {
  const map = {
    0: '待产',
    1: '进行',
    2: '暂停',
    3: '完工',
    4: '关闭'
  }
  return map[status] || status
}

// 获取状态样式类
const getStatusClass = (status) => {
  const map = {
    0: 'pending',
    1: 'running',
    2: 'paused',
    3: 'completed',
    4: 'cancelled'
  }
  return map[status] || 'pending'
}

// 获取状态图标
const getStatusIcon = (status) => {
  const map = {
    0: 'fas fa-clock',
    1: 'fas fa-spinner fa-spin',
    2: 'fas fa-pause-circle',
    3: 'fas fa-check-circle',
    4: 'fas fa-times-circle'
  }
  return map[status] || 'fas fa-circle'
}

// 获取良品率样式类
const getQualityClass = (rate) => {
  if (rate >= 95) return 'excellent'
  if (rate >= 90) return 'good'
  if (rate >= 80) return 'normal'
  return 'poor'
}

// 获取效率样式类
const getEfficiencyClass = (efficiency) => {
  if (efficiency >= 90) return 'excellent'
  if (efficiency >= 75) return 'good'
  if (efficiency >= 60) return 'normal'
  return 'poor'
}

// 格式化日期时间
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

// 格式化用时
const formatUsedTime = (order) => {
  let totalSeconds = order.used_seconds || 0
  
  // 如果正在生产中且有开始时间，加上当前段的用时
  if (order.status === 1 && order.current_start_time) {
    const currentStart = new Date(order.current_start_time)
    const elapsed = Math.floor((Date.now() - currentStart.getTime()) / 1000)
    totalSeconds += elapsed
  }
  
  if (totalSeconds === 0) return '-'
  
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  
  if (hours > 0) {
    return `${hours}h${minutes}m`
  } else if (minutes > 0) {
    return `${minutes}m${seconds}s`
  } else {
    return `${seconds}s`
  }
}

// 将秒数转换为分钟（用于详情显示）
const formatSecondsToMinutes = (seconds) => {
  if (!seconds || seconds <= 0) return '-'
  return (seconds / 60).toFixed(2)
}

// 显示提示框
const showToast = (message, type = 'success', duration = 2000) => {
  toast.value.message = message
  toast.value.type = type
  toast.value.duration = duration
  toast.value.show = true
}

// 保存系数设置
const saveCoefficient = async () => {
  if (productionCoefficient.value <= 0) {
    showToast('系数必须大于0', 'warning')
    return
  }
  
  try {
    // 保存到后端配置文件
    if (window.go && window.go.main && window.go.main.App) {
      await window.go.main.App.SetProductionCoefficient(productionCoefficient.value)
    }
    
    // 重新计算所有工单的效率
    allOrders.value = allOrders.value.map(order => ({
      ...order,
      efficiency: calculateEfficiency(order)
    }))
    
    applyFilters() // 重新应用筛选
    showToast('系数设置已保存', 'success')
  } catch (e) {
    console.error('保存系数失败:', e)
    showToast('保存失败: ' + e, 'error')
  }
}

// 查看工单详情
const viewOrder = (order) => {
  // 计算实际用时
  let actualSeconds = order.used_seconds || 0
  if (order.status === 1 && order.current_start_time) {
    const currentStart = new Date(order.current_start_time)
    actualSeconds += Math.floor((Date.now() - currentStart.getTime()) / 1000)
  }
  const actualTime = formatSecondsToMinutes(actualSeconds)
  
  // 计算理论时间
  const theoreticalTime = order.ok_qty > 0 
    ? (order.ok_qty * productionCoefficient.value / 60).toFixed(2) 
    : '-'
  
  confirmDialog.value = {
    show: true,
    type: 'info',
    title: '📋 工单详情',
    message: '查看工单详细信息',
    details: [
      { icon: '📋', label: '工单号', value: order.order_no },
      { icon: '📦', label: '产品编码', value: order.product_code },
      { icon: '🏭', label: '目标设备', value: getDeviceName(order.target_device_id) },
      { icon: '🎯', label: '计划数量', value: order.plan_qty },
      { icon: '📊', label: '实际产量', value: order.actual_qty },
      { icon: '✅', label: '良品数', value: order.ok_qty },
      { icon: '❌', label: '不良品数', value: order.ng_qty },
      { icon: '📈', label: '良品率', value: `${order.quality_rate}%` },
      { icon: '⏱️', label: '实际用时', value: `${actualTime} 分钟` },
      { icon: '⏰', label: '理论时间', value: `${theoreticalTime} 分钟` },
      { icon: '🚀', label: '性能稼动率', value: `${order.efficiency}%` },
      { icon: '🔖', label: '状态', value: getStatusText(order.status) }
    ],
    warnings: [],
    confirmText: '关闭',
    cancelText: '',
    confirmIcon: 'fas fa-times',
    onConfirm: () => {
      confirmDialog.value.show = false
    },
    onCancel: () => {
      confirmDialog.value.show = false
    }
  }
}

// 显示工单弹窗
const showOrderModal = (mode, order = null) => {
  planModal.value.mode = mode
  planModal.value.form = order ? { ...order } : {
    order_no: '',
    product_code: '',
    target_device_id: null,
    plan_qty: 0,
    status: 0
  }
  planModal.value.show = true
}

// 关闭工单弹窗
const closePlanModal = () => {
  planModal.value.show = false
}

// 保存工单
const savePlan = async (plan) => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      if (planModal.value.mode === 'add') {
        // 新增工单
        await window.go.main.App.CreateOrder(
          plan.order_no,
          plan.product_code,
          plan.plan_qty,
          plan.target_device_id || null
        )
        showToast('工单创建成功', 'success')
      } else {
        // 编辑工单
        await window.go.main.App.UpdateOrder(
          plan.id,
          plan.product_code || null,
          plan.plan_qty || null,
          plan.status !== undefined ? plan.status : null,
          plan.target_device_id !== undefined ? plan.target_device_id : null
        )
        showToast('工单更新成功', 'success')
      }
      closePlanModal()
      await loadOrders()
    }
  } catch (e) {
    console.error('保存工单失败:', e)
    showToast('保存失败: ' + e, 'error')
  }
}

// 确认删除工单（从 PlanModal 触发）
const confirmDeleteOrder = (plan) => {
  closePlanModal()
  deleteOrder(plan)
}

// 删除工单
const deleteOrder = (order) => {
  confirmDialog.value = {
    show: true,
    type: 'danger',
    title: '⚠️ 确认删除',
    message: '确认要删除这个工单吗？',
    details: [
      { icon: '📋', label: '工单号', value: order.order_no },
      { icon: '📦', label: '产品编码', value: order.product_code },
      { icon: '✅', label: '良品数', value: order.ok_qty }
    ],
    warnings: [
      '删除后无法恢复',
      '所有生产数据将被清除'
    ],
    warningTitle: '⚠️ 警告',
    confirmText: '确认删除',
    cancelText: '取消',
    confirmIcon: 'fas fa-trash',
    onConfirm: async () => {
      confirmDialog.value.show = false
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.DeleteOrder(order.id)
          showToast('工单已删除', 'success')
          await loadOrders()
        }
      } catch (e) {
        console.error('删除工单失败:', e)
        showToast('删除失败: ' + e, 'error')
      }
    },
    onCancel: () => {
      confirmDialog.value.show = false
    }
  }
}

// 导出CSV
const exportData = () => {
  if (orders.value.length === 0) {
    showToast('没有数据可导出', 'warning')
    return
  }
  
  try {
    // 构建CSV内容
    let csv = '\uFEFF' // UTF-8 BOM
    
    // 表头
    const headers = [
      '序号', '工单号', '产品编码', '设备', '计划数', '实际数', '良品数', '不良数',
      '良品率(%)', '性能稼动率(%)', '状态', '实际用时(分钟)', '开工时间', '完工时间', '理论时间(分钟)'
    ]
    csv += headers.join(',') + '\n'
    
    // 数据行
    orders.value.forEach((order, index) => {
      // 计算实际用时
      let actualSeconds = order.used_seconds || 0
      if (order.status === 1 && order.current_start_time) {
        const currentStart = new Date(order.current_start_time)
        actualSeconds += Math.floor((Date.now() - currentStart.getTime()) / 1000)
      }
      const actualTime = formatSecondsToMinutes(actualSeconds)
      
      const theoreticalTime = order.ok_qty > 0 
        ? (order.ok_qty * productionCoefficient.value / 60).toFixed(2) 
        : '-'
      
      const values = [
        index + 1,
        order.order_no,
        order.product_code,
        getDeviceName(order.target_device_id),
        order.plan_qty,
        order.actual_qty,
        order.ok_qty,
        order.ng_qty,
        order.quality_rate,
        order.efficiency,
        getStatusText(order.status),
        actualTime,
        formatDateTime(order.start_time),
        formatDateTime(order.end_time),
        theoreticalTime
      ]
      csv += values.join(',') + '\n'
    })
    
    // 下载文件
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5)
    link.download = `工单及不良品数据_${timestamp}.csv`
    link.click()
    
    showToast(`导出成功，共 ${orders.value.length} 条记录`, 'success')
  } catch (error) {
    console.error('导出失败:', error)
    showToast('导出失败: ' + error, 'error')
  }
}

// 加载理论节拍系数
const loadProductionCoefficient = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const coefficient = await window.go.main.App.GetProductionCoefficient()
      productionCoefficient.value = coefficient
      console.log('加载理论节拍系数:', coefficient)
    }
  } catch (e) {
    console.error('加载系数失败:', e)
    // 使用默认值
    productionCoefficient.value = 10.0
  }
}

onMounted(async () => {
  // 加载理论节拍系数
  await loadProductionCoefficient()
  
  await loadDevices()
  await loadOrders()
})
</script>

<style scoped>
/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

/* 统计卡片 */
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

/* 工业低调配色 */
.stat-card.card-total {
  background: linear-gradient(135deg, #556070 0%, #6e7e8e 100%);
}

.stat-card.card-quality {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

.stat-card.card-efficiency {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
}

.stat-card.card-production {
  background: linear-gradient(135deg, #5a7080 0%, #6e8e9e 100%);
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
  background: rgba(20, 30, 48, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  overflow: hidden;
  margin-bottom: 24px;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
}

.card-header h3 i {
  color: #546e7a;
}

.action-btn-mini {
  padding: 6px 12px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
}

.action-btn-mini.secondary {
  background: rgba(84, 110, 122, 0.2);
  border: 1px solid rgba(84, 110, 122, 0.3);
  color: rgba(255,255,255,0.8);
}

.action-btn-mini.secondary:hover {
  background: rgba(84, 110, 122, 0.4);
  border-color: rgba(84, 110, 122, 0.5);
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(84, 110, 122, 0.3);
}

.badge-count {
  background: rgba(84, 110, 122, 0.3);
  color: #fff;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
  margin-left: 8px;
}

.card-content {
  padding: 24px;
}

/* 筛选行 */
.filter-row {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-group label {
  font-size: 14px;
  color: rgba(255,255,255,0.8);
  font-weight: 500;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 6px;
}

.filter-group label i {
  color: #546e7a;
}

.custom-select {
  position: relative;
  display: inline-block;
  min-width: 140px;
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
}

.custom-select select {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  width: 100%;
  padding: 8px 32px 8px 16px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.custom-select select:hover {
  border-color: rgba(84, 110, 122, 0.5);
  background: rgba(255,255,255,0.08);
}

.custom-select select:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}

.custom-select select option {
  background: #1a1f3a;
  color: #fff;
  padding: 10px;
}

/* 高级设置区域 */
.advanced-settings {
  margin-top: 20px;
  padding-top: 20px;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.settings-divider {
  height: 1px;
  background: linear-gradient(90deg, 
    transparent 0%, 
    rgba(84, 110, 122, 0.3) 20%, 
    rgba(84, 110, 122, 0.3) 80%, 
    transparent 100%
  );
  margin-bottom: 20px;
}

/* 设置行 */
.setting-row {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 16px;
}

.setting-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  font-weight: 600;
  min-width: 200px;
}

.setting-label i {
  color: #546e7a;
  font-size: 16px;
}

.setting-input {
  display: flex;
  align-items: center;
  gap: 8px;
}

.setting-input input {
  width: 150px;
  padding: 8px 16px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  transition: all 0.2s;
}

.setting-input input:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}

.setting-input .unit {
  color: rgba(255,255,255,0.6);
  font-size: 13px;
}

.setting-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
  padding: 12px 16px;
  background: rgba(84, 110, 122, 0.1);
  border-radius: 8px;
  border-left: 3px solid #546e7a;
}

.setting-hint i {
  color: #546e7a;
}

/* 表格容器 */
.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  min-width: 1400px;
}

.data-table thead {
  background: rgba(84, 110, 122, 0.1);
}

.data-table th {
  padding: 14px 12px;
  text-align: center;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  border-bottom: 2px solid rgba(84, 110, 122, 0.3);
  border-right: 1px solid rgba(255,255,255,0.08);
  white-space: nowrap;
  font-size: 13px;
}

.data-table th:last-child {
  border-right: none;
}

.data-table td {
  padding: 14px 12px;
  text-align: left;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  border-right: 1px solid rgba(255,255,255,0.08);
  color: rgba(255,255,255,0.8);
}

.data-table td:last-child {
  border-right: none;
}

.data-table tbody tr {
  transition: all 0.2s;
}

.data-table tbody tr:hover {
  background: rgba(84, 110, 122, 0.08);
}

.empty-row {
  text-align: center;
  color: rgba(255,255,255,0.5);
  padding: 60px 20px !important;
}

.empty-row i {
  font-size: 48px;
  display: block;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-row p {
  font-size: 16px;
  margin: 0;
}

/* 文本对齐 */
.text-center {
  text-align: center;
}

.text-right {
  text-align: right;
}

.text-success {
  color: #2ecc71 !important;
  font-weight: 600;
}

.text-danger {
  color: #e74c3c !important;
  font-weight: 600;
}

/* 表格元素样式 */
.order-code {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: #667eea;
}

.product-code {
  font-weight: 500;
  color: rgba(255,255,255,0.9);
}

.device-name {
  font-size: 12px;
  color: rgba(255,255,255,0.7);
}

.time-text {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  color: rgba(255,255,255,0.7);
}

.time-badge {
  display: inline-block;
  padding: 4px 12px;
  background: rgba(52, 73, 94, 0.2);
  border: 1px solid rgba(52, 73, 94, 0.4);
  border-radius: 12px;
  font-size: 12px;
  color: #ecf0f1;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

/* 良品率徽章 */
.quality-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.quality-badge.excellent {
  background: rgba(46, 204, 113, 0.2);
  color: #2ecc71;
  border: 1px solid #2ecc71;
}

.quality-badge.good {
  background: rgba(52, 152, 219, 0.2);
  color: #3498db;
  border: 1px solid #3498db;
}

.quality-badge.normal {
  background: rgba(243, 156, 18, 0.2);
  color: #f39c12;
  border: 1px solid #f39c12;
}

.quality-badge.poor {
  background: rgba(231, 76, 60, 0.2);
  color: #e74c3c;
  border: 1px solid #e74c3c;
}

/* 效率徽章 */
.efficiency-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.efficiency-badge.excellent {
  background: rgba(46, 204, 113, 0.2);
  color: #2ecc71;
  border: 1px solid #2ecc71;
}

.efficiency-badge.good {
  background: rgba(52, 152, 219, 0.2);
  color: #3498db;
  border: 1px solid #3498db;
}

.efficiency-badge.normal {
  background: rgba(243, 156, 18, 0.2);
  color: #f39c12;
  border: 1px solid #f39c12;
}

.efficiency-badge.poor {
  background: rgba(231, 76, 60, 0.2);
  color: #e74c3c;
  border: 1px solid #e74c3c;
}

/* 状态徽章 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.status-badge.pending {
  background: rgba(52, 152, 219, 0.2);
  color: #3498db;
  border: 1px solid #3498db;
}

.status-badge.running {
  background: rgba(46, 204, 113, 0.2);
  color: #2ecc71;
  border: 1px solid #2ecc71;
}

.status-badge.paused {
  background: rgba(241, 196, 15, 0.2);
  color: #f1c40f;
  border: 1px solid #f1c40f;
}

.status-badge.completed {
  background: rgba(155, 89, 182, 0.2);
  color: #9b59b6;
  border: 1px solid #9b59b6;
}

.status-badge.cancelled {
  background: rgba(231, 76, 60, 0.2);
  color: #e74c3c;
  border: 1px solid #e74c3c;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.table-btn {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  background: rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.8);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.table-btn:hover {
  transform: scale(1.05);
}

.table-btn.info:hover {
  background: #3498db;
  color: #fff;
}

.table-btn.edit:hover {
  background: #667eea;
  color: #fff;
}

.table-btn.delete:hover {
  background: #e74c3c;
  color: #fff;
}
</style>
