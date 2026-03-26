<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-bell"></i>
          报警管理
        </div>
        <div class="page-subtitle">Alarm Management System</div>
      </div>
      <div class="header-actions">
        <button class="action-btn refresh" @click="loadData" title="刷新数据">
          <i class="fas fa-sync-alt"></i> 刷新数据
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card card-total">
        <div class="stat-icon">
          <i class="fas fa-bell"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.today_total || 0 }}</div>
          <div class="stat-label">今日报警</div>
        </div>
      </div>
      <div class="stat-card card-acked">
        <div class="stat-icon">
          <i class="fas fa-check-circle"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.today_acked || 0 }}</div>
          <div class="stat-label">已确认</div>
        </div>
      </div>
      <div class="stat-card card-pending">
        <div class="stat-icon">
          <i class="fas fa-exclamation-circle"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.today_pending || 0 }}</div>
          <div class="stat-label">待确认</div>
        </div>
      </div>
      <div class="stat-card card-comparison">
        <div class="stat-icon">
          <i :class="getComparisonIcon()"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value-mini">
            <span v-if="stats.comparison === 'up'" class="trend-up">
              ↑ {{ stats.comparison_value }}
            </span>
            <span v-else-if="stats.comparison === 'down'" class="trend-down">
              ↓ {{ stats.comparison_value }}
            </span>
            <span v-else class="trend-equal">
              - 持平
            </span>
          </div>
          <div class="stat-label">较昨日</div>
        </div>
      </div>
    </div>

    <!-- 报警记录查询 -->
    <div class="card">
      <div class="card-header">
        <h3><i class="fas fa-list"></i> 报警历史记录</h3>
        <span class="badge">{{ filteredAlarms.length }} 条记录</span>
      </div>
      
      <!-- 查询条件 -->
      <n-space :size="12" style="margin-bottom: 20px;">
        <n-select
          v-model:value="alarmFilter.ackStatus"
          :options="ackStatusOptions"
          placeholder="确认状态"
          clearable
          style="width: 150px;"
        >
          <template #prefix>
            <i class="fas fa-filter" style="margin-right: 8px; color: #78909c;"></i>
          </template>
        </n-select>
        
        <n-select
          v-model:value="alarmFilter.alarmType"
          :options="alarmTypeOptions"
          placeholder="报警类型"
          clearable
          style="width: 150px;"
        >
          <template #prefix>
            <i class="fas fa-exclamation-triangle" style="margin-right: 8px; color: #78909c;"></i>
          </template>
        </n-select>
        
        <n-select
          v-model:value="alarmFilter.varID"
          :options="variableOptions"
          placeholder="选择变量"
          clearable
          filterable
          style="width: 250px;"
        >
          <template #prefix>
            <i class="fas fa-tag" style="margin-right: 8px; color: #78909c;"></i>
          </template>
        </n-select>
        
        <n-date-picker
          v-model:value="alarmFilter.startTimeValue"
          type="datetime"
          placeholder="开始时间"
          clearable
          style="width: 200px;"
        />
        
        <n-date-picker
          v-model:value="alarmFilter.endTimeValue"
          type="datetime"
          placeholder="结束时间"
          clearable
          style="width: 200px;"
        />
        
        <n-button type="primary" @click="loadAlarms">
          <template #icon>
            <i class="fas fa-search"></i>
          </template>
          查询
        </n-button>
      </n-space>

      <!-- 报警记录表格 -->
      <n-data-table
        :columns="columns"
        :data="filteredAlarms"
        :loading="loading"
        :pagination="paginationConfig"
        :max-height="600"
        :row-class-name="rowClassName"
        striped
        size="small"
      />
    </div>

    <!-- AI 诊断弹窗 -->
    <div v-if="showAIModal" class="modal-overlay" @click.self="closeAIModal">
      <div class="modal-container">
        <div class="modal-header">
          <h3>
            <i class="fas fa-robot"></i>
            AI 智能诊断
          </h3>
          <button class="modal-close" @click="closeAIModal">
            <i class="fas fa-times"></i>
          </button>
        </div>
        
        <div class="modal-body">
          <!-- 报警信息摘要 -->
          <div v-if="aiDiagnosisData.alarm" class="alarm-summary">
            <div class="summary-header">
              <i class="fas fa-exclamation-triangle"></i>
              报警信息
            </div>
            <div class="summary-grid">
              <div class="summary-item">
                <span class="summary-label">变量名称</span>
                <span class="summary-value">{{ aiDiagnosisData.alarm.var_name }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">当前值</span>
                <span class="summary-value highlight">{{ Math.round(aiDiagnosisData.alarm.val) }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">报警类型</span>
                <span class="summary-value">{{ aiDiagnosisData.alarm.alarm_type }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">限值</span>
                <span class="summary-value">{{ aiDiagnosisData.alarm.limit_value === 0 || aiDiagnosisData.alarm.limit_value === null || aiDiagnosisData.alarm.limit_value === undefined ? '-' : aiDiagnosisData.alarm.limit_value.toFixed(2) }}</span>
              </div>
              <div class="summary-item full-width">
                <span class="summary-label">报警消息</span>
                <span class="summary-value">{{ aiDiagnosisData.alarm.msg }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">持续时长</span>
                <span class="summary-value">{{ aiDiagnosisData.alarm.duration }}</span>
              </div>
            </div>
          </div>

          <!-- AI 分析结果 -->
          <div class="ai-result">
            <div class="result-header">
              <i class="fas fa-lightbulb"></i>
              AI 分析结果
              <span v-if="aiDiagnosisData.fromFeedback" class="feedback-badge" :class="getFeedbackBadgeClass()">
                <span v-html="getFeedbackIcon()"></span> 来自点赞反馈 ({{ aiDiagnosisData.feedbackLikeCount }}人点赞)
              </span>
            </div>

            <!-- 排队提示 -->
            <div v-if="aiDiagnosisData.queued" class="queue-notice">
              <i class="fas fa-hourglass-half"></i>
              <span>当前有其他查询正在处理，您的查询排在第 <strong>{{ aiDiagnosisData.queuePosition }}</strong> 位，请稍候...</span>
            </div>

            <!-- 思考中动画（只在没有答案内容时显示） -->
            <div v-if="aiDiagnosisData.isThinking && !aiDiagnosisData.queued && !aiDiagnosisData.answer" class="thinking-section">
              <div class="thinking-animation">
                <div class="thinking-text">{{ aiDiagnosisData.thinkingText }}</div>
                <div class="thinking-dots">
                  <span></span><span></span><span></span>
                </div>
              </div>
              <button class="stop-btn" @click="stopAIDiagnosis" title="停止生成">
                <i class="fas fa-stop-circle"></i> 停止回答
              </button>
            </div>

            <!-- 错误信息 -->
            <div v-if="aiDiagnosisData.error" class="error-section">
              <i class="fas fa-exclamation-circle"></i>
              {{ aiDiagnosisData.error }}
            </div>

            <!-- AI 回答卡片 -->
            <div v-if="aiDiagnosisData.answer" class="answer-card">
              <div class="card-header-ai">
                <i class="fas fa-robot"></i>
                <span>AI 诊断结果</span>
                <!-- 流式输出时显示停止按钮 -->
                <button v-if="aiDiagnosisData.isThinking" class="stop-btn-inline" @click="stopAIDiagnosis" title="停止生成">
                  <i class="fas fa-stop-circle"></i> 停止
                </button>
              </div>
              <div class="answer-content">
                <div class="answer-text">{{ aiDiagnosisData.answer }}</div>
                
                <!-- 如果是来自优质回答，提供重新生成选项 -->
                <div v-if="aiDiagnosisData.fromFeedback && !aiDiagnosisData.isThinking" class="regenerate-section">
                  <p class="regenerate-hint">💡 这是经过验证的优质答案。如果不满意，可以重新生成个性化回答。</p>
                  <button class="regenerate-btn" @click="regenerateAnswer" :disabled="aiDiagnosisData.regenerating">
                    <i :class="aiDiagnosisData.regenerating ? 'fas fa-spinner fa-spin' : 'fas fa-sync-alt'"></i>
                    {{ aiDiagnosisData.regenerating ? '正在重新生成...' : '重新生成' }}
                  </button>
                </div>
                
                <!-- 如果是重新生成的答案，提供恢复原答案和点赞替换的提示 -->
                <div v-if="aiDiagnosisData.isRegeneratedAnswer && !aiDiagnosisData.isThinking" class="regenerate-section">
                  <p class="regenerate-hint">
                    ✨ 这是重新生成的答案。如果满意，请点赞以替换原优质回答；如果不满意，可以恢复原答案。
                  </p>
                  <div class="action-buttons">
                    <button class="restore-btn" @click="restoreOriginalAnswer">
                      <i class="fas fa-undo"></i>
                      恢复原答案
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- 参考文档卡片 -->
            <div v-if="aiDiagnosisData.relevantDocs.length > 0" class="docs-card">
              <div class="card-header-docs">
                <i class="fas fa-book"></i>
                <span>参考文档 ({{ aiDiagnosisData.relevantDocs.length }})</span>
              </div>
              <div class="docs-list">
                <div 
                  v-for="(doc, idx) in aiDiagnosisData.relevantDocs" 
                  :key="idx"
                  class="doc-item"
                >
                  <div class="doc-header">
                    <span class="doc-source">
                      <i class="fas fa-tag"></i>
                      {{ doc.source || '未知来源' }}
                    </span>
                    <span class="doc-similarity">
                      相似度: {{ (doc.similarity * 100).toFixed(0) }}%
                    </span>
                  </div>
                  <div class="doc-content">{{ doc.content }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <button 
            v-if="aiDiagnosisData.answer && !aiDiagnosisData.isThinking" 
            class="modal-btn like" 
            @click="likeAnswer"
            :disabled="aiDiagnosisData.liked"
            :title="aiDiagnosisData.isRegeneratedAnswer ? '点赞以采纳新答案并替换原优质回答' : (aiDiagnosisData.fromFeedback ? '为这个优质答案点赞，增加点赞数' : '点赞此回答，帮助AI学习')"
          >
            <i :class="aiDiagnosisData.liked ? 'fas fa-thumbs-up' : 'far fa-thumbs-up'"></i>
            {{ aiDiagnosisData.liked ? '已点赞' : (aiDiagnosisData.isRegeneratedAnswer ? '采纳新答案' : '点赞') }}
          </button>
          <button class="modal-btn cancel" @click="closeAIModal">
            <i class="fas fa-times"></i> 关闭
          </button>
        </div>
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, h } from 'vue'
import { NButton, NSpace } from 'naive-ui'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { QueryAIStream, StopAIStream } from '../../wailsjs/go/main/App'
import Toast from '../components/Toast.vue'

const alarmRecords = ref([])
const stats = ref({
  today_total: 0,
  today_acked: 0,
  today_pending: 0,
  yesterday_total: 0,
  comparison: 'equal',
  comparison_value: 0
})
const loading = ref(false)
const alarmFilter = ref({
  ackStatus: null,
  alarmType: null,
  varID: null,
  startTime: '',
  endTime: '',
  startTimeValue: null,
  endTimeValue: null
})
const variableOptions = ref([])
let refreshTimer = null

// Toast 提示
const toast = ref({
  show: false,
  message: '',
  type: 'success',
  duration: 2000
})

// 分页配置 - 使用响应式对象让 n-data-table 自动处理分页
const paginationConfig = reactive({
  page: 1,
  pageSize: 50,
  pageSlot: 7,
  pageSizes: [20, 50, 100, 200],
  showSizePicker: true,
  showQuickJumper: true,
  prefix: (info) => `共 ${info.itemCount} 条记录`,
  onChange: (page) => {
    paginationConfig.page = page
  },
  onUpdatePageSize: (pageSize) => {
    paginationConfig.pageSize = pageSize
    paginationConfig.page = 1
  }
})

// 确认状态选项
const ackStatusOptions = [
  { label: '待确认', value: 0 },
  { label: '已确认', value: 1 }
]

// 报警类型选项
const alarmTypeOptions = [
  { label: '上上限(HH)', value: 'HH' },
  { label: '上限(H)', value: 'H' },
  { label: '下限(L)', value: 'L' },
  { label: '下下限(LL)', value: 'LL' }
]

// 表格列配置
const columns = [
  {
    title: '#',
    key: 'index',
    width: 60,
    render: (row, index) => {
      // n-data-table 会自动处理分页，index 已经是当前页的索引
      // 所以需要加上前面页的数量
      return (paginationConfig.page - 1) * paginationConfig.pageSize + index + 1
    }
  },
  {
    title: '变量名称',
    key: 'var_name',
    width: 150,
    render: (row) => h('div', { class: 'var-name' }, [
      h('i', { class: 'fas fa-tag' }),
      h('strong', row.var_name)
    ])
  },
  {
    title: '报警类型',
    key: 'alarm_type',
    width: 100,
    render: (row) => h('span', {
      class: ['alarm-type-badge', getAlarmTypeClass(row.alarm_type)]
    }, getAlarmTypeName(row.alarm_type))
  },
  {
    title: '触发值',
    key: 'val',
    width: 100,
    render: (row) => h('span', { class: 'value-text' }, Math.round(row.val).toString())
  },
  {
    title: '阈值',
    key: 'limit_value',
    width: 100,
    render: (row) => {
      // 阈值为0或null时显示为空
      if (row.limit_value === 0 || row.limit_value === null || row.limit_value === undefined) {
        return h('span', { class: 'limit-text text-muted' }, '-')
      }
      return h('span', { class: 'limit-text' }, row.limit_value.toFixed(2))
    }
  },
  {
    title: '开始时间',
    key: 'start_time',
    width: 140,
    render: (row) => formatDateTime(row.start_time)
  },
  {
    title: '恢复时间',
    key: 'end_time',
    width: 140,
    render: (row) => h('span', {
      class: row.end_time ? 'recovered' : 'alarm-ongoing'
    }, formatDateTime(row.end_time) || '报警中')
  },
  {
    title: '持续时长',
    key: 'duration',
    width: 100,
    render: (row) => h('span', {
      class: row.end_time ? 'duration-text' : 'duration-ongoing'
    }, row.duration)
  },
  {
    title: '确认状态',
    key: 'ack_status',
    width: 100,
    render: (row) => h('span', {
      class: ['ack-badge', row.ack_status === 1 ? 'acked' : 'pending']
    }, [
      h('i', { class: row.ack_status === 1 ? 'fas fa-check-circle' : 'fas fa-clock' }),
      row.ack_status === 1 ? ' 已确认' : ' 待确认'
    ])
  },
  {
    title: '报警消息',
    key: 'msg',
    width: 250,
    ellipsis: {
      tooltip: true
    },
    render: (row) => h('span', { class: 'msg-text' }, row.msg)
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right',
    render: (row) => {
      if (row.ack_status === 0) {
        return h(NSpace, { size: 'small' }, [
          h(NButton, {
            size: 'small',
            type: 'info',
            onClick: () => showAIDiagnosis(row)
          }, {
            icon: () => h('i', { class: 'fas fa-robot' }),
            default: () => 'AI'
          }),
          h(NButton, {
            size: 'small',
            type: 'warning',
            onClick: () => ackAlarm(row)
          }, {
            icon: () => h('i', { class: 'fas fa-check' }),
            default: () => '确认'
          })
        ])
      } else {
        // 已确认的报警也显示 AI 按钮
        return h(NButton, {
          size: 'small',
          type: 'info',
          onClick: () => showAIDiagnosis(row)
        }, {
          icon: () => h('i', { class: 'fas fa-robot' }),
          default: () => 'AI'
        })
      }
    }
  }
]

// 行类名 - 只有未恢复且未确认的报警才显示红色
const rowClassName = (row) => {
  return (!row.end_time && row.ack_status === 0) ? 'alarm-active' : ''
}

// 加载变量选项
const loadVariableOptions = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetVariableOptions()
      if (result && Array.isArray(result)) {
        variableOptions.value = result.map(v => ({
          label: `${v.gateway_name} - ${v.device_name} - ${v.display_name || v.var_name}`,
          value: v.var_id
        }))
      }
    }
  } catch (e) {
    console.error('加载变量列表失败:', e)
  }
}

// 加载报警统计
const loadStats = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAlarmStats()
      stats.value = result || {}
    }
  } catch (e) {
    console.error('加载统计数据失败:', e)
  }
}

