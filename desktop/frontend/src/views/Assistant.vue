<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-robot"></i>
          AI 智能助手
        </div>
        <div class="page-subtitle">AI Assistant & Knowledge Management</div>
      </div>
      <div class="header-actions">
        <div class="service-status" :class="{ online: aiServiceOnline, offline: !aiServiceOnline }">
          <div class="status-dot"></div>
          <span>{{ aiServiceOnline ? 'AI 服务在线' : 'AI 服务离线' }}</span>
        </div>
        <button class="action-btn refresh" @click="checkServiceAndRefresh" title="刷新">
          <i class="fas fa-sync-alt"></i> 刷新
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
          <div class="stat-value">{{ knowledgeStats.total }}</div>
          <div class="stat-label">知识总数</div>
        </div>
      </div>
      <div class="stat-card card-active">
        <div class="stat-icon">
          <i class="fas fa-comments"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ chatStats.totalQuestions }}</div>
          <div class="stat-label">问答次数</div>
        </div>
      </div>
      <div class="stat-card card-success">
        <div class="stat-icon">
          <i class="fas fa-check-circle"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ chatStats.successRate }}%</div>
          <div class="stat-label">成功率</div>
        </div>
      </div>
      <div class="stat-card card-response">
        <div class="stat-icon">
          <i class="fas fa-clock"></i>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ chatStats.avgResponseTime }}s</div>
          <div class="stat-label">平均响应</div>
        </div>
      </div>
    </div>

    <!-- Tab 导航 -->
    <div class="tab-container">
      <div class="tab-nav">
        <button 
          v-for="tab in tabs" 
          :key="tab.id"
          :class="['tab-btn', { active: activeTab === tab.id }]"
          @click="activeTab = tab.id"
        >
          <i :class="tab.icon"></i>
          {{ tab.label }}
          <span v-if="tab.badge" class="tab-badge">{{ tab.badge }}</span>
        </button>
      </div>

      <!-- Tab 内容区域 -->
      <div class="tab-content">
        <!-- Tab 1: 知识库管理 -->
        <div v-show="activeTab === 'knowledge'" class="tab-pane">
          <div class="knowledge-grid">
            <!-- 左侧：添加知识表单 -->
            <div class="card">
              <div class="card-header">
                <h3><i class="fas fa-plus-circle"></i> 添加知识</h3>
              </div>
              <div class="form-container">
                <div class="form-group">
                  <label><i class="fas fa-file-alt"></i> 知识内容 <span class="required">*</span></label>
                  <textarea 
                    v-model="newKnowledge.content"
                    placeholder="请输入知识内容（支持多行）&#10;例如：当温度超过80°C时，应立即检查冷却系统..."
                    rows="8"
                    class="form-textarea"
                  ></textarea>
                  <div class="char-count">{{ newKnowledge.content.length }} / 10000</div>
                </div>
                <div class="form-group">
                  <label><i class="fas fa-tag"></i> 知识来源</label>
                  <input 
                    v-model="newKnowledge.source"
                    type="text"
                    placeholder="例如：设备操作手册、维护规程等"
                    class="form-input"
                  />
                </div>
                <div class="form-actions">
                  <button 
                    class="btn-primary" 
                    @click="addKnowledge"
                    :disabled="!newKnowledge.content.trim() || addingKnowledge"
                  >
                    <i class="fas fa-plus"></i>
                    {{ addingKnowledge ? '添加中...' : '添加知识' }}
                  </button>
                  <button class="btn-secondary" @click="clearForm">
                    <i class="fas fa-eraser"></i>
                    清空
                  </button>
                </div>
              </div>
            </div>

            <!-- 右侧：知识列表 -->
            <div class="card">
              <div class="card-header">
                <h3><i class="fas fa-list"></i> 知识库列表</h3>
                <span class="badge">{{ knowledgeList.length }} 条</span>
              </div>
              <div class="knowledge-list">
                <div v-if="loadingKnowledge" class="loading-state">
                  <i class="fas fa-spinner fa-spin"></i>
                  <p>加载中...</p>
                </div>
                <div v-else-if="knowledgeList.length === 0" class="empty-state">
                  <i class="fas fa-inbox"></i>
                  <p>暂无知识数据</p>
                  <small>请在左侧添加知识</small>
                </div>
                <div v-else class="knowledge-items">
                  <div 
                    v-for="item in knowledgeList" 
                    :key="item.id"
                    class="knowledge-item"
                  >
                    <div class="knowledge-header">
                      <div class="knowledge-meta">
                        <span class="knowledge-id">ID: {{ item.id.substring(0, 8) }}</span>
                        <span v-if="item.source" class="knowledge-source">
                          <i class="fas fa-tag"></i>
                          {{ item.source }}
                        </span>
                      </div>
                      <button 
                        class="btn-delete" 
                        @click="deleteKnowledge(item)"
                        title="删除"
                      >
                        <i class="fas fa-trash"></i>
                      </button>
                    </div>
                    <div class="knowledge-content">{{ item.content }}</div>
                    <div class="knowledge-footer">
                      <span v-if="item.created_at" class="knowledge-time">
                        <i class="fas fa-clock"></i>
                        {{ formatTime(item.created_at) }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 2: AI 问答 -->
        <div v-show="activeTab === 'chat'" class="tab-pane">
          <div class="chat-container">
            <div class="card chat-card">
              <div class="card-header">
                <h3><i class="fas fa-comments"></i> AI 对话</h3>
                <button class="btn-clear" @click="clearChat">
                  <i class="fas fa-broom"></i>
                  清空对话
                </button>
              </div>
              
              <!-- 聊天消息区域 -->
              <div class="chat-messages" ref="chatMessages">
                <div v-if="chatHistory.length === 0" class="chat-welcome">
                  <i class="fas fa-robot"></i>
                  <h3>你好！我是 AI 助手</h3>
                  <p>我可以基于知识库回答你的问题</p>
                  <div class="quick-questions">
                    <button 
                      v-for="q in quickQuestions" 
                      :key="q"
                      class="quick-btn"
                      @click="askQuestion(q)"
                    >
                      {{ q }}
                    </button>
                  </div>
                </div>

                <!-- 聊天记录 -->
                <div 
                  v-for="(msg, index) in chatHistory" 
                  :key="index"
                  :class="['chat-message', msg.role]"
                >
                  <div class="message-avatar">
                    <i :class="msg.role === 'user' ? 'fas fa-user' : 'fas fa-robot'"></i>
                  </div>
                  <div class="message-content">
                    <div class="message-text">{{ msg.content }}</div>
                    
                    <!-- 显示相关文档 -->
                    <div v-if="msg.relevantDocs && msg.relevantDocs.length > 0" class="relevant-docs">
                      <div class="docs-header">
                        <i class="fas fa-book"></i>
                        相关文档 ({{ msg.relevantDocs.length }})
                      </div>
                      <div class="docs-list">
                        <div 
                          v-for="(doc, idx) in msg.relevantDocs" 
                          :key="idx"
                          class="doc-item"
                        >
                          <div class="doc-header">
                            <span class="doc-source">{{ doc.source || '未知来源' }}</span>
                            <span class="doc-similarity">{{ (doc.similarity * 100).toFixed(0) }}%</span>
                          </div>
                          <div class="doc-content">{{ doc.content }}</div>
                        </div>
                      </div>
                    </div>

                    <div class="message-time">{{ msg.time }}</div>
                  </div>
                </div>

                <!-- 思考中动画 -->
                <div v-if="isThinking" class="chat-message assistant thinking">
                  <div class="message-avatar">
                    <i class="fas fa-robot"></i>
                  </div>
                  <div class="message-content">
                    <div class="thinking-animation">
                      <div class="thinking-text">{{ thinkingText }}</div>
                      <div class="thinking-dots">
                        <span></span><span></span><span></span>
                      </div>
                    </div>
                    <button class="stop-btn" @click="stopGeneration" title="停止生成">
                      <i class="fas fa-stop-circle"></i> 停止回答
                    </button>
                  </div>
                </div>
              </div>

              <!-- 输入区域 -->
              <div class="chat-input-area">
                <textarea 
                  v-model="currentQuestion"
                  @keydown.enter.exact.prevent="sendMessage"
                  placeholder="输入你的问题... (Enter 发送, Shift+Enter 换行)"
                  rows="2"
                  class="chat-input"
                  :disabled="isThinking"
                ></textarea>
                <button 
                  class="btn-send" 
                  @click="sendMessage"
                  :disabled="!currentQuestion.trim() || isThinking"
                >
                  <i class="fas fa-paper-plane"></i>
                  发送
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 3: 批量导入 -->
        <div v-show="activeTab === 'import'" class="tab-pane">
          <!-- 错误码导入卡片 -->
          <div class="card error-code-import-card">
            <div class="card-header">
              <h3><i class="fas fa-file-excel"></i> 错误码批量导入</h3>
            </div>
            <div class="import-container">
              <!-- 步骤1: 选择文件 -->
              <div v-if="importStep === 1" class="import-step-1">
                <div class="import-info">
                  <p><i class="fas fa-info-circle"></i> 支持批量导入错误码知识（错误码、来源、现象、原因、解决方案、预防措施）</p>
                  <p><i class="fas fa-lightbulb"></i> 数据将自动同步到MySQL和知识库，来源字段用于区分知识来源（如：设备手册、维修记录等）</p>
                </div>
                <div class="import-actions">
                  <button class="btn-download" @click="downloadTemplate">
                    <i class="fas fa-download"></i>
                    下载模板
                  </button>
                  <label class="btn-upload">
                    <i class="fas fa-upload"></i>
                    选择Excel文件
                    <input 
                      type="file" 
                      accept=".xlsx,.xls" 
                      @change="handleFileSelect"
                      ref="fileInput"
                      style="display: none;"
                    />
                  </label>
                </div>
              </div>

              <!-- 步骤2: 预览和编辑 -->
              <div v-if="importStep === 2" class="import-step-2">
                <div class="preview-header">
                  <h4>数据预览（共 {{ errorCodeData.length }} 条）</h4>
                  <div class="preview-actions">
                    <button class="btn-secondary" @click="cancelImport">
                      <i class="fas fa-times"></i> 取消
                    </button>
                    <button class="btn-primary" @click="confirmImport">
                      <i class="fas fa-check"></i> 确认导入
                    </button>
                  </div>
                </div>
                <div class="preview-table-container">
                  <table class="preview-table">
                    <thead>
                      <tr>
                        <th style="width: 80px;">错误码</th>
                        <th style="width: 120px;">来源</th>
                        <th style="width: 180px;">现象</th>
                        <th style="width: 180px;">原因</th>
                        <th style="width: 180px;">解决方案</th>
                        <th style="width: 180px;">预防措施</th>
                        <th style="width: 80px;">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(row, idx) in errorCodeData" :key="idx">
                        <td><input v-model.number="row.errorCode" type="number" class="table-input" /></td>
                        <td><input v-model="row.source" type="text" class="table-input" placeholder="来源" /></td>
                        <td><textarea v-model="row.phenomenon" class="table-textarea" rows="2"></textarea></td>
                        <td><textarea v-model="row.cause" class="table-textarea" rows="2"></textarea></td>
                        <td><textarea v-model="row.solution" class="table-textarea" rows="2"></textarea></td>
                        <td><textarea v-model="row.prevention" class="table-textarea" rows="2"></textarea></td>
                        <td>
                          <button class="btn-delete" @click="removeRow(idx)" title="删除">
                            <i class="fas fa-trash"></i>
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- 步骤3: 导入进度 -->
              <div v-if="importStep === 3" class="import-step-3">
                <div class="import-progress">
                  <div class="progress-info">
                    <i class="fas fa-spinner fa-spin"></i>
                    <span>正在导入... {{ importProgress.current }}/{{ importProgress.total }}</span>
                  </div>
                  <div class="progress-bar-container">
                    <div class="progress-bar" :style="{ width: (importProgress.current / importProgress.total * 100) + '%' }"></div>
                  </div>
                  <div class="progress-details">
                    <p><i class="fas fa-database"></i> MySQL: {{ importProgress.mysql }} 条</p>
                    <p><i class="fas fa-brain"></i> 知识库: {{ importProgress.knowledge }} 条</p>
                    <p v-if="importProgress.errors.length > 0" class="progress-errors">
                      <i class="fas fa-exclamation-triangle"></i> 失败: {{ importProgress.errors.length }} 条
                    </p>
                  </div>
                </div>
              </div>

              <!-- 步骤4: 完成 -->
              <div v-if="importStep === 4" class="import-step-4">
                <div class="import-result" :class="importProgress.errors.length === 0 ? 'success' : 'warning'">
                  <i :class="importProgress.errors.length === 0 ? 'fas fa-check-circle' : 'fas fa-exclamation-circle'"></i>
                  <div class="result-content">
                    <h4>导入完成！</h4>
                    <p>成功导入 {{ importProgress.current }} 条错误码知识</p>
                    <p>MySQL: {{ importProgress.mysql }} 条 | 知识库: {{ importProgress.knowledge }} 条</p>
                    <div v-if="importProgress.errors.length > 0" class="error-list">
                      <p>失败 {{ importProgress.errors.length }} 条：</p>
                      <ul>
                        <li v-for="(error, idx) in importProgress.errors" :key="idx">{{ error }}</li>
                      </ul>
                    </div>
                  </div>
                </div>
                <button class="btn-primary" @click="resetImport">
                  <i class="fas fa-redo"></i> 继续导入
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 4: 错误码对照 -->
        <div v-show="activeTab === 'error-codes'" class="tab-pane">
          <div class="card">
            <div class="card-header">
              <h3><i class="fas fa-list-alt"></i> 错误码对照表</h3>
              <div class="header-actions">
                <input 
                  v-model="errorCodeSearch" 
                  type="text" 
                  placeholder="搜索错误码或现象..." 
                  class="search-input"
                  @input="searchErrorCodes"
                />
                <button class="btn-refresh" @click="loadErrorCodes" :disabled="loadingErrorCodes">
                  <i :class="loadingErrorCodes ? 'fas fa-spinner fa-spin' : 'fas fa-sync-alt'"></i>
                  刷新
                </button>
              </div>
            </div>
            <div class="error-codes-container">
              <div v-if="loadingErrorCodes" class="loading-state">
                <i class="fas fa-spinner fa-spin"></i>
                <p>加载中...</p>
              </div>
              <div v-else-if="filteredErrorCodes.length === 0" class="empty-state">
                <i class="fas fa-inbox"></i>
                <p>暂无错误码数据</p>
                <small>请先在"批量导入"中导入错误码</small>
              </div>
              <div v-else class="error-codes-table-container">
                <table class="error-codes-table">
                  <thead>
                    <tr>
                      <th style="width: 120px;">错误码</th>
                      <th>现象描述</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(code, index) in filteredErrorCodes" :key="code?.ErrorCode || index">
                      <td class="error-code-cell">
                        <span class="error-code-badge">{{ code?.ErrorCode || '-' }}</span>
                      </td>
                      <td class="error-msg-cell">{{ code?.ErrorMsg || '-' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div v-if="!loadingErrorCodes && filteredErrorCodes.length > 0" class="table-footer">
                <span>共 {{ filteredErrorCodes.length }} 条错误码</span>
              </div>
            </div>
          </div>
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
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { 
  QueryAI, 
  QueryAIStream,
  StopAIStream,
  AddKnowledge,
  DeleteKnowledge,
  DeleteKnowledgeBySource,
  GetKnowledgeList,
  CheckAIServiceHealth
} from '../../wailsjs/go/main/App'
import Toast from '../components/Toast.vue'

// ========================================================
// 状态管理
// ========================================================

const aiServiceOnline = ref(false)
const activeTab = ref('knowledge')

// Toast 提示
const toast = ref({
  show: false,
  message: '',
  type: 'success',
  duration: 2000
})

// Tab 配置
const tabs = computed(() => [
  { id: 'knowledge', label: '知识库管理', icon: 'fas fa-database', badge: knowledgeList.value.length },
  { id: 'chat', label: 'AI 问答', icon: 'fas fa-comments' },
  { id: 'import', label: '批量导入', icon: 'fas fa-file-import' },
  { id: 'error-codes', label: '错误码对照', icon: 'fas fa-list-alt' }
])

// 知识库统计
const knowledgeStats = reactive({
  total: 0
})

// 聊天统计
const chatStats = reactive({
  totalQuestions: 0,
  successRate: 100,
  avgResponseTime: 2.3
})

// ========================================================
// 知识库管理
// ========================================================

const knowledgeList = ref([])
const loadingKnowledge = ref(false)
const addingKnowledge = ref(false)

const newKnowledge = reactive({
  content: '',
  source: ''
})

// 错误码导入
const importStep = ref(1) // 1: 选择文件, 2: 预览编辑, 3: 导入中, 4: 完成
const errorCodeData = ref([])
const fileInput = ref(null)
const importProgress = reactive({
  current: 0,
  total: 0,
  mysql: 0,
  knowledge: 0,
  errors: []
})

// 错误码对照
const errorCodes = ref([])
const filteredErrorCodes = ref([])
const errorCodeSearch = ref('')
const loadingErrorCodes = ref(false)

// 加载知识列表
const loadKnowledgeList = async () => {
  loadingKnowledge.value = true
  try {
    const result = await GetKnowledgeList()
    if (result.success) {
      knowledgeList.value = result.data || []
      knowledgeStats.total = result.total || 0
    }
  } catch (e) {
    console.error('加载知识列表失败:', e)
    alert('加载失败: ' + e)
  } finally {
    loadingKnowledge.value = false
  }
}

// 添加知识
const addKnowledge = async () => {
  if (!newKnowledge.content.trim()) {
    alert('请输入知识内容')
    return
  }

  addingKnowledge.value = true
  try {
    const source = newKnowledge.source.trim() || null
    const result = await AddKnowledge(newKnowledge.content, source)
    
    if (result.success) {
      alert('✅ 知识添加成功')
      clearForm()
      await loadKnowledgeList()
    }
  } catch (e) {
    console.error('添加知识失败:', e)
    alert('添加失败: ' + e)
  } finally {
    addingKnowledge.value = false
  }
}

// 删除知识
const deleteKnowledge = async (item) => {
  if (!confirm(`确定要删除这条知识吗？\n\n${item.content.substring(0, 50)}...`)) {
    return
  }

  try {
    const result = await DeleteKnowledge(item.id)
    if (result.success) {
      alert('✅ 删除成功')
      await loadKnowledgeList()
    }
  } catch (e) {
    console.error('删除知识失败:', e)
    alert('删除失败: ' + e)
  }
}

// 清空表单
const clearForm = () => {
  newKnowledge.content = ''
  newKnowledge.source = ''
}

// 下载错误码模板
const downloadTemplate = () => {
  try {
    // 使用xlsx库创建模板
    import('xlsx').then(XLSX => {
      const template = [
        {
          '错误码': 400,
          '来源': '设备手册',
          '现象': '按下急停后设备停止，无法恢复运行',
          '原因': '急停按钮触发后安全回路未复位',
          '解决方案': '1. 检查急停按钮是否复位\n2. 检查安全回路\n3. 重启设备',
          '预防措施': '定期检查急停系统'
        },
        {
          '错误码': 401,
          '来源': '维修记录',
          '现象': '使用抛光液清洁时，镜片表面颜色发生变化',
          '原因': '抛光液与镜片材质发生化学反应',
          '解决方案': '1. 停止使用该抛光液\n2. 更换镜片\n3. 使用专用清洁剂',
          '预防措施': '使用厂家推荐的清洁剂'
        }
      ]
      
      const ws = XLSX.utils.json_to_sheet(template)
      const wb = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(wb, ws, '错误码')
      XLSX.writeFile(wb, '错误码导入模板.xlsx')
      
      showToast('模板下载成功', 'success')
    })
  } catch (error) {
    console.error('下载模板失败:', error)
    showToast('下载模板失败: ' + error.message, 'error')
  }
}

// 处理文件选择
const handleFileSelect = async (event) => {
  const file = event.target.files[0]
  if (!file) return
  
  // 检查文件类型
  if (!file.name.endsWith('.xlsx') && !file.name.endsWith('.xls')) {
    showToast('只支持Excel文件(.xlsx, .xls)', 'error')
    event.target.value = ''
    return
  }
  
  try {
    // 动态导入xlsx库
    const XLSX = await import('xlsx')
    
    // 读取文件
    const data = await file.arrayBuffer()
    const workbook = XLSX.read(data)
    const worksheet = workbook.Sheets[workbook.SheetNames[0]]
    const jsonData = XLSX.utils.sheet_to_json(worksheet)
    
    // 验证数据
    if (jsonData.length === 0) {
      throw new Error('Excel文件为空')
    }
    
    // 验证必需列
    const requiredColumns = ['错误码', '来源', '现象', '原因', '解决方案', '预防措施']
    const firstRow = jsonData[0]
    const missingColumns = requiredColumns.filter(col => !(col in firstRow))
    
    if (missingColumns.length > 0) {
      throw new Error(`Excel缺少必需的列: ${missingColumns.join(', ')}`)
    }
    
    // 解析数据
    errorCodeData.value = jsonData.map(row => ({
      errorCode: parseInt(row['错误码']) || 0,
      source: String(row['来源'] || '未知来源'),
      phenomenon: String(row['现象'] || ''),
      cause: String(row['原因'] || ''),
      solution: String(row['解决方案'] || ''),
      prevention: String(row['预防措施'] || '')
    }))
    
    // 进入预览步骤
    importStep.value = 2
    showToast(`成功读取 ${errorCodeData.value.length} 条数据`, 'success')
    
  } catch (error) {
    console.error('文件解析失败:', error)
    showToast('文件解析失败: ' + error.message, 'error')
  } finally {
    event.target.value = '' // 清空input
  }
}

// 删除行
const removeRow = (index) => {
  errorCodeData.value.splice(index, 1)
  showToast('已删除', 'info')
}

// 取消导入
const cancelImport = () => {
  if (confirm('确定要取消导入吗？')) {
    resetImport()
  }
}

// 重置导入状态
const resetImport = () => {
  importStep.value = 1
  errorCodeData.value = []
  importProgress.current = 0
  importProgress.total = 0
  importProgress.mysql = 0
  importProgress.knowledge = 0
  importProgress.errors = []
}

// 确认导入（替换式）
const confirmImport = async () => {
  if (errorCodeData.value.length === 0) {
    showToast('没有数据可导入', 'warning')
    return
  }
  
  // 进入导入步骤
  importStep.value = 3
  importProgress.total = errorCodeData.value.length
  importProgress.current = 0
  importProgress.mysql = 0
  importProgress.knowledge = 0
  importProgress.errors = []
  
  for (const row of errorCodeData.value) {
    try {
      // 验证数据
      if (!row.errorCode || !row.phenomenon) {
        throw new Error('错误码和现象不能为空')
      }
      
      // 1. 同步到MySQL（通过Wails调用Go）- 已经是UPSERT逻辑
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.SyncErrorCode(row.errorCode, row.phenomenon)
          importProgress.mysql++
        }
      } catch (e) {
        console.error(`MySQL同步失败 [${row.errorCode}]:`, e)
        throw new Error(`MySQL同步失败: ${e}`)
      }
      
      // 2. 同步到知识库（替换式：先删除旧的，再添加新的）
      try {
        const knowledgeContent = `【错误码 ${row.errorCode}】

现象：
${row.phenomenon}

原因分析：
${row.cause}

解决方案：
${row.solution}

预防措施：
${row.prevention}`
        
        // 使用用户指定的来源，格式：来源-错误码
        const source = row.source ? `${row.source}-${row.errorCode}` : `错误码-${row.errorCode}`
        
        // 🔥 替换式导入：先删除该来源的旧知识
        try {
          await DeleteKnowledgeBySource(source)
          console.log(`🗑️ [替换式导入] 已删除来源 '${source}' 的旧知识`)
        } catch (deleteError) {
          console.warn(`⚠️ [替换式导入] 删除旧知识失败（可能不存在）: ${deleteError}`)
          // 删除失败不影响后续添加
        }
        
        // 添加新知识
        const result = await AddKnowledge(knowledgeContent, source)
        
        if (!result || !result.success) {
          throw new Error(result?.message || '知识库同步失败')
        }
        
        importProgress.knowledge++
      } catch (e) {
        console.error(`知识库同步失败 [${row.errorCode}]:`, e)
        throw new Error(`知识库同步失败: ${e}`)
      }
      
      importProgress.current++
      
    } catch (error) {
      console.error(`导入错误码 ${row.errorCode} 失败:`, error)
      importProgress.errors.push(`错误码 ${row.errorCode}: ${error.message}`)
      importProgress.current++
    }
  }
  
  // 进入完成步骤
  importStep.value = 4
  
  // 刷新知识库列表
  await loadKnowledgeList()
  
  if (importProgress.errors.length === 0) {
    showToast(`导入完成！成功 ${importProgress.current} 条`, 'success', 3000)
  } else {
    showToast(`导入完成，但有 ${importProgress.errors.length} 条失败`, 'warning', 3000)
  }
}

// ========================================================
// 错误码对照
// ========================================================

// 加载错误码列表
const loadErrorCodes = async () => {
  loadingErrorCodes.value = true
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const codes = await window.go.main.App.GetAllErrorCodes()
      console.log('🔍 [错误码] 原始返回数据:', codes)
      console.log('🔍 [错误码] 数据类型:', typeof codes, '是否数组:', Array.isArray(codes))
      
      if (Array.isArray(codes) && codes.length > 0) {
        console.log('🔍 [错误码] 第一条数据:', codes[0])
        console.log('🔍 [错误码] 第一条数据的键:', Object.keys(codes[0]))
        errorCodes.value = codes
        filteredErrorCodes.value = codes
        console.log('✅ [错误码] 加载成功:', codes.length, '条')
      } else if (Array.isArray(codes) && codes.length === 0) {
        console.warn('⚠️ [错误码] 返回空数组')
        errorCodes.value = []
        filteredErrorCodes.value = []
      } else {
        console.warn('⚠️ [错误码] 数据格式不正确，不是数组:', codes)
        errorCodes.value = []
        filteredErrorCodes.value = []
      }
    }
  } catch (error) {
    console.error('❌ [错误码] 加载失败:', error)
    showToast('加载错误码失败: ' + error, 'error')
    errorCodes.value = []
    filteredErrorCodes.value = []
  } finally {
    loadingErrorCodes.value = false
  }
}

