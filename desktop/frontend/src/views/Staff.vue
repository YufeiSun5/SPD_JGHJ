<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-users"></i>
          人员管理
        </div>
        <div class="page-subtitle">Staff Management System</div>
      </div>
      <div class="header-actions">
        <button 
          class="action-btn warning" 
          @click="showShiftModal" 
          title="设备班次登记"
        >
          <i class="fas fa-clipboard-check"></i> 
          班次登记
        </button>
        <button 
          class="action-btn info" 
          @click="showActiveDevices" 
          title="查看在线设备"
        >
          <i class="fas fa-desktop"></i> 
          在线设备 <span class="badge-mini">{{ activeSessions.length }}</span>
        </button>
        <button class="action-btn secondary" @click="showSessionHistory" title="班次历史">
          <i class="fas fa-history"></i> 班次历史
        </button>
        <button class="action-btn info" @click="showTeamModal('add')" title="班组管理">
          <i class="fas fa-layer-group"></i> 班组管理
        </button>
        <button class="action-btn primary" @click="showStaffModal('add')" title="新增员工">
          <i class="fas fa-user-plus"></i> 新增员工
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card card-total">
        <div class="stat-icon">
          <i class="fas fa-users"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.total }}</div>
          <div class="stat-label">员工总数</div>
        </div>
      </div>
      <div class="stat-card card-active">
        <div class="stat-icon">
          <i class="fas fa-user-check"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.active }}</div>
          <div class="stat-label">在职人数</div>
        </div>
      </div>
      <div class="stat-card card-teams">
        <div class="stat-icon">
          <i class="fas fa-layer-group"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ teams.length }}</div>
          <div class="stat-label">班组数量</div>
        </div>
      </div>
      <div class="stat-card card-devices">
        <div class="stat-icon">
          <i class="fas fa-desktop"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ activeSessions.length }}</div>
          <div class="stat-label">在线设备</div>
        </div>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-group">
        <label><i class="fas fa-filter"></i> 班组筛选：</label>
        <div class="custom-select">
          <select v-model="filters.teamId" @change="loadStaff">
            <option :value="null">全部班组</option>
            <option v-for="team in teams" :key="team.id" :value="team.id">
              {{ team.team_name }}
            </option>
          </select>
        </div>
      </div>
      <div class="filter-group">
        <label><i class="fas fa-user-check"></i> 在职状态：</label>
        <div class="custom-select">
          <select v-model="filters.isActive" @change="loadStaff">
            <option :value="null">全部</option>
            <option :value="1">在职</option>
            <option :value="0">离职</option>
          </select>
        </div>
      </div>
      <div class="filter-group search-box">
        <i class="fas fa-search"></i>
        <input 
          v-model="searchKeyword" 
          type="text" 
          placeholder="搜索员工姓名或工号..."
          @input="filterStaff"
        />
      </div>
    </div>

    <!-- 员工列表 -->
    <div class="card">
      <div class="card-header">
        <h3><i class="fas fa-list"></i> 员工列表</h3>
        <span class="badge">{{ filteredStaff.length }} 人</span>
      </div>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th width="80">序号</th>
              <th width="120">工号</th>
              <th width="120">姓名</th>
              <th width="150">所属班组</th>
              <th width="150">班组长</th>
              <th width="100">状态</th>
              <th width="180">入职时间</th>
              <th width="200">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredStaff.length === 0">
              <td colspan="8" class="empty-row">
                <i class="fas fa-inbox"></i>
                <p>暂无员工数据</p>
              </td>
            </tr>
            <tr v-for="(staff, index) in filteredStaff" :key="staff.id">
              <td>{{ index + 1 }}</td>
              <td>
                <span class="staff-code">{{ staff.staff_code }}</span>
              </td>
              <td>
                <div class="staff-name">
                  <i class="fas fa-user-circle"></i>
                  {{ staff.name }}
                </div>
              </td>
              <td>
                <span v-if="staff.current_team" class="team-badge">
                  <i class="fas fa-users"></i>
                  {{ staff.current_team.team_name }}
                </span>
                <span v-else class="no-team">未分配</span>
              </td>
              <td>
                <span v-if="staff.current_team?.leader_name" class="leader-name">
                  {{ staff.current_team.leader_name }}
                </span>
                <span v-else class="text-muted">-</span>
              </td>
              <td>
                <span :class="['status-badge', staff.is_active === 1 ? 'active' : 'inactive']">
                  <i :class="staff.is_active === 1 ? 'fas fa-check-circle' : 'fas fa-times-circle'"></i>
                  {{ staff.is_active === 1 ? '在职' : '离职' }}
                </span>
              </td>
              <td>{{ formatDateTime(staff.created_at) }}</td>
              <td>
                <div class="action-buttons">
                  <button 
                    class="table-btn edit" 
                    @click="showStaffModal('edit', staff)" 
                    title="编辑基本信息"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button 
                    class="table-btn delete" 
                    @click="deleteStaff(staff)" 
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

    <!-- 员工新增/编辑弹窗 -->
    <StaffModal
      v-if="staffModal.show"
      :mode="staffModal.mode"
      :staff="staffModal.form"
      :teams="teams"
      @close="closeStaffModal"
      @save="saveStaff"
    />

    <!-- 班组管理弹窗 -->
    <TeamModal
      v-if="teamModal.show"
      :mode="teamModal.mode"
      :team="teamModal.form"
      @close="closeTeamModal"
      @save="saveTeam"
      @refresh="loadTeams"
    />


    <!-- 设备上班弹窗 -->
    <ShiftModal
      v-if="shiftModal.show"
      :teams="teams"
      :active-sessions="activeSessions"
      @close="closeShiftModal"
      @shift="handleShift"
    />

    <!-- 在班设备列表弹窗 -->
    <ActiveDevicesModal
      v-if="activeDevicesModal.show"
      :sessions="activeSessions"
      @close="closeActiveDevicesModal"
      @logout="handleLogout"
    />

    <!-- 班次历史记录弹窗 -->
    <SessionHistoryModal
      v-if="sessionHistoryModal.show"
      @close="closeSessionHistoryModal"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import StaffModal from '../components/StaffModal.vue'
