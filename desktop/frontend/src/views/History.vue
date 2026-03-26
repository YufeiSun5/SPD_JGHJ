<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-history"></i>
          历史数据查询
        </div>
        <div class="page-subtitle">History Data Query</div>
      </div>
      <div class="header-actions">
        <button 
          class="action-btn refresh" 
          @click="loadTags" 
          title="刷新变量列表"
        >
          <i class="fas fa-sync-alt"></i> 刷新变量
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card card-total">
        <div class="stat-icon">
          <i class="fas fa-database"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ totalRecords }}</div>
          <div class="stat-label">查询记录数</div>
        </div>
      </div>
      <div class="stat-card card-tags">
        <div class="stat-icon">
          <i class="fas fa-tags"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ availableTags.length }}</div>
          <div class="stat-label">可用变量数</div>
        </div>
      </div>
      <div class="stat-card card-selected">
        <div class="stat-icon">
          <i class="fas fa-check-square"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ selectedTagIds.length }}</div>
          <div class="stat-label">已选变量数</div>
        </div>
      </div>
      <div class="stat-card card-time">
        <div class="stat-icon">
          <i class="fas fa-clock"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ timeRangeDays }}</div>
          <div class="stat-label">查询天数</div>
        </div>
      </div>
    </div>
    
    <!-- 查询条件卡片 -->
    <div class="card">
      <div class="card-header">
        <h3><i class="fas fa-filter"></i> 查询条件</h3>
      </div>
      
      <div class="card-content">
        <!-- 时间选择区域 -->
        <div class="filter-section">
          <div class="filter-label">
            <i class="fas fa-calendar-alt"></i>
            <span>时间范围</span>
          </div>
          <n-space :size="12" align="center">
            <n-date-picker 
              v-model:value="startTimeStamp"
              type="datetime"
              clearable
              style="width: 200px"
              placeholder="开始时间"
            />
            <span style="color: rgba(255,255,255,0.6);">至</span>
            <n-date-picker 
              v-model:value="endTimeStamp"
              type="datetime"
              clearable
              style="width: 200px"
              placeholder="结束时间"
            />
            <button class="action-btn secondary" @click="setToday">
              <i class="fas fa-calendar-day"></i> 今天
            </button>
            <button class="action-btn secondary" @click="setYesterday">
              <i class="fas fa-calendar-minus"></i> 昨天
            </button>
            <button class="action-btn secondary" @click="setLast7Days">
              <i class="fas fa-calendar-week"></i> 最近7天
            </button>
          </n-space>
        </div>
        
        <!-- 变量选择区域 -->
        <div class="filter-section">
          <div class="filter-label">
            <i class="fas fa-tags"></i>
            <span>选择变量</span>
            <div style="margin-left: auto; display: flex; gap: 8px;">
              <button class="action-btn secondary" @click="selectAllTags" style="padding: 6px 12px; font-size: 13px;">
                <i class="fas fa-check-square"></i> 全选
              </button>
              <button class="action-btn secondary" @click="clearAllTags" style="padding: 6px 12px; font-size: 13px;">
                <i class="fas fa-square"></i> 清空
              </button>
            </div>
          </div>
          <div class="tags-container">
            <n-checkbox-group v-model:value="selectedTagIds">
              <n-space :size="[12, 12]">
                <div 
                  v-for="tag in availableTags" 
                  :key="tag.var_id"
                  class="tag-checkbox-wrapper"
                >
                  <n-checkbox 
                    :value="tag.var_id"
                    :label="`${tag.display_name || tag.var_name}${tag.unit ? ' (' + tag.unit + ')' : ''}`"
                  />
                </div>
              </n-space>
            </n-checkbox-group>
            <div v-if="availableTags.length === 0" class="empty-tags">
              <i class="fas fa-info-circle"></i>
              <span>暂无可用变量，请先配置存储变量</span>
            </div>
          </div>
          <div class="selected-info">
            <span>已选择 </span>
            <span class="count-badge">{{ selectedTagIds.length }}</span>
            <span> / {{ availableTags.length }} 个变量</span>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="action-buttons">
          <button 
            class="action-btn primary"
            @click="handleQueryClick" 
            :disabled="selectedTagIds.length === 0 || loading"
          >
            <i class="fas fa-search"></i>
            <span v-if="!loading">查询历史数据</span>
            <span v-else>查询中...</span>
          </button>
          <button 
            class="action-btn success"
            @click="exportData" 
            :disabled="totalRecords === 0 || loading"
          >
            <i class="fas fa-file-excel"></i> 导出CSV
          </button>
        </div>
      </div>
    </div>

    
    <!-- 数据表格卡片 -->
    <div class="card" v-if="historyData.length > 0">
      <div class="card-header">
        <h3>
          <i class="fas fa-table"></i> 查询结果
          <span class="badge-count">{{ totalRecords }} 条</span>
        </h3>
        <div style="flex: 1;"></div>
        <n-pagination 
          v-model:page="currentPage"
          :page-count="totalPages"
          :page-size="pageSize"
          show-size-picker
          :page-sizes="[50, 100, 200, 500]"
          @update:page-size="handlePageSizeChange"
        />
      </div>
      
      <div class="table-container">
        <n-data-table
          :columns="tableColumns"
          :data="displayData"
          :max-height="500"
          :striped="true"
          size="small"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h, watch } from 'vue'
