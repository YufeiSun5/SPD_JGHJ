<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container large">
      <div class="modal-header">
        <h3>
          <i class="fas fa-history"></i>
          班次换班记录
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <!-- 开发中遮罩层 -->
        <!-- <div class="development-overlay">
          <div class="development-content">
            <div class="development-icon">
              <i class="fas fa-tools"></i>
            </div>
            <h3>功能开发中</h3>
            <p>班次历史记录功能正在开发中，敬请期待...</p>
            <div class="development-progress">
              <div class="progress-bar">
                <div class="progress-fill" style="width: 65%"></div>
              </div>
              <span class="progress-text">开发进度: 65%</span>
            </div>
          </div>
        </div> -->

        <!-- 原有内容 -->
        <div>
          <n-config-provider :theme-overrides="themeOverrides">
            <!-- 筛选栏 -->
            <div class="filter-section">
              <n-space>
                <n-select
                  v-model:value="filters.deviceId"
                  :options="deviceOptions"
                  placeholder="选择设备"
                  style="width: 150px"
                  clearable
                  @update:value="loadHistory"
                />
                <n-select
                  v-model:value="filters.teamId"
                  :options="teamOptions"
                  placeholder="选择班组"
                  style="width: 150px"
                  clearable
                  @update:value="loadHistory"
                />
                <n-date-picker
                  v-model:value="filters.dateRange"
                  type="daterange"
                  clearable
                  placeholder="选择日期范围"
                  @update:value="loadHistory"
                />
                <n-button type="primary" @click="loadHistory">
                  <template #icon>
                    <i class="fas fa-search"></i>
                  </template>
                  查询
                </n-button>
              </n-space>
            </div>

            <!-- 数据表格 -->
            <n-data-table
              :columns="columns"
              :data="historyData"
              :pagination="pagination"
              :loading="loading"
              :bordered="false"
              :single-line="false"
              striped
              size="small"
              class="history-table"
            />
          </n-config-provider>
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
import { ref, h, onMounted } from 'vue'
import { NSpace, NTag, NDataTable, NSelect, NDatePicker, NButton, NConfigProvider } from 'naive-ui'

const emit = defineEmits(['close'])

// Naive UI 深色主题配置
const themeOverrides = {
  common: {
    primaryColor: '#667eea',
    primaryColorHover: '#764ba2',
    primaryColorPressed: '#5568d3',
    textColorBase: '#fff',
    textColor1: 'rgba(255,255,255,0.9)',
    textColor2: 'rgba(255,255,255,0.8)',
    textColor3: 'rgba(255,255,255,0.6)',
    baseColor: '#0a0e27',
    cardColor: 'rgba(20, 30, 48, 0.6)',
    modalColor: 'rgba(20, 30, 48, 0.98)',
    bodyColor: '#0a0e27',
    borderColor: 'rgba(255,255,255,0.1)'
  },
  DataTable: {
    thColor: 'rgba(102, 126, 234, 0.15)',
    tdColor: 'transparent',
    tdColorHover: 'rgba(255,255,255,0.05)',
    borderColor: 'rgba(255,255,255,0.1)',
    thTextColor: 'rgba(255,255,255,0.9)',
    tdTextColor: 'rgba(255,255,255,0.8)'
  }
}

const loading = ref(false)
const historyData = ref([])
const filters = ref({
  deviceId: null,
  teamId: null,
  dateRange: null
})

const deviceOptions = ref([])
const teamOptions = ref([])

// 表格列定义
const columns = [
  {
    title: '设备',
    key: 'device_id',
    width: 80,
    render: (row) => `${row.device_id}号`
  },
  {
    title: '班组',
    key: 'team',
    width: 120,
    render: (row) => row.team?.team_name || '-'
  },
  {
    title: '班组长',
    key: 'leader',
    width: 100,
    render: (row) => row.team?.leader_name || '-'
  },
  {
    title: '当班人员',
    key: 'staff_ids',
    width: 200,
    render: (row) => {
      try {
        const staffIds = JSON.parse(row.staff_ids)
        return h(
          NSpace,
          { size: 'small' },
          {
            default: () => staffIds.map(id => 
              h(NTag, { type: 'info', size: 'small' }, { default: () => getStaffName(id) })
            )
          }
        )
      } catch {
        return '-'
      }
    }
  },
  {
    title: '入班时间',
    key: 'login_time',
    width: 150,
    render: (row) => formatDateTime(row.login_time)
  },
  {
    title: '出班时间',
    key: 'logout_time',
    width: 150,
    render: (row) => row.logout_time ? formatDateTime(row.logout_time) : h(
      NTag,
      { type: 'success' },
      { default: () => '进行中' }
    )
  },
  {
    title: '时长',
    key: 'duration_min',
    width: 100,
    render: (row) => {
      if (row.logout_time) {
        return formatDuration(row.duration_min)
      } else {
        const start = new Date(row.login_time)
        const now = new Date()
        const minutes = Math.floor((now - start) / (1000 * 60))
        return formatDuration(minutes)
      }
    }
  }
]

