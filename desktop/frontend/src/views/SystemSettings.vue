<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-sliders-h"></i>
          系统参数配置
        </div>
        <div class="page-subtitle">System Parameter Configuration</div>
      </div>
      <div class="header-actions">
        <button class="action-btn secondary" @click="loadConfig" title="刷新配置">
          <i class="fas fa-sync-alt"></i> 刷新
        </button>
        <button class="action-btn success" @click="saveAllConfig" title="保存所有配置">
          <i class="fas fa-save"></i> 保存全部
        </button>
      </div>
    </div>

    <!-- 配置卡片区域 -->
    <div class="settings-grid">
      <!-- 生产参数配置 -->
      <div class="card full-width">
        <div class="card-header">
          <h3><i class="fas fa-industry"></i> 生产参数</h3>
          <div class="card-hint">Production Parameters</div>
        </div>
        <div class="card-body">
          <div class="setting-item">
            <div class="setting-label">
              <i class="fas fa-clock"></i>
              <div>
                <div class="label-text">单件加工时间（理论节拍）</div>
                <div class="label-hint">用于计算性能稼动率 = 理论时间 / 实际时间 × 100%</div>
              </div>
            </div>
            <div class="setting-control">
              <input 
                v-model.number="config.productionCoefficient" 
                type="number" 
                min="0.1" 
                step="0.1"
                class="input-field"
                placeholder="例如：100"
              />
              <span class="unit">秒/件</span>
              <button class="btn-save" @click="saveProductionCoefficient">
                <i class="fas fa-save"></i> 保存
              </button>
            </div>
          </div>
          
          <div class="setting-divider"></div>
          
          <div class="setting-item">
            <div class="setting-label">
              <i class="fas fa-business-time"></i>
              <div>
                <div class="label-text">每日理论工作时间</div>
                <div class="label-hint">扣除休息时间后的实际工作时长，用于计算人员稼动率</div>
              </div>
            </div>
            <div class="setting-control">
              <input 
                v-model.number="config.dailyWorkMinutes" 
                type="number" 
                min="60" 
                max="1440"
                step="1"
                class="input-field"
                placeholder="例如：460"
              />
              <span class="unit">分钟</span>
              <span class="unit-hint">({{ (config.dailyWorkMinutes / 60).toFixed(2) }} 小时)</span>
              <button class="btn-save" @click="saveDailyWorkMinutes">
                <i class="fas fa-save"></i> 保存
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 休息时间段配置 -->
      <div class="card full-width">
        <div class="card-header">
          <h3><i class="fas fa-coffee"></i> 休息时间段配置</h3>
          <div class="card-hint">Break Time Configuration</div>
          <button class="action-btn primary" @click="addBreakTime">
            <i class="fas fa-plus"></i> 添加时间段
          </button>
        </div>
        <div class="card-body">
          <div v-if="config.breakTimes.length === 0" class="empty-state">
            <i class="fas fa-inbox"></i>
            <p>暂无休息时间段配置</p>
            <button class="action-btn primary" @click="addBreakTime">
              <i class="fas fa-plus"></i> 添加第一个时间段
            </button>
          </div>
          
          <div v-else class="break-times-list">
            <div 
              v-for="(breakTime, index) in config.breakTimes" 
              :key="breakTime.id"
              class="break-time-item"
            >
              <div class="break-time-number">{{ index + 1 }}</div>
              
              <div class="break-time-field">
                <label>名称</label>
                <input 
                  v-model="breakTime.name" 
                  type="text" 
                  class="input-field"
                  placeholder="例如：午餐休息"
                />
              </div>
              
              <div class="break-time-field">
                <label>开始时间</label>
                <div class="time-input-group">
                  <input 
                    v-model.number="breakTime.start_hour" 
                    type="number" 
                    min="0" 
                    max="23"
                    class="input-field time-input"
                    placeholder="时"
                  />
                  <span class="time-separator">:</span>
                  <input 
                    v-model.number="breakTime.start_min" 
                    type="number" 
                    min="0" 
                    max="59"
                    class="input-field time-input"
                    placeholder="分"
                  />
                </div>
              </div>
              
              <div class="break-time-field">
                <label>结束时间</label>
                <div class="time-input-group">
                  <input 
                    v-model.number="breakTime.end_hour" 
                    type="number" 
                    min="0" 
                    max="23"
                    class="input-field time-input"
                    placeholder="时"
                  />
                  <span class="time-separator">:</span>
                  <input 
                    v-model.number="breakTime.end_min" 
                    type="number" 
                    min="0" 
                    max="59"
                    class="input-field time-input"
                    placeholder="分"
                  />
                </div>
              </div>
              
              <div class="break-time-duration">
                <i class="fas fa-hourglass-half"></i>
                {{ calculateDuration(breakTime) }} 分钟
              </div>
              
              <div class="break-time-actions">
                <button 
                  class="btn-icon danger" 
                  @click="deleteBreakTime(index)"
                  title="删除"
                >
                  <i class="fas fa-trash"></i>
                </button>
              </div>
            </div>
          </div>
          
          <div v-if="config.breakTimes.length > 0" class="break-times-summary">
            <div class="summary-item">
              <i class="fas fa-list"></i>
              <span>共 {{ config.breakTimes.length }} 个休息时间段</span>
            </div>
            <div class="summary-item">
              <i class="fas fa-clock"></i>
              <span>总休息时长：{{ totalBreakMinutes }} 分钟 ({{ (totalBreakMinutes / 60).toFixed(2) }} 小时)</span>
            </div>
            <button class="action-btn success" @click="saveBreakTimes">
              <i class="fas fa-save"></i> 保存休息时间配置
            </button>
          </div>
        </div>
      </div>

      <!-- 配置说明 -->
      <div class="card info-card">
        <div class="card-header">
          <h3><i class="fas fa-info-circle"></i> 配置说明</h3>
        </div>
        <div class="card-body">
          <div class="info-section">
            <h4><i class="fas fa-industry"></i> 单件加工时间（理论节拍）</h4>
            <ul>
              <li>定义：完成一件产品的理论加工时间（秒）</li>
              <li>用途：用于计算设备性能稼动率</li>
              <li>公式：性能稼动率 = (理论节拍 × 良品数) / 实际加工时间 × 100%</li>
              <li>示例：如果理论上每100秒完成1件，则填写 100</li>
            </ul>
          </div>
          
          <div class="info-section">
            <h4><i class="fas fa-business-time"></i> 每日理论工作时间</h4>
            <ul>
              <li>定义：扣除休息时间后的每日实际工作时长（分钟）</li>
              <li>用途：用于计算人员稼动率</li>
              <li>公式：人员稼动率 = 实际工作时长 / 理论工作时长 × 100%</li>
              <li>示例：工作时间 7:40-16:20，扣除休息60分钟，则填写 460 分钟</li>
            </ul>
          </div>
          
          <div class="info-section">
            <h4><i class="fas fa-coffee"></i> 休息时间段</h4>
            <ul>
              <li>定义：每日固定的休息时间段（如午餐、茶歇等）</li>
              <li>用途：计算人员稼动率时自动扣除这些时间</li>
              <li>支持：可以添加多个休息时间段，每个时间段可自定义名称</li>
              <li>示例：午餐休息 11:40-12:20，上午茶歇 9:40-9:50</li>
            </ul>
          </div>
          
          <div class="info-section warning">
            <h4><i class="fas fa-exclamation-triangle"></i> 注意事项</h4>
            <ul>
              <li>修改配置后需要点击"保存"按钮才能生效</li>
              <li>配置修改后会立即影响驾驶舱和各报表的计算结果</li>
              <li>休息时间段不能重叠，结束时间必须晚于开始时间</li>
              <li>建议在非生产时间段进行配置修改</li>
            </ul>
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

    <!-- 确认对话框 -->
    <ConfirmDialog
      v-if="confirmDialog.show"
      :show="true"
      :type="confirmDialog.type"
      :title="confirmDialog.title"
      :message="confirmDialog.message"
      :details="confirmDialog.details"
      :confirm-text="confirmDialog.confirmText"
      :cancel-text="confirmDialog.cancelText"
      @confirm="confirmDialog.onConfirm"
      @cancel="confirmDialog.onCancel"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Toast from '../components/Toast.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'