import { GetAllTags, GetHistoryData } from '../../wailsjs/go/main/App'

// 获取今天的开始时间（00:00:00）
function getTodayStart() {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return today.getTime()
}

const startTimeStamp = ref(getTodayStart())
const endTimeStamp = ref(Date.now())
const availableTags = ref([])
const selectedTagIds = ref([])
const historyData = ref([])
const loading = ref(false)
const hasQueried = ref(false)

// 分页相关
const currentPage = ref(1)
const pageSize = ref(100)
const totalRecords = ref(0) // 总记录数

// 计算时间范围天数
const timeRangeDays = computed(() => {
  if (!startTimeStamp.value || !endTimeStamp.value) return 0
  const days = Math.ceil((endTimeStamp.value - startTimeStamp.value) / (1000 * 60 * 60 * 24))
  return days
})

onMounted(async () => {
  await loadTags()
  // 加载完标签后，默认查询当天数据
  if (selectedTagIds.value.length > 0) {
    queryHistory()
  }
})

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(totalRecords.value / pageSize.value)
})

// 当前页显示的数据就是 historyData（已经是分页后的数据）
const displayData = computed(() => {
  return historyData.value
})

// 动态生成表格列
const tableColumns = computed(() => {
  const columns = [
    {
      title: '序号',
      key: 'index',
      width: 70,
      fixed: 'left',
      align: 'center',
      render: (row, index) => {
        const globalIndex = (currentPage.value - 1) * pageSize.value + index + 1
        return h('span', { 
          style: { 
            color: '#9ca3af', 
            fontWeight: '500',
            fontFamily: 'monospace'
          } 
        }, globalIndex)
      }
    },
    {
      title: '时间',
      key: 'timestamp',
      width: 180,
      fixed: 'left',
      render: (row) => h('span', { 
        style: { 
          color: '#667eea', 
          fontFamily: 'monospace',
          fontWeight: '500'
        } 
      }, row.timestamp)
    }
  ]
  
  selectedTagIds.value.forEach(tagId => {
    const tag = availableTags.value.find(t => t.var_id === tagId)
    const tagName = tag ? (tag.display_name || tag.var_name) : `Tag ${tagId}`
    const unit = tag?.unit
    
    columns.push({
      title: unit ? `${tagName} (${unit})` : tagName,
      key: `value_${tagId}`,
      width: 140,
      align: 'right',
      render: (row) => {
        const value = row.values[tagId]
        return h('span', {
          style: {
            color: value !== null && value !== undefined ? '#10b981' : '#6b7280',
            fontWeight: '500',
            fontFamily: 'monospace'
          }
        }, formatValue(value))
      }
    })
  })
  
  return columns
})

// 加载所有可用的标签
async function loadTags() {
  try {
    const tags = await GetAllTags()
    availableTags.value = tags.filter(tag => tag.store_mode > 0)
    // 默认全选所有变量
    if (availableTags.value.length > 0) {
      selectedTagIds.value = availableTags.value.map(tag => tag.var_id)
      window.$message?.success(`已加载 ${availableTags.value.length} 个变量`)
    }
  } catch (error) {
    console.error('加载变量列表失败:', error)
    window.$message?.error('加载变量列表失败: ' + error)
  }
}

// 全选变量
function selectAllTags() {
  selectedTagIds.value = availableTags.value.map(tag => tag.var_id)
}

// 清空选择
function clearAllTags() {
  selectedTagIds.value = []
}

// 设置今天
function setToday() {
  startTimeStamp.value = getTodayStart()
  endTimeStamp.value = Date.now()
}

// 设置昨天
function setYesterday() {
  const yesterday = new Date()
  yesterday.setDate(yesterday.getDate() - 1)
  yesterday.setHours(0, 0, 0, 0)
  startTimeStamp.value = yesterday.getTime()
  
  const yesterdayEnd = new Date(yesterday)
  yesterdayEnd.setHours(23, 59, 59, 999)
  endTimeStamp.value = yesterdayEnd.getTime()
}

// 设置最近7天
function setLast7Days() {
  const sevenDaysAgo = new Date()
  sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7)
  sevenDaysAgo.setHours(0, 0, 0, 0)
  startTimeStamp.value = sevenDaysAgo.getTime()
  endTimeStamp.value = Date.now()
}