// 加载报警记录
const loadAlarms = async () => {
  try {
    loading.value = true
    console.log('🔍 [报警管理] 开始加载报警记录...')
    
    // 转换时间戳为字符串
    let startTime = alarmFilter.value.startTime
    let endTime = alarmFilter.value.endTime
    
    if (alarmFilter.value.startTimeValue) {
      const date = new Date(alarmFilter.value.startTimeValue)
      startTime = formatDateTimeLocal(date)
    }
    
    if (alarmFilter.value.endTimeValue) {
      const date = new Date(alarmFilter.value.endTimeValue)
      endTime = formatDateTimeLocal(date)
    }
    
    console.log('📅 [报警管理] 查询参数:', {
      ackStatus: alarmFilter.value.ackStatus,
      startTime,
      endTime,
      varID: alarmFilter.value.varID
    })
    
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAlarmRecords(
        alarmFilter.value.ackStatus,
        startTime,
        endTime,
        alarmFilter.value.varID
      )
      
      console.log('📦 [报警管理] 收到原始数据:', result)
      
      // 去重：根据 id 去重，避免数据重复
      const uniqueRecords = []
      const idSet = new Set()
      
      if (result && Array.isArray(result)) {
        result.forEach(record => {
          if (!idSet.has(record.id)) {
            idSet.add(record.id)
            uniqueRecords.push(record)
          }
        })
      }
      
      alarmRecords.value = uniqueRecords
      console.log('✅ [报警管理] 加载完成，共', uniqueRecords.length, '条记录')
    } else {
      console.error('❌ [报警管理] window.go 未加载')
    }
  } catch (e) {
    console.error('❌ [报警管理] 加载报警记录失败:', e)
  } finally {
    loading.value = false
  }
}