// 搜索错误码
const searchErrorCodes = () => {
  const keyword = errorCodeSearch.value.trim().toLowerCase()
  if (!keyword) {
    filteredErrorCodes.value = errorCodes.value
    return
  }
  
  filteredErrorCodes.value = errorCodes.value.filter(code => {
    if (!code) return false
    const codeStr = String(code.ErrorCode || '').toLowerCase()
    const msgStr = String(code.ErrorMsg || '').toLowerCase()
    return codeStr.includes(keyword) || msgStr.includes(keyword)
  })
}

// ========================================================
// 工具函数
// ========================================================

// 显示提示框
const showToast = (message, type = 'success', duration = 2000) => {
  toast.value.message = message
  toast.value.type = type
  toast.value.duration = duration
  toast.value.show = true
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', { 
    year: 'numeric', 
    month: '2-digit', 
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// ========================================================
// AI 问答
// ========================================================

const chatHistory = ref([])
const currentQuestion = ref('')
const isThinking = ref(false)
const thinkingText = ref('正在思考...')
const chatMessages = ref(null)

// 快捷问题
const quickQuestions = [
  '如何解决温度过高问题？',
  '设备维护流程是什么？',
  '报警如何处理？'
]

// 当前流式数据状态
let currentStreamingAnswer = ''
let currentRelevantDocs = []

// 发送消息
const sendMessage = async () => {
  const question = currentQuestion.value.trim()
  if (!question || isThinking.value) return

  // 添加用户消息
  chatHistory.value.push({
    role: 'user',
    content: question,
    time: new Date().toLocaleTimeString()
  })

  currentQuestion.value = ''
  isThinking.value = true
  thinkingText.value = '🤔 正在思考...'

  // 重置流式数据状态
  currentStreamingAnswer = ''
  currentRelevantDocs = []

  // 滚动到底部
  nextTick(() => {
    if (chatMessages.value) {
      chatMessages.value.scrollTop = chatMessages.value.scrollHeight
    }
  })

  try {
    // 调用流式 API
    console.log('🚀 [前端] 调用 QueryAIStream:', question)
    await QueryAIStream(question)
    console.log('✅ [前端] QueryAIStream 调用完成')
    chatStats.totalQuestions++
  } catch (e) {
    isThinking.value = false
    console.error('❌ [前端] AI 问答失败:', e)
    chatHistory.value.push({
      role: 'assistant',
      content: '❌ 发生错误: ' + e,
      time: new Date().toLocaleTimeString()
    })
  }
}

// 快捷提问
const askQuestion = (question) => {
  currentQuestion.value = question
  sendMessage()
}

// 停止生成
const stopGeneration = async () => {
  console.log('🛑 [AI助手] 用户请求停止生成')
  try {
    const stopped = await StopAIStream()
    if (stopped) {
      isThinking.value = false
      // 如果有部分回答，保存它
      if (currentStreamingAnswer) {
        chatHistory.value.push({
          role: 'assistant',
          content: currentStreamingAnswer + '\n\n⚠️ (已停止生成)',
          relevantDocs: currentRelevantDocs,
          time: new Date().toLocaleTimeString()
        })
        currentStreamingAnswer = ''
        currentRelevantDocs = []
      }
      console.log('✅ [AI助手] 已停止生成')
    }
  } catch (e) {
    console.error('❌ [AI助手] 停止失败:', e)
  }
}

// 清空对话
const clearChat = () => {
  if (chatHistory.value.length > 0 && !confirm('确定要清空所有对话记录吗？')) {
    return
  }
  chatHistory.value = []
}

// ========================================================
// 服务检查
// ========================================================

const checkServiceAndRefresh = async () => {
  try {
    aiServiceOnline.value = await CheckAIServiceHealth()
    if (aiServiceOnline.value) {
      await loadKnowledgeList()
    } else {
      alert('⚠️ AI 服务离线，请确保 FastAPI 服务运行在 8006')
    }
  } catch (e) {
    aiServiceOnline.value = false
    console.error('服务检查失败:', e)
  }
}

// ========================================================
// 生命周期
// ========================================================

onMounted(async () => {
  // 注册全局事件监听器（只注册一次）
  console.log('🎧 [前端] 注册 AI 流式事件监听器')
  
  // 加载错误码列表
  await loadErrorCodes()
  
  EventsOn('ai-stream-chunk', (chunk) => {
    console.log('📦 [前端] 收到数据块:', chunk)
    if (chunk.type === 'thinking') {
      thinkingText.value = chunk.data
    } else if (chunk.type === 'docs') {
      // 收到文档，保存但不立即创建消息
      currentRelevantDocs = chunk.data || []
      console.log('📚 [前端] 收到相关文档:', currentRelevantDocs.length, '个')
      
      // 如果已经创建了消息框（token先到了），更新文档
      if (chatHistory.value.length > 0 && chatHistory.value[chatHistory.value.length - 1].role === 'assistant') {
        const lastMsg = chatHistory.value[chatHistory.value.length - 1]
        lastMsg.relevantDocs = currentRelevantDocs
        
        // 自动滚动
        nextTick(() => {
          if (chatMessages.value) {
            chatMessages.value.scrollTop = chatMessages.value.scrollHeight
          }
        })
      }
    } else if (chunk.type === 'token') {
      // 收到 token（逐字输出）
      if (currentStreamingAnswer === '') {
        // 第一次收到答案，创建消息框（包含已收到的文档）
        isThinking.value = false
        chatHistory.value.push({
          role: 'assistant',
          content: '',
          relevantDocs: currentRelevantDocs,
          time: new Date().toLocaleTimeString()
        })
      }
      currentStreamingAnswer += chunk.data
      // 更新最后一条消息
      const lastMsg = chatHistory.value[chatHistory.value.length - 1]
      lastMsg.content = currentStreamingAnswer
      
      // 自动滚动
      nextTick(() => {
        if (chatMessages.value) {
          chatMessages.value.scrollTop = chatMessages.value.scrollHeight
        }
      })
    } else if (chunk.type === 'end') {
      // 流结束，确保文档已更新
      if (chatHistory.value.length > 0 && chatHistory.value[chatHistory.value.length - 1].role === 'assistant') {
        const lastMsg = chatHistory.value[chatHistory.value.length - 1]
        lastMsg.relevantDocs = currentRelevantDocs
      }
      console.log('🏁 [前端] 流式输出结束')
    }
  })

  EventsOn('ai-stream-error', (error) => {
    console.error('❌ [前端] 收到错误:', error)
    isThinking.value = false
    chatHistory.value.push({
      role: 'assistant',
      content: '❌ ' + error,
      time: new Date().toLocaleTimeString()
    })
  })
  
  EventsOn('ai-stream-cancelled', () => {
    console.log('🛑 [AI助手] 流式传输已取消')
    isThinking.value = false
    // 如果有部分回答，保存它
    if (currentStreamingAnswer) {
      chatHistory.value.push({
        role: 'assistant',
        content: currentStreamingAnswer + '\n\n⚠️ (已停止生成)',
        relevantDocs: currentRelevantDocs,
        time: new Date().toLocaleTimeString()
      })
      currentStreamingAnswer = ''
      currentRelevantDocs = []
    }
  })

  // 检查服务并刷新数据
  await checkServiceAndRefresh()
})

// 页面卸载前停止正在进行的生成
onBeforeUnmount(() => {
  console.log('🔄 [AI助手] 页面即将卸载，停止正在进行的生成')
  if (isThinking.value) {
    StopAIStream().catch(err => {
      console.error('❌ [AI助手] 停止失败:', err)
    })
  }
})
</script>

<style scoped>
/* 继承全局样式 */
.page-container {
  padding: 40px;
  height: 100vh;
  overflow-y: auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title i {
  color: #78909c;
}

.page-subtitle {
  font-size: 14px;
  color: rgba(255,255,255,0.5);
  margin-top: 4px;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

/* 服务状态 */
.service-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
}

.service-status.online {
  background: rgba(94, 139, 126, 0.2);
  color: #7ea896;
}

.service-status.offline {
  background: rgba(158, 126, 94, 0.2);
  color: #b8967e;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.service-status.online .status-dot {
  background: #7ea896;
}

.service-status.offline .status-dot {
  background: #b8967e;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 按钮样式 */
.action-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn.refresh {
  background: rgba(120, 144, 156, 0.2);
  color: #78909c;
}

.action-btn.refresh:hover {
  background: rgba(120, 144, 156, 0.3);
  transform: translateY(-2px);
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

.stat-card.card-total {
  background: linear-gradient(135deg, #7a5858 0%, #8e6e6e 100%);
}

.stat-card.card-active {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
}

.stat-card.card-success {
  background: linear-gradient(135deg, #556070 0%, #6e7e8e 100%);
}

.stat-card.card-response {
  background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%);
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

/* Tab 容器 */
.tab-container {
  background: rgba(30, 40, 60, 0.6);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  overflow: hidden;
}

.tab-nav {
  display: flex;
  background: rgba(20, 30, 50, 0.4);
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.tab-btn {
  flex: 1;
  padding: 16px 24px;
  border: none;
  background: transparent;
  color: rgba(255,255,255,0.6);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  position: relative;
}

.tab-btn:hover {
  background: rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.9);
}

.tab-btn.active {
  background: rgba(120, 144, 156, 0.2);
  color: #78909c;
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: #78909c;
}

.tab-badge {
  background: rgba(255,255,255,0.2);
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
}

.tab-content {
  padding: 24px;
  min-height: 500px;
}

/* 卡片样式 */
.card {
  background: rgba(30, 40, 60, 0.4);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  overflow: hidden;
}

.card-header {
  padding: 16px 20px;
  background: rgba(20, 30, 50, 0.4);
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  font-size: 16px;
  font-weight: 500;
  color: rgba(255,255,255,0.9);
  display: flex;
  align-items: center;
  gap: 8px;
}

.badge {
  background: rgba(120, 144, 156, 0.2);
  color: #78909c;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

/* 错误码导入卡片 */
.error-code-import-card {
  margin-bottom: 24px;
  background: linear-gradient(135deg, rgba(103, 58, 183, 0.1) 0%, rgba(63, 81, 181, 0.1) 100%);
  border: 1px solid rgba(103, 58, 183, 0.3);
}

.error-code-import-card .card-header {
  background: rgba(103, 58, 183, 0.2);
  border-bottom: 1px solid rgba(103, 58, 183, 0.3);
}

.import-container {
  padding: 20px;
  min-height: 200px;
}

/* 步骤1: 选择文件 */
.import-step-1 .import-info {
  margin-bottom: 16px;
}

.import-info p {
  color: rgba(255,255,255,0.8);
  font-size: 14px;
  margin: 8px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.import-info i {
  color: rgba(103, 58, 183, 0.8);
}

.import-actions {
  display: flex;
  gap: 12px;
}

.btn-download,
.btn-upload {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-download {
  background: linear-gradient(135deg, #42a5f5 0%, #1976d2 100%);
  color: #fff;
}

.btn-download:hover {
  box-shadow: 0 4px 12px rgba(66, 165, 245, 0.4);
  transform: translateY(-2px);
}

.btn-upload {
  background: linear-gradient(135deg, #66bb6a 0%, #43a047 100%);
  color: #fff;
  position: relative;
}

.btn-upload:hover {
  box-shadow: 0 4px 12px rgba(102, 187, 106, 0.4);
  transform: translateY(-2px);
}

/* 步骤2: 预览编辑 */
.import-step-2 {
  animation: fadeIn 0.3s ease;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.preview-header h4 {
  color: rgba(255,255,255,0.9);
  font-size: 16px;
  font-weight: 500;
}

.preview-actions {
  display: flex;
  gap: 12px;
}

.preview-table-container {
  max-height: 500px;
  overflow: auto;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
}

.preview-table {
  width: 100%;
  border-collapse: collapse;
  background: rgba(20, 30, 50, 0.4);
}

.preview-table thead {
  position: sticky;
  top: 0;
  background: rgba(103, 58, 183, 0.3);
  z-index: 10;
}

.preview-table th {
  padding: 12px 8px;
  text-align: left;
  color: rgba(255,255,255,0.9);
  font-size: 13px;
  font-weight: 500;
  border-bottom: 2px solid rgba(103, 58, 183, 0.5);
}

.preview-table td {
  padding: 8px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.table-input,
.table-textarea {
  width: 100%;
  padding: 6px 8px;
  background: rgba(20, 30, 50, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 4px;
  color: #fff;
  font-size: 13px;
  font-family: inherit;
}

.table-input:focus,
.table-textarea:focus {
  outline: none;
  border-color: rgba(103, 58, 183, 0.8);
  background: rgba(20, 30, 50, 0.8);
}

.table-textarea {
  resize: vertical;
  min-height: 40px;
}

.btn-delete {
  padding: 6px 10px;
  background: rgba(244, 67, 54, 0.2);
  border: 1px solid rgba(244, 67, 54, 0.5);
  border-radius: 4px;
  color: #f44336;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-delete:hover {
  background: rgba(244, 67, 54, 0.3);
  transform: scale(1.1);
}

/* 步骤3: 导入进度 */
.import-step-3 {
  animation: fadeIn 0.3s ease;
}

.import-progress {
  text-align: center;
  padding: 40px 20px;
}

.progress-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 20px;
  color: rgba(255,255,255,0.9);
  font-size: 16px;
}

.progress-bar-container {
  width: 100%;
  height: 8px;
  background: rgba(20, 30, 50, 0.6);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 20px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #42a5f5 0%, #66bb6a 100%);
  transition: width 0.3s ease;
}

.progress-details {
  display: flex;
  justify-content: center;
  gap: 24px;
  color: rgba(255,255,255,0.8);
  font-size: 14px;
}

.progress-details p {
  display: flex;
  align-items: center;
  gap: 6px;
}

.progress-errors {
  color: #f44336 !important;
}

/* 步骤4: 完成 */
.import-step-4 {
  animation: fadeIn 0.3s ease;
  text-align: center;
  padding: 20px;
}

.import-result {
  padding: 24px;
  border-radius: 12px;
  margin-bottom: 20px;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  text-align: left;
}

.import-result.success {
  background: rgba(76, 175, 80, 0.2);
  border: 1px solid rgba(76, 175, 80, 0.5);
}

.import-result.warning {
  background: rgba(255, 152, 0, 0.2);
  border: 1px solid rgba(255, 152, 0, 0.5);
}

.import-result i {
  font-size: 32px;
  margin-top: 4px;
}

.import-result.success i {
  color: #4caf50;
}

.import-result.warning i {
  color: #ff9800;
}

.result-content h4 {
  color: rgba(255,255,255,0.9);
  font-size: 18px;
  margin-bottom: 8px;
}

.result-content p {
  color: rgba(255,255,255,0.8);
  font-size: 14px;
  margin: 4px 0;
}

.error-list {
  margin-top: 12px;
  padding: 12px;
  background: rgba(244, 67, 54, 0.1);
  border-radius: 6px;
  max-height: 200px;
  overflow-y: auto;
}

.error-list ul {
  list-style: none;
  padding: 0;
  margin: 8px 0 0 0;
}

.error-list li {
  color: rgba(255,255,255,0.7);
  font-size: 13px;
  padding: 4px 0;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

/* 错误码对照表 */
.error-codes-container {
  padding: 20px;
  min-height: 400px;
}

.search-input {
  padding: 8px 16px;
  background: rgba(20, 30, 50, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  width: 300px;
  transition: all 0.3s;
}

.search-input:focus {
  outline: none;
  border-color: rgba(103, 58, 183, 0.8);
  background: rgba(20, 30, 50, 0.8);
}

.btn-refresh {
  padding: 8px 16px;
  background: rgba(103, 58, 183, 0.2);
  border: 1px solid rgba(103, 58, 183, 0.5);
  border-radius: 6px;
  color: rgba(103, 58, 183, 0.9);
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s;
}

.btn-refresh:hover:not(:disabled) {
  background: rgba(103, 58, 183, 0.3);
  transform: translateY(-1px);
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-codes-table-container {
  max-height: 600px;
  overflow: auto;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
}

.error-codes-table {
  width: 100%;
  border-collapse: collapse;
  background: rgba(20, 30, 50, 0.4);
}

.error-codes-table thead {
  position: sticky;
  top: 0;
  background: rgba(103, 58, 183, 0.3);
  z-index: 10;
}

.error-codes-table th {
  padding: 12px 16px;
  text-align: left;
  color: rgba(255,255,255,0.9);
  font-size: 14px;
  font-weight: 500;
  border-bottom: 2px solid rgba(103, 58, 183, 0.5);
}

.error-codes-table td {
  padding: 12px 16px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
  font-size: 14px;
}

.error-codes-table tbody tr:hover {
  background: rgba(103, 58, 183, 0.1);
}

.error-code-cell {
  text-align: center;
}

.error-code-badge {
  display: inline-block;
  padding: 4px 12px;
  background: linear-gradient(135deg, #f44336 0%, #e91e63 100%);
  color: #fff;
  border-radius: 6px;
  font-weight: 600;
  font-size: 14px;
}

.error-msg-cell {
  line-height: 1.6;
}

.time-cell {
  color: rgba(255,255,255,0.6);
  font-size: 13px;
}

.table-footer {
  padding: 12px 16px;
  text-align: right;
  color: rgba(255,255,255,0.6);
  font-size: 13px;
  border-top: 1px solid rgba(255,255,255,0.1);
}

/* 知识库管理布局 */
.knowledge-grid {
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: 20px;
}

/* 表单样式 */
.form-container {
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: rgba(255,255,255,0.8);
  font-size: 14px;
  font-weight: 500;
}

.required {
  color: #d4a8a8;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 12px;
  background: rgba(20, 30, 50, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  transition: all 0.3s ease;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #78909c;
  background: rgba(20, 30, 50, 0.8);
}

.form-textarea {
  resize: vertical;
  font-family: inherit;
}

.char-count {
  text-align: right;
  font-size: 12px;
  color: rgba(255,255,255,0.5);
  margin-top: 4px;
}

.form-actions {
  display: flex;
  gap: 12px;
}

.btn-primary,
.btn-secondary {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-primary {
  background: rgba(120, 144, 156, 0.3);
  color: #78909c;
  flex: 1;
}

.btn-primary:hover:not(:disabled) {
  background: rgba(120, 144, 156, 0.4);
  transform: translateY(-2px);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.7);
}

.btn-secondary:hover {
  background: rgba(255,255,255,0.15);
}

/* 知识列表 */
.knowledge-list {
  padding: 20px;
  max-height: 600px;
  overflow-y: auto;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: rgba(255,255,255,0.5);
}

.loading-state i,
.empty-state i {
  font-size: 48px;
  margin-bottom: 16px;
  display: block;
}

.knowledge-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.knowledge-item {
  background: rgba(20, 30, 50, 0.4);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  padding: 16px;
  transition: all 0.3s ease;
}

.knowledge-item:hover {
  background: rgba(20, 30, 50, 0.6);
  border-color: rgba(255,255,255,0.2);
}

.knowledge-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.knowledge-meta {
  display: flex;
  gap: 12px;
  align-items: center;
}

.knowledge-id {
  font-size: 11px;
  color: rgba(255,255,255,0.4);
  font-family: monospace;
}

.knowledge-source {
  font-size: 12px;
  color: #78909c;
  background: rgba(120, 144, 156, 0.2);
  padding: 2px 8px;
  border-radius: 4px;
}

.btn-delete {
  padding: 6px 12px;
  background: rgba(142, 110, 110, 0.2);
  border: none;
  border-radius: 6px;
  color: #d4a8a8;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-delete:hover {
  background: rgba(142, 110, 110, 0.3);
}

.knowledge-content {
  color: rgba(255,255,255,0.8);
  font-size: 14px;
  line-height: 1.6;
  margin-bottom: 8px;
}

.knowledge-footer {
  display: flex;
  justify-content: flex-end;
}

.knowledge-time {
  font-size: 11px;
  color: rgba(255,255,255,0.4);
}

/* 聊天容器 */
.chat-container {
  max-width: 1200px;
  margin: 0 auto;
}

.chat-card {
  display: flex;
  flex-direction: column;
  height: 700px;
}

.btn-clear {
  padding: 8px 16px;
  background: rgba(142, 110, 110, 0.2);
  border: none;
  border-radius: 6px;
  color: #d4a8a8;
  cursor: pointer;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s ease;
}

.btn-clear:hover {
  background: rgba(142, 110, 110, 0.3);
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.chat-welcome {
  text-align: center;
  padding: 60px 20px;
  color: rgba(255,255,255,0.7);
}

.chat-welcome i {
  font-size: 64px;
  color: #78909c;
  margin-bottom: 20px;
}

.chat-welcome h3 {
  font-size: 24px;
  margin-bottom: 8px;
  color: #fff;
}

.quick-questions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
  margin-top: 24px;
}

.quick-btn {
  padding: 10px 20px;
  background: rgba(120, 144, 156, 0.2);
  border: 1px solid rgba(120, 144, 156, 0.3);
  border-radius: 20px;
  color: #78909c;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.3s ease;
}

.quick-btn:hover {
  background: rgba(120, 144, 156, 0.3);
  transform: translateY(-2px);
}

.chat-message {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.chat-message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.chat-message.user .message-avatar {
  background: rgba(120, 144, 156, 0.3);
  color: #78909c;
}

.chat-message.assistant .message-avatar {
  background: rgba(94, 139, 126, 0.3);
  color: #7ea896;
}

.message-content {
  flex: 1;
  max-width: 70%;
}

.chat-message.user .message-content {
  text-align: right;
}

.message-text {
  background: rgba(20, 30, 50, 0.6);
  padding: 12px 16px;
  border-radius: 12px;
  color: rgba(255,255,255,0.9);
  font-size: 14px;
  line-height: 1.6;
  display: inline-block;
  max-width: 100%;
  word-wrap: break-word;
}

.chat-message.user .message-text {
  background: rgba(120, 144, 156, 0.2);
}

.message-time {
  font-size: 11px;
  color: rgba(255,255,255,0.4);
  margin-top: 4px;
}

/* 思考动画 */
.thinking-animation {
  background: rgba(20, 30, 50, 0.6);
  padding: 12px 16px;
  border-radius: 12px;
  display: inline-block;
}

.thinking-text {
  color: #78909c;
  font-size: 14px;
  margin-bottom: 8px;
}

.thinking-dots {
  display: flex;
  gap: 4px;
}

.thinking-dots span {
  width: 8px;
  height: 8px;
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
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
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
  display: inline-flex;
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

/* 相关文档 */
.relevant-docs {
  margin-top: 12px;
  background: rgba(20, 30, 50, 0.4);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  padding: 12px;
}

.docs-header {
  font-size: 13px;
  color: #78909c;
  margin-bottom: 8px;
  font-weight: 500;
}

.docs-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.doc-item {
  background: rgba(20, 30, 50, 0.6);
  border: 1px solid rgba(255,255,255,0.05);
  border-radius: 6px;
  padding: 10px;
}

.doc-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.doc-source {
  font-size: 12px;
  color: #78909c;
}

.doc-similarity {
  font-size: 11px;
  color: rgba(255,255,255,0.5);
  background: rgba(255,255,255,0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

.doc-content {
  font-size: 13px;
  color: rgba(255,255,255,0.7);
  line-height: 1.5;
}

/* 输入区域 */
.chat-input-area {
  padding: 20px;
  background: rgba(20, 30, 50, 0.4);
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  gap: 12px;
}

.chat-input {
  flex: 1;
  padding: 12px;
  background: rgba(20, 30, 50, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  resize: none;
  font-family: inherit;
}

.chat-input:focus {
  outline: none;
  border-color: #78909c;
}

.btn-send {
  padding: 12px 24px;
  background: rgba(120, 144, 156, 0.3);
  border: none;
  border-radius: 8px;
  color: #78909c;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s ease;
}

.btn-send:hover:not(:disabled) {
  background: rgba(120, 144, 156, 0.4);
  transform: translateY(-2px);
}

.btn-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 开发中页面 */
.coming-soon {
  text-align: center;
  padding: 80px 20px;
  color: rgba(255,255,255,0.6);
}

.coming-soon i {
  font-size: 64px;
  color: #78909c;
  margin-bottom: 20px;
}

.coming-soon h3 {
  font-size: 24px;
  margin-bottom: 12px;
  color: #fff;
}

.feature-list {
  list-style: none;
  margin-top: 24px;
  display: inline-block;
  text-align: left;
}

.feature-list li {
  padding: 8px 0;
  color: rgba(255,255,255,0.7);
}

.feature-list i {
  color: #7ea896;
  margin-right: 8px;
}

/* 设置页面 */
.settings-container {
  padding: 20px;
}

.setting-item {
  padding: 20px;
  background: rgba(20, 30, 50, 0.4);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.setting-label {
  font-size: 14px;
  color: rgba(255,255,255,0.8);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.setting-value {
  flex: 1;
  max-width: 400px;
  margin-left: 20px;
}

.badge-info {
  background: rgba(120, 144, 156, 0.2);
  color: #78909c;
  padding: 6px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
}

/* 响应式 */
@media (max-width: 1400px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .knowledge-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .tab-nav {
    flex-wrap: wrap;
  }
  
  .tab-btn {
    flex: 1 1 50%;
  }
}
</style>