// 处理分页大小变化
function handlePageSizeChange(newPageSize) {
  pageSize.value = newPageSize
  currentPage.value = 1 // 重置到第一页
  queryHistory() // 重新查询数据
}

// 监听页码变化，自动查询
watch(currentPage, () => {
  if (hasQueried.value) {
    queryHistory()
  }
})

// 处理查询按钮点击
function handleQueryClick() {
  currentPage.value = 1 // 重置到第一页
  queryHistory()
}

// 查询历史数据（分页）
async function queryHistory() {
  if (selectedTagIds.value.length === 0) {
    window.$message?.warning('请至少选择一个变量')
    return
  }
  
  loading.value = true
  hasQueried.value = true
  
  try {
    const start = formatTimestamp(startTimeStamp.value)
    const end = formatTimestamp(endTimeStamp.value)
    
    // 为每个选中的变量查询历史数据（分页）
    const promises = selectedTagIds.value.map(async (tagId) => {
      try {
        const response = await GetHistoryData(tagId, start, end, currentPage.value, pageSize.value)
        return {
          tagId,
          records: response.records || [],
          total: response.total || 0
        }
      } catch (error) {
        console.error(`查询变量 ${tagId} 历史数据失败:`, error)
        return {
          tagId,
          records: [],
          total: 0
        }
      }
    })
    
    const results = await Promise.all(promises)
    
    // 使用第一个变量的总数作为总记录数（假设所有变量的记录数相同）
    // 如果需要更精确，可以取最大值
    totalRecords.value = results.length > 0 ? Math.max(...results.map(r => r.total)) : 0
    
    // 合并数据 - 以时间为主轴
    const timeMap = new Map()
    
    results.forEach(result => {
      result.records.forEach(record => {
        const timestamp = record.timestamp
        if (!timeMap.has(timestamp)) {
          timeMap.set(timestamp, {
            timestamp,
            values: {}
          })
        }
        timeMap.get(timestamp).values[result.tagId] = record.value
      })
    })
    
    // 转换为数组并按时间倒序排序
    historyData.value = Array.from(timeMap.values()).sort((a, b) => {
      return new Date(b.timestamp) - new Date(a.timestamp)
    })
    
    if (currentPage.value === 1) {
      // 只在第一页时显示总记录数提示
      if (totalRecords.value === 0) {
        window.$message?.warning('未查询到历史数据，请确认时间范围和变量配置')
      } else {
        window.$message?.success(`查询成功，共 ${totalRecords.value} 条记录`)
      }
    }
    
  } catch (error) {
    console.error('查询历史数据失败:', error)
    window.$message?.error('查询失败: ' + error)
  } finally {
    loading.value = false
  }
}