// 过滤后的报警记录
const filteredAlarms = computed(() => {
  let result = alarmRecords.value
  
  if (alarmFilter.value.alarmType) {
    result = result.filter(a => a.alarm_type === alarmFilter.value.alarmType)
  }
  
  // 当筛选条件变化时，重置到第一页
  if (result.length > 0 && paginationConfig.page > Math.ceil(result.length / paginationConfig.pageSize)) {
    paginationConfig.page = 1
  }
  
  return result
})

// 显示提示框
const showToast = (message, type = 'success', duration = 2000) => {
  toast.value.message = message
  toast.value.type = type
  toast.value.duration = duration
  toast.value.show = true
}

// 确认报警（无需二次确认，直接确认）
const ackAlarm = async (alarm) => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      await window.go.main.App.AckAlarm(alarm.id)
      showToast(`报警【${alarm.var_name}】已确认`, 'success')
      await loadAlarms()
      await loadStats()
    }
  } catch (e) {
    console.error('确认报警失败:', e)
    showToast('确认失败: ' + e, 'error')
  }
}

// ========================================================
// AI 诊断功能
// ========================================================

const showAIModal = ref(false)
const aiDiagnosisData = reactive({
  alarm: null,
  isThinking: false,
  thinkingText: '正在思考...',
  answer: '',
  relevantDocs: [],
  error: null,
  liked: false,  // 是否已点赞
  queued: false,  // 是否在排队
  queuePosition: 0,  // 排队位置
  fromFeedback: false,  // 是否来自点赞反馈
  feedbackLikeCount: 0,  // 反馈点赞数
  feedbackId: '',  // 反馈ID（用于减少点赞）
  regenerating: false,  // 是否正在重新生成
  originalAnswer: '',  // 保存原始优质回答
  originalDocs: [],  // 保存原始文档
  originalLikeCount: 0,  // 保存原始点赞数
  isRegeneratedAnswer: false  // 当前答案是否为重新生成的
})

