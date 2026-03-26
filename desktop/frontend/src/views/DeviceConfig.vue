<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-wrench"></i>
          采集配置管理
        </div>
        <div class="page-subtitle">Data Acquisition Configuration</div>
      </div>
      <div class="header-actions">
        <button class="action-btn secondary" @click="loadVariables">
          <i class="fas fa-sync-alt" :class="{ 'fa-spin': loading }"></i> 
          刷新
        </button>
        <button 
          class="action-btn danger" 
          @click="batchDelete" 
          :disabled="selectedIds.size === 0"
        >
          <i class="fas fa-trash"></i> 
          批量删除
          <span v-if="selectedIds.size > 0" class="badge-count">{{ selectedIds.size }}</span>
        </button>
        <button class="action-btn success" @click="addNewVariable">
          <i class="fas fa-plus"></i> 
          新增变量
        </button>
        <button 
          class="action-btn primary" 
          @click="saveAll" 
          :disabled="saving || !hasChanges"
        >
          <i class="fas fa-save"></i> 
          保存所有修改
          <span v-if="hasChanges" class="badge-count">{{ modifiedIds.size }}</span>
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
          placeholder="搜索变量名称、显示名称..."
        />
      </div>
      <div class="filter-group">
        <label><i class="fas fa-cogs"></i> 设备筛选：</label>
        <div class="custom-select">
          <select v-model="filterDeviceId">
            <option :value="null">全部设备</option>
            <option v-for="device in devices" :key="device.id" :value="device.id">
              {{ device.device_name }} (ID:{{ device.id }})
            </option>
          </select>
        </div>
      </div>
      <div class="filter-group">
        <label><i class="fas fa-filter"></i> 数据类型：</label>
        <div class="custom-select">
          <select v-model="filterDataType">
            <option :value="null">全部类型</option>
            <option value="BOOL">BOOL (布尔)</option>
            <option value="INT16">INT16 (短整型)</option>
            <option value="INT32">INT32 (整型)</option>
            <option value="INT64">INT64 (长整型)</option>
            <option value="FLOAT">FLOAT (浮点)</option>
            <option value="DOUBLE">DOUBLE (双精度)</option>
            <option value="STRING">STRING (字符串)</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 配置表格 -->
    <div class="card">
      <div class="card-header">
        <h3><i class="fas fa-list"></i> 变量配置列表</h3>
        <span class="badge">{{ filteredVariables.length }} 个变量</span>
      </div>
      
      <div class="table-wrapper">
        <table class="config-table">
          <thead>
            <tr>
              <th width="50">
                <label class="checkbox-wrapper">
                  <input 
                    type="checkbox" 
                    @change="toggleSelectAll"
                    :checked="isAllSelected"
                  />
                </label>
              </th>
              <th width="50">ID</th>
              <th width="80">设备ID</th>
              <th width="150">变量名 *</th>
              <th width="150">显示名称</th>
              <th width="250">JSON路径 *</th>
              <th width="100">数据类型</th>
              <th width="80">读写</th>
              <th width="80">单位</th>
              <th width="80">缩放</th>
              <th width="80">偏移</th>
              <th width="100">存储模式</th>
              <th width="100">存储周期(秒)</th>
              <th width="100">存储死区</th>
              <th width="80">报警</th>
              <th width="100">上上限(HH)</th>
              <th width="100">上限(H)</th>
              <th width="100">下限(L)</th>
              <th width="100">下下限(LL)</th>
              <th width="100">死区</th>
              <th width="150">报警消息</th>
              <th width="100">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredVariables.length === 0">
              <td colspan="22" class="empty-row">
                <i class="fas fa-inbox"></i>
                <p>暂无配置数据</p>
              </td>
            </tr>
            <tr 
              v-for="row in filteredVariables" 
              :key="row.ID"
              :class="{ 
                'modified-row': modifiedIds.has(row.ID),
                'new-row': row.ID < 0
              }"
            >
              <td>
                <label class="checkbox-wrapper">
                  <input 
                    type="checkbox" 
                    :checked="selectedIds.has(row.ID)"
                    @change="toggleSelect(row.ID)"
                    :disabled="row.ID < 0"
                  />
                </label>
              </td>
              <td>{{ row.ID > 0 ? row.ID : '新增' }}</td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.DeviceID" 
                  @input="markModified(row.ID)"
                  class="input-field mini"
                />
              </td>
              <td>
                <input 
                  type="text" 
                  v-model="row.VarName" 
                  @input="markModified(row.ID)"
                  placeholder="变量名"
                  class="input-field"
                  required
                />
              </td>
              <td>
                <input 
                  type="text" 
                  v-model="row.DisplayName" 
                  @input="markModified(row.ID)"
                  placeholder="显示名称"
                  class="input-field"
                />
              </td>
              <td>
                <input 
                  type="text" 
                  v-model="row.JSONPath" 
                  @input="markModified(row.ID)"
                  placeholder="JSON路径"
                  class="input-field"
                  required
                />
              </td>
              <td>
                <select 
                  v-model="row.DataType" 
                  @change="markModified(row.ID)"
                  class="select-field"
                >
                  <option value="BOOL">BOOL</option>
                  <option value="INT16">INT16</option>
                  <option value="INT32">INT32</option>
                  <option value="INT64">INT64</option>
                  <option value="FLOAT">FLOAT</option>
                  <option value="DOUBLE">DOUBLE</option>
                  <option value="STRING">STRING</option>
                </select>
              </td>
              <td>
                <select 
                  v-model="row.RWMode" 
                  @change="markModified(row.ID)"
                  class="select-field mini"
                >
                  <option value="R">R</option>
                  <option value="W">W</option>
                  <option value="RW">RW</option>
                </select>
              </td>
              <td>
                <input 
                  type="text" 
                  v-model="row.Unit" 
                  @input="markModified(row.ID)"
                  placeholder="单位"
                  class="input-field mini"
                />
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.ScaleFactor" 
                  @input="markModified(row.ID)"
                  step="0.01"
                  class="input-field mini"
                />
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.OffsetVal" 
                  @input="markModified(row.ID)"
                  step="0.01"
                  class="input-field mini"
                />
              </td>
              <td>
                <select 
                  v-model.number="row.StoreMode" 
                  @change="markModified(row.ID)"
                  class="select-field"
                >
                  <option :value="0">不存储</option>
                  <option :value="1">变化存储</option>
                  <option :value="2">定时存储</option>
                  <option :value="3">混合存储</option>
                </select>
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.StoreCycle" 
                  @input="markModified(row.ID)"
                  :disabled="row.StoreMode === 0 || row.StoreMode === 1"
                  class="input-field mini"
                  min="0"
                />
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.StoreDeadband" 
                  @input="markModified(row.ID)"
                  :disabled="row.StoreMode === 0 || row.StoreMode === 2"
                  step="0.01"
                  class="input-field mini"
                  min="0"
                />
              </td>
              <td>
                <label class="switch">
                  <input 
                    type="checkbox" 
                    v-model="row.AlarmEnable" 
                    @change="markModified(row.ID)"
                  />
                  <span class="slider"></span>
                </label>
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.LimitHH" 
                  @input="markModified(row.ID)"
                  :disabled="!row.AlarmEnable"
                  step="0.01"
                  class="input-field mini"
                />
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.LimitH" 
                  @input="markModified(row.ID)"
                  :disabled="!row.AlarmEnable"
                  step="0.01"
                  class="input-field mini"
                />
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.LimitL" 
                  @input="markModified(row.ID)"
                  :disabled="!row.AlarmEnable"
                  step="0.01"
                  class="input-field mini"
                />
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.LimitLL" 
                  @input="markModified(row.ID)"
                  :disabled="!row.AlarmEnable"
                  step="0.01"
                  class="input-field mini"
                />
              </td>
              <td>
                <input 
                  type="number" 
                  v-model.number="row.Deadband" 
                  @input="markModified(row.ID)"
                  :disabled="!row.AlarmEnable"
                  step="0.01"
                  class="input-field mini"
                  min="0"
                />
              </td>
              <td>
                <input 
                  type="text" 
                  v-model="row.AlarmMsg" 
                  @input="markModified(row.ID)"
                  :disabled="!row.AlarmEnable"
                  placeholder="报警消息"
                  class="input-field"
                />
              </td>
              <td>
                <button 
                  v-if="row.ID < 0"
                  class="table-btn save"
                  @click="saveNewVariable(row)"
                  title="保存新变量"
                >
                  <i class="fas fa-check"></i>
                </button>
                <button 
                  v-if="row.ID < 0"
                  class="table-btn cancel"
                  @click="cancelNewVariable(row.ID)"
                  title="取消新增"
                >
                  <i class="fas fa-times"></i>
                </button>
                <button 
                  v-if="row.ID > 0"
                  class="table-btn delete"
                  @click="deleteVariable(row.ID)"
                  title="删除"
                >
                  <i class="fas fa-trash"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const variables = ref([])