import TeamModal from '../components/TeamModal.vue'
import ShiftModal from '../components/ShiftModal.vue'
import ActiveDevicesModal from '../components/ActiveDevicesModal.vue'
import SessionHistoryModal from '../components/SessionHistoryModal.vue'

const staffList = ref([])
const teams = ref([])
const searchKeyword = ref('')
const activeSessions = ref([]) // 改为数组，支持多设备班次
const filters = ref({
  teamId: null,
  isActive: null
})
const activeDevicesModal = ref({
  show: false
})

const staffModal = ref({
  show: false,
  mode: 'add',
  form: {}
})

const teamModal = ref({
  show: false,
  mode: 'add',
  form: {}
})


const shiftModal = ref({
  show: false
})

const sessionHistoryModal = ref({
  show: false
})

// 统计数据
const stats = computed(() => {
  const total = staffList.value.length
  const active = staffList.value.filter(s => s.is_active === 1).length
  
  return { total, active }
})

// 显示在班设备列表
const showActiveDevices = () => {
  activeDevicesModal.value.show = true
}

// 关闭在班设备弹窗
const closeActiveDevicesModal = () => {
  activeDevicesModal.value.show = false
}

// 筛选后的员工列表
const filteredStaff = computed(() => {
  return staffList.value.filter(staff => {
    if (searchKeyword.value) {
      const keyword = searchKeyword.value.toLowerCase()
      if (!staff.name.toLowerCase().includes(keyword) && 
          !staff.staff_code.toLowerCase().includes(keyword)) {
        return false
      }
    }
    return true
  })
})

// 加载员工列表
const loadStaff = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const teamId = filters.value.teamId || null
      const isActive = filters.value.isActive !== null ? filters.value.isActive : null
      
      const result = await window.go.main.App.GetAllStaff(teamId, isActive)
      staffList.value = result || []
    }
  } catch (e) {
    console.error('加载员工列表失败:', e)
  }
}

// 加载班组列表
const loadTeams = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllTeams(1) // 只加载启用的班组
      teams.value = result || []
    }
  } catch (e) {
    console.error('加载班组列表失败:', e)
  }
}

// 筛选员工
const filterStaff = () => {
  // computed 会自动响应
}

// 显示员工弹窗
const showStaffModal = (mode, staff = null) => {
  staffModal.value.mode = mode
  staffModal.value.form = staff ? { ...staff } : {
    staff_code: '',
    name: '',
    current_team_id: null,
    is_active: 1
  }
  staffModal.value.show = true
}

// 关闭员工弹窗
const closeStaffModal = () => {
  staffModal.value.show = false
}

