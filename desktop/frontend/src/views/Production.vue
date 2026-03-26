<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-clipboard-list"></i>
          生产计划管理
        </div>
        <div class="page-subtitle">Production Planning & Management</div>
      </div>
      <div class="header-actions">
        <button class="action-btn primary" @click="showPlanModal('add')" title="新增计划">
          <i class="fas fa-plus"></i> 新增计划
        </button>
        <button class="action-btn warning" @click="showQuickShift" title="快捷换班">
          <i class="fas fa-exchange-alt"></i> 快捷换班
        </button>
        <button class="action-btn danger" @click="showActiveDevices" title="结束班次">
          <i class="fas fa-power-off"></i> 结束班次
        </button>
      </div>
    </div>

    <!-- 顶部统计卡片 - 左右布局 -->
    <div class="stats-row">
      <div class="stat-card card-total">
        <div class="stat-icon">
          <i class="fas fa-clipboard-list"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.total }}</div>
          <div class="stat-label">项目中计划</div>
        </div>
      </div>
      <div class="stat-card card-completed">
        <div class="stat-icon">
          <i class="fas fa-check-circle"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.completed }}</div>
          <div class="stat-label">已完成计划</div>
        </div>
      </div>
      <!-- <div class="stat-card card-workers">
        <div class="stat-icon">
          <i class="fas fa-boxes"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.totalQty }}</div>
          <div class="stat-label">累计数量</div>
        </div>
      </div>
      <div class="stat-card card-efficiency">
        <div class="stat-icon">
          <i class="fas fa-chart-line"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.avgRate }}%</div>
          <div class="stat-label">平均完成率</div>
        </div>
      </div> -->
      
      <!-- 一号机操作卡片 -->
      <div class="stat-card card-device-control device-1">
        <div class="device-info">
          <div class="device-icon">
            <i class="fas fa-industry"></i>
          </div>
          <div class="device-title">一号机</div>
        </div>
        <div class="device-controls">
          <button class="device-btn ng-plus" @click="triggerDeviceTask(25)" title="NG数+1">
            <div class="btn-icon">
              <i class="fas fa-plus"></i>
            </div>
            <div class="btn-label">NG+1</div>
          </button>
          <button class="device-btn ng-minus" @click="triggerDeviceTask(32)" title="NG数-1">
            <div class="btn-icon">
              <i class="fas fa-minus"></i>
            </div>
            <div class="btn-label">NG-1</div>
          </button>
        </div>
      </div>
      
      <!-- 二号机操作卡片 -->
      <div class="stat-card card-device-control device-2">
        <div class="device-info">
          <div class="device-icon">
            <i class="fas fa-industry"></i>
          </div>
          <div class="device-title">二号机</div>
        </div>
        <div class="device-controls">
          <button class="device-btn ng-plus" @click="triggerDeviceTask(31)" title="NG数+1">
            <div class="btn-icon">
              <i class="fas fa-plus"></i>
            </div>
            <div class="btn-label">NG+1</div>
          </button>
          <button class="device-btn ng-minus" @click="triggerDeviceTask(33)" title="NG数-1">
            <div class="btn-icon">
              <i class="fas fa-minus"></i>
            </div>
            <div class="btn-label">NG-1</div>
          </button>
        </div>
      </div>
    </div>

    <!-- 主要内容区域 - 4个模块 -->
    <div class="content-grid">
      <!-- 左上：人员信息 -->
      <div class="card module-card">
        <div class="card-header">
          <h3><i class="fas fa-users"></i> 人员信息</h3>
          <div class="header-actions-mini">
            <button class="btn-quick-shift" @click="showQuickShift" title="快捷换班">
              <i class="fas fa-exchange-alt"></i>
            </button>
            <button class="btn-active-devices" @click="showActiveDevices" title="结束班次">
              <i class="fas fa-power-off"></i>
            </button>
            <button class="btn-refresh" @click="loadAllStaff" title="刷新">
              <i class="fas fa-sync-alt"></i>
            </button>
          </div>
        </div>
        <div class="staff-container">
          <!-- 在班人员 -->
          <div v-if="onDutyStaffList.length > 0" class="staff-section">
            <div class="section-title">
              <i class="fas fa-user-check"></i>
              在班人员 ({{ onDutyStaffList.length }})
            </div>
            <div class="staff-grid">
              <div 
                v-for="staff in onDutyStaffList" 
                :key="staff.id" 
                class="staff-card on-duty"
              >
                <div class="staff-avatar online">
                  <i class="fas fa-user-circle"></i>
                </div>
                <div class="staff-info">
                  <div class="staff-name-line">{{ staff.name }}</div>
                  <div class="staff-code-line">{{ staff.staff_code }}</div>
                  <div class="staff-team-line">
                    <i class="fas fa-users"></i>
                    {{ staff.team_name }}
                  </div>
                </div>
                <div class="staff-status online">
                  <div class="status-dot"></div>
                  <span>在班</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 不在班人员 -->
          <div v-if="offDutyStaffList.length > 0" class="staff-section">
            <div class="section-title">
              <i class="fas fa-user-clock"></i>
              不在班人员 ({{ offDutyStaffList.length }})
            </div>
            <div class="staff-grid">
              <div 
                v-for="staff in offDutyStaffList" 
                :key="staff.id" 
                class="staff-card off-duty"
              >
                <div class="staff-avatar offline">
                  <i class="fas fa-user-circle"></i>
                </div>
                <div class="staff-info">
                  <div class="staff-name-line">{{ staff.name }}</div>
                  <div class="staff-code-line">{{ staff.staff_code }}</div>
                  <div class="staff-team-line">
                    <i class="fas fa-users"></i>
                    {{ staff.current_team?.team_name || '未分配' }}
                  </div>
                </div>
                <div class="staff-status offline">
                  <div class="status-dot"></div>
                  <span>离班</span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="onDutyStaffList.length === 0 && offDutyStaffList.length === 0" class="empty-staff">
            <i class="fas fa-user-slash"></i>
            <p>暂无员工数据</p>
          </div>
        </div>
      </div>

      <!-- 右上：基本设置 -->
      <div class="card module-card">
        <div class="card-header">
          <h3><i class="fas fa-cog"></i> 基本设置</h3>
          <button class="btn-refresh" @click="loadTeams" title="刷新">
            <i class="fas fa-sync-alt"></i>
          </button>
        </div>
        <div class="settings-content">
          <!-- 班制信息 -->
          <div class="setting-section">
            <div class="setting-label">
              <i class="fas fa-users-cog"></i> 班制信息
            </div>
            <div class="table-container-mini">
              <table class="data-table mini">
                <thead>
                  <tr>
                    <th>#</th>
                    <th>班组名称</th>
                    <th>开始</th>
                    <th>结束</th>
                    <th>时长</th>
                    <th>状态</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="shiftInfoData.length === 0">
                    <td colspan="7" class="empty-mini">暂无班制数据</td>
                  </tr>
                  <tr v-for="shift in shiftInfoData" :key="shift.id">
                    <td>{{ shift.index }}</td>
                    <td>
                      <span class="team-name">{{ shift.teamName }}</span>
                    </td>
                    <td>{{ shift.startTime }}</td>
                    <td>{{ shift.endTime }}</td>
                    <td>{{ shift.duration }}</td>
                    <td>
                      <span :class="['mini-status-badge', shift.isActive ? 'active' : 'inactive']">
                        <i :class="shift.isActive ? 'fas fa-check-circle' : 'fas fa-times-circle'"></i>
                      </span>
                    </td>
                    <td>
                      <button class="table-btn info mini" @click="goToStaffPage" title="查看">
                        <i class="fas fa-external-link-alt"></i>
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 能耗信息 -->
          <div class="setting-section">
            <div class="setting-label">
              <i class="fas fa-bolt"></i> 能耗信息
            </div>
            <div class="table-container-mini">
              <table class="data-table mini">
                <thead>
                  <tr>
                    <th>#</th>
                    <th>设备名称</th>
                    <th>当前功率</th>
                    <th>今日总用能</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="energyData.length === 0">
                    <td colspan="5" class="empty-mini">暂无能耗数据</td>
                  </tr>
                  <tr v-for="device in energyData" :key="device.id">
                    <td>{{ device.id }}</td>
                    <td>
                      <span class="device-name">{{ device.deviceName }}</span>
                    </td>
                    <td>
                      <span class="power-value">{{ device.currentPower }}</span>
                    </td>
                    <td>
                      <span class="consumption-value">{{ device.totalConsumption }}</span>
                    </td>
                    <td>
                      <button class="table-btn info mini" @click="goToDevicePage" title="查看">
                        <i class="fas fa-external-link-alt"></i>
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- 左下：计划列表配置 -->
      <div class="card module-card full-width">
        <div class="card-header">
          <h3><i class="fas fa-calendar-alt"></i> 计划列表配置</h3>
          <button class="btn-refresh" @click="loadPlans" title="刷新">
            <i class="fas fa-sync-alt"></i>
          </button>
        </div>
        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>序号</th>
                <th>工单号</th>
                <th>项目名称</th>
                <th>状态</th>
                <th>设备</th>
                <th>用时</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="filteredPlans.length === 0">
                <td colspan="7" class="empty-row">
                  <i class="fas fa-inbox"></i>
                  <p>暂无计划数据</p>
                </td>
              </tr>
              <tr v-for="(plan, index) in filteredPlans.slice(0, 3)" :key="plan.id">
                <td>{{ index + 1 }}</td>
                <td>
                  <span class="plan-code">{{ plan.order_no }}</span>
                </td>
                <td>
                  <span class="project-name">{{ plan.product_code }}</span>
                </td>
                <td>
                  <span :class="['status-badge', getStatusClass(plan.status)]">
                    <i :class="getStatusIcon(plan.status)"></i>
                    {{ getStatusText(plan.status) }}
                  </span>
                </td>
                <td>
                  <span class="device-tag">{{ getDeviceName(plan.target_device_id) }}</span>
                </td>
                <td>
                  <span class="time-badge">{{ formatUsedTime(plan) }}</span>
                </td>
                <td>
                  <div class="action-buttons">
                    <!-- 待产状态：显示"开工"按钮 -->
                    <button 
                      v-if="plan.status === 0" 
                      class="table-btn success start-btn" 
                      @click="startProduction(plan)" 
                      title="开工生产"
                    >
                      <i class="fas fa-play"></i>
                    </button>
                    
                    <!-- 暂停状态：显示"继续"按钮 -->
                    <button 
                      v-else-if="plan.status === 2" 
                      class="table-btn warning resume-btn" 
                      @click="startProduction(plan)" 
                      title="继续生产"
                    >
                      <i class="fas fa-play"></i>
                    </button>
                    
                    <!-- 生产中/暂停状态：良品数>=计划数时显示"结束"按钮 -->
                    <button 
                      v-else-if="(plan.status === 1 || plan.status === 2) && (plan.ok_qty >= plan.plan_qty)" 
                      class="table-btn complete complete-btn" 
                      @click="completeProduction(plan)" 
                      title="完成生产"
                    >
                      <i class="fas fa-check"></i>
                    </button>
                    
                    <!-- 生产中状态：显示"暂停"按钮 -->
                    <button 
                      v-else-if="plan.status === 1" 
                      class="table-btn warning pause-btn" 
                      @click="pauseProduction(plan)" 
                      title="暂停生产"
                    >
                      <i class="fas fa-pause"></i>
                    </button>
                    
                    <!-- 其他状态：显示"查看"按钮 -->
                    <button 
                      v-else
                      class="table-btn info view-btn" 
                      @click="viewPlan(plan)" 
                      title="查看详情"
                    >
                      <i class="fas fa-eye"></i>
                    </button>
                    
                    <button class="table-btn edit" @click="showPlanModal('edit', plan)" title="编辑">
                      <i class="fas fa-edit"></i>
                    </button>
                    
                    <!-- 强制结束按钮 - 仅在生产中或暂停状态显示 -->
                    <button 
                      v-if="plan.status === 1 || plan.status === 2"
                      class="table-btn danger force-end-btn" 
                      @click="forceCompleteProduction(plan)" 
                      title="强制结束"
                    >
                      <i class="fas fa-flag-checkered"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 右下：任务单配置 -->
      <div class="card module-card">
        <div class="card-header">
          <h3><i class="fas fa-tasks"></i> 任务单配置</h3>
          <button class="btn-refresh" @click="loadPlans" title="刷新">
            <i class="fas fa-sync-alt"></i>
          </button>
        </div>
        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>序号</th>
                <th>编号</th>
                <th>计划</th>
                <th>实际</th>
                <th>不良</th>
                <th>完成率</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="tasks.length === 0">
                <td colspan="7" class="empty-row">
                  <i class="fas fa-inbox"></i>
                  <p>暂无任务数据</p>
                </td>
              </tr>
              <tr v-for="(task, index) in tasks.slice(0, 3)" :key="task.id">
                <td>{{ index + 1 }}</td>
                <td>
                  <span class="task-code">{{ task.order_no }}</span>
                </td>
                <td>
                  <span class="qty-badge">{{ task.plan_qty }}</span>
                </td>
                <td>
                  <span class="qty-badge ok">{{ task.actual_qty }}</span>
                </td>
                <td>
                  <span class="qty-badge ng">{{ task.ng_qty }}</span>
                </td>
                <td>
                  <div class="progress-wrapper">
                    <div class="progress-bar-mini">
                      <div class="progress-fill-mini" :style="{ width: task.progress + '%' }"></div>
                    </div>
                    <span class="progress-text-mini">{{ task.progress }}%</span>
                  </div>
                </td>
                <td>
                  <div class="action-buttons">
                    <button class="table-btn info" @click="viewTask(task)" title="查看">
                      <i class="fas fa-eye"></i>
                    </button>
                    <button class="table-btn edit" @click="editTask(task)" title="编辑">
                      <i class="fas fa-edit"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 新增/编辑计划弹窗 -->
    <PlanModal
      v-if="planModal.show"
      :mode="planModal.mode"
      :plan="planModal.form"
      @close="closePlanModal"
      @save="savePlan"
      @delete="confirmDelete"
    />

    <!-- 快捷换班弹窗 -->
    <ShiftModal
      v-if="quickShiftModal.show"
      :teams="teams"
      :active-sessions="activeSessions"
      @close="closeQuickShift"
      @shift="handleQuickShift"
    />

    <!-- 在线设备弹窗 -->
    <ActiveDevicesModal
      v-if="activeDevicesModal.show"
      :sessions="activeSessions"
      @close="closeActiveDevices"
      @logout="handleLogout"
    />

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
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import PlanModal from '../components/PlanModal.vue'
import ShiftModal from '../components/ShiftModal.vue'
import ActiveDevicesModal from '../components/ActiveDevicesModal.vue'
import Toast from '../components/Toast.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'