const devices = ref([])
const searchText = ref('')
const filterDeviceId = ref(null)
const filterDataType = ref(null)
const loading = ref(false)
const saving = ref(false)
const modifiedIds = ref(new Set())
const originalData = ref({})
const selectedIds = ref(new Set())
let nextTempId = -1 // 用于新增行的临时ID

// 加载设备列表
const loadDevices = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllDevices()
      devices.value = result || []
    }
  } catch (e) {
    console.error('加载设备列表失败:', e)
  }
}

// 加载变量配置
const loadVariables = async () => {
  try {
    loading.value = true
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllVariables()
      variables.value = result || []
      
      // 保存原始数据用于比较
      originalData.value = {}
      result.forEach(v => {
        originalData.value[v.ID] = JSON.parse(JSON.stringify(v))
      })
      
      // 清空修改标记
      modifiedIds.value.clear()
    }
  } catch (e) {
    console.error('加载变量配置失败:', e)
    alert('加载失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 过滤后的变量
const filteredVariables = computed(() => {
  let result = variables.value
  
  // 设备筛选
  if (filterDeviceId.value !== null) {
    result = result.filter(v => v.DeviceID === filterDeviceId.value)
  }
  
  // 数据类型筛选
  if (filterDataType.value) {
    result = result.filter(v => v.DataType === filterDataType.value)
  }
  
  // 搜索过滤
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    result = result.filter(v => 
      (v.VarName && v.VarName.toLowerCase().includes(search)) ||
      (v.DisplayName && v.DisplayName.toLowerCase().includes(search))
    )
  }
  
  return result
})

// 是否有修改
const hasChanges = computed(() => modifiedIds.value.size > 0)

// 是否全选
const isAllSelected = computed(() => {
  const selectableRows = filteredVariables.value.filter(v => v.ID > 0)
  return selectableRows.length > 0 && selectableRows.every(v => selectedIds.value.has(v.ID))
})

// 标记修改
const markModified = (id) => {
  modifiedIds.value.add(id)
}

// 切换选择
const toggleSelect = (id) => {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id)
  } else {
    selectedIds.value.add(id)
  }
}