// 配置数据
const config = ref({
  productionCoefficient: 100.0,
  dailyWorkMinutes: 460,
  breakTimes: []
})

// Toast 提示
const toast = ref({
  show: false,
  message: '',
  type: 'success',
  duration: 2000
})

// 确认对话框
const confirmDialog = ref({
  show: false,
  type: 'warning',
  title: '确认删除',
  message: '',
  details: [],
  confirmText: '确认',
  cancelText: '取消',
  onConfirm: () => {},
  onCancel: () => {}
})

// 计算总休息时长
const totalBreakMinutes = computed(() => {
  return config.value.breakTimes.reduce((total, breakTime) => {
    return total + calculateDuration(breakTime)
  }, 0)
})

// 计算单个休息时间段的时长
const calculateDuration = (breakTime) => {
  const startMinutes = breakTime.start_hour * 60 + breakTime.start_min
  const endMinutes = breakTime.end_hour * 60 + breakTime.end_min
  return Math.max(0, endMinutes - startMinutes)
}

// 显示提示
const showToast = (message, type = 'success') => {
  toast.value.message = message
  toast.value.type = type
  toast.value.show = true
}

// 加载配置
const loadConfig = async () => {
  try {
    if (window.go?.main?.App?.GetSystemConfig) {
      const result = await window.go.main.App.GetSystemConfig()
      if (result) {
        config.value = {
          productionCoefficient: result.production_coefficient || 100.0,
          dailyWorkMinutes: result.daily_work_minutes || 460,
          breakTimes: (result.break_times || []).map(bt => ({
            id: bt.id,
            name: bt.name,
            start_hour: bt.start_hour,
            start_min: bt.start_min,
            end_hour: bt.end_hour,
            end_min: bt.end_min
          }))
        }
        console.log('✅ 加载配置成功:', config.value)
      }
    }
  } catch (e) {
    console.error('❌ 加载配置失败:', e)
    showToast('加载配置失败: ' + e.message, 'error')
  }
}