const router = useRouter()
const plans = ref([])
const devices = ref([])
const teams = ref([])
const activeSessions = ref([])

const planModal = ref({
  show: false,
  mode: 'add',
  form: {}
})

const quickShiftModal = ref({
  show: false
})

const activeDevicesModal = ref({
  show: false
})

const toast = ref({
  show: false,
  message: '',
  type: 'success', // success, error, warning, info
  duration: 2000
})

// 连续点击计数器
const clickCounter = ref({
  taskId: null,      // 当前点击的任务ID
  count: 0,          // 累计点击次数
  timer: null        // 重置计时器
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

// 所有员工列表
const allStaff = ref([])
// 在班人员列表（从活动班次中提取）
const onDutyStaffIds = ref(new Set())

// 在班人员
const onDutyStaffList = computed(() => {
  return allStaff.value.filter(staff => onDutyStaffIds.value.has(staff.id))
})

// 不在班人员
const offDutyStaffList = computed(() => {
  return allStaff.value.filter(staff => !onDutyStaffIds.value.has(staff.id))
})

// 班制信息数据（基于真实的班组和活动班次，使用真实时间）
const shiftInfoData = computed(() => {
  return teams.value.map((team, index) => {
    // 查找该班组的活动班次
    const activeSession = activeSessions.value.find(s => s.team_id === team.id)
    
    let startTime = '-'
    let endTime = '-'
    let duration = '-'
    
    if (activeSession && activeSession.login_time) {
      // 解析登录时间
      const loginDate = new Date(activeSession.login_time)
      startTime = loginDate.toLocaleTimeString('zh-CN', { 
        hour: '2-digit', 
        minute: '2-digit',
        hour12: false 
      })
      
      // 如果有登出时间，计算时长
      if (activeSession.logout_time) {
        const logoutDate = new Date(activeSession.logout_time)
        endTime = logoutDate.toLocaleTimeString('zh-CN', { 
          hour: '2-digit', 
          minute: '2-digit',
          hour12: false 
        })
        const durationMin = activeSession.duration_min || 0
        duration = `${Math.floor(durationMin / 60)}h${durationMin % 60}m`
      } else {
        endTime = '进行中'
        // 计算当前已工作时长
        const now = new Date()
        const diffMs = now - loginDate
        const diffMin = Math.floor(diffMs / 60000)
        duration = `${Math.floor(diffMin / 60)}h${diffMin % 60}m`
      }
    }
    
    return {
      id: team.id,
      index: index + 1,
      teamName: team.team_name,
      startTime,
      endTime,
      duration,
      isActive: !!activeSession,
      sessionId: activeSession?.id
    }
  })
})

// 真实能耗数据
const deviceEnergyData = ref([])

// 加载设备能耗数据
const loadDeviceEnergyData = async () => {
  try {
    if (window.go?.main?.App) {
      const data = await window.go.main.App.GetAllDevicesEnergyData()
      deviceEnergyData.value = data || []
    }
  } catch (e) {
    console.error('❌ 加载能耗数据失败:', e)
  }
}

// 能耗信息数据（使用真实数据）
const energyData = computed(() => {
  if (deviceEnergyData.value.length > 0) {
    return deviceEnergyData.value.map(device => ({
      id: device.device_id,
      deviceName: device.device_name,
      deviceCode: `#${device.device_id}`,
      currentPower: `${device.real_time_power.toFixed(1)} ${device.power_unit}`,
      totalConsumption: `${device.today_consumption.toFixed(1)} ${device.energy_unit}`,
      efficiency: `${(75 + Math.random() * 20).toFixed(0)}%`
    }))
  }
  
  // 兜底：使用设备列表
  return devices.value.slice(0, 2).map((device, index) => ({
    id: device.id,
    deviceName: device.device_name,
    deviceCode: device.device_code,
    currentPower: `0.0 kW`,
    totalConsumption: `0.0 kWh`,
    efficiency: `-`
  }))
})

// 过滤工单：只显示未完工和完工一天内的
const filteredPlans = computed(() => {
  const now = new Date()
  const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000) // 24小时前
  
  return plans.value.filter(order => {
    // 状态 0=待产, 1=生产中, 2=暂停, 3=完工, 4=关闭
    
    // 未完工的（待产、生产中、暂停）直接显示
    if (order.status === 0 || order.status === 1 || order.status === 2) {
      return true
    }
    
    // 完工的，检查完工时间（使用 end_time 字段）
    if (order.status === 3) {
      // 如果有 end_time 字段，使用它判断
      if (order.end_time) {
        const endTime = new Date(order.end_time)
        return endTime >= oneDayAgo
      }
      // 如果没有 end_time，使用 created_at 判断（兜底逻辑）
      if (order.created_at) {
        const createdTime = new Date(order.created_at)
        return createdTime >= oneDayAgo
      }
      // 如果都没有时间字段，默认不显示完工的
      return false
    }
    
    // 关闭状态的不显示
    return false
  })
})

// 任务数据（基于过滤后的工单）
const tasks = computed(() => {
  return filteredPlans.value.map(order => ({
    id: order.id,
    order_no: order.order_no,
    product_code: order.product_code,
    target_device_id: order.target_device_id,
    plan_qty: order.plan_qty,
    actual_qty: order.actual_qty || 0,
    ng_qty: order.ng_qty || 0,
    status: order.status,
    progress: order.plan_qty > 0 ? Math.round((order.actual_qty || 0) / order.plan_qty * 100) : 0
  }))
})

// （已删除所有 Naive UI 表格列配置，全部改用原生表格）

// 统计数据（基于过滤后的工单）
const stats = computed(() => {
  const total = filteredPlans.value.length
  const completed = filteredPlans.value.filter(p => p.status === 3).length
  // const totalQty = filteredPlans.value.reduce((sum, p) => sum + (p.actual_qty || 0), 0)
  
  // let totalRate = 0
  // let validCount = 0
  // filteredPlans.value.forEach(p => {
  //   if (p.plan_qty > 0) {
  //     totalRate += (p.actual_qty || 0) / p.plan_qty * 100
  //     validCount++
  //   }
  // })
  // const avgRate = validCount > 0 ? Math.round(totalRate / validCount) : 0
  
  return {
    total,
    completed,
    // totalQty,
    // avgRate
  }
})

// 触发设备任务（通过API手动触发）
const triggerDeviceTask = async (taskId) => {
  try {
    if (!window.go?.main?.App) {
      showToast('系统未就绪', 'error')
      return
    }
    
    // 获取设备ID和操作类型
    const taskConfig = {
      25: { deviceId: 1, operation: 'NG+1', deviceName: '一号机', type: 'error' },    // 红色
      31: { deviceId: 2, operation: 'NG+1', deviceName: '二号机', type: 'error' },    // 红色
      32: { deviceId: 1, operation: 'NG-1', deviceName: '一号机', type: 'success' },  // 绿色
      33: { deviceId: 2, operation: 'NG-1', deviceName: '二号机', type: 'success' }   // 绿色
    }
    
    const config = taskConfig[taskId]
    if (!config) {
      showToast('未知的任务ID', 'error')
      return
    }
    
    // 处理连续点击计数
    if (clickCounter.value.taskId === taskId) {
      // 同一个按钮，累加计数
      clickCounter.value.count++
    } else {
      // 不同按钮，重置计数
      clickCounter.value.taskId = taskId
      clickCounter.value.count = 1
    }
    
    // 清除之前的重置计时器
    if (clickCounter.value.timer) {
      clearTimeout(clickCounter.value.timer)
    }
    
    // 2秒后重置计数器
    clickCounter.value.timer = setTimeout(() => {
      clickCounter.value.taskId = null
      clickCounter.value.count = 0
      clickCounter.value.timer = null
    }, 2000)
    
    // 查找该设备当前生产中的工单
    const activeOrder = filteredPlans.value.find(
      order => order.target_device_id === config.deviceId && order.status === 1
    )
    
    // 构建显示的操作文本（带累计数量）
    const operationText = clickCounter.value.count > 1 
      ? `${config.operation.replace('1', clickCounter.value.count)}` 
      : config.operation
    
    let message = ''
    if (activeOrder) {
      // 有生产中的工单：几号机 什么工单 怎么样了
      message = `${config.deviceName} 工单【${activeOrder.order_no}】${operationText}`
    } else {
      // 没有生产中的工单
      message = `${config.deviceName} 当前无生产工单 ${operationText}`
    }
    
    // 调用后端API手动触发任务
    await window.go.main.App.TriggerTaskManually(taskId)
    
    // NG+1 显示红色（error），NG-1 显示绿色（success）
    showToast(message, config.type, 2500)
    
    // 刷新工单数据
    setTimeout(() => {
      loadPlans()
    }, 500)
    
  } catch (e) {
    console.error('❌ 触发任务失败:', e)
    showToast('操作失败: ' + e, 'error')
  }
}

// （已删除员工表格列配置，改用卡片展示）


// 验证工单数据完整性
const validatePlanData = (plan) => {
  if (!plan) return false
  if (!plan.id || plan.id <= 0) return false
  if (!plan.order_no) return false
  return true
}

// 加载工单列表
const loadPlans = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const orders = await window.go.main.App.GetAllOrders()
      
      // 过滤和验证工单数据
      const validOrders = (orders || []).filter(order => {
        const isValid = validatePlanData(order)
        if (!isValid) {
          console.warn('⚠️ 发现无效工单数据:', order)
        }
        return isValid
      })
      
      plans.value = validOrders
      console.log('✅ 加载工单列表:', plans.value.length, '条有效工单')
      
      if (plans.value.length > 0) {
        console.log('第一条工单:', plans.value[0])
        
        // 检查是否有重复ID
        const ids = plans.value.map(p => p.id)
        const uniqueIds = [...new Set(ids)]
        if (ids.length !== uniqueIds.length) {
          console.warn('⚠️ 发现重复的工单ID')
        }
      }
    }
  } catch (e) {
    console.error('❌ 加载工单失败:', e)
    showToast('加载工单列表失败: ' + e, 'error')
  }
}