// 全选/取消全选
const toggleSelectAll = () => {
  const selectableRows = filteredVariables.value.filter(v => v.ID > 0)
  if (isAllSelected.value) {
    selectableRows.forEach(v => selectedIds.value.delete(v.ID))
  } else {
    selectableRows.forEach(v => selectedIds.value.add(v.ID))
  }
}

// 保存所有修改
const saveAll = async () => {
  if (!hasChanges.value) return
  
  try {
    saving.value = true
    
    // 收集修改的变量（只保存已存在的，不包括新增的）
    const modifiedVars = variables.value.filter(v => v.ID > 0 && modifiedIds.value.has(v.ID))
    
    if (!window.go || !window.go.main || !window.go.main.App) {
      throw new Error('Wails API 不可用')
    }
    
    // 批量保存
    await window.go.main.App.BatchUpdateVariables(modifiedVars)
    
    alert(`成功保存 ${modifiedVars.length} 个变量配置`)
    
    // 重新加载数据
    await loadVariables()
    
  } catch (error) {
    console.error('保存失败:', error)
    alert('保存失败: ' + error)
  } finally {
    saving.value = false
  }
}

// 新增变量
const addNewVariable = () => {
  const newVar = {
    ID: nextTempId--,
    DeviceID: filterDeviceId.value || 1,
    VarName: '',
    DisplayName: '',
    JSONPath: '',
    DataType: 'FLOAT',
    RWMode: 'R',
    Unit: '',
    ScaleFactor: 1.0,
    OffsetVal: 0,
    AlarmEnable: false,
    LimitHH: null,
    LimitH: null,
    LimitL: null,
    LimitLL: null,
    Deadband: null,
    AlarmMsg: '',
    StoreMode: 0,
    StoreCycle: 0,
    StoreDeadband: 0
  }
  
  variables.value.unshift(newVar)
  markModified(newVar.ID)
}