// 保存单件加工时间
const saveProductionCoefficient = async () => {
  try {
    if (config.value.productionCoefficient <= 0) {
      showToast('单件加工时间必须大于0', 'warning')
      return
    }
    
    if (window.go?.main?.App?.SetProductionCoefficient) {
      await window.go.main.App.SetProductionCoefficient(config.value.productionCoefficient)
      showToast('保存成功！单件加工时间已更新', 'success')
      console.log('✅ 保存单件加工时间:', config.value.productionCoefficient)
    }
  } catch (e) {
    console.error('❌ 保存失败:', e)
    showToast('保存失败: ' + e.message, 'error')
  }
}

// 保存每日工作时间
const saveDailyWorkMinutes = async () => {
  try {
    if (config.value.dailyWorkMinutes <= 0 || config.value.dailyWorkMinutes > 1440) {
      showToast('每日工作时间必须在1-1440分钟之间', 'warning')
      return
    }
    
    if (window.go?.main?.App?.SetDailyWorkMinutes) {
      await window.go.main.App.SetDailyWorkMinutes(config.value.dailyWorkMinutes)
      showToast('保存成功！每日工作时间已更新', 'success')
      console.log('✅ 保存每日工作时间:', config.value.dailyWorkMinutes)
    }
  } catch (e) {
    console.error('❌ 保存失败:', e)
    showToast('保存失败: ' + e.message, 'error')
  }
}

// 添加休息时间段
const addBreakTime = () => {
  const newId = config.value.breakTimes.length > 0 
    ? Math.max(...config.value.breakTimes.map(bt => bt.id)) + 1 
    : 1
    
  config.value.breakTimes.push({
    id: newId,
    name: '休息时间' + newId,
    start_hour: 12,
    start_min: 0,
    end_hour: 13,
    end_min: 0
  })
}

// 删除休息时间段
const deleteBreakTime = (index) => {
  const breakTime = config.value.breakTimes[index]
  
  confirmDialog.value = {
    show: true,
    type: 'warning',
    title: '确认删除',
    message: `确定要删除休息时间段"${breakTime.name}"吗？`,
    details: [
      `时间段：${String(breakTime.start_hour).padStart(2, '0')}:${String(breakTime.start_min).padStart(2, '0')} - ${String(breakTime.end_hour).padStart(2, '0')}:${String(breakTime.end_min).padStart(2, '0')}`,
      `时长：${calculateDuration(breakTime)} 分钟`
    ],
    confirmText: '确认删除',
    cancelText: '取消',
    onConfirm: () => {
      config.value.breakTimes.splice(index, 1)
      confirmDialog.value.show = false
      showToast('删除成功！记得保存配置', 'success')
    },
    onCancel: () => {
      confirmDialog.value.show = false
    }
  }
}

