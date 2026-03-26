<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-tasks"></i>
          任务管理
        </div>
        <div class="page-subtitle">Task Management - 定时任务、数据改变任务、条件事件任务</div>
      </div>
      <div class="header-actions">
        <button class="action-btn secondary" @click="loadTasks">
          <i class="fas fa-sync-alt" :class="{ 'fa-spin': loading }"></i> 
          刷新
        </button>
        <button class="action-btn success" @click="showAddDialog">
          <i class="fas fa-plus"></i> 
          新增任务
        </button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-group search-box">
        <i class="fas fa-search"></i>
        <input 
          v-model="searchText" 
          type="text" 
          placeholder="搜索任务名称、描述..."
        />
      </div>
      <div class="filter-group">
        <label><i class="fas fa-filter"></i> 任务类型：</label>
        <select v-model="filterType">
          <option :value="null">全部类型</option>
          <option :value="1">定时任务</option>
          <option :value="2">数据改变任务</option>
          <option :value="3">条件事件任务</option>
        </select>
      </div>
      <div class="filter-group">
        <label><i class="fas fa-toggle-on"></i> 状态：</label>
        <select v-model="filterEnabled">
          <option :value="null">全部状态</option>
          <option :value="1">已启用</option>
          <option :value="0">已禁用</option>
        </select>
      </div>
    </div>

    <!-- 任务列表 -->
    <div class="card">
      <div class="card-header">
        <h3><i class="fas fa-list"></i> 任务列表</h3>
        <span class="badge">{{ filteredTasks.length }} 个任务</span>
      </div>
      
      <div class="table-wrapper">
        <table class="task-table">
          <thead>
            <tr>
              <th width="60">ID</th>
              <th width="200">任务名称</th>
              <th width="120">类型</th>
              <th width="80">状态</th>
              <th width="180">触发条件</th>
              <th width="100">变化类型</th>
              <th width="100">动作类型</th>
              <th width="250">描述</th>
              <th width="150">更新时间</th>
              <th width="180">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredTasks.length === 0">
              <td colspan="10" class="empty-row">
                <i class="fas fa-inbox"></i>
                <p>暂无任务数据</p>
              </td>
            </tr>
            <tr v-for="task in filteredTasks" :key="task.task_id">
              <td>{{ task.task_id }}</td>
              <td class="task-name">{{ task.task_name }}</td>
              <td>
                <span class="type-badge" :class="getTypeClass(task.task_type)">
                  {{ getTypeName(task.task_type) }}
                  <span v-if="task.task_type === 3 && task.interval_sec" class="interval-hint">
                    ({{ task.interval_sec }}秒)
                  </span>
                </span>
              </td>
              <td>
                <span class="status-badge" :class="task.is_enabled ? 'enabled' : 'disabled'">
                  {{ task.is_enabled ? '启用' : '禁用' }}
                </span>
              </td>
              <td class="var-name">
                <span v-if="task.task_type === 3 && task.condition_expr" class="condition-expr" :title="task.condition_expr">
                  {{ truncateCondition(task.condition_expr) }}
                </span>
                <span v-else>{{ task.trigger_var_name || '-' }}</span>
              </td>
              <td>
                <span class="change-type-badge" :class="getChangeTypeClass(task.change_type)">
                  {{ getChangeTypeName(task.change_type) }}
                </span>
              </td>
              <td>{{ getActionName(task.action_type) }}</td>
              <td class="description">{{ task.description || '-' }}</td>
              <td>{{ formatTime(task.updated_at) }}</td>
              <td>
                <div class="action-buttons">
                  <button 
                    class="table-btn edit" 
                    @click="showEditDialog(task)"
                    title="编辑"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button 
                    v-if="task.is_enabled"
                    class="table-btn warning" 
                    @click="toggleTaskStatus(task)"
                    title="禁用"
                  >
                    <i class="fas fa-pause"></i>
                  </button>
                  <button 
                    v-else
                    class="table-btn success" 
                    @click="toggleTaskStatus(task)"
                    title="启用"
                  >
                    <i class="fas fa-play"></i>
                  </button>
                  <button 
                    class="table-btn delete" 
                    @click="deleteTask(task)"
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

    <!-- 编辑/新增对话框 -->
    <div v-if="showDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{{ isEditing ? '编辑任务' : '新增任务' }}</h3>
          <button class="close-btn" @click="closeDialog">
            <i class="fas fa-times"></i>
          </button>
        </div>
        
        <div class="dialog-body">
          <div class="form-row">
            <div class="form-group">
              <label>任务名称 *</label>
              <input v-model="formData.task_name" type="text" placeholder="例如：设备1系统错误信息-记录系统报警" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>任务类型 *</label>
              <select v-model.number="formData.task_type">
                <option :value="1">定时任务</option>
                <option :value="2">数据改变任务</option>
                <option :value="3">条件事件任务</option>
              </select>
            </div>
            <div class="form-group">
              <label>状态 *</label>
              <select v-model.number="formData.is_enabled">
                <option :value="1">启用</option>
                <option :value="0">禁用</option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group full-width">
              <label>任务描述</label>
              <input v-model="formData.description" type="text" placeholder="例如：当设备1系统错误信息变化时，查询错误码表并记录系统报警" />
            </div>
          </div>

          <!-- 数据改变任务配置 -->
          <div v-if="formData.task_type === 2" class="config-section">
            <h4>数据改变任务配置</h4>
            <div class="form-row">
              <div class="form-group">
                <label>触发变量ID *</label>
                <input v-model.number="formData.trigger_var_id" type="number" placeholder="例如：88" />
              </div>
              <div class="form-group">
                <label>触发变量名称 *</label>
                <input v-model="formData.trigger_var_name" type="text" placeholder="例如：设备1系统错误信息" />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>变化类型 *</label>
                <select v-model="formData.change_type">
                  <option value="ANY">ANY - 任意变化</option>
                  <option value="INCREASE">INCREASE - 增加</option>
                  <option value="DECREASE">DECREASE - 减少</option>
                  <option value="THRESHOLD">THRESHOLD - 阈值</option>
                  <option value="FALSE_TO_TRUE">FALSE_TO_TRUE - 0→1</option>
                  <option value="TRUE_TO_FALSE">TRUE_TO_FALSE - 1→0</option>
                  <option value="QUALITY_GOOD">QUALITY_GOOD - 设备上线 (质量码0→1)</option>
                  <option value="QUALITY_BAD">QUALITY_BAD - 设备离线 (质量码1→0)</option>
                </select>
              </div>
              <div class="form-group" v-if="formData.change_type === 'THRESHOLD'">
                <label>变化阈值</label>
                <input v-model.number="formData.change_threshold" type="number" step="0.01" />
              </div>
            </div>
            
            <!-- 质量码监控说明 -->
            <div v-if="formData.change_type === 'QUALITY_GOOD' || formData.change_type === 'QUALITY_BAD'" class="info-box">
              <i class="fas fa-info-circle"></i>
              <div>
                <strong>质量码监控说明：</strong>
                <ul>
                  <li><strong>QUALITY_GOOD (设备上线)</strong>：当变量的质量码从 0 变为 1 时触发（设备从离线恢复在线）</li>
                  <li><strong>QUALITY_BAD (设备离线)</strong>：当变量的质量码从 1 变为 0 时触发（设备从在线变为离线）</li>
                  <li>质量码 1 = 在线（MQTT质量码192），质量码 0 = 离线（MQTT质量码非192）</li>
                </ul>
              </div>
            </div>
          </div>

          <!-- 定时任务配置 -->
          <div v-if="formData.task_type === 1" class="config-section">
            <h4>定时任务配置</h4>
            <div class="form-row">
              <div class="form-group">
                <label>间隔时间(秒)</label>
                <input v-model.number="formData.interval_sec" type="number" placeholder="例如：300 (5分钟)" />
              </div>
              <div class="form-group">
                <label>Cron表达式</label>
                <input v-model="formData.cron_expr" type="text" placeholder="例如：*/5 * * * *" />
              </div>
            </div>
          </div>

          <!-- 条件事件任务配置 -->
          <div v-if="formData.task_type === 3" class="config-section">
            <h4>条件事件任务配置</h4>
            <div class="form-row">
              <div class="form-group full-width">
                <label>条件表达式 *</label>
                <input v-model="formData.condition_expr" type="text" placeholder="例如：temp>50 AND pressure>100" />
              </div>
            </div>
          </div>

          <!-- 动作配置 -->
          <div class="config-section">
            <h4>动作配置</h4>
            <div class="form-row">
              <div class="form-group">
                <label>动作类型 *</label>
                <select v-model.number="formData.action_type">
                  <option :value="1">HTTP请求</option>
                  <option :value="2">MQTT发布</option>
                  <option :value="3">数据库操作</option>
                  <option :value="4">执行脚本</option>
                  <option :value="5">写日志</option>
                </select>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group full-width">
                <label>动作配置 (JSON) *</label>
                <textarea 
                  v-model="formData.action_config" 
                  rows="10" 
                  :placeholder="getActionConfigPlaceholder()"
                ></textarea>
              </div>
            </div>
            
            <!-- 动作配置说明 -->
            <div class="info-box" v-if="formData.action_type === 3">
              <i class="fas fa-lightbulb"></i>
              <div>
                <strong>数据库操作类型：</strong>
                <ul>
                  <li><code>update_device_status</code> - 更新设备状态</li>
                  <li><code>update_device_status_conditional</code> - 条件更新设备状态（检查多个变量）</li>
                  <li><code>end_device_status</code> - 结束设备状态</li>
                  <li><code>increment_production_qty</code> - 增加产量</li>
                  <li><code>log_system_alarm</code> - 记录系统报警</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <div class="dialog-footer">
          <button class="btn secondary" @click="closeDialog">取消</button>
          <button class="btn primary" @click="saveTask">
            <i class="fas fa-save"></i> 保存
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const tasks = ref([])
const loading = ref(false)
const searchText = ref('')
const filterType = ref(null)
const filterEnabled = ref(null)
const showDialog = ref(false)
const isEditing = ref(false)
const formData = ref({
  task_id: null,
  task_name: '',
  task_type: 2,
  is_enabled: 1,
  description: '',
  trigger_var_id: null,
  trigger_var_name: '',
  change_type: 'ANY',
  change_threshold: null,
  condition_expr: '',
  interval_sec: null,
  cron_expr: '',
  action_type: 3,
  action_config: ''
})