// 加载所有员工并标记在班状态
const loadAllStaff = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      // 1. 获取所有员工
      const staffResult = await window.go.main.App.GetAllStaff(null, 1)
      
      // 2. 获取活动班次，提取在班人员ID
      await loadActiveSessions()
      const onDutyIds = new Set()
      const staffTeamMap = new Map() // 存储在班员工的班组信息
      
      for (const session of activeSessions.value) {
        try {
          const staffIds = JSON.parse(session.staff_ids || '[]')
          staffIds.forEach(id => {
            onDutyIds.add(id)
            staffTeamMap.set(id, session.team?.team_name || '未知班组')
          })
        } catch (e) {
          console.error('解析员工ID失败:', e)
        }
      }
      
      // 3. 更新在班人员ID集合
      onDutyStaffIds.value = onDutyIds
      
      // 4. 为所有员工添加班组信息
      allStaff.value = (staffResult || []).map(staff => ({
        ...staff,
        team_name: staffTeamMap.get(staff.id) || staff.current_team?.team_name || '未分配',
        is_on_duty: onDutyIds.has(staff.id)
      }))
      
      console.log('员工总数:', allStaff.value.length, '在班:', onDutyStaffList.value.length, '离班:', offDutyStaffList.value.length)
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
      console.log('加载设备列表:', devices.value.length, '台')
    }
  } catch (e) {
    console.error('加载设备失败:', e)
  }
}

// 加载班组列表
const loadTeams = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllTeams(1)
      teams.value = result || []
    }
  } catch (e) {
    console.error('加载班组失败:', e)
  }
}

// 加载活动班次
const loadActiveSessions = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const sessions = await window.go.main.App.GetAllActiveSessions()
      activeSessions.value = sessions || []
    }
  } catch (e) {
    activeSessions.value = []
  }
}

// 显示计划弹窗
const showPlanModal = (mode, plan = null) => {
  console.log('📝 显示计划弹窗:', { mode, plan })
  
  planModal.value.mode = mode
  
  if (mode === 'edit' && plan) {
    // 编辑模式：验证工单数据完整性
    if (!plan.id || plan.id <= 0) {
      console.error('❌ 编辑模式但工单ID无效:', plan)
      showToast('工单数据异常，无法编辑', 'error')
      return
    }
    
    // 确保所有必要字段都存在
    planModal.value.form = {
      id: plan.id,
      order_no: plan.order_no || '',
      product_code: plan.product_code || '',
      plan_qty: plan.plan_qty || 0,
      target_device_id: plan.target_device_id || null,
      status: plan.status !== undefined ? plan.status : 0,
      // 保留其他可能需要的字段
      actual_qty: plan.actual_qty || 0,
      ok_qty: plan.ok_qty || 0,
      ng_qty: plan.ng_qty || 0
    }
    
    console.log('✅ 编辑表单数据:', planModal.value.form)
  } else {
    // 新增模式
    planModal.value.form = {
      order_no: '',
      product_code: '',
      plan_qty: 0,
      target_device_id: null,
      status: 0
    }
  }
  
  planModal.value.show = true
}

// 关闭计划弹窗
const closePlanModal = () => {
  planModal.value.show = false
}