// 保存休息时间配置
const saveBreakTimes = async () => {
  try {
    // 验证时间段
    for (const breakTime of config.value.breakTimes) {
      if (!breakTime.name || breakTime.name.trim() === '') {
        showToast('休息时间段名称不能为空', 'warning')
        return
      }
      
      if (breakTime.start_hour < 0 || breakTime.start_hour > 23 || 
          breakTime.end_hour < 0 || breakTime.end_hour > 23) {
        showToast('小时必须在0-23之间', 'warning')
        return
      }
      
      if (breakTime.start_min < 0 || breakTime.start_min > 59 || 
          breakTime.end_min < 0 || breakTime.end_min > 59) {
        showToast('分钟必须在0-59之间', 'warning')
        return
      }
      
      const duration = calculateDuration(breakTime)
      if (duration <= 0) {
        showToast(`"${breakTime.name}"的结束时间必须晚于开始时间`, 'warning')
        return
      }
    }
    
    if (window.go?.main?.App?.SetBreakTimes) {
      // 转换为后端需要的格式
      const breakTimesData = config.value.breakTimes.map(bt => ({
        id: bt.id,
        name: bt.name,
        start_hour: bt.start_hour,
        start_min: bt.start_min,
        end_hour: bt.end_hour,
        end_min: bt.end_min
      }))
      
      await window.go.main.App.SetBreakTimes(breakTimesData)
      showToast('保存成功！休息时间配置已更新', 'success')
      console.log('✅ 保存休息时间配置:', breakTimesData)
    }
  } catch (e) {
    console.error('❌ 保存失败:', e)
    showToast('保存失败: ' + e.message, 'error')
  }
}