// 过滤后的任务列表
const filteredTasks = computed(() => {
  let result = tasks.value

  if (filterType.value !== null) {
    result = result.filter(t => t.task_type === filterType.value)
  }

  if (filterEnabled.value !== null) {
    result = result.filter(t => t.is_enabled === filterEnabled.value)
  }

  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    result = result.filter(t => 
      (t.task_name && t.task_name.toLowerCase().includes(search)) ||
      (t.description && t.description.toLowerCase().includes(search))
    )
  }

  return result
})

// 加载任务列表
const loadTasks = async () => {
  try {
    loading.value = true
    if (window.go?.main?.App?.GetAllTasks) {
      const result = await window.go.main.App.GetAllTasks()
      tasks.value = result || []
    }
  } catch (e) {
    console.error('加载任务失败:', e)
    alert('加载任务失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 获取类型名称
const getTypeName = (type) => {
  const names = { 1: '定时', 2: '数据改变', 3: '条件事件' }
  return names[type] || '未知'
}

// 获取类型样式
const getTypeClass = (type) => {
  const classes = { 1: 'type-scheduled', 2: 'type-data', 3: 'type-condition' }
  return classes[type] || ''
}

// 获取动作名称
const getActionName = (type) => {
  const names = { 1: 'HTTP', 2: 'MQTT', 3: '数据库', 4: '脚本', 5: '日志' }
  return names[type] || '未知'
}

// 获取变化类型名称
const getChangeTypeName = (type) => {
  const names = {
    'ANY': '任意变化',
    'INCREASE': '增加',
    'DECREASE': '减少',
    'THRESHOLD': '阈值',
    'FALSE_TO_TRUE': '0→1',
    'TRUE_TO_FALSE': '1→0',
    'QUALITY_GOOD': '设备上线',
    'QUALITY_BAD': '设备离线'
  }
  return names[type] || type || '-'
}

// 获取变化类型样式
const getChangeTypeClass = (type) => {
  if (type === 'QUALITY_GOOD') return 'quality-good'
  if (type === 'QUALITY_BAD') return 'quality-bad'
  if (type === 'FALSE_TO_TRUE') return 'true'
  if (type === 'TRUE_TO_FALSE') return 'false'
  return ''
}

// 截断条件表达式（用于显示）
const truncateCondition = (expr) => {
  if (!expr) return '-'
  if (expr.length <= 30) return expr
  return expr.substring(0, 30) + '...'
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN', { 
    year: 'numeric', 
    month: '2-digit', 
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 获取动作配置的占位符文本
const getActionConfigPlaceholder = () => {
  const actionType = formData.value.action_type
  
  if (actionType === 3) { // 数据库操作
    return `示例1 - 更新设备状态：
{
  "operation": "update_device_status",
  "op_params": {
    "device_id": 1,
    "status": 1,
    "remark": "设备开机"
  }
}

示例2 - 条件更新设备状态（空闲检测）：
{
  "operation": "update_device_status_conditional",
  "op_params": {
    "device_id": 1,
    "status": 3,
    "remark": "#1设备进入空闲状态",
    "conditions": {
      "开机状态_var_id": 76,
      "开机状态_value": 1,
      "开机状态_quality": 1,
      "停机状态_var_id": 73,
      "停机状态_value": 1
    }
  }
}

示例3 - 记录系统报警：
{
  "operation": "log_system_alarm",
  "op_params": {
    "var_id": 88,
    "var_name": "#1焊机错误信息",
    "error_code": "{{new_value}}"
  }
}`
  } else if (actionType === 1) { // HTTP请求
    return `{
  "method": "POST",
  "url": "http://localhost:8080/api/devices/update",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "device_id": 1,
    "value": "{{new_value}}"
  }
}`
  } else if (actionType === 2) { // MQTT发布
    return `{
  "topic": "device/status",
  "payload": {
    "device_id": 1,
    "status": "{{new_value}}"
  },
  "qos": 1,
  "retained": false
}`
  }
  
  return '请输入JSON格式的配置'
}

// 显示新增对话框
const showAddDialog = () => {
  isEditing.value = false
  formData.value = {
    task_id: null,
    task_name: '',
    task_type: 2,
    is_enabled: 1,
    description: '',
    trigger_var_id: null,
    trigger_var_name: '',
    change_type: 'ANY',
    change_threshold: null,
    condition_expr: '',
    interval_sec: null,
    cron_expr: '',
    action_type: 3,
    action_config: ''
  }
  showDialog.value = true
}

// 显示编辑对话框
const showEditDialog = (task) => {
  isEditing.value = true
  formData.value = { ...task }
  showDialog.value = true
}

// 关闭对话框
const closeDialog = () => {
  showDialog.value = false
}

// 保存任务
const saveTask = async () => {
  // 验证必填字段
  if (!formData.value.task_name) {
    alert('请输入任务名称')
    return
  }

  // 验证JSON格式
  try {
    JSON.parse(formData.value.action_config)
  } catch (e) {
    alert('动作配置JSON格式错误: ' + e.message)
    return
  }

  try {
    if (!window.go?.main?.App) {
      throw new Error('Wails API 不可用')
    }

    if (isEditing.value) {
      // 更新任务
      await window.go.main.App.UpdateTask(formData.value.task_id, formData.value)
      alert('任务更新成功')
    } else {
      // 创建任务
      await window.go.main.App.CreateTask(formData.value)
      alert('任务创建成功')
    }

    closeDialog()
    await loadTasks()
  } catch (error) {
    console.error('保存任务失败:', error)
    alert('保存任务失败: ' + error)
  }
}

// 切换任务状态
const toggleTaskStatus = async (task) => {
  try {
    if (!window.go?.main?.App) {
      throw new Error('Wails API 不可用')
    }

    const newStatus = task.is_enabled ? 0 : 1
    if (newStatus === 1) {
      await window.go.main.App.EnableTask(task.task_id)
    } else {
      await window.go.main.App.DisableTask(task.task_id)
    }

    alert(`任务已${newStatus ? '启用' : '禁用'}`)
    await loadTasks()
  } catch (error) {
    console.error('切换任务状态失败:', error)
    alert('切换任务状态失败: ' + error)
  }
}

// 删除任务
const deleteTask = async (task) => {
  if (!confirm(`确定要删除任务"${task.task_name}"吗？`)) {
    return
  }

  try {
    if (!window.go?.main?.App) {
      throw new Error('Wails API 不可用')
    }

    await window.go.main.App.DeleteTask(task.task_id)
    alert('任务删除成功')
    await loadTasks()
  } catch (error) {
    console.error('删除任务失败:', error)
    alert('删除任务失败: ' + error)
  }
}

onMounted(() => {
  loadTasks()
})
</script>

<style scoped>
/* 页面容器 */
.page-container {
  padding: 40px;
  min-height: 100vh;
}

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

.action-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.action-btn.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.action-btn.secondary {
  background: rgba(255,255,255,0.1);
  color: white;
}

.action-btn.success {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: white;
}

.action-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}

/* 筛选栏 */
.filter-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-group label {
  color: rgba(255,255,255,0.8);
  font-size: 14px;
  white-space: nowrap;
}

.filter-group label i {
  margin-right: 4px;
  color: #667eea;
}

.search-box {
  flex: 1;
  min-width: 300px;
  position: relative;
}

.search-box i {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255,255,255,0.5);
}

.search-box input {
  width: 100%;
  padding: 10px 12px 10px 38px;
  background: rgba(30, 40, 60, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
}

.search-box input:focus {
  outline: none;
  border-color: #667eea;
}

.filter-group select {
  padding: 8px 32px 8px 12px;
  background: rgba(30, 40, 60, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  min-width: 150px;
}

.filter-group select:focus {
  outline: none;
  border-color: #667eea;
}

/* 卡片 */
.card {
  background: rgba(30, 40, 60, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 24px;
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
  color: #667eea;
}

.badge {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}

/* 表格 */
.table-wrapper {
  overflow-x: auto;
  overflow-y: auto;
  max-height: 600px;
  border-radius: 8px;
  background: rgba(20, 30, 48, 0.4);
}

.task-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 13px;
}

.task-table thead {
  position: sticky;
  top: 0;
  z-index: 10;
  background: rgba(102, 126, 234, 0.15);
}

.task-table th {
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  border-bottom: 2px solid rgba(102, 126, 234, 0.3);
  white-space: nowrap;
}

.task-table tbody tr {
  border-bottom: 1px solid rgba(255,255,255,0.05);
  transition: all 0.2s;
}

.task-table tbody tr:hover {
  background: rgba(102, 126, 234, 0.08);
}

.task-table td {
  padding: 12px 8px;
  color: rgba(255,255,255,0.8);
}

.task-name {
  font-weight: 500;
  color: #fff;
}

.var-name {
  color: #00ffaa;
}

.condition-expr {
  color: #667eea;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
}

.interval-hint {
  font-size: 10px;
  opacity: 0.7;
  margin-left: 4px;
}

.description {
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

/* 徽章 */
.type-badge, .status-badge, .change-type-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
}

.type-badge.type-scheduled {
  background: rgba(255, 152, 0, 0.2);
  color: #ff9800;
}

.type-badge.type-data {
  background: rgba(56, 239, 125, 0.2);
  color: #38ef7d;
}

.type-badge.type-condition {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
}

.status-badge.enabled {
  background: rgba(56, 239, 125, 0.2);
  color: #38ef7d;
}

.status-badge.disabled {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
}

.change-type-badge {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
}

.change-type-badge.quality-good {
  background: rgba(56, 239, 125, 0.2);
  color: #38ef7d;
}

.change-type-badge.quality-bad {
  background: rgba(238, 90, 111, 0.2);
  color: #ee5a6f;
}

.change-type-badge.true {
  background: rgba(56, 239, 125, 0.15);
  color: #38ef7d;
}

.change-type-badge.false {
  background: rgba(238, 90, 111, 0.15);
  color: #ee5a6f;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 4px;
}

.table-btn {
  padding: 6px 10px;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.table-btn:hover {
  transform: translateY(-1px);
}

.table-btn.edit {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
}

.table-btn.edit:hover {
  background: rgba(102, 126, 234, 0.3);
}

.table-btn.success {
  background: rgba(56, 239, 125, 0.2);
  color: #38ef7d;
}

.table-btn.success:hover {
  background: rgba(56, 239, 125, 0.3);
}

.table-btn.warning {
  background: rgba(255, 152, 0, 0.2);
  color: #ff9800;
}

.table-btn.warning:hover {
  background: rgba(255, 152, 0, 0.3);
}

.table-btn.delete {
  background: rgba(238, 90, 111, 0.2);
  color: #ee5a6f;
}

.table-btn.delete:hover {
  background: rgba(238, 90, 111, 0.3);
}

/* 对话框 */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: rgba(30, 40, 60, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 12px;
  width: 800px;
  max-width: 90vw;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  border: 1px solid rgba(255,255,255,0.1);
}

.dialog-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-header h3 {
  margin: 0;
  color: #fff;
  font-size: 18px;
}

.close-btn {
  background: none;
  border: none;
  color: rgba(255,255,255,0.6);
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
  transition: all 0.2s;
}

.close-btn:hover {
  color: #fff;
}

.dialog-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.config-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid rgba(255,255,255,0.1);
}

.config-section h4 {
  margin: 0 0 16px 0;
  color: #667eea;
  font-size: 14px;
}

.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.form-group {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group.full-width {
  flex: 1 1 100%;
}

.form-group label {
  color: rgba(255,255,255,0.8);
  font-size: 13px;
  font-weight: 500;
}

.form-group input,
.form-group select,
.form-group textarea {
  padding: 10px 12px;
  background: rgba(20, 30, 48, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 6px;
  color: #fff;
  font-size: 13px;
  font-family: 'Consolas', 'Monaco', monospace;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #667eea;
}

.form-group textarea {
  resize: vertical;
  min-height: 120px;
}

.dialog-footer {
  padding: 16px 24px;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn.secondary {
  background: rgba(255,255,255,0.1);
  color: white;
}

.btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

/* 信息提示框 */
.info-box {
  background: rgba(102, 126, 234, 0.1);
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 8px;
  padding: 12px 16px;
  margin-top: 12px;
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.info-box i {
  color: #667eea;
  font-size: 16px;
  margin-top: 2px;
  flex-shrink: 0;
}

.info-box strong {
  color: #667eea;
  display: block;
  margin-bottom: 8px;
}

.info-box ul {
  margin: 0;
  padding-left: 20px;
  color: rgba(255, 255, 255, 0.8);
  font-size: 12px;
  line-height: 1.6;
}

.info-box li {
  margin-bottom: 4px;
}

.info-box code {
  background: rgba(0, 0, 0, 0.3);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', monospace;
  color: #00ffaa;
  font-size: 11px;
}
</style>