// 显示 AI 诊断弹窗
const showAIDiagnosis = (alarm) => {
  // 重置状态
  aiDiagnosisData.alarm = alarm
  aiDiagnosisData.isThinking = true
  aiDiagnosisData.thinkingText = '🤔 正在启动 AI 分析...'
  aiDiagnosisData.answer = ''
  aiDiagnosisData.relevantDocs = []
  aiDiagnosisData.error = null
  aiDiagnosisData.liked = false
  aiDiagnosisData.queued = false
  aiDiagnosisData.queuePosition = 0
  aiDiagnosisData.fromFeedback = false
  aiDiagnosisData.feedbackLikeCount = 0
  aiDiagnosisData.originalAnswer = ''
  aiDiagnosisData.originalDocs = []
  aiDiagnosisData.originalLikeCount = 0
  aiDiagnosisData.isRegeneratedAnswer = false
  
  showAIModal.value = true
  
  // 开始 AI 诊断
  startAIDiagnosis(alarm)
}

// 格式化报警信息为 AI 提示词
const formatAlarmForAI = (alarm) => {
  // 只传递错误码（当前值）和错误信息，用于精确匹配
  return `错误码 ${alarm.val}：${alarm.msg}`
}

// AI 诊断的临时数据
let aiTempRelevantDocs = []

// 开始 AI 诊断
const startAIDiagnosis = async (alarm) => {
  const question = formatAlarmForAI(alarm)
  console.log('🚀 [报警AI诊断] 开始诊断:', question)
  
  // 重置临时数据
  aiTempRelevantDocs = []
  
  try {
    console.log('📞 [报警AI诊断] 调用 QueryAIStreamWithQueue（带排队）')
    // 使用带排队的流式查询
    await window.go.main.App.QueryAIStreamWithQueue(question)
    console.log('✅ [报警AI诊断] QueryAIStreamWithQueue 完成')
  } catch (e) {
    aiDiagnosisData.isThinking = false
    aiDiagnosisData.error = '发生错误: ' + e
    console.error('❌ [报警AI诊断] 失败:', e)
  }
}

// 停止 AI 诊断
const stopAIDiagnosis = async () => {
  console.log('🛑 [报警AI诊断] 用户请求停止')
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const stopped = await window.go.main.App.StopAIStream()
      if (stopped) {
        aiDiagnosisData.isThinking = false
        aiDiagnosisData.queued = false
        if (!aiDiagnosisData.answer) {
          aiDiagnosisData.answer = '⚠️ 已停止生成'
        }
        console.log('✅ [报警AI诊断] 已停止')
      }
    }
  } catch (e) {
    console.error('❌ [报警AI诊断] 停止失败:', e)
  }
}

// 点赞 AI 回答
const likeAnswer = async () => {
  try {
    if (!aiDiagnosisData.answer || !aiDiagnosisData.alarm) {
      showToast('没有可点赞的回答', 'warning')
      return
    }
    
    console.log('👍 [报警AI诊断] 用户点赞回答')
    
    if (window.go && window.go.main && window.go.main.App) {
      // 构造问题
      const question = formatAlarmForAI(aiDiagnosisData.alarm)
      
      // 如果是重新生成的答案被点赞，说明用户认可新答案，用它替换原来的优质回答
      if (aiDiagnosisData.isRegeneratedAnswer) {
        console.log('✨ [报警AI诊断] 重新生成的答案被点赞，将替换原优质回答')
        showToast('新答案已被采纳并替换原优质回答！', 'success', 3000)
      } else {
        console.log('👍 [报警AI诊断] 点赞回答')
        showToast('点赞成功！感谢您的反馈', 'success', 3000)
      }
      
      await window.go.main.App.LikeAIAnswer(
        question,
        aiDiagnosisData.answer,
        aiDiagnosisData.relevantDocs
      )
      
      aiDiagnosisData.liked = true
      
      // 如果是重新生成的答案，点赞后清除原始备份（因为已经替换）
      if (aiDiagnosisData.isRegeneratedAnswer) {
        aiDiagnosisData.originalAnswer = ''
        aiDiagnosisData.originalDocs = []
        aiDiagnosisData.originalLikeCount = 0
        aiDiagnosisData.isRegeneratedAnswer = false
        // 更新点赞数为1（新的优质回答）
        aiDiagnosisData.feedbackLikeCount = 1
        aiDiagnosisData.fromFeedback = true
      }
    }
  } catch (e) {
    console.error('❌ [报警AI诊断] 点赞失败:', e)
    showToast('点赞失败: ' + e, 'error')
  }
}