// 保存计划
const savePlan = async (plan) => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      if (planModal.value.mode === 'add') {
        await window.go.main.App.CreateOrder(
          plan.order_no,
          plan.product_code,
          plan.plan_qty,
          plan.target_device_id || null
        )
        console.log('创建工单成功')
        showToast('工单创建成功', 'success')
      } else {
        // 编辑模式：添加数据验证和调试信息
        console.log('🔧 编辑工单:', {
          id: plan.id,
          mode: planModal.value.mode,
          plan: plan
        })
        
        // 验证工单ID是否存在
        if (!plan.id || plan.id <= 0) {
          console.error('❌ 工单ID无效:', plan.id)
          showToast('工单ID无效，请刷新页面重试', 'error')
          return
        }
        
        // 检查工单是否还存在于当前列表中
        const existingPlan = plans.value.find(p => p.id === plan.id)
        if (!existingPlan) {
          console.error('❌ 工单在当前列表中不存在:', plan.id)
          showToast('工单可能已被删除，请刷新页面', 'warning')
          closePlanModal()
          await loadPlans()
          return
        }
        
        // 在更新前重新加载最新数据，确保工单仍然存在
        console.log('🔄 更新前验证工单存在性...')
        try {
          await loadPlans() // 重新加载最新数据
          const latestPlan = plans.value.find(p => p.id === plan.id)
          if (!latestPlan) {
            console.error('❌ 工单已被删除:', plan.id)
            showToast('工单已被删除，无法更新', 'error')
            closePlanModal()
            return
          }
          console.log('✅ 工单存在性验证通过')
        } catch (verifyError) {
          console.error('❌ 验证工单存在性失败:', verifyError)
          showToast('无法验证工单状态，请重试', 'error')
          return
        }
        
        // 准备更新参数并记录详细信息
        const updateParams = {
          id: plan.id,
          product_code: plan.product_code || null,
          plan_qty: plan.plan_qty || null,
          status: plan.status !== undefined ? plan.status : null,
          target_device_id: plan.target_device_id !== undefined ? plan.target_device_id : null
        }
        
        console.log('📤 发送更新请求:', updateParams)
        console.log('📋 原始工单数据:', existingPlan)
        
        // 检查是否有实际变化
        const hasChanges = 
          (updateParams.product_code !== null && updateParams.product_code !== existingPlan.product_code) ||
          (updateParams.plan_qty !== null && updateParams.plan_qty !== existingPlan.plan_qty) ||
          (updateParams.status !== null && updateParams.status !== existingPlan.status) ||
          (updateParams.target_device_id !== null && updateParams.target_device_id !== existingPlan.target_device_id)
        
        console.log('🔍 是否有实际变化:', hasChanges)
        console.log('🔧 字段变化详情:', {
          product_code: { old: existingPlan.product_code, new: updateParams.product_code, changed: updateParams.product_code !== null && updateParams.product_code !== existingPlan.product_code },
          plan_qty: { old: existingPlan.plan_qty, new: updateParams.plan_qty, changed: updateParams.plan_qty !== null && updateParams.plan_qty !== existingPlan.plan_qty },
          status: { old: existingPlan.status, new: updateParams.status, changed: updateParams.status !== null && updateParams.status !== existingPlan.status },
          target_device_id: { old: existingPlan.target_device_id, new: updateParams.target_device_id, changed: updateParams.target_device_id !== null && updateParams.target_device_id !== existingPlan.target_device_id }
        })
        
        if (!hasChanges) {
          console.log('ℹ️ 没有数据变化，跳过更新')
          showToast('没有数据变化', 'info')
          closePlanModal()
          return
        }
        
        await window.go.main.App.UpdateOrder(
          updateParams.id,
          updateParams.product_code,
          updateParams.plan_qty,
          updateParams.status,
          updateParams.target_device_id
        )
        console.log('✅ 更新工单成功')
        showToast('工单更新成功', 'success')
      }
      closePlanModal()
      await loadPlans()
    }
  } catch (e) {
    console.error('❌ 保存工单失败:', e)
    
    // 更友好的错误提示
    let errorMessage = '保存失败'
    if (e.toString().includes('工单不存在')) {
      errorMessage = '工单不存在，可能已被删除。页面将自动刷新'
      // 自动刷新数据
      setTimeout(async () => {
        await loadPlans()
      }, 1500)
    } else if (e.toString().includes('网络')) {
      errorMessage = '网络连接失败，请检查网络后重试'
    } else {
      errorMessage = `保存失败: ${e}`
    }
    
    showToast(errorMessage, 'error', 3000)
  }
}

// 确认删除（二次确认）
const confirmDelete = (plan) => {
  // 关闭弹窗
  closePlanModal()
  
  const completionRate = plan.plan_qty > 0 ? Math.round((plan.actual_qty || 0) / plan.plan_qty * 100) : 0
  
  // 第一次确认
  confirmDialog.value = {
    show: true,
    type: 'danger',
    title: '⚠️ 确认删除',
    message: '确认要删除这个工单吗？',
    details: [
      { icon: '📋', label: '工单号', value: plan.order_no },
      { icon: '📦', label: '产品', value: plan.product_code },
      { icon: '🎯', label: '计划数量', value: plan.plan_qty },
      { icon: '✅', label: '已完成', value: plan.actual_qty || 0 },
      { icon: '📊', label: '状态', value: getStatusText(plan.status) },
      { icon: '📈', label: '完成率', value: `${completionRate}%` }
    ],
    warnings: [
      '删除后无法恢复',
      '所有生产数据将被清除',
      '此操作不可撤销'
    ],
    warningTitle: '⚠️ 警告',
    confirmText: '确认删除',
    cancelText: '取消',
    confirmIcon: 'fas fa-trash',
    onConfirm: () => {
      confirmDialog.value.show = false
      
      // 延迟一下，等第一个对话框关闭动画完成
      setTimeout(() => {
        // 第二次确认
        confirmDialog.value = {
          show: true,
          type: 'danger',
          title: '⚠️ 再次确认',
          message: '真的要删除这个工单吗？',
          details: [
            { icon: '📋', label: '工单号', value: plan.order_no }
          ],
          warnings: [
            '这是最后一次确认',
            '删除后将无法恢复'
          ],
          warningTitle: '⚠️ 最后警告',
          confirmText: '确定删除',
          cancelText: '取消',
          confirmIcon: 'fas fa-exclamation-triangle',
          onConfirm: () => {
            confirmDialog.value.show = false
            deletePlan(plan)
          },
          onCancel: () => {
            confirmDialog.value.show = false
          }
        }
      }, 300)
    },
    onCancel: () => {
      confirmDialog.value.show = false
    }
  }
}

// 删除计划
const deletePlan = async (plan) => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      await window.go.main.App.DeleteOrder(plan.id)
      console.log('删除工单成功')
      showToast('工单已删除', 'success')
      await loadPlans()
    }
  } catch (e) {
    console.error('删除工单失败:', e)
    showToast('删除失败: ' + e, 'error')
  }
}

// 开工生产
const startProduction = async (order) => {
  try {
    if (!window.go?.main?.App) {
      showToast('系统未就绪，请稍后重试', 'error')
      return
    }
    
    // 1. 检查该设备是否有活动班次
    const session = activeSessions.value.find(s => s.device_id === order.target_device_id)
    if (!session) {
      const deviceName = getDeviceName(order.target_device_id)
      
      confirmDialog.value = {
        show: true,
        type: 'warning',
        title: '班次未登记',
        message: `设备【${deviceName}】还没有班次登记`,
        details: [],
        warnings: ['需要先进行班次登记才能开始生产'],
        warningTitle: '提示',
        confirmText: '前往登记',
        cancelText: '取消',
        confirmIcon: 'fas fa-arrow-right',
        onConfirm: () => {
          confirmDialog.value.show = false
          router.push({ name: 'staff' })
        },
        onCancel: () => {
          confirmDialog.value.show = false
        }
      }
      return
    }
    
    // 2. 解析班次员工信息
    let staffIds = []
    try {
      staffIds = JSON.parse(session.staff_ids || '[]')
    } catch (e) {
      console.error('解析员工ID失败:', e)
      showToast('班次数据异常，请重新登记', 'error')
      return
    }
    
    if (staffIds.length === 0) {
      showToast('当前班次没有员工，请重新登记班次', 'warning')
      return
    }
    
    // 3. 构建确认信息
    const deviceName = getDeviceName(order.target_device_id)
    const teamName = session.team?.team_name || '未知班组'
    
    confirmDialog.value = {
      show: true,
      type: 'success',
      title: '确认开工',
      message: '请确认开工信息',
      details: [
        { icon: '📋', label: '工单号', value: order.order_no },
        { icon: '📦', label: '产品', value: order.product_code },
        { icon: '🎯', label: '计划数量', value: order.plan_qty },
        { icon: '🏭', label: '设备', value: deviceName },
        { icon: '👥', label: '班组', value: teamName },
        { icon: '👷', label: '人员', value: `${staffIds.length} 人` }
      ],
      warnings: [
        '该设备其他"生产中"的工单将自动暂停',
        '产量脉冲将自动累加到此工单'
      ],
      warningTitle: '注意',
      confirmText: '开始生产',
      cancelText: '取消',
      confirmIcon: 'fas fa-play',
      onConfirm: async () => {
        confirmDialog.value.show = false
        try {
          // 4. 调用智能开工API
          console.log('🚀 调用智能开工:', { order_id: order.id })
          
          await window.go.main.App.StartProductionSmart(order.id)
          
          showToast('开工成功！该设备的其他工单已自动暂停', 'success', 3000)
          
          // 5. 刷新工单列表
          await loadPlans()
        } catch (e) {
          console.error('❌ 开工失败:', e)
          showToast('开工失败: ' + e, 'error')
        }
      },
      onCancel: () => {
        confirmDialog.value.show = false
      }
    }
    
  } catch (e) {
    console.error('❌ 开工失败:', e)
    showToast('开工失败: ' + e, 'error')
  }
}