// 保存新变量
const saveNewVariable = async (row) => {
  // 验证必填字段
  if (!row.VarName || !row.VarName.trim()) {
    alert('变量名不能为空')
    return
  }
  if (!row.JSONPath || !row.JSONPath.trim()) {
    alert('JSON路径不能为空')
    return
  }
  
  try {
    saving.value = true
    
    if (!window.go || !window.go.main || !window.go.main.App) {
      throw new Error('Wails API 不可用')
    }
    
    // 创建新变量（不包含ID字段，让数据库自动生成）
    const newVarData = {
      DeviceID: row.DeviceID,
      VarName: row.VarName,
      DisplayName: row.DisplayName,
      JSONPath: row.JSONPath,
      DataType: row.DataType,
      RWMode: row.RWMode,
      Unit: row.Unit,
      ScaleFactor: row.ScaleFactor,
      OffsetVal: row.OffsetVal,
      AlarmEnable: row.AlarmEnable,
      LimitHH: row.LimitHH,
      LimitH: row.LimitH,
      LimitL: row.LimitL,
      LimitLL: row.LimitLL,
      Deadband: row.Deadband,
      AlarmMsg: row.AlarmMsg,
      StoreMode: row.StoreMode,
      StoreCycle: row.StoreCycle,
      StoreDeadband: row.StoreDeadband
    }
    
    await window.go.main.App.CreateVariable(newVarData)
    
    alert('成功创建新变量')
    
    // 重新加载数据
    await loadVariables()
    
  } catch (error) {
    console.error('创建失败:', error)
    alert('创建失败: ' + error)
  } finally {
    saving.value = false
  }
}

// 取消新增
const cancelNewVariable = (id) => {
  variables.value = variables.value.filter(v => v.ID !== id)
  modifiedIds.value.delete(id)
}

// 删除单个变量
const deleteVariable = async (id) => {
  if (!confirm('确定要删除这个变量吗？')) {
    return
  }
  
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      throw new Error('Wails API 不可用')
    }
    
    await window.go.main.App.DeleteVariable(id)
    
    alert('删除成功')
    
    // 重新加载数据
    await loadVariables()
    
  } catch (error) {
    console.error('删除失败:', error)
    alert('删除失败: ' + error)
  }
}

// 批量删除
const batchDelete = async () => {
  if (selectedIds.value.size === 0) {
    alert('请先选择要删除的变量')
    return
  }
  
  if (!confirm(`确定要删除选中的 ${selectedIds.value.size} 个变量吗？`)) {
    return
  }
  
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      throw new Error('Wails API 不可用')
    }
    
    const idsArray = Array.from(selectedIds.value)
    await window.go.main.App.BatchDeleteVariables(idsArray)
    
    alert(`成功删除 ${idsArray.length} 个变量`)
    
    // 清空选择
    selectedIds.value.clear()
    
    // 重新加载数据
    await loadVariables()
    
  } catch (error) {
    console.error('批量删除失败:', error)
    alert('批量删除失败: ' + error)
  }
}