const pagination = {
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  showQuickJumper: true
}

const allStaff = ref([])

// 加载员工数据
const loadStaff = async () => {
  console.log('📋 加载员工数据...')
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllStaff(null, null)
      console.log('✅ 员工数据加载成功:', result?.length, '人')
      allStaff.value = result || []
    }
  } catch (e) {
    console.error('❌ 加载员工失败:', e)
  }
}

// 加载设备数据
const loadDevices = async () => {
  console.log('🖥️ 加载设备数据...')
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const devs = await window.go.main.App.GetAllDevices()
      console.log('✅ 设备数据加载成功:', devs)
      deviceOptions.value = devs.map(d => ({
        label: `${d.device_name} (${d.device_code})`,
        value: d.id
      }))
      console.log('设备选项:', deviceOptions.value)
    }
  } catch (e) {
    console.error('❌ 加载设备失败:', e)
  }
}

// 加载班组数据
const loadTeams = async () => {
  console.log('👥 加载班组数据...')
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const teams = await window.go.main.App.GetAllTeams(null)
      console.log('✅ 班组数据加载成功:', teams)
      teamOptions.value = teams.map(t => ({
        label: t.team_name,
        value: t.id
      }))
      console.log('班组选项:', teamOptions.value)
    }
  } catch (e) {
    console.error('❌ 加载班组失败:', e)
  }
}

// 加载历史记录
const loadHistory = async () => {
  loading.value = true
  console.log('开始加载历史记录...')
  
  try {
    if (window.go && window.go.main && window.go.main.App) {
      let startDate = null
      let endDate = null
      
      if (filters.value.dateRange && filters.value.dateRange.length === 2) {
        startDate = new Date(filters.value.dateRange[0]).toISOString().split('T')[0]
        endDate = new Date(filters.value.dateRange[1]).toISOString().split('T')[0]
      }
      
      console.log('查询参数:', {
        deviceId: filters.value.deviceId || null,
        teamId: filters.value.teamId || null,
        startDate,
        endDate
      })
      
      const history = await window.go.main.App.GetSessionHistory(
        filters.value.deviceId || null,
        filters.value.teamId || null,
        startDate || "",
        endDate || ""
      )
      
      console.log('加载到的历史记录:', history)
      historyData.value = history || []
    }
  } catch (e) {
    console.error('加载历史记录失败:', e)
    alert('加载失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 获取员工姓名
const getStaffName = (staffId) => {
  const staff = allStaff.value.find(s => s.id === staffId)
  return staff ? staff.name : `员工${staffId}`
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

// 格式化时长
const formatDuration = (minutes) => {
  if (!minutes) return '0分钟'
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return hours > 0 ? `${hours}小时${mins}分钟` : `${mins}分钟`
}

onMounted(async () => {
  await loadStaff()
  await loadDevices()
  await loadTeams()
  await loadHistory()
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

.modal-container {
  background: rgba(20, 30, 48, 0.98);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 95%;
  max-width: 1200px;
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
  position: relative;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: flex-end;
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

.modal-btn.confirm {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.modal-btn.confirm:hover {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  transform: translateY(-2px);
}

.filter-section {
  margin-bottom: 20px;
  padding: 16px;
  background: rgba(255,255,255,0.03);
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.1);
}

/* 开发中遮罩层 */
.development-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(10, 14, 39, 0.85);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  border-radius: 16px;
}

.development-content {
  text-align: center;
  padding: 40px;
  max-width: 400px;
}

.development-icon {
  font-size: 64px;
  color: #667eea;
  margin-bottom: 24px;
  animation: rotate 3s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.development-content h3 {
  font-size: 28px;
  color: #fff;
  margin: 0 0 16px 0;
  font-weight: 600;
}

.development-content p {
  font-size: 16px;
  color: rgba(255,255,255,0.8);
  margin: 0 0 32px 0;
  line-height: 1.6;
}

.development-progress {
  margin-top: 24px;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: rgba(255,255,255,0.1);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 12px;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 4px;
  transition: width 0.3s ease;
  animation: progress-glow 2s ease-in-out infinite alternate;
}

@keyframes progress-glow {
  from { box-shadow: 0 0 10px rgba(102, 126, 234, 0.5); }
  to { box-shadow: 0 0 20px rgba(102, 126, 234, 0.8); }
}

.progress-text {
  font-size: 14px;
  color: rgba(255,255,255,0.7);
  font-weight: 500;
}

/* 模糊内容 */
.blurred-content {
  filter: blur(3px);
  opacity: 0.9;
  pointer-events: none;
  user-select: none;
}

</style>