// 暂停生产
const pauseProduction = async (order) => {
  try {
    if (!window.go?.main?.App) {
      showToast('系统未就绪', 'error')
      return
    }
    
    const deviceName = getDeviceName(order.target_device_id)
    
    confirmDialog.value = {
      show: true,
      type: 'warning',
      title: '确认暂停',
      message: '确认暂停生产吗？',
      details: [
        { icon: '📋', label: '工单号', value: order.order_no },
        { icon: '📦', label: '产品', value: order.product_code },
        { icon: '🏭', label: '设备', value: deviceName }
      ],
      warnings: [
        '工单状态将改为"暂停"',
        '产量脉冲将停止累加',
        '可以稍后点击"继续"恢复生产'
      ],
      warningTitle: '注意',
      confirmText: '暂停生产',
      cancelText: '取消',
      confirmIcon: 'fas fa-pause',
      onConfirm: async () => {
        confirmDialog.value.show = false
        try {
          // 调用更新工单状态API
          await window.go.main.App.UpdateOrder(order.id, null, null, 2, null) // status=2 暂停
          
          showToast('生产已暂停', 'success')
          await loadPlans()
        } catch (e) {
          console.error('❌ 暂停失败:', e)
          showToast('暂停失败: ' + e, 'error')
        }
      },
      onCancel: () => {
        confirmDialog.value.show = false
      }
    }
    
  } catch (e) {
    console.error('❌ 暂停失败:', e)
    showToast('暂停失败: ' + e, 'error')
  }
}

// 完成生产（良品数>=计划数）
const completeProduction = async (order) => {
  try {
    if (!window.go?.main?.App) {
      showToast('系统未就绪', 'error')
      return
    }
    
    const completionRate = order.plan_qty > 0 ? Math.round((order.ok_qty || 0) / order.plan_qty * 100) : 0
    
    confirmDialog.value = {
      show: true,
      type: 'success',
      title: '确认完成',
      message: '工单已达到计划数量，确认完成生产吗？',
      details: [
        { icon: '📋', label: '工单号', value: order.order_no },
        { icon: '📦', label: '产品', value: order.product_code },
        { icon: '🎯', label: '计划数量', value: order.plan_qty },
        { icon: '✅', label: '良品数', value: order.ok_qty || 0 },
        { icon: '❌', label: '不良品数', value: order.ng_qty || 0 },
        { icon: '📊', label: '完成率', value: `${completionRate}%` }
      ],
      warnings: [
        '工单将标记为"完工"状态'
      ],
      warningTitle: '提示',
      confirmText: '完成生产',
      cancelText: '取消',
      confirmIcon: 'fas fa-check',
      onConfirm: async () => {
        confirmDialog.value.show = false
        try {
          // 调用更新工单状态API
          await window.go.main.App.UpdateOrder(order.id, null, null, 3, null) // status=3 完工
          
          showToast('工单已完成！', 'success')
          await loadPlans()
        } catch (e) {
          console.error('❌ 完成失败:', e)
          showToast('完成失败: ' + e, 'error')
        }
      },
      onCancel: () => {
        confirmDialog.value.show = false
      }
    }
    
  } catch (e) {
    console.error('❌ 完成失败:', e)
    showToast('完成失败: ' + e, 'error')
  }
}

// 强制结束生产
const forceCompleteProduction = async (order) => {
  try {
    if (!window.go?.main?.App) {
      showToast('系统未就绪', 'error')
      return
    }
    
    const completionRate = order.plan_qty > 0 ? Math.round((order.ok_qty || 0) / order.plan_qty * 100) : 0
    
    confirmDialog.value = {
      show: true,
      type: 'danger',
      title: '⚠️ 强制结束',
      message: '确认要强制结束生产吗？',
      details: [
        { icon: '📋', label: '工单号', value: order.order_no },
        { icon: '📦', label: '产品', value: order.product_code },
        { icon: '🎯', label: '计划数量', value: order.plan_qty },
        { icon: '✅', label: '当前良品数', value: order.ok_qty || 0 },
        { icon: '❌', label: '不良品数', value: order.ng_qty || 0 },
        { icon: '📊', label: '完成率', value: `${completionRate}%` }
      ],
      warnings: [
        '工单将强制标记为"完工"',
        '即使未达到计划数量也会结束',
        '此操作不可撤销'
      ],
      warningTitle: '⚠️ 警告',
      confirmText: '强制结束',
      cancelText: '取消',
      confirmIcon: 'fas fa-flag-checkered',
      onConfirm: async () => {
        confirmDialog.value.show = false
        try {
          // 调用更新工单状态API
          await window.go.main.App.UpdateOrder(order.id, null, null, 3, null) // status=3 完工
          
          showToast('工单已强制结束', 'warning', 2500)
          await loadPlans()
        } catch (e) {
          console.error('❌ 强制结束失败:', e)
          showToast('强制结束失败: ' + e, 'error')
        }
      },
      onCancel: () => {
        confirmDialog.value.show = false
      }
    }
    
  } catch (e) {
    console.error('❌ 强制结束失败:', e)
    showToast('强制结束失败: ' + e, 'error')
  }
}

// 查看计划
const viewPlan = (plan) => {
  const completionRate = plan.plan_qty > 0 ? Math.round((plan.actual_qty || 0) / plan.plan_qty * 100) : 0
  const deviceName = getDeviceName(plan.target_device_id)
  
  confirmDialog.value = {
    show: true,
    type: 'info',
    title: '📋 工单详情',
    message: '查看工单详细信息',
    details: [
      { icon: '📋', label: '工单号', value: plan.order_no },
      { icon: '📦', label: '产品型号', value: plan.product_code },
      { icon: '🏭', label: '目标设备', value: deviceName },
      { icon: '🎯', label: '计划数量', value: plan.plan_qty },
      { icon: '📊', label: '实际产出', value: plan.actual_qty || 0 },
      { icon: '✅', label: '良品数', value: plan.ok_qty || 0 },
      { icon: '❌', label: '不良品数', value: plan.ng_qty || 0 },
      { icon: '📈', label: '完成率', value: `${completionRate}%` },
      { icon: '🔖', label: '状态', value: getStatusText(plan.status) }
    ],
    warnings: [],
    confirmText: '关闭',
    cancelText: '',  // 不显示取消按钮
    confirmIcon: 'fas fa-times',
    onConfirm: () => {
      confirmDialog.value.show = false
    },
    onCancel: () => {
      confirmDialog.value.show = false
    }
  }
}

// （已删除员工查看和删除功能）

// 跳转到人员管理页面
const goToStaffPage = () => {
  router.push({ name: 'staff' })
}

// 跳转到设备状态页面
const goToDevicePage = () => {
  router.push({ name: 'device' })
}

// 查看任务
const viewTask = (task) => {
  const deviceName = getDeviceName(task.target_device_id)
  
  confirmDialog.value = {
    show: true,
    type: 'info',
    title: '📋 任务详情',
    message: '查看任务详细信息',
    details: [
      { icon: '📋', label: '工单号', value: task.order_no },
      { icon: '📦', label: '产品型号', value: task.product_code },
      { icon: '🏭', label: '目标设备', value: deviceName },
      { icon: '🎯', label: '计划数量', value: task.plan_qty },
      { icon: '📊', label: '实际数量', value: task.actual_qty },
      { icon: '❌', label: '不良品数', value: task.ng_qty },
      { icon: '📈', label: '完成率', value: `${task.progress}%` },
      { icon: '🔖', label: '状态', value: getStatusText(task.status) }
    ],
    warnings: [],
    confirmText: '关闭',
    cancelText: '',  // 不显示取消按钮
    confirmIcon: 'fas fa-times',
    onConfirm: () => {
      confirmDialog.value.show = false
    },
    onCancel: () => {
      confirmDialog.value.show = false
    }
  }
}

// 编辑任务
const editTask = (task) => {
  console.log('编辑任务:', task)
  // 可以跳转到工单编辑
  const plan = plans.value.find(p => p.id === task.id)
  if (plan) {
    showPlanModal('edit', plan)
  }
}

// 获取设备名称
const getDeviceName = (deviceId) => {
  if (!deviceId) return '-'
  const device = devices.value.find(d => d.id === deviceId)
  return device ? device.device_name : `设备${deviceId}`
}