// 页面加载时初始化
onMounted(() => {
  loadDevices()
  loadVariables()
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
  position: relative;
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

.action-btn.danger {
  background: linear-gradient(135deg, #ee5a6f 0%, #f29263 100%);
  color: white;
}

.action-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}

.action-btn.success:hover:not(:disabled) {
  box-shadow: 0 4px 15px rgba(56, 239, 125, 0.4);
}

.action-btn.danger:hover:not(:disabled) {
  box-shadow: 0 4px 15px rgba(238, 90, 111, 0.4);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.badge-count {
  background: #ee5a6f;
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: bold;
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

.custom-select {
  position: relative;
}

.custom-select select {
  padding: 8px 32px 8px 12px;
  background: rgba(30, 40, 60, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  min-width: 150px;
}

.custom-select select:focus {
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

/* 表格容器 */
.table-wrapper {
  overflow-x: auto;
  overflow-y: auto;
  max-height: 600px;
  border-radius: 8px;
  background: rgba(20, 30, 48, 0.4);
}

/* 配置表格 */
.config-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 13px;
}

.config-table thead {
  position: sticky;
  top: 0;
  z-index: 10;
  background: rgba(102, 126, 234, 0.15);
}

.config-table th {
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  border-bottom: 2px solid rgba(102, 126, 234, 0.3);
  white-space: nowrap;
}

.config-table tbody tr {
  border-bottom: 1px solid rgba(255,255,255,0.05);
  transition: all 0.2s;
}

.config-table tbody tr:hover {
  background: rgba(102, 126, 234, 0.08);
}

.config-table tbody tr.modified-row {
  background: rgba(247, 151, 30, 0.15);
  border-left: 3px solid #f7971e;
}

.config-table tbody tr.new-row {
  background: rgba(56, 239, 125, 0.1);
  border-left: 3px solid #38ef7d;
}

.config-table td {
  padding: 8px;
  color: rgba(255,255,255,0.8);
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

/* 表单控件 */
.input-field, .select-field {
  width: 100%;
  padding: 6px 8px;
  background: rgba(30, 40, 60, 0.8);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 4px;
  color: #fff;
  font-size: 13px;
}

.input-field.mini, .select-field.mini {
  max-width: 70px;
}

.input-field:focus, .select-field:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(30, 40, 60, 0.95);
}

.input-field:disabled, .select-field:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 开关按钮 */
.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 22px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255,255,255,0.2);
  transition: 0.3s;
  border-radius: 22px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .slider {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

input:checked + .slider:before {
  transform: translateX(22px);
}

/* 复选框 */
.checkbox-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
}

.checkbox-wrapper input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #667eea;
}

/* 表格操作按钮 */
.table-btn {
  padding: 4px 8px;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  margin: 0 2px;
}

.table-btn.save {
  background: rgba(56, 239, 125, 0.2);
  color: #38ef7d;
  border: 1px solid #38ef7d;
}

.table-btn.save:hover {
  background: rgba(56, 239, 125, 0.3);
}

.table-btn.cancel {
  background: rgba(255, 152, 0, 0.2);
  color: #ff9800;
  border: 1px solid #ff9800;
}

.table-btn.cancel:hover {
  background: rgba(255, 152, 0, 0.3);
}

.table-btn.delete {
  background: rgba(238, 90, 111, 0.2);
  color: #ee5a6f;
  border: 1px solid #ee5a6f;
}

.table-btn.delete:hover {
  background: rgba(238, 90, 111, 0.3);
}

/* 响应式 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .filter-bar {
    flex-direction: column;
  }
  
  .search-box {
    width: 100%;
  }
}
</style>