// 关闭 AI 诊断弹窗
const closeAIModal = async () => {
  console.log('🔄 [报警AI诊断] 关闭弹窗，停止正在进行的生成')
  // 关闭弹窗时也停止正在进行的生成
  if (aiDiagnosisData.isThinking) {
    await stopAIDiagnosis()
  }
  showAIModal.value = false
}

// 重新生成答案（需要点赞后才替换原优质回答）
const regenerateAnswer = async () => {
  if (!aiDiagnosisData.alarm) return
  
  console.log('🔄 [报警AI诊断] 用户请求重新生成')
  
  // 保存原始优质回答（如果还没保存的话）
  if (!aiDiagnosisData.originalAnswer && aiDiagnosisData.fromFeedback) {
    aiDiagnosisData.originalAnswer = aiDiagnosisData.answer
    aiDiagnosisData.originalDocs = [...aiDiagnosisData.relevantDocs]
    aiDiagnosisData.originalLikeCount = aiDiagnosisData.feedbackLikeCount
    console.log('💾 [报警AI诊断] 已保存原始优质回答（点赞数:', aiDiagnosisData.originalLikeCount, '）')
  }
  
  // 重置状态
  aiDiagnosisData.regenerating = true
  aiDiagnosisData.isThinking = true
  aiDiagnosisData.thinkingText = '🔄 正在重新生成个性化回答...'
  aiDiagnosisData.answer = ''
  aiDiagnosisData.relevantDocs = []
  aiDiagnosisData.fromFeedback = false
  aiDiagnosisData.feedbackLikeCount = 0
  aiDiagnosisData.feedbackId = ''
  aiDiagnosisData.liked = false
  aiDiagnosisData.isRegeneratedAnswer = false
  
  // 重置临时数据
  aiTempRelevantDocs = []
  
  try {
    // 使用普通的流式查询（不使用缓存）
    const question = formatAlarmForAI(aiDiagnosisData.alarm)
    console.log('📞 [报警AI诊断] 调用 QueryAIStream（强制重新生成）')
    await window.go.main.App.QueryAIStream(question)
    console.log('✅ [报警AI诊断] 重新生成完成')
    
    // 标记当前答案为重新生成的
    aiDiagnosisData.isRegeneratedAnswer = true
  } catch (e) {
    aiDiagnosisData.isThinking = false
    aiDiagnosisData.error = '重新生成失败: ' + e
    console.error('❌ [报警AI诊断] 重新生成失败:', e)
  } finally {
    aiDiagnosisData.regenerating = false
  }
}

// 恢复原始优质回答
const restoreOriginalAnswer = () => {
  if (!aiDiagnosisData.originalAnswer) return
  
  console.log('↩️ [报警AI诊断] 恢复原始优质回答')
  aiDiagnosisData.answer = aiDiagnosisData.originalAnswer
  aiDiagnosisData.relevantDocs = [...aiDiagnosisData.originalDocs]
  aiDiagnosisData.feedbackLikeCount = aiDiagnosisData.originalLikeCount
  aiDiagnosisData.fromFeedback = true
  aiDiagnosisData.liked = false
  aiDiagnosisData.isRegeneratedAnswer = false
  
  showToast('已恢复原优质回答', 'info', 2000)
}

// 加载所有数据
const loadData = async () => {
  await loadStats()
  
  // 默认加载今天的报警记录
  if (!alarmFilter.value.startTime) {
    const now = new Date()
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    alarmFilter.value.startTime = formatDateTimeLocal(todayStart)
    alarmFilter.value.endTime = formatDateTimeLocal(now)
    alarmFilter.value.startTimeValue = todayStart.getTime()
    alarmFilter.value.endTimeValue = now.getTime()
  }
  
  await loadAlarms()
}

// 获取反馈图标（根据点赞数）
const getFeedbackIcon = () => {
  const count = aiDiagnosisData.feedbackLikeCount
  if (count >= 15) {
    // 15次以上：三个金星
    return '<i class="fas fa-star gold-star"></i><i class="fas fa-star gold-star"></i><i class="fas fa-star gold-star"></i>'
  } else if (count >= 10) {
    // 10-14次：三个大拇指
    return '<i class="fas fa-thumbs-up"></i><i class="fas fa-thumbs-up"></i><i class="fas fa-thumbs-up"></i>'
  } else if (count >= 3) {
    // 3-9次：两个大拇指
    return '<i class="fas fa-thumbs-up"></i><i class="fas fa-thumbs-up"></i>'
  } else {
    // 1-2次：一个大拇指
    return '<i class="fas fa-thumbs-up"></i>'
  }
}

// 获取反馈徽章样式类
const getFeedbackBadgeClass = () => {
  const count = aiDiagnosisData.feedbackLikeCount
  if (count >= 15) {
    return 'badge-legendary'  // 传奇级别（金星）
  } else if (count >= 10) {
    return 'badge-excellent'  // 优秀级别（三个赞）
  } else if (count >= 3) {
    return 'badge-good'       // 良好级别（两个赞）
  } else {
    return 'badge-normal'     // 普通级别（一个赞）
  }
}