// 保存员工
const saveStaff = async (staff) => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      if (staffModal.value.mode === 'add') {
        await window.go.main.App.CreateStaff(
          staff.staff_code,
          staff.name,
          staff.current_team_id || null
        )
      } else {
        await window.go.main.App.UpdateStaff(
          staff.id,
          staff.name || null,
          staff.current_team_id !== undefined ? staff.current_team_id : null,
          staff.is_active !== undefined ? staff.is_active : null
        )
      }
      closeStaffModal()
      await loadStaff()
    }
  } catch (e) {
    console.error('保存员工失败:', e)
    alert('保存失败: ' + e)
  }
}

// 删除员工
const deleteStaff = async (staff) => {
  if (confirm(`确定要删除员工 ${staff.name}（${staff.staff_code}）吗？`)) {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        await window.go.main.App.DeleteStaff(staff.id)
        await loadStaff()
      }
    } catch (e) {
      console.error('删除员工失败:', e)
      alert('删除失败: ' + e)
    }
  }
}

// 显示班组弹窗
const showTeamModal = (mode, team = null) => {
  teamModal.value.mode = mode
  teamModal.value.form = team ? { ...team } : {
    team_name: '',
    leader_name: '',
    status: 1
  }
  teamModal.value.show = true
}

// 关闭班组弹窗
const closeTeamModal = () => {
  teamModal.value.show = false
}

// 保存班组
const saveTeam = async (team) => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      if (teamModal.value.mode === 'add') {
        await window.go.main.App.CreateTeam(
          team.team_name,
          team.leader_name || null
        )
      } else {
        await window.go.main.App.UpdateTeam(
          team.id,
          team.team_name || null,
          team.leader_name || null,
          team.status !== undefined ? team.status : null
        )
      }
      closeTeamModal()
      await loadTeams()
      await loadStaff() // 刷新员工列表以更新关联的班组信息
    }
  } catch (e) {
    console.error('保存班组失败:', e)
    alert('保存失败: ' + e)
  }
}


// 显示班次历史记录
const showSessionHistory = () => {
  sessionHistoryModal.value.show = true
}

// 关闭班次历史记录弹窗
const closeSessionHistoryModal = () => {
  sessionHistoryModal.value.show = false
}

// 加载所有活动班次（支持多设备）
const loadActiveSessions = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const sessions = await window.go.main.App.GetAllActiveSessions()
      activeSessions.value = sessions || []
      console.log('当前活动班次:', activeSessions.value.length, '个')
    }
  } catch (e) {
    // 没有活动班次是正常的
    activeSessions.value = []
    console.log('暂无活动班次')
  }
}

// 显示换班弹窗
const showShiftModal = () => {
  shiftModal.value.show = true
}

// 关闭换班弹窗
const closeShiftModal = () => {
  shiftModal.value.show = false
}

// 处理换班/上班
const handleShift = async (data) => {
  console.log('🔥 handleShift 被调用，收到数据:', data)
  
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      console.error('❌ Wails Go 对象未加载')
      alert('系统未就绪，请稍后重试')
      return
    }
    
    console.log('✅ Wails Go 对象已加载')
    
    if (window.go && window.go.main && window.go.main.App) {
      // 1. 检查该设备是否已有活动班次，如果有则先下班
      const existingSession = activeSessions.value.find(s => s.device_id === data.device_id)
      if (existingSession) {
        console.log('📤 设备', data.device_id, '已有活动班次，正在下班:', existingSession.id)
        await window.go.main.App.DeviceLogout(data.device_id)
        console.log('✅ 下班成功')
      }
      
      // 2. 更新员工班组归属
      console.log('🔍 [换班] 开始处理员工班组归属...')
      
      // 2.1 重新加载完整的员工列表（不带过滤）
      let allStaff = []
      try {
        allStaff = await window.go.main.App.GetAllStaff(null, 1) // null=所有班组, 1=在职
        console.log('📊 [换班] 重新加载完整员工列表，共', allStaff.length, '人')
      } catch (e) {
        console.error('❌ [换班] 加载完整员工列表失败:', e)
        allStaff = staffList.value // 降级使用当前列表
      }
      
      // 2.2 获取目标班组的原有成员ID
      const originalTeamMemberIds = allStaff
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
        const staff = allStaff.find(s => s.id === staffId)
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
        const staff = allStaff.find(s => s.id === staffId)
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
      
      const result = await window.go.main.App.DeviceLogin(
        data.device_id,
        data.team_id,
        data.staff_ids
      )
      
      console.log('✅ 登录成功，返回结果:', result)
      
      console.log('🔄 刷新数据...')
      closeShiftModal()
      alert('班次登记成功！')
      
      // 等待一下确保数据库已更新
      await new Promise(resolve => setTimeout(resolve, 500))
      
      await loadActiveSessions()
      console.log('✅ 活动班次已刷新:', activeSessions.value)
      
      await loadStaff() // 刷新员工列表以显示新的班组归属
      console.log('✅ 员工列表已刷新')
      
      console.log('🎉 换班流程完成')
      
      // 生成更详细的提示信息
      let message = '班次登记成功！'
      if (joinCount > 0 || leaveCount > 0) {
        const details = []
        if (joinCount > 0) details.push(`${joinCount}人加入班组`)
        if (leaveCount > 0) details.push(`${leaveCount}人移出班组`)
        message += ` (${details.join('，')})`
      }
      alert(message)
    }
  } catch (e) {
    console.error('❌ 换班失败:', e)
    alert('换班失败: ' + e)
  }
}