// 格式化时间戳
function formatTimestamp(timestamp) {
  const date = new Date(timestamp)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 格式化值
function formatValue(value) {
  if (value === null || value === undefined) {
    return '-'
  }
  if (typeof value === 'number') {
    return value.toFixed(2)
  }
  return value
}

// 导出CSV（导出所有数据，不仅仅是当前页）
async function exportData() {
  if (totalRecords.value === 0) {
    window.$message?.warning('没有数据可导出')
    return
  }
  
  // 如果数据量很大，提示用户
  if (totalRecords.value > 10000) {
    const confirmed = confirm(`即将导出 ${totalRecords.value} 条记录，数据量较大，可能需要一些时间。是否继续？`)
    if (!confirmed) return
  }
  
  loading.value = true
  
  try {
    const start = formatTimestamp(startTimeStamp.value)
    const end = formatTimestamp(endTimeStamp.value)
    
    // 获取所有数据（不分页）
    const promises = selectedTagIds.value.map(async (tagId) => {
      try {
        // 使用一个很大的 pageSize 来获取所有数据
        const response = await GetHistoryData(tagId, start, end, 1, 1000000)
        return {
          tagId,
          records: response.records || []
        }
      } catch (error) {
        console.error(`查询变量 ${tagId} 历史数据失败:`, error)
        return {
          tagId,
          records: []
        }
      }
    })
    
    const results = await Promise.all(promises)
    
    // 合并数据 - 以时间为主轴
    const timeMap = new Map()
    
    results.forEach(result => {
      result.records.forEach(record => {
        const timestamp = record.timestamp
        if (!timeMap.has(timestamp)) {
          timeMap.set(timestamp, {
            timestamp,
            values: {}
          })
        }
        timeMap.get(timestamp).values[result.tagId] = record.value
      })
    })
    
    // 转换为数组并按时间倒序排序
    const allData = Array.from(timeMap.values()).sort((a, b) => {
      return new Date(b.timestamp) - new Date(a.timestamp)
    })
    
    // 构建CSV内容
    let csv = '\uFEFF' // UTF-8 BOM
    
    // 表头
    const headers = ['序号', '时间']
    selectedTagIds.value.forEach(tagId => {
      const tag = availableTags.value.find(t => t.var_id === tagId)
      const name = tag ? (tag.display_name || tag.var_name) : `Tag ${tagId}`
      const unit = tag?.unit
      headers.push(unit ? `${name}(${unit})` : name)
    })
    csv += headers.join(',') + '\n'
    
    // 数据行
    allData.forEach((row, index) => {
      const values = [index + 1, row.timestamp]
      selectedTagIds.value.forEach(tagId => {
        values.push(formatValue(row.values[tagId]))
      })
      csv += values.join(',') + '\n'
    })
    
    // 下载文件
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    const timestamp = formatTimestamp(Date.now()).replace(/[:\s]/g, '-')
    link.download = `历史数据_${timestamp}.csv`
    link.click()
    
    window.$message?.success(`导出成功，共 ${allData.length} 条记录`)
  } catch (error) {
    console.error('导出失败:', error)
    window.$message?.error('导出失败: ' + error)
  } finally {
    loading.value = false
  }
}
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

/* 统计卡片 - 与其他页面完全一致 */
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

.stat-card.card-tags {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

.stat-card.card-selected {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
}

.stat-card.card-time {
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

/* 卡片头部 */
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

/* 卡片内容 */
.card-content {
  padding: 24px;
}

/* 筛选区域 */
.filter-section {
  margin-bottom: 24px;
}

.filter-section:last-child {
  margin-bottom: 0;
}

.filter-label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  font-weight: 600;
}

.filter-label i {
  color: #546e7a;
  font-size: 16px;
}

/* 变量选择容器 */
.tags-container {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 16px;
  min-height: 100px;
  max-height: 220px;
  overflow-y: auto;
}

.tag-checkbox-wrapper {
  background: rgba(255, 255, 255, 0.05);
  padding: 8px 12px;
  border-radius: 6px;
  transition: all 0.2s;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.tag-checkbox-wrapper:hover {
  background: rgba(84, 110, 122, 0.2);
  border-color: rgba(84, 110, 122, 0.4);
}

/* 修复复选框文字颜色 */
:deep(.n-checkbox) {
  color: rgba(255, 255, 255, 0.9) !important;
}

:deep(.n-checkbox__label) {
  color: rgba(255, 255, 255, 0.9) !important;
  font-size: 14px;
  font-weight: 500;
}

.empty-tags {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.4);
  font-size: 14px;
  padding: 30px;
}

.empty-tags i {
  font-size: 16px;
  color: #546e7a;
}

/* 已选择信息 */
.selected-info {
  margin-top: 12px;
  color: rgba(255, 255, 255, 0.7);
  font-size: 13px;
}

.count-badge {
  display: inline-block;
  background: rgba(84, 110, 122, 0.3);
  color: #fff;
  padding: 2px 10px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 14px;
  margin: 0 4px;
}

/* 操作按钮组 */
.action-buttons {
  display: flex;
  gap: 12px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

/* 表格容器 */
.table-container {
  overflow-x: auto;
  margin-top: 20px;
}

/* 表格样式 - 与其他页面统一 */
:deep(.n-data-table) {
  background: transparent;
}

:deep(.n-data-table-th) {
  background: rgba(84, 110, 122, 0.1) !important;
  color: rgba(255,255,255,0.9) !important;
  font-weight: 600;
  border-bottom: 2px solid rgba(84, 110, 122, 0.3) !important;
  font-size: 13px;
}

:deep(.n-data-table-td) {
  background: transparent !important;
  color: rgba(255,255,255,0.8) !important;
  border-bottom: 1px solid rgba(255,255,255,0.05) !important;
  font-size: 13px;
}

:deep(.n-data-table-tr):hover .n-data-table-td {
  background: rgba(84, 110, 122, 0.08) !important;
}

:deep(.n-data-table-tr--striped) .n-data-table-td {
  background: rgba(255, 255, 255, 0.02) !important;
}

/* 分页样式 - 与其他页面统一 */
:deep(.n-pagination) {
  margin-top: 0;
  justify-content: flex-end;
}

:deep(.n-pagination-item) {
  background: rgba(30, 40, 60, 0.6) !important;
  border: 1px solid rgba(255,255,255,0.1) !important;
  color: rgba(255,255,255,0.8) !important;
}

:deep(.n-pagination-item--active) {
  background: rgba(84, 110, 122, 0.4) !important;
  border-color: rgba(84, 110, 122, 0.5) !important;
  color: #fff !important;
}

:deep(.n-pagination-item:not(.n-pagination-item--disabled)):hover {
  background: rgba(84, 110, 122, 0.3) !important;
  border-color: rgba(84, 110, 122, 0.4) !important;
}
</style>