// 保存所有配置
const saveAllConfig = async () => {
  try {
    await saveProductionCoefficient()
    await saveDailyWorkMinutes()
    await saveBreakTimes()
    showToast('所有配置保存成功！', 'success')
  } catch (e) {
    console.error('❌ 保存配置失败:', e)
    showToast('保存配置失败: ' + e.message, 'error')
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
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
  color: #546e7a;
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
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s ease;
}

.action-btn.primary {
  background: rgba(84, 110, 122, 0.3);
  color: #546e7a;
  border: 1px solid rgba(84, 110, 122, 0.3);
}

.action-btn.primary:hover {
  background: rgba(84, 110, 122, 0.4);
  transform: translateY(-2px);
}

.action-btn.secondary {
  background: rgba(120, 144, 156, 0.2);
  color: #78909c;
  border: 1px solid rgba(120, 144, 156, 0.2);
}

.action-btn.secondary:hover {
  background: rgba(120, 144, 156, 0.3);
  transform: translateY(-2px);
}

.action-btn.success {
  background: rgba(94, 139, 126, 0.3);
  color: #7ea896;
  border: 1px solid rgba(94, 139, 126, 0.3);
}

.action-btn.success:hover {
  background: rgba(94, 139, 126, 0.4);
  transform: translateY(-2px);
}

/* 配置网格 */
.settings-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card {
  background: rgba(30, 40, 60, 0.6);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  overflow: hidden;
}

.card.full-width {
  grid-column: 1 / -1;
}

.card.info-card {
  grid-column: 1 / -1;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(20, 30, 50, 0.4);
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

.card-hint {
  font-size: 12px;
  color: rgba(255,255,255,0.5);
  margin-left: auto;
  margin-right: 15px;
}

.card-body {
  padding: 24px;
}

/* 设置项 */
.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 32px;
  padding: 20px 0;
}

.setting-label {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.setting-label > i {
  font-size: 20px;
  color: #546e7a;
  margin-top: 3px;
  flex-shrink: 0;
}

.label-text {
  font-size: 15px;
  font-weight: 500;
  color: rgba(255,255,255,0.9);
  margin-bottom: 6px;
}

.label-hint {
  font-size: 13px;
  color: rgba(255,255,255,0.5);
  line-height: 1.6;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.input-field {
  padding: 12px 18px;
  background: rgba(20, 30, 50, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 15px;
  outline: none;
  transition: all 0.3s ease;
  width: 180px;
  font-weight: 500;
}

.input-field:focus {
  border-color: #546e7a;
  background: rgba(20, 30, 50, 0.8);
  box-shadow: 0 0 0 3px rgba(84, 110, 122, 0.1);
}

.input-field::placeholder {
  color: rgba(255, 255, 255, 0.3);
}

.unit {
  font-size: 15px;
  color: rgba(255,255,255,0.7);
  white-space: nowrap;
  font-weight: 500;
  min-width: 50px;
}

.unit-hint {
  font-size: 13px;
  color: rgba(255,255,255,0.5);
  white-space: nowrap;
}

.btn-save {
  padding: 12px 24px;
  background: rgba(94, 139, 126, 0.3);
  border: 1px solid rgba(94, 139, 126, 0.3);
  border-radius: 8px;
  color: #7ea896;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.btn-save:hover {
  background: rgba(94, 139, 126, 0.4);
  transform: translateY(-2px);
}

.setting-divider {
  height: 1px;
  background: rgba(255,255,255,0.05);
  margin: 20px 0;
}

/* 休息时间段列表 */
.break-times-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.break-time-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: rgba(20, 30, 50, 0.4);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  transition: all 0.3s ease;
}

.break-time-item:hover {
  background: rgba(20, 30, 50, 0.6);
  border-color: rgba(255,255,255,0.2);
}

.break-time-number {
  width: 36px;
  height: 36px;
  background: rgba(84, 110, 122, 0.3);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  color: #546e7a;
  flex-shrink: 0;
}

.break-time-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.break-time-field label {
  font-size: 12px;
  color: rgba(255,255,255,0.6);
  font-weight: 500;
}

.break-time-field .input-field {
  width: 200px;
}

.time-input-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.time-input {
  width: 70px !important;
  text-align: center;
  font-size: 15px;
  font-weight: 500;
}

.time-separator {
  font-size: 16px;
  color: rgba(255,255,255,0.5);
  font-weight: 600;
}

.break-time-duration {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: rgba(94, 139, 126, 0.2);
  border: 1px solid rgba(94, 139, 126, 0.3);
  border-radius: 6px;
  color: #7ea896;
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  margin-left: auto;
}

.break-time-actions {
  display: flex;
  gap: 8px;
}

.btn-icon {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.3s ease;
}

.btn-icon.danger {
  background: rgba(142, 110, 110, 0.2);
  color: #d4a8a8;
  border: 1px solid rgba(142, 110, 110, 0.3);
}

.btn-icon.danger:hover {
  background: rgba(142, 110, 110, 0.3);
  transform: scale(1.05);
}

/* 休息时间汇总 */
.break-times-summary {
  margin-top: 20px;
  padding: 20px;
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255,255,255,0.7);
  font-size: 14px;
}

.summary-item i {
  color: #546e7a;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: rgba(255,255,255,0.5);
}

.empty-state i {
  font-size: 48px;
  opacity: 0.3;
  margin-bottom: 16px;
  color: #546e7a;
}

.empty-state p {
  font-size: 14px;
  margin-bottom: 20px;
}

/* 配置说明 */
.info-section {
  margin-bottom: 20px;
  padding: 20px;
  background: rgba(20, 30, 50, 0.4);
  border-left: 3px solid #546e7a;
  border-radius: 6px;
}

.info-section.warning {
  border-left-color: #9e7e5e;
}

.info-section h4 {
  font-size: 15px;
  font-weight: 600;
  color: #546e7a;
  margin: 0 0 12px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-section.warning h4 {
  color: #9e7e5e;
}

.info-section ul {
  margin: 0;
  padding-left: 20px;
  color: rgba(255,255,255,0.7);
  font-size: 13px;
  line-height: 1.8;
}

.info-section li {
  margin-bottom: 6px;
}

/* 响应式 */
@media (max-width: 1200px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .setting-control {
    width: 100%;
  }
  
  .input-field {
    flex: 1;
    min-width: 200px;
  }
}

@media (max-width: 768px) {
  .page-container {
    padding: 20px;
  }
  
  .break-time-item {
    flex-wrap: wrap;
    gap: 12px;
  }
  
  .break-time-duration {
    margin-left: 0;
    width: 100%;
  }
  
  .break-time-field .input-field {
    width: 100%;
  }
}
</style>