// 处理下班（单个设备）
const handleLogout = async (deviceId) => {
  const session = activeSessions.value.find(s => s.device_id === deviceId)
  if (!session) {
    alert('该设备当前没有活动班次')
    return
  }

  // 获取设备名称
  let deviceName = `设备${deviceId}`
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const devices = await window.go.main.App.GetAllDevices()
      const device = devices.find(d => d.id === deviceId)
      if (device) {
        deviceName = device.device_name
      }
    }
  } catch (e) {
    console.error('获取设备名称失败:', e)
  }

  if (!confirm(`确定要结束【${deviceName}】的当前班次吗？`)) {
    return
  }

  try {
    if (window.go && window.go.main && window.go.main.App) {
      await window.go.main.App.DeviceLogout(deviceId)
      await loadActiveSessions()
      closeActiveDevicesModal()
      alert('班次已结束！')
    }
  } catch (e) {
    console.error('结束班次失败:', e)
    alert('结束班次失败: ' + e)
  }
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
    minute: '2-digit'
  })
}

onMounted(async () => {
  await loadTeams()
  await loadStaff()
  await loadActiveSessions()
  
  // 每30秒刷新一次活动班次状态
  setInterval(loadActiveSessions, 30000)
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

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
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

/* 工业低调配色 - 渐变背景 */
.stat-card.card-total {
  background: linear-gradient(135deg, #556070 0%, #6e7e8e 100%);
}

.stat-card.card-active {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

.stat-card.card-teams {
  background: linear-gradient(135deg, #6a6860 0%, #7e7e6e 100%);
}

.stat-card.card-devices {
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

.stat-value-mini {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  line-height: 1.2;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stat-label {
  font-size: 14px;
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

/* 筛选栏 */
.filter-bar {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
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
  border-color: rgba(102, 126, 234, 0.5);
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

.search-box {
  flex: 1;
  max-width: 400px;
  position: relative;
}

.search-box i {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255,255,255,0.5);
  font-size: 14px;
}

.search-box input {
  width: 100%;
  padding: 8px 16px 8px 40px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  transition: all 0.2s;
}

.search-box input:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}

.search-box input::placeholder {
  color: rgba(255,255,255,0.4);
}

/* 卡片 */
.card {
  background: rgba(20, 30, 48, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  overflow: hidden;
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
}

.badge {
  padding: 4px 12px;
  background: rgba(102, 126, 234, 0.2);
  border: 1px solid #667eea;
  border-radius: 12px;
  font-size: 13px;
  color: #667eea;
  font-weight: 600;
}

/* 表格 */
.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.data-table thead {
  background: rgba(102, 126, 234, 0.15);
}

.data-table th {
  padding: 16px;
  text-align: left;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  border-bottom: 2px solid rgba(255,255,255,0.1);
  white-space: nowrap;
}

.data-table td {
  padding: 16px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
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
.staff-code {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: #667eea;
}

.staff-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.staff-name i {
  color: #667eea;
  font-size: 16px;
}

.team-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background: rgba(102, 126, 234, 0.2);
  border: 1px solid rgba(102, 126, 234, 0.4);
  border-radius: 12px;
  font-size: 13px;
  color: #667eea;
}

.no-team {
  color: rgba(255,255,255,0.4);
  font-style: italic;
}

.leader-name {
  color: rgba(255,255,255,0.7);
}

.text-muted {
  color: rgba(255,255,255,0.4);
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
}

.status-badge.inactive {
  background: rgba(231, 76, 60, 0.2);
  color: #e74c3c;
  border: 1px solid #e74c3c;
}

/* 按钮样式统一使用全局样式 naive-theme.css */
.badge-mini {
  background: rgba(255,255,255,0.2);
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  margin-left: 4px;
}

.action-btn.success.pulse {
  animation: pulse-glow 2s infinite;
}


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

.table-btn.warning:hover {
  background: #f39c12;
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