// 格式化日期时间为 datetime-local 格式
const formatDateTimeLocal = (date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day}T${hours}:${minutes}`
}

// 格式化日期时间显示
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 获取同比图标
const getComparisonIcon = () => {
  if (stats.value.comparison === 'up') return 'fas fa-arrow-up'
  if (stats.value.comparison === 'down') return 'fas fa-arrow-down'
  return 'fas fa-minus'
}

// 获取报警类型样式类
const getAlarmTypeClass = (type) => {
  const classMap = {
    'HH': 'type-hh',
    'H': 'type-h',
    'L': 'type-l',
    'LL': 'type-ll'
  }
  return classMap[type] || ''
}

// 获取报警类型名称
const getAlarmTypeName = (type) => {
  const nameMap = {
    'HH': '上上限',
    'H': '上限',
    'L': '下限',
    'LL': '下下限'
  }
  return nameMap[type] || type
}

// 页面挂载时加载数据
onMounted(() => {
  loadVariableOptions()  // 先加载变量选项
  loadData()
  
  // 注册 AI 流式事件监听器（只注册一次）
  console.log('🎧 [报警AI诊断] 注册事件监听器')
  EventsOn('ai-stream-chunk', (chunk) => {
    console.log('📦 [报警AI诊断] 收到数据块:', chunk)
    if (chunk.type === 'thinking') {
      aiDiagnosisData.thinkingText = chunk.data
    } else if (chunk.type === 'queue_info') {
      // 收到排队信息
      aiDiagnosisData.queued = chunk.data.queued
      aiDiagnosisData.queuePosition = chunk.data.position
      aiDiagnosisData.thinkingText = `⏳ 正在排队中... 您的位置：第 ${chunk.data.position} 位`
      aiDiagnosisData.isThinking = true  // 确保显示思考状态
      console.log('⏳ [报警AI诊断] 排队中，位置:', chunk.data.position)
    } else if (chunk.type === 'feedback_info') {
      // 收到点赞反馈信息
      aiDiagnosisData.fromFeedback = chunk.data.from_feedback
      aiDiagnosisData.feedbackLikeCount = chunk.data.like_count
      aiDiagnosisData.feedbackId = chunk.data.feedback_id || ''
      aiDiagnosisData.thinkingText = `正在加载优质回答...`
      console.log('✨ [报警AI诊断] 使用点赞反馈，点赞数:', chunk.data.like_count, 'ID:', chunk.data.feedback_id)
    } else if (chunk.type === 'docs') {
      // 收到文档，立即显示
      aiTempRelevantDocs = chunk.data || []
      aiDiagnosisData.relevantDocs = aiTempRelevantDocs
      console.log('📚 [报警AI诊断] 收到相关文档:', aiTempRelevantDocs.length, '个')
    } else if (chunk.type === 'token') {
      // 收到 token（逐字输出）
      if (aiDiagnosisData.answer === '') {
        aiDiagnosisData.isThinking = false
        aiDiagnosisData.queued = false
        console.log('📝 [报警AI诊断] 开始接收token，清除思考状态')
      }
      aiDiagnosisData.answer += chunk.data
      // 强制触发响应式更新
      nextTick(() => {
        // DOM已更新
      })
    } else if (chunk.type === 'end') {
      // 流结束
      aiDiagnosisData.isThinking = false
      aiDiagnosisData.queued = false
      console.log('🏁 [报警AI诊断] 流式输出结束')
    }
  })
  
  EventsOn('ai-stream-error', (error) => {
    console.error('❌ [报警AI诊断] 收到错误:', error)
    aiDiagnosisData.isThinking = false
    aiDiagnosisData.error = error
  })
  
  EventsOn('ai-stream-cancelled', () => {
    console.log('🛑 [报警AI诊断] 流式传输已取消')
    aiDiagnosisData.isThinking = false
    if (!aiDiagnosisData.answer) {
      aiDiagnosisData.answer = '⚠️ 已停止生成'
    }
  })
  
  // 设置自动刷新（每30秒）
  refreshTimer = setInterval(() => {
    loadStats()
    loadAlarms()
  }, 30000)
})

// 页面卸载时清除定时器和停止AI生成
onUnmounted(() => {
  console.log('🔄 [报警管理] 页面即将卸载')
  
  // 清除定时器
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  
  // 如果有正在进行的AI诊断，停止它
  if (aiDiagnosisData.isThinking || aiDiagnosisData.queued) {
    console.log('🛑 [报警管理] 停止正在进行的AI诊断')
    if (window.go && window.go.main && window.go.main.App) {
      window.go.main.App.StopAIStream().catch(err => {
        console.error('❌ [报警管理] 停止失败:', err)
      })
    }
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
.stat-card.card-total {
  background: linear-gradient(135deg, #7a5858 0%, #8e6e6e 100%);
}

.stat-card.card-acked {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

.stat-card.card-pending {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
}

.stat-card.card-comparison {
  background: linear-gradient(135deg, #556070 0%, #6e7e8e 100%);
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
  font-size: 28px;
  font-weight: bold;
  color: #fff;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

.trend-up {
  color: #d4a8a8;
}

.trend-down {
  color: #a8d5c7;
}

.trend-equal {
  color: rgba(255,255,255,0.9);
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

/* Naive UI 表格自定义样式 */
:deep(.n-data-table) {
  background: transparent;
}

:deep(.n-data-table-th) {
  background: rgba(120, 144, 156, 0.1) !important;
  color: rgba(255,255,255,0.9) !important;
  font-weight: 600;
  border-bottom: 2px solid rgba(120, 144, 156, 0.3) !important;
}

:deep(.n-data-table-td) {
  background: transparent !important;
  color: rgba(255,255,255,0.8) !important;
  border-bottom: 1px solid rgba(255,255,255,0.05) !important;
}

:deep(.n-data-table-tr):hover .n-data-table-td {
  background: rgba(120, 144, 156, 0.08) !important;
}

:deep(.n-data-table-tr.alarm-active) {
  background: rgba(142, 110, 110, 0.15) !important;
}

:deep(.n-data-table-tr.alarm-active) .n-data-table-td {
  background: rgba(142, 110, 110, 0.15) !important;
  border-left: 3px solid #8e6e6e;
}

:deep(.n-data-table__empty) {
  color: rgba(255,255,255,0.4);
  padding: 60px 20px;
}

/* 分页样式 */
:deep(.n-pagination) {
  margin-top: 16px;
  justify-content: flex-end;
}

/* 分页前缀文本 */
:deep(.n-pagination-prefix) {
  color: rgba(255,255,255,0.8) !important;
  margin-right: 12px;
}

/* 分页按钮基础样式 */
:deep(.n-pagination-item) {
  background: rgba(30, 40, 60, 0.6) !important;
  border: 1px solid rgba(255,255,255,0.1) !important;
  color: rgba(255,255,255,0.8) !important;
  min-width: 32px;
  height: 32px;
}

:deep(.n-pagination-item__link) {
  color: rgba(255,255,255,0.8) !important;
}

/* 上一页/下一页按钮 */
:deep(.n-pagination-item--button) {
  background: rgba(30, 40, 60, 0.6) !important;
  border: 1px solid rgba(255,255,255,0.1) !important;
}

:deep(.n-pagination-item--button .n-base-icon) {
  color: rgba(255,255,255,0.8) !important;
}

/* 悬停效果 */
:deep(.n-pagination-item:not(.n-pagination-item--disabled):hover) {
  background: rgba(120, 144, 156, 0.3) !important;
  border-color: rgba(120, 144, 156, 0.4) !important;
}

/* 激活状态 */
:deep(.n-pagination-item--active) {
  background: rgba(120, 144, 156, 0.4) !important;
  border-color: rgba(120, 144, 156, 0.5) !important;
}

:deep(.n-pagination-item--active .n-pagination-item__link) {
  color: #fff !important;
  font-weight: 600;
}

/* 禁用状态 */
:deep(.n-pagination-item--disabled) {
  background: rgba(30, 40, 60, 0.3) !important;
  border-color: rgba(255,255,255,0.05) !important;
  opacity: 0.4;
}

/* 省略号 */
:deep(.n-pagination-item--disabled .n-pagination-item__link) {
  color: rgba(255,255,255,0.4) !important;
}

/* 快速跳转输入框 */
:deep(.n-pagination-quick-jumper) {
  color: rgba(255,255,255,0.8) !important;
}

:deep(.n-pagination-quick-jumper .n-input) {
  background: rgba(30, 40, 60, 0.6) !important;
  border-color: rgba(255,255,255,0.1) !important;
}

:deep(.n-pagination-quick-jumper .n-input__input-el) {
  color: rgba(255,255,255,0.8) !important;
}

/* 每页条数选择器 */
:deep(.n-pagination-size-picker) {
  color: rgba(255,255,255,0.8) !important;
}

:deep(.n-pagination-size-picker .n-base-selection) {
  background: rgba(30, 40, 60, 0.6) !important;
  border-color: rgba(255,255,255,0.1) !important;
}

:deep(.n-pagination-size-picker .n-base-selection-label) {
  color: rgba(255,255,255,0.8) !important;
}

:deep(.n-pagination-size-picker .n-base-suffix .n-base-icon) {
  color: rgba(255,255,255,0.6) !important;
}

/* 下拉菜单 */
:deep(.n-base-select-menu) {
  background: rgba(30, 40, 60, 0.95) !important;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255,255,255,0.1) !important;
}

:deep(.n-base-select-option) {
  color: rgba(255,255,255,0.8) !important;
}

:deep(.n-base-select-option:hover),
:deep(.n-base-select-option--selected) {
  background: rgba(120, 144, 156, 0.3) !important;
  color: #fff !important;
}

/* 变量名 */
.var-name {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255,255,255,0.9);
}

.var-name i {
  color: #78909c;
}

/* 报警类型徽章 */
.alarm-type-badge {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  display: inline-block;
}

.alarm-type-badge.type-hh {
  background: rgba(142, 110, 110, 0.2);
  color: #d4a8a8;
}

.alarm-type-badge.type-h {
  background: rgba(158, 126, 94, 0.2);
  color: #d4c4a8;
}

.alarm-type-badge.type-l {
  background: rgba(110, 126, 142, 0.2);
  color: #b8c4d4;
}

.alarm-type-badge.type-ll {
  background: rgba(94, 139, 126, 0.2);
  color: #a8d5c7;
}

/* 数值显示 */
.value-text, .limit-text {
  font-family: 'Courier New', monospace;
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

/* 状态文本 */
.recovered {
  color: rgba(255,255,255,0.7);
}

.alarm-ongoing {
  color: #d4a8a8;
  font-weight: 500;
}

.duration-text {
  color: rgba(255,255,255,0.7);
}

.duration-ongoing {
  color: #d4a8a8;
  font-weight: 500;
}

/* 确认状态徽章 */
.ack-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.ack-badge.acked {
  background: rgba(94, 139, 126, 0.2);
  color: #7ea896;
}

.ack-badge.pending {
  background: rgba(158, 126, 94, 0.2);
  color: #b8967e;
}

.text-muted {
  color: rgba(255,255,255,0.4);
}

/* 消息文本 */
.msg-text {
  color: rgba(255,255,255,0.8);
  font-size: 13px;
  display: inline-block;
  max-width: 250px;
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
}

/* ======================================================== */
/* AI 诊断弹窗样式 */
/* ======================================================== */

/* 弹窗遮罩层 */
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

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* 弹窗容器 */
.modal-container {
  background: rgba(20, 30, 48, 0.98);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 90%;
  max-width: 900px;
  max-height: 85vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  animation: slideUp 0.3s;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 弹窗头部 */
.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
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

/* 弹窗主体 */
.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

/* 弹窗底部 */
.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
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

.modal-btn.cancel {
  background: rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.8);
}

.modal-btn.cancel:hover {
  background: rgba(255,255,255,0.15);
}

.modal-btn.confirm {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.modal-btn.confirm:hover {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  transform: translateY(-2px);
}

.modal-btn.like {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: #fff;
}

.modal-btn.like:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
  transform: translateY(-2px);
}

.modal-btn.like:disabled {
  background: rgba(16, 185, 129, 0.5);
  cursor: not-allowed;
  opacity: 0.6;
}

/* 报警信息摘要 */
.alarm-summary {
  background: rgba(158, 126, 94, 0.1);
  border: 1px solid rgba(158, 126, 94, 0.3);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
}

.summary-header {
  font-size: 14px;
  font-weight: 600;
  color: #b8967e;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-item.full-width {
  grid-column: 1 / -1;
}

.summary-label {
  font-size: 12px;
  color: rgba(255,255,255,0.6);
}

.summary-value {
  font-size: 14px;
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

.summary-value.highlight {
  color: #d4a8a8;
  font-size: 16px;
  font-weight: 600;
}

/* AI 分析结果 */
.ai-result {
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  padding: 20px;
}

.result-header {
  font-size: 14px;
  font-weight: 600;
  color: #78909c;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.feedback-badge {
  color: #fff;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
  transition: all 0.3s ease;
}

/* 普通级别（1-2个赞）- 绿色 */
.feedback-badge.badge-normal {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

/* 良好级别（3-9个赞）- 蓝色 */
.feedback-badge.badge-good {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  animation: pulse-good 2s ease-in-out infinite;
}

/* 优秀级别（10-14个赞）- 紫色 */
.feedback-badge.badge-excellent {
  background: linear-gradient(135deg, #a855f7 0%, #9333ea 100%);
  animation: pulse-excellent 2s ease-in-out infinite;
  box-shadow: 0 0 20px rgba(168, 85, 247, 0.5);
}

/* 传奇级别（15+个赞）- 金色 */
.feedback-badge.badge-legendary {
  background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%);
  animation: shine 2s ease-in-out infinite;
  box-shadow: 0 0 30px rgba(251, 191, 36, 0.8);
}

/* 金星样式 */
.gold-star {
  color: #ffd700;
  text-shadow: 0 0 10px rgba(255, 215, 0, 0.8);
  animation: twinkle 1.5s ease-in-out infinite;
}

/* 脉动动画（良好级别） */
@keyframes pulse-good {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 0 10px rgba(59, 130, 246, 0.3);
  }
  50% {
    transform: scale(1.05);
    box-shadow: 0 0 20px rgba(59, 130, 246, 0.6);
  }
}

/* 脉动动画（优秀级别） */
@keyframes pulse-excellent {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

/* 闪耀动画（传奇级别） */
@keyframes shine {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 0 30px rgba(251, 191, 36, 0.8);
  }
  50% {
    transform: scale(1.08);
    box-shadow: 0 0 40px rgba(251, 191, 36, 1);
  }
}

/* 星星闪烁动画 */
@keyframes twinkle {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.2);
  }
}

.queue-notice {
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.3);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fbbf24;
  font-size: 14px;
  animation: pulse 2s ease-in-out infinite;
}

.queue-notice i {
  font-size: 20px;
  animation: spin 2s linear infinite;
}

.queue-notice strong {
  color: #fbbf24;
  font-size: 16px;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 思考中动画 */
.thinking-section {
  padding: 24px;
  text-align: center;
}

.thinking-animation {
  display: inline-block;
}

.thinking-text {
  color: #78909c;
  font-size: 14px;
  margin-bottom: 12px;
}

.thinking-dots {
  display: flex;
  gap: 6px;
  justify-content: center;
}

.thinking-dots span {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #78909c;
  animation: bounce 1.4s infinite ease-in-out;
}

.thinking-dots span:nth-child(1) {
  animation-delay: -0.32s;
}

.thinking-dots span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes bounce {
  0%, 80%, 100% { 
    transform: scale(0);
    opacity: 0.3;
  }
  40% { 
    transform: scale(1);
    opacity: 1;
  }
}

/* 滑入动画 */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.stop-btn {
  margin-top: 12px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s ease;
}

.stop-btn:hover {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.stop-btn:active {
  transform: translateY(0);
}

/* 内联停止按钮（在答案卡片标题栏） */
.stop-btn-inline {
  margin-left: auto;
  padding: 6px 12px;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.2);
}

.stop-btn-inline:hover {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.stop-btn-inline:active {
  transform: translateY(0);
}

/* 错误信息 */
.error-section {
  padding: 16px;
  background: rgba(142, 110, 110, 0.2);
  border: 1px solid rgba(142, 110, 110, 0.3);
  border-radius: 6px;
  color: #d4a8a8;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* AI 回答 */
/* AI 回答卡片 */
.answer-card {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, rgba(37, 99, 235, 0.05) 100%);
  border: 2px solid rgba(59, 130, 246, 0.3);
  border-radius: 16px;
  padding: 0;
  margin-bottom: 20px;
  box-shadow: 0 8px 32px rgba(59, 130, 246, 0.2);
  animation: slideIn 0.4s ease-out;
}

.card-header-ai {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.3) 0%, rgba(37, 99, 235, 0.2) 100%);
  padding: 12px 20px;
  border-radius: 14px 14px 0 0;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: #60a5fa;
  border-bottom: 1px solid rgba(59, 130, 246, 0.2);
}

.card-header-ai i {
  font-size: 18px;
}

.answer-content {
  padding: 20px;
}

.answer-text {
  color: rgba(255,255,255,0.95);
  font-size: 15px;
  line-height: 1.8;
  white-space: pre-wrap;
  word-wrap: break-word;
  letter-spacing: 0.3px;
}

/* 重新生成区域 */
.regenerate-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid rgba(255,255,255,0.1);
}

.regenerate-hint {
  color: rgba(255,255,255,0.7);
  font-size: 13px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.regenerate-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s;
}

.regenerate-btn:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  transform: translateY(-2px);
}

.regenerate-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 操作按钮组 */
.action-buttons {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

/* 恢复按钮 */
.restore-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #78909c 0%, #607d8b 100%);
  color: #fff;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s;
}

.restore-btn:hover {
  box-shadow: 0 4px 12px rgba(120, 144, 156, 0.4);
  transform: translateY(-2px);
}

.restore-btn:active {
  transform: translateY(0);
}

/* 相关文档 */
/* 参考文档卡片 */
.docs-card {
  background: linear-gradient(135deg, rgba(168, 85, 247, 0.08) 0%, rgba(147, 51, 234, 0.04) 100%);
  border: 1px solid rgba(168, 85, 247, 0.2);
  border-radius: 16px;
  padding: 0;
  margin-top: 20px;
  box-shadow: 0 4px 16px rgba(168, 85, 247, 0.1);
  animation: slideIn 0.5s ease-out;
}

.card-header-docs {
  background: linear-gradient(135deg, rgba(168, 85, 247, 0.15) 0%, rgba(147, 51, 234, 0.1) 100%);
  padding: 10px 20px;
  border-radius: 14px 14px 0 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #c084fc;
  border-bottom: 1px solid rgba(168, 85, 247, 0.15);
}

.card-header-docs i {
  font-size: 14px;
}

.docs-list {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.doc-item {
  background: rgba(20, 30, 50, 0.3);
  border: 1px solid rgba(168, 85, 247, 0.15);
  border-radius: 8px;
  padding: 12px;
  transition: all 0.3s ease;
}

.doc-item:hover {
  background: rgba(30, 40, 60, 0.5);
  border-color: rgba(168, 85, 247, 0.3);
  transform: translateX(4px);
}

.doc-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.doc-source {
  font-size: 12px;
  color: #78909c;
  display: flex;
  align-items: center;
  gap: 6px;
}

.doc-similarity {
  font-size: 11px;
  color: rgba(255,255,255,0.5);
  background: rgba(255,255,255,0.1);
  padding: 2px 8px;
  border-radius: 4px;
}

.doc-content {
  font-size: 13px;
  color: rgba(255,255,255,0.8);
  line-height: 1.6;
}
</style>