// 获取状态文本 - 简化版本
const getStatusText = (status) => {
  const map = {
    0: '待产',
    1: '进行',  // 简化"生产中"为"进行"
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

// 格式化用时
const formatUsedTime = (plan) => {
  let totalSeconds = plan.used_seconds || 0
  
  // 如果正在生产中且有开始时间，加上当前段的用时
  if (plan.status === 1 && plan.current_start_time) {
    const currentStart = new Date(plan.current_start_time)
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

// 显示提示框
const showToast = (message, type = 'success', duration = 2000) => {
  toast.value.message = message
  toast.value.type = type
  toast.value.duration = duration
  toast.value.show = true
}

// 显示快捷换班弹窗
const showQuickShift = () => {
  quickShiftModal.value.show = true
}

// 关闭快捷换班弹窗
const closeQuickShift = () => {
  quickShiftModal.value.show = false
}

// 显示在线设备弹窗
const showActiveDevices = () => {
  activeDevicesModal.value.show = true
}

// 关闭在线设备弹窗
const closeActiveDevices = () => {
  activeDevicesModal.value.show = false
}

// 处理快捷换班（从 ShiftModal 触发）
const handleQuickShift = async (data) => {
  console.log('🔥 handleQuickShift 被调用，收到数据:', data)
  
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      console.error('❌ Wails Go 对象未加载')
      showToast('系统未就绪，请稍后重试', 'error')
      return
    }
    
    // 1. 检查该设备是否已有活动班次，如果有则先下班
    const existingSession = activeSessions.value.find(s => s.device_id === data.device_id)
    if (existingSession) {
      console.log('📤 设备', data.device_id, '已有活动班次，正在下班:', existingSession.id)
      await window.go.main.App.DeviceLogout(data.device_id)
      console.log('✅ 下班成功')
    }
    
    // 2. 更新员工班组归属
    console.log('🔍 [换班] 开始处理员工班组归属...')
    
    // 2.1 重新加载完整的员工列表（确保数据最新）
    try {
      await loadAllStaff()
      console.log('📊 [换班] 重新加载完整员工列表，共', allStaff.value.length, '人')
    } catch (e) {
      console.error('❌ [换班] 重新加载员工列表失败:', e)
    }
    
    // 2.2 获取目标班组的原有成员ID
    const originalTeamMemberIds = allStaff.value
      .filter(s => s.current_team_id === data.team_id && s.is_active === 1)
      .map(s => s.id)
    
    console.log('📋 [换班] 目标班组原有成员ID:', originalTeamMemberIds)
    console.log('👥 [换班] 选中员工ID:', data.staff_ids)
    
    // 2.3 找出需要加入班组的员工（选中了但不在班组内的）
    const staffToJoin = data.staff_ids.filter(staffId => !originalTeamMemberIds.includes(staffId))
    
    // 2.4 找出需要移出班组的员工（在班组内但没被选中的 - 这种情况用户手动取消选择）
    const staffToLeave = originalTeamMemberIds.filter(staffId => !data.staff_ids.includes(staffId))
    
    console.log(`📥 [换班] 需要加入班组的员工:`, staffToJoin)
    console.log(`📤 [换班] 需要移出班组的员工:`, staffToLeave)
    
    let joinCount = 0
    let leaveCount = 0
    
    // 2.5 将需要加入的员工调入目标班组
    for (const staffId of staffToJoin) {
      const staff = allStaff.value.find(s => s.id === staffId)
      if (staff) {
        try {
          const oldTeamName = staff.current_team?.team_name || '无班组'
          console.log(`  📥 加入: ${staff.name}(id:${staffId}) (${oldTeamName} → 目标班组)`)
          
          await window.go.main.App.UpdateStaff(
            staffId,
            null, // 不更新姓名
            data.team_id, // 更新到目标班组
            null // 不更新状态
          )
          joinCount++
          console.log(`  ✅ ${staff.name} 已加入班组`)
        } catch (e) {
          console.error(`  ❌ 更新员工 ${staff.name}(id:${staffId}) 失败:`, e)
        }
      } else {
        console.error(`  ⚠️ 在员工列表中找不到 id=${staffId} 的员工`)
      }
    }
    
    // 2.6 将需要移出的员工从班组中移除（设为 null）
    for (const staffId of staffToLeave) {
      const staff = allStaff.value.find(s => s.id === staffId)
      if (staff) {
        try {
          console.log(`  📤 移出: ${staff.name}(id:${staffId}) (从目标班组移出)`)
          
          await window.go.main.App.UpdateStaff(
            staffId,
            null, // 不更新姓名
            -1, // -1 表示清空班组
            null // 不更新状态
          )
          leaveCount++
          console.log(`  ✅ ${staff.name} 已移出班组`)
        } catch (e) {
          console.error(`  ❌ 移出员工 ${staff.name}(id:${staffId}) 失败:`, e)
        }
      } else {
        console.error(`  ⚠️ 在员工列表中找不到 id=${staffId} 的员工`)
      }
    }
    
    console.log(`✅ [换班] 班组调整完成: ${joinCount}人加入, ${leaveCount}人移出`)
    
    // 3. 登录新班次
    console.log('🔐 开始登录新班次:', {
      device_id: data.device_id,
      team_id: data.team_id,
      staff_ids: data.staff_ids
    })
    
    await window.go.main.App.DeviceLogin(
      data.device_id,
      data.team_id,
      data.staff_ids
    )
    
    console.log('✅ 登录成功')
    
    closeQuickShift()
    
    // 等待一下确保数据库已更新
    await new Promise(resolve => setTimeout(resolve, 500))
    
    await loadActiveSessions()
    await loadAllStaff() // 刷新人员信息
    
    // 生成更详细的提示信息
    let message = '班次登记成功！'
    if (joinCount > 0 || leaveCount > 0) {
      const details = []
      if (joinCount > 0) details.push(`${joinCount}人加入班组`)
      if (leaveCount > 0) details.push(`${leaveCount}人移出班组`)
      message += ` (${details.join('，')})`
    }
    showToast(message, 'success', 3000)
    
  } catch (e) {
    console.error('❌ 换班失败:', e)
    showToast('换班失败: ' + e, 'error')
  }
}

// 处理下班（从 ActiveDevicesModal 触发）
const handleLogout = async (deviceId) => {
  const session = activeSessions.value.find(s => s.device_id === deviceId)
  if (!session) {
    showToast('该设备当前没有活动班次', 'warning')
    return
  }

  // 获取设备名称
  let deviceName = `设备${deviceId}`
  const device = devices.value.find(d => d.id === deviceId)
  if (device) {
    deviceName = device.device_name
  }

  // 获取班组名称
  const teamName = session.team?.team_name || '未知班组'
  
  // 解析员工信息
  let staffIds = []
  try {
    staffIds = JSON.parse(session.staff_ids || '[]')
  } catch (e) {
    console.error('解析员工ID失败:', e)
  }
  
  // 获取员工姓名列表
  const staffNames = staffIds
    .map(id => {
      const staff = allStaff.value.find(s => s.id === id)
      return staff ? staff.name : `员工${id}`
    })
    .join('、')
  
  // 计算班次时长
  let duration = '-'
  if (session.login_time) {
    const loginDate = new Date(session.login_time)
    const now = new Date()
    const diffMs = now - loginDate
    const diffMin = Math.floor(diffMs / 60000)
    duration = `${Math.floor(diffMin / 60)}h${diffMin % 60}m`
  }

  // 显示确认对话框
  confirmDialog.value = {
    show: true,
    type: 'danger',
    title: '⚠️ 确认结束班次',
    message: `确定要结束【${deviceName}】的当前班次吗？`,
    details: [
      { icon: '🏭', label: '设备', value: deviceName },
      { icon: '👥', label: '班组', value: teamName },
      { icon: '👷', label: '人员', value: staffNames || '无' },
      { icon: '⏱️', label: '已工作', value: duration }
    ],
    warnings: [
      '班次将被标记为结束并记录工作时长',
      '该设备可以开始新的班次',
      '员工仍保持在原班组中'
    ],
    warningTitle: '💡 提示',
    confirmText: '确认结束',
    cancelText: '取消',
    confirmIcon: 'fas fa-power-off',
    onConfirm: async () => {
      confirmDialog.value.show = false
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.DeviceLogout(deviceId)
          await loadActiveSessions()
          await loadAllStaff() // 刷新人员信息
          closeActiveDevices()
          // 显示结束成功提示
          showToast(`【${deviceName}】班次已结束`, 'success')
        }
      } catch (e) {
        console.error('结束班次失败:', e)
        showToast('结束班次失败: ' + e, 'error')
      }
    },
    onCancel: () => {
      confirmDialog.value.show = false
    }
  }
}

// 定时刷新能耗数据
let energyRefreshTimer = null

const startEnergyRefresh = () => {
  loadDeviceEnergyData() // 立即加载
  energyRefreshTimer = setInterval(() => {
    loadDeviceEnergyData()
  }, 5000) // 5秒刷新
}

const stopEnergyRefresh = () => {
  if (energyRefreshTimer) {
    clearInterval(energyRefreshTimer)
    energyRefreshTimer = null
  }
}

onMounted(async () => {
  console.log('生产计划管理页面加载')
  await loadDevices()
  await loadTeams()
  await loadActiveSessions()
  await loadPlans()
  await loadAllStaff()
  startEnergyRefresh() // 启动能耗数据刷新
})

onUnmounted(() => {
  stopEnergyRefresh()
})
</script>

<style scoped>
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

/* 刷新按钮 */
/* 卡片头部操作按钮组 */
.header-actions-mini {
  display: flex;
  gap: 8px;
  align-items: center;
}

.btn-refresh {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, rgba(84, 110, 122, 0.2) 0%, rgba(96, 125, 139, 0.2) 100%);
  border: 1px solid rgba(84, 110, 122, 0.3);
  color: #546e7a;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.btn-refresh:hover {
  background: linear-gradient(135deg, #546e7a 0%, #607d8b 100%);
  color: #fff;
  transform: rotate(90deg);
  box-shadow: 0 4px 12px rgba(84, 110, 122, 0.3);
}

.btn-refresh:active {
  transform: rotate(90deg) scale(0.95);
}

/* 快捷换班按钮 */
.btn-quick-shift {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, rgba(243, 156, 18, 0.2) 0%, rgba(230, 126, 34, 0.2) 100%);
  border: 1px solid rgba(243, 156, 18, 0.3);
  color: #f39c12;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.btn-quick-shift:hover {
  background: linear-gradient(135deg, #f39c12 0%, #e67e22 100%);
  color: #fff;
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(243, 156, 18, 0.4);
}

.btn-quick-shift:active {
  transform: scale(1.05);
}

/* 结束班次按钮 */
.btn-active-devices {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, rgba(231, 76, 60, 0.2) 0%, rgba(192, 57, 43, 0.2) 100%);
  border: 1px solid rgba(231, 76, 60, 0.3);
  color: #e74c3c;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.btn-active-devices:hover {
  background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
  color: #fff;
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.4);
}

.btn-active-devices:active {
  transform: scale(1.05);
}

/* 统计卡片 - 左右布局，四个卡片宽度一致，高度与人员管理页面一致 */
.stats-row {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
  flex-wrap: nowrap;
}

.stat-card {
  padding: 24px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 20px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.3);
  transition: all 0.3s;
  flex: 1;
  min-width: 0;
  height: auto;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0,0,0,0.4);
}

/* 工业低调配色 - 与人员管理页面统一 */
.stat-card.card-total {
  background: linear-gradient(135deg, #556070 0%, #6e7e8e 100%);
}

.stat-card.card-completed {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

/* .stat-card.card-workers {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
}

.stat-card.card-efficiency {
  background: linear-gradient(135deg, #5a7080 0%, #6e8e9e 100%);
} */

/* 设备操作卡片 - 左右布局，保持与其他卡片相同的 padding */
.stat-card.card-device-control {
  padding: 24px;
  display: flex;
  flex-direction: row;
  gap: 20px;
  align-items: center;
  position: relative;
  overflow: hidden;
}

.stat-card.card-device-control.device-1 {
  background: linear-gradient(135deg, #5a7a8f 0%, #6e8fa8 100%);
  border: 2px solid rgba(90, 122, 143, 0.3);
}

.stat-card.card-device-control.device-2 {
  background: linear-gradient(135deg, #7a6a5f 0%, #9e8a7e 100%);
  border: 2px solid rgba(122, 106, 95, 0.3);
}

/* 设备信息 - 左侧 */
.device-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.device-icon {
  width: 64px;
  height: 64px;
  background: rgba(255,255,255,0.2);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  flex-shrink: 0;
}

.device-title {
  font-size: 20px;
  font-weight: bold;
  color: #fff;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

/* 设备控制按钮组 - 右侧 */
.device-controls {
  display: flex;
  gap: 10px;
  margin-left: auto;
  flex-shrink: 0;
}

/* 设备操作按钮 - 紧凑型 */
.device-btn {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 16px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
  box-shadow: 0 3px 10px rgba(0,0,0,0.2);
  flex-shrink: 0;
  min-width: 90px;
}

/* NG+1 按钮 - 红色渐变，更大 */
.device-btn.ng-plus {
  background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
  border: 2px solid rgba(231, 76, 60, 0.4);
}

.device-btn.ng-plus .btn-icon {
  font-size: 24px;
  color: #fff;
}

.device-btn.ng-plus .btn-label {
  font-size: 15px;
  font-weight: bold;
  color: #fff;
}

/* NG-1 按钮 - 绿色渐变，稍小 */
.device-btn.ng-minus {
  background: linear-gradient(135deg, #27ae60 0%, #229954 100%);
  border: 2px solid rgba(39, 174, 96, 0.4);
}

.device-btn.ng-minus .btn-icon {
  font-size: 20px;
  color: #fff;
}

.device-btn.ng-minus .btn-label {
  font-size: 13px;
  font-weight: 600;
  color: #fff;
}

/* 按钮悬停效果 */
.device-btn:hover {
  transform: scale(1.1) translateY(-2px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.3);
}

.device-btn.ng-plus:hover {
  background: linear-gradient(135deg, #ec7063 0%, #e74c3c 100%);
  border-color: rgba(231, 76, 60, 0.6);
  box-shadow: 0 8px 24px rgba(231, 76, 60, 0.4);
}

.device-btn.ng-minus:hover {
  background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
  border-color: rgba(46, 204, 113, 0.6);
  box-shadow: 0 8px 24px rgba(46, 204, 113, 0.4);
}

/* 按钮点击效果 */
.device-btn:active {
  transform: scale(1.05) translateY(0);
}

/* 按钮图标容器 */
.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: rgba(255,255,255,0.2);
  border-radius: 50%;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.device-btn:hover .btn-icon {
  background: rgba(255,255,255,0.3);
  transform: rotate(180deg);
}

/* 按钮标签 */
.btn-label {
  letter-spacing: 1px;
  text-shadow: 0 2px 4px rgba(0,0,0,0.2);
}

/* 按钮波纹效果 */
.device-btn::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 0;
  height: 0;
  background: rgba(255,255,255,0.3);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  transition: width 0.6s, height 0.6s;
}

.device-btn:active::before {
  width: 200%;
  height: 200%;
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
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
  min-width: 0;
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
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 内容网格 - 4个模块 */
.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.module-card {
  min-height: 200px;
}

.module-card.full-width {
  grid-column: 1 / 2;
}

/* 卡片 */
.card {
  background: rgba(20, 30, 48, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.1);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.card-header {
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.card-header h3 {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}

/* 表格容器 - 自适应优化 */
.table-container {
  flex: 1;
  overflow-x: auto;  /* 水平滚动 */
  overflow-y: hidden;
  min-width: 0;      /* 允许收缩 */
}

/* 表格最小宽度 */
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
  table-layout: fixed;
  min-width: 700px;  /* 设置最小宽度，小于此宽度时出现滚动条 */
}

/* Naive UI 表格样式覆盖 */
.table-container :deep(.n-data-table) {
  --n-th-color: rgba(102, 126, 234, 0.12);
  --n-td-color: transparent;
  --n-border-color: rgba(255,255,255,0.05);
  --n-th-text-color: rgba(255,255,255,0.85);
  --n-td-text-color: rgba(255,255,255,0.8);
  --n-font-size: 13px;
}

.table-container :deep(.n-data-table-th) {
  font-weight: 600;
  font-size: 12px;
}

.table-container :deep(.n-data-table-tr:hover) {
  background: rgba(255,255,255,0.03);
}

/* 人员容器 */
.staff-container {
  overflow-y: auto;
  max-height: 240px;
  padding: 12px;
  flex: 1;
}

.staff-section {
  margin-bottom: 16px;
}

.staff-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: rgba(255,255,255,0.85);
  margin-bottom: 12px;
  padding: 8px 12px;
  background: rgba(255,255,255,0.05);
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-left: 3px solid #667eea;
}

/* 人员卡片网格 */
.staff-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 10px;
}

/* 在班员工卡片 */
.staff-card.on-duty {
  background: linear-gradient(135deg, rgba(46, 204, 113, 0.1) 0%, rgba(39, 174, 96, 0.1) 100%);
  border: 1px solid rgba(46, 204, 113, 0.3);
  border-radius: 10px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
}

.staff-card.on-duty::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 3px;
  height: 100%;
  background: linear-gradient(180deg, #2ecc71 0%, #27ae60 100%);
}

/* 离班员工卡片 */
.staff-card.off-duty {
  background: linear-gradient(135deg, rgba(127, 140, 141, 0.1) 0%, rgba(95, 106, 106, 0.1) 100%);
  border: 1px solid rgba(127, 140, 141, 0.3);
  border-radius: 10px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
  opacity: 0.85;
}

.staff-card.off-duty::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 3px;
  height: 100%;
  background: rgba(127, 140, 141, 0.5);
}

.staff-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.2);
}

.staff-card.on-duty:hover {
  border-color: rgba(46, 204, 113, 0.6);
  box-shadow: 0 4px 16px rgba(46, 204, 113, 0.2);
}

.staff-card.off-duty:hover {
  opacity: 1;
  border-color: rgba(127, 140, 141, 0.5);
}

/* 头像 */
.staff-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #fff;
  margin: 0 auto;
}

.staff-avatar.online {
  background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
  box-shadow: 0 4px 12px rgba(46, 204, 113, 0.3);
}

.staff-avatar.offline {
  background: linear-gradient(135deg, #95a5a6 0%, #7f8c8d 100%);
}

.staff-info {
  text-align: center;
  flex: 1;
}

.staff-name-line {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.staff-code-line {
  font-size: 11px;
  font-family: 'Courier New', monospace;
  color: rgba(255,255,255,0.5);
  margin-bottom: 6px;
}

.staff-team-line {
  font-size: 11px;
  color: rgba(255,255,255,0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 3px 8px;
  background: rgba(255,255,255,0.05);
  border-radius: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.staff-team-line i {
  font-size: 10px;
}

/* 状态标签 */
.staff-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 5px 10px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 500;
}

.staff-status.online {
  background: rgba(46, 204, 113, 0.15);
  border: 1px solid rgba(46, 204, 113, 0.4);
  color: #2ecc71;
}

.staff-status.offline {
  background: rgba(127, 140, 141, 0.15);
  border: 1px solid rgba(127, 140, 141, 0.4);
  color: #95a5a6;
}

.staff-status .status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.staff-status.online .status-dot {
  background: #2ecc71;
  animation: pulse 2s infinite;
}

.staff-status.offline .status-dot {
  background: #95a5a6;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    box-shadow: 0 0 0 0 rgba(46, 204, 113, 0.7);
  }
  50% {
    opacity: 0.8;
    box-shadow: 0 0 0 6px rgba(46, 204, 113, 0);
  }
}

.empty-staff {
  text-align: center;
  padding: 40px 20px;
  color: rgba(255,255,255,0.4);
}

.empty-staff i {
  font-size: 40px;
  display: block;
  margin-bottom: 12px;
  opacity: 0.5;
}

.empty-staff p {
  font-size: 14px;
  margin: 0;
}

/* 原生表格样式（与人员管理页面统一） */
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
  table-layout: fixed;
}

/* 表格列宽优化 - 自适应布局 */
.data-table th:nth-child(1) { width: 50px; }    /* 序号 */
.data-table th:nth-child(2) { width: 130px; }   /* 工单号 */
.data-table th:nth-child(3) { width: auto; }    /* 项目名称 - 自适应 */
.data-table th:nth-child(4) { width: 85px; }    /* 状态 */
.data-table th:nth-child(5) { width: 85px; }    /* 设备 */
.data-table th:nth-child(6) { width: 80px; }    /* 用时 */
.data-table th:nth-child(7) { width: 180px; }   /* 操作 */

.data-table thead {
  background: rgba(102, 126, 234, 0.15);
}

.data-table th {
  padding: 12px 8px;  /* 减少内边距 */
  text-align: center;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  border-bottom: 2px solid rgba(255,255,255,0.1);
  white-space: nowrap;
  font-size: 13px;    /* 稍微小一点 */
}

.data-table td {
  padding: 12px 8px;  /* 减少内边距 */
  border-bottom: 1px solid rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
  text-align: center;
  white-space: nowrap;  /* 防止所有单元格换行 */
  overflow: hidden;     /* 隐藏溢出 */
  text-overflow: ellipsis; /* 超长显示省略号 */
}

.data-table tbody tr {
  transition: all 0.2s;
}

.data-table tbody tr:hover {
  background: rgba(255,255,255,0.03);
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

/* 表格元素样式 */
.plan-code {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: #667eea;
}

.project-name {
  font-weight: 500;
  color: rgba(255,255,255,0.9);
}

.device-tag {
  font-size: 12px;
  color: rgba(255,255,255,0.7);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 80px;  /* 限制最大宽度 */
  display: inline-block;
}

.task-code {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: #667eea;
}

.qty-badge {
  display: inline-block;
  padding: 4px 12px;
  background: rgba(52, 152, 219, 0.2);
  border: 1px solid rgba(52, 152, 219, 0.4);
  border-radius: 12px;
  font-size: 13px;
  color: #3498db;
  font-weight: 600;
}

.qty-badge.ok {
  background: rgba(46, 204, 113, 0.2);
  border-color: rgba(46, 204, 113, 0.4);
  color: #2ecc71;
}

.qty-badge.ng {
  background: rgba(231, 76, 60, 0.2);
  border-color: rgba(231, 76, 60, 0.4);
  color: #e74c3c;
}

.time-badge {
  display: inline-block;
  padding: 4px 12px;
  background: rgba(52, 73, 94, 0.2);
  border: 1px solid rgba(52, 73, 94, 0.4);
  border-radius: 12px;
  font-size: 13px;
  color: #ecf0f1;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

/* 状态徽章 - 防止换行优化 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;  /* 防止换行 */
  min-width: 60px;      /* 最小宽度 */
  justify-content: center;
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

/* 进度条 */
.progress-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-bar-mini {
  flex: 1;
  height: 8px;
  background: rgba(255,255,255,0.1);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill-mini {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  border-radius: 4px;
  transition: width 0.3s;
}

.progress-text-mini {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255,255,255,0.7);
  min-width: 40px;
  text-align: right;
}

/* 设置内容区域 */
.settings-content {
  padding: 16px 20px;
  overflow-y: auto;
  max-height: 240px;
  flex: 1;
}

.setting-section {
  margin-bottom: 12px;
}

.setting-section:last-child {
  margin-bottom: 0;
}

.setting-label {
  font-size: 13px;
  font-weight: 600;
  color: rgba(255,255,255,0.85);
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.table-container-mini {
  overflow-y: auto;
  max-height: 160px;
}

/* 迷你表格样式 */
.data-table.mini {
  font-size: 12px;
  table-layout: fixed;
}

/* 班制信息表格列宽 */
.setting-section:first-child .data-table.mini th:nth-child(1) { width: 40px; }   /* # */
.setting-section:first-child .data-table.mini th:nth-child(2) { width: auto; }   /* 班组名称 - 自适应 */
.setting-section:first-child .data-table.mini th:nth-child(3) { width: 60px; }   /* 开始 */
.setting-section:first-child .data-table.mini th:nth-child(4) { width: 70px; }   /* 结束 */
.setting-section:first-child .data-table.mini th:nth-child(5) { width: 65px; }   /* 时长 */
.setting-section:first-child .data-table.mini th:nth-child(6) { width: 50px; }   /* 状态 */
.setting-section:first-child .data-table.mini th:nth-child(7) { width: 50px; }   /* 操作 */

/* 能耗信息表格列宽 */
.setting-section:last-child .data-table.mini th:nth-child(1) { width: 40px; }    /* # */
.setting-section:last-child .data-table.mini th:nth-child(2) { width: auto; }    /* 设备名称 - 自适应 */
.setting-section:last-child .data-table.mini th:nth-child(3) { width: 75px; }    /* 当前功率 */
.setting-section:last-child .data-table.mini th:nth-child(4) { width: 110px; }   /* 总能耗 - 支持10位数 */
.setting-section:last-child .data-table.mini th:nth-child(5) { width: 50px; }    /* 操作 */

.data-table.mini th {
  padding: 10px 6px;
  font-size: 11px;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.data-table.mini td {
  padding: 10px 6px;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.empty-mini {
  text-align: center;
  color: rgba(255,255,255,0.4);
  padding: 20px !important;
  font-size: 12px;
}

/* 迷你状态徽章 */
.mini-status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  font-size: 12px;
}

.mini-status-badge.active {
  color: #2ecc71;
  background: rgba(46, 204, 113, 0.2);
}

.mini-status-badge.inactive {
  color: #95a5a6;
  background: rgba(149, 165, 166, 0.2);
}

/* 迷你按钮 */
.table-btn.mini {
  padding: 4px 10px;
  font-size: 12px;
}

.team-name {
  font-weight: 500;
  color: rgba(255,255,255,0.9);
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.device-name {
  font-weight: 500;
  color: rgba(255,255,255,0.9);
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.power-value {
  color: #f1c40f;
  font-weight: 600;
  font-size: 11px;
  white-space: nowrap;
  font-family: 'Courier New', monospace;
}

.consumption-value {
  color: #3498db;
  font-weight: 600;
  font-size: 11px;
  white-space: nowrap;
  font-family: 'Courier New', monospace;
}

/* 表格操作按钮 - 优化版 */
.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
  min-width: 140px; /* 确保有足够空间 */
}

.table-btn {
  padding: 6px 12px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.8);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap; /* 防止文字换行 */
  flex-shrink: 0;      /* 防止按钮被压缩 */
  min-width: 32px;
  height: 32px;
}

.table-btn:hover {
  transform: scale(1.05);
}

.table-btn.info:hover {
  background: #3498db;
  color: #fff;
  box-shadow: 0 4px 12px rgba(52, 152, 219, 0.3);
}

.table-btn.warning:hover {
  background: #f39c12;
  color: #fff;
  box-shadow: 0 4px 12px rgba(243, 156, 18, 0.3);
}

.table-btn.edit:hover {
  background: #667eea;
  color: #fff;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.table-btn.delete:hover {
  background: #e74c3c;
  color: #fff;
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.3);
}

/* 开工按钮 - 圆形设计 */
.table-btn.start-btn {
  width: 36px;
  height: 36px;
  padding: 0;
  border-radius: 50%;
  background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
  border: 2px solid rgba(46, 204, 113, 0.3);
  color: #fff;
  font-size: 14px;
  position: relative;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(46, 204, 113, 0.2);
}

.table-btn.start-btn::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 0;
  height: 0;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  transition: all 0.3s ease;
}

.table-btn.start-btn:hover::before {
  width: 100%;
  height: 100%;
}

.table-btn.start-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 20px rgba(46, 204, 113, 0.4);
  border-color: rgba(46, 204, 113, 0.6);
}

.table-btn.start-btn:active {
  transform: scale(1.05);
}

/* 查看按钮 - 圆形设计 */
.table-btn.view-btn {
  width: 36px;
  height: 36px;
  padding: 0;
  border-radius: 50%;
  background: rgba(52, 152, 219, 0.2);
  border: 2px solid rgba(52, 152, 219, 0.3);
  color: #3498db;
  font-size: 14px;
}

.table-btn.view-btn:hover {
  background: #3498db;
  color: #fff;
  transform: scale(1.1);
  box-shadow: 0 6px 20px rgba(52, 152, 219, 0.3);
}

/* 继续按钮 - 橙色圆形 */
.table-btn.resume-btn {
  width: 36px;
  height: 36px;
  padding: 0;
  border-radius: 50%;
  background: linear-gradient(135deg, #f39c12 0%, #e67e22 100%);
  border: 2px solid rgba(243, 156, 18, 0.3);
  color: #fff;
  font-size: 14px;
  box-shadow: 0 4px 12px rgba(243, 156, 18, 0.2);
}

.table-btn.resume-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 20px rgba(243, 156, 18, 0.4);
  border-color: rgba(243, 156, 18, 0.6);
}

/* 暂停按钮 - 黄色圆形 */
.table-btn.pause-btn {
  width: 36px;
  height: 36px;
  padding: 0;
  border-radius: 50%;
  background: linear-gradient(135deg, #f1c40f 0%, #f39c12 100%);
  border: 2px solid rgba(241, 196, 15, 0.3);
  color: #fff;
  font-size: 14px;
  box-shadow: 0 4px 12px rgba(241, 196, 15, 0.2);
}

.table-btn.pause-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 20px rgba(241, 196, 15, 0.4);
  border-color: rgba(241, 196, 15, 0.6);
}

/* 完成按钮 - 紫色圆形 */
.table-btn.complete-btn {
  width: 36px;
  height: 36px;
  padding: 0;
  border-radius: 50%;
  background: linear-gradient(135deg, #9b59b6 0%, #8e44ad 100%);
  border: 2px solid rgba(155, 89, 182, 0.3);
  color: #fff;
  font-size: 14px;
  box-shadow: 0 4px 12px rgba(155, 89, 182, 0.2);
}

.table-btn.complete-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 20px rgba(155, 89, 182, 0.4);
  border-color: rgba(155, 89, 182, 0.6);
}

/* 强制结束按钮 - 方形红色 */
.table-btn.force-end-btn {
  padding: 6px 12px;
  background: rgba(231, 76, 60, 0.2);
  border: 1px solid rgba(231, 76, 60, 0.4);
  color: #e74c3c;
}

.table-btn.force-end-btn:hover {
  background: #e74c3c;
  color: #fff;
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.3);
}
</style>
