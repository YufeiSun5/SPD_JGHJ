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
                <div class="label-text">默认单件加工时间（全局理论节拍）</div>
                <div class="label-hint">设备未单独设置时使用此默认值，可在时间安排的设备列表中为每台设备单独配置</div>
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

          <!-- 逻辑日工时汇总（由班次配置自动计算，只读展示） -->
          <div class="setting-item">
            <div class="setting-label">
              <i class="fas fa-business-time"></i>
              <div>
                <div class="label-text">逻辑日工时统计</div>
                <div class="label-hint">由下方班次配置自动计算，无需手动填写</div>
              </div>
            </div>
          </div>

          <div v-if="dailyWorkSummary.rows.length === 0" class="daily-work-empty">
            <i class="fas fa-info-circle"></i>
            暂未配置任何启用的班次，请在下方「工作安排」中添加班次
          </div>

          <div v-else class="daily-work-table">
            <div class="dwt-header">
              <span class="dwt-col name">班次</span>
              <span class="dwt-col window">时间段</span>
              <span class="dwt-col dur">总时长</span>
              <span class="dwt-col brk">休息</span>
              <span class="dwt-col net">净工时</span>
            </div>
            <div
              v-for="row in dailyWorkSummary.rows"
              :key="row.name"
              class="dwt-row"
            >
              <span class="dwt-col name">{{ row.name }}</span>
              <span class="dwt-col window">{{ row.window }}</span>
              <span class="dwt-col dur">{{ fmtMins(row.shiftMins) }}</span>
              <span class="dwt-col brk">- {{ row.breakMins }} 分</span>
              <span class="dwt-col net highlight">{{ fmtMins(row.netMins) }}</span>
            </div>
            <div class="dwt-total">
              <span class="dwt-col name">逻辑日合计</span>
              <span class="dwt-col window"></span>
              <span class="dwt-col dur"></span>
              <span class="dwt-col brk"></span>
              <span class="dwt-col net highlight bold">
                {{ fmtMins(dailyWorkSummary.totalNet) }}
                <span class="net-mins">（{{ dailyWorkSummary.totalNet }} 分钟）</span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 工作安排（两层：时间安排组 → 班次） -->
      <div class="card full-width">
        <div class="card-header">
          <h3><i class="fas fa-layer-group"></i> 工作安排</h3>
          <div class="card-hint">时间安排 → 班次 → 休息段 · 设备关联到「时间安排」</div>
          <button class="action-btn primary" @click="addSchedule">
            <i class="fas fa-plus"></i> 添加时间安排
          </button>
        </div>
        <div class="card-body">
          <div v-if="schedules.length === 0" class="empty-state">
            <i class="fas fa-layer-group"></i>
            <p>暂无时间安排，点击右上角「添加时间安排」开始配置</p>
            <button class="action-btn primary" @click="addSchedule">
              <i class="fas fa-plus"></i> 添加第一个时间安排
            </button>
          </div>

          <!-- 时间安排组列表 -->
          <div v-else class="schedules-list">
            <div
              v-for="(sched, gi) in schedules"
              :key="sched.id || gi"
              class="schedule-block"
            >
              <!-- 时间安排组标题行 -->
              <div class="schedule-header" @click="toggleSchedule(gi)">
                <div class="shift-badge sched-badge">{{ gi + 1 }}</div>
                <div class="shift-meta">
                  <span class="shift-name-display">{{ sched.name || '未命名时间安排' }}</span>
                  <span class="shift-window-tag">{{ sched.shifts.length }} 个班次</span>
                  <span :class="['shift-status-tag', sched.is_active ? 'active' : 'inactive']">
                    {{ sched.is_active ? '启用' : '停用' }}
                  </span>
                  <span v-if="scheduleDeviceMap[sched.id] && scheduleDeviceMap[sched.id].length" class="device-count-tag">
                    <i class="fas fa-microchip"></i>
                    {{ scheduleDeviceMap[sched.id].length }} 台设备
                  </span>
                </div>
                <div class="shift-header-right" @click.stop>
                  <button class="btn-icon danger" @click="deleteSchedule(gi)" title="删除时间安排">
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
                <div class="shift-expand-icon" :class="{ expanded: expandedSchedules.has(gi) }">
                  <i class="fas fa-chevron-down"></i>
                </div>
              </div>

              <!-- 展开内容 -->
              <div v-if="expandedSchedules.has(gi)" class="schedule-body">
                <!-- 时间安排基本信息 -->
                <div class="shift-fields-row sched-fields">
                  <div class="shift-field">
                    <label>名称</label>
                    <input v-model="sched.name" type="text" class="input-field" placeholder="如：三班制" />
                  </div>
                  <div class="shift-field shift-field-toggle">
                    <label>是否启用</label>
                    <button :class="['toggle-btn', sched.is_active ? 'on' : 'off']" @click="sched.is_active = !sched.is_active">
                      <i :class="sched.is_active ? 'fas fa-check-circle' : 'fas fa-times-circle'"></i>
                      {{ sched.is_active ? '启用' : '停用' }}
                    </button>
                  </div>
                </div>

                <!-- 设备关联区（关联到安排组级别） -->
                <div v-if="sched.id > 0" class="shift-devices-section">
                  <div class="shift-devices-header">
                    <span><i class="fas fa-microchip"></i> 关联设备</span>
                    <span class="section-hint">一台设备只能属于一个时间安排，点击切换</span>
                  </div>
                  <div class="device-assign-list">
                    <div
                      v-for="dev in allDevices"
                      :key="dev.id"
                      :class="['device-assign-item', deviceAssignedToSchedule(dev.id, sched.id) ? 'assigned' : '']"
                    >
                      <div class="device-assign-left" @click="toggleDeviceSchedule(dev.id, sched.id)"
                        :title="deviceAssignedToSchedule(dev.id, sched.id) ? '点击取消关联' : '点击关联到此时间安排'">
                        <i :class="deviceAssignedToSchedule(dev.id, sched.id) ? 'fas fa-check-circle' : 'far fa-circle'"></i>
                        <span class="device-assign-name">{{ dev.device_name }}</span>
                        <span class="device-assign-code">{{ dev.device_code }}</span>
                      </div>
                      <div class="device-ct-inline" @click.stop>
                        <span class="ct-label">CT</span>
                        <input
                          type="number" min="0" step="0.1"
                          class="ct-input-small"
                          :value="dev.cycle_time || ''"
                          :placeholder="String(config.productionCoefficient || 100)"
                          @change="saveDeviceCT(dev, $event)"
                        />
                        <span class="ct-unit">s</span>
                      </div>
                    </div>
                    <div v-if="!allDevices.length" class="breaks-empty">暂无设备</div>
                  </div>
                </div>
                <div v-else class="shift-devices-section">
                  <div class="breaks-empty"><i class="fas fa-info-circle"></i> 请先保存后再关联设备</div>
                </div>

                <!-- 班次列表（第二层） -->
                <div class="sched-shifts-section">
                  <div class="shift-breaks-header">
                    <span><i class="fas fa-clock"></i> 班次列表</span>
                    <button class="action-btn primary small" @click="addShiftToSchedule(gi)">
                      <i class="fas fa-plus"></i> 添加班次
                    </button>
                  </div>

                  <div v-if="sched.shifts.length === 0" class="breaks-empty">
                    暂无班次，点击右侧按钮添加
                  </div>

                  <div v-else class="shifts-list inner-shifts">
                    <div
                      v-for="(shift, si) in sched.shifts"
                      :key="shift.id || si"
                      class="shift-block inner-shift-block"
                      :class="{ 'drag-over': dragOverIndex === gi * 100 + si }"
                      draggable="true"
                      @dragstart="onShiftDragStart(gi, si, $event)"
                      @dragover.prevent="onShiftDragOver(gi, si)"
                      @drop.prevent="onShiftDrop(gi, si)"
                      @dragend="onShiftDragEnd"
                    >
                      <!-- 班次标题行 -->
                      <div class="shift-header" @click="toggleShift(gi, si)">
                        <div class="shift-drag-handle" @click.stop title="拖拽排序">
                          <i class="fas fa-grip-vertical"></i>
                        </div>
                        <div class="shift-badge">{{ si + 1 }}</div>
                        <div class="shift-meta">
                          <span class="shift-name-display">{{ shift.name || '未命名班次' }}</span>
                          <span class="shift-window-tag">
                            {{ padTime(shift.start_hour, shift.start_min) }}
                            <i class="fas fa-long-arrow-alt-right"></i>
                            {{ padTime(shift.end_hour, shift.end_min) }}
                            <span class="shift-dur-tag">{{ shiftDurationLabel(shift) }}</span>
                          </span>
                          <span :class="['shift-status-tag', shift.is_active ? 'active' : 'inactive']">
                            {{ shift.is_active ? '启用' : '停用' }}
                          </span>
                        </div>
                        <div class="shift-header-right" @click.stop>
                          <span class="break-count-tag">
                            <i class="fas fa-coffee"></i> {{ shift.breaks.length }} 个休息段
                          </span>
                          <button class="btn-icon danger" @click="deleteShiftFromSchedule(gi, si)" title="删除班次">
                            <i class="fas fa-trash"></i>
                          </button>
                        </div>
                        <div class="shift-expand-icon" :class="{ expanded: expandedShifts.has(gi * 1000 + si) }">
                          <i class="fas fa-chevron-down"></i>
                        </div>
                      </div>

                      <!-- 班次展开内容 -->
                      <div v-if="expandedShifts.has(gi * 1000 + si)" class="shift-body">
                        <div class="shift-fields-row">
                          <div class="shift-field">
                            <label>班次名称</label>
                            <input v-model="shift.name" type="text" class="input-field shift-name-input" placeholder="如：早班" />
                          </div>
                          <div class="shift-field">
                            <label>开始时间</label>
                            <input type="time" :value="padTime(shift.start_hour, shift.start_min)"
                              @change="applyTime($event.target.value, shift, 'start')" class="input-field time-combined" />
                          </div>
                          <div class="shift-field">
                            <label>结束时间</label>
                            <input type="time" :value="padTime(shift.end_hour, shift.end_min)"
                              @change="applyTime($event.target.value, shift, 'end')" class="input-field time-combined" />
                          </div>
                          <div class="shift-field shift-field-toggle">
                            <label>是否启用</label>
                            <button :class="['toggle-btn', shift.is_active ? 'on' : 'off']" @click="shift.is_active = !shift.is_active">
                              <i :class="shift.is_active ? 'fas fa-check-circle' : 'fas fa-times-circle'"></i>
                              {{ shift.is_active ? '启用' : '停用' }}
                            </button>
                          </div>
                          <div class="shift-duration-display">
                            <i class="fas fa-hourglass-half"></i> 班次时长：{{ shiftDurationLabel(shift) }}
                          </div>
                        </div>

                        <!-- 休息时间段 -->
                        <div class="shift-breaks-section">
                          <div class="shift-breaks-header">
                            <span><i class="fas fa-coffee"></i> 班内休息时间段</span>
                            <button class="action-btn primary small" @click="addBreakToShift(gi, si)">
                              <i class="fas fa-plus"></i> 添加休息段
                            </button>
                          </div>
                          <div v-if="shift.breaks.length === 0" class="breaks-empty">暂无休息时间段</div>
                          <div v-else class="break-times-list">
                            <div v-for="(brk, bi) in shift.breaks" :key="bi" class="break-time-item">
                              <div class="break-time-number">{{ bi + 1 }}</div>
                              <div class="break-time-field">
                                <label>名称</label>
                                <input v-model="brk.name" type="text" class="input-field" placeholder="如：午餐休息" />
                              </div>
                              <div class="break-time-field">
                                <label>开始</label>
                                <input type="time" :value="padTime(brk.start_hour, brk.start_min)"
                                  @change="applyTime($event.target.value, brk, 'start')" class="input-field time-combined" />
                              </div>
                              <div class="break-time-field">
                                <label>结束</label>
                                <input type="time" :value="padTime(brk.end_hour, brk.end_min)"
                                  @change="applyTime($event.target.value, brk, 'end')" class="input-field time-combined" />
                              </div>
                              <div class="break-time-duration">
                                <i class="fas fa-hourglass-half"></i> {{ calculateDuration(brk) }} 分钟
                              </div>
                              <div class="break-time-actions">
                                <button class="btn-icon danger" @click="deleteBreakFromShift(gi, si, bi)" title="删除">
                                  <i class="fas fa-trash"></i>
                                </button>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 底部保存按钮 -->
          <div v-if="schedules.length > 0" class="shifts-footer">
            <div class="summary-item">
              <i class="fas fa-layer-group"></i>
              <span>共 {{ schedules.length }} 个时间安排，合计 {{ schedules.reduce((s,g)=>s+g.shifts.length,0) }} 个班次</span>
            </div>
            <button class="action-btn success" @click="saveSchedules">
              <i class="fas fa-save"></i> 保存时间安排配置
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
            <h4><i class="fas fa-business-time"></i> 逻辑日工时统计</h4>
            <ul>
              <li>定义：所有启用班次的净工时之和（班次时长 − 各班次内休息时长）</li>
              <li>用途：自动用于计算人员稼动率，无需手动填写</li>
              <li>公式：净工时 = 班次结束时间 − 班次开始时间 − 该班次所有休息时长</li>
              <li>示例：早班 07:40-16:20（8小时40分钟），扣除 60 分钟休息，净工时 = 460 分钟</li>
            </ul>
          </div>
          
          <div class="info-section">
            <h4><i class="fas fa-layer-group"></i> 工作安排（班次 + 休息段）</h4>
            <ul>
              <li>一个"班次"对应一段连续的工作时间，如早班 07:40–16:20</li>
              <li>每个班次可以单独配置多个休息时间段（如午餐 11:40–12:20）</li>
              <li>班次可设为"停用"，停用后不参与工时计算但保留配置</li>
              <li>可配置多个班次（如早班+晚班），各班次净工时累加为逻辑日工时</li>
            </ul>
          </div>

          <div class="info-section warning">
            <h4><i class="fas fa-exclamation-triangle"></i> 注意事项</h4>
            <ul>
              <li>修改后点击「保存班次配置」或页面顶部「保存全部」才能生效</li>
              <li>保存后会自动更新驾驶舱人员稼动率的基准工时</li>
              <li>各班次时间不应相互重叠；休息时间段应落在所属班次窗口内</li>
              <li>结束时间必须晚于开始时间（暂不支持跨零点班次）</li>
            </ul>
          </div>
        </div>
      </div>
      <!-- 设备独立节拍（CT）配置 -->
      <div class="card full-width">
        <div class="card-header">
          <h3><i class="fas fa-tachometer-alt"></i> 设备独立节拍（CT）配置</h3>
          <div class="card-hint">Device Cycle-Time Configuration</div>
          <button class="action-btn secondary small" style="margin-left:auto" @click="refreshDeviceCT">
            <i class="fas fa-sync-alt"></i> 刷新
          </button>
        </div>
        <div class="card-body">
          <div class="info-section info" style="margin-bottom:12px">
            <p style="margin:0;font-size:13px">
              <i class="fas fa-info-circle"></i>
              为每台设备单独配置理论节拍（CT）；留空表示使用上方「全局默认 CT」。修改后即时生效，无需点击「保存全部」。
            </p>
          </div>

          <div v-if="!allDevices.length" class="breaks-empty">
            <i class="fas fa-info-circle"></i> 暂无设备，请先在设备管理中添加设备
          </div>

          <div v-else class="device-ct-table">
            <div class="device-ct-header">
              <span class="dct-col dct-name">设备名称</span>
              <span class="dct-col dct-code">设备编号</span>
              <span class="dct-col dct-sched">时间安排组</span>
              <span class="dct-col dct-ct">独立 CT（秒/件）</span>
              <span class="dct-col dct-eff">实际生效 CT</span>
            </div>
            <div
              v-for="dev in allDevices"
              :key="dev.id"
              class="device-ct-row"
              :class="{ 'ct-overridden': dev.cycle_time && dev.cycle_time > 0 }"
            >
              <span class="dct-col dct-name">
                <i class="fas fa-microchip" style="margin-right:5px;color:rgba(100,180,220,0.7)"></i>
                {{ dev.device_name }}
              </span>
              <span class="dct-col dct-code">{{ dev.device_code || '—' }}</span>
              <span class="dct-col dct-sched">
                <template v-if="dev.schedule_id">
                  <span class="sched-tag">{{ schedules.find(s => s.id === dev.schedule_id)?.name || '时间安排 #' + dev.schedule_id }}</span>
                </template>
                <span v-else class="no-sched-tag">未分配</span>
              </span>
              <span class="dct-col dct-ct">
                <div class="ct-edit-row">
                  <input
                    type="number" min="0" step="0.1"
                    class="ct-input-full"
                    :value="dev.cycle_time || ''"
                    :placeholder="String(config.productionCoefficient || 100)"
                    @change="saveDeviceCT(dev, $event)"
                  />
                  <span class="ct-unit">s</span>
                  <button
                    v-if="dev.cycle_time && dev.cycle_time > 0"
                    class="btn-clear-ct"
                    title="清除，恢复为全局默认"
                    @click="clearDeviceCT(dev)"
                  >
                    <i class="fas fa-times"></i>
                  </button>
                </div>
              </span>
              <span class="dct-col dct-eff">
                <span :class="['eff-ct-val', dev.cycle_time && dev.cycle_time > 0 ? 'eff-custom' : 'eff-global']">
                  {{ dev.cycle_time && dev.cycle_time > 0 ? dev.cycle_time + ' s' : (config.productionCoefficient || 100) + ' s（全局）' }}
                </span>
              </span>
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

// ─── 班次时间配置 ────────────────────────────────────────
// shifts: 班次列表，每项含 name/start_hour/start_min/end_hour/end_min/is_active/breaks[]
// ─── 时间安排组状态 ──────────────────────────────────────
const schedules = ref([])
// expandedSchedules: 当前展开的时间安排组索引
const expandedSchedules = ref(new Set())
// expandedShifts: key = gi*1000+si，展开的班次
const expandedShifts = ref(new Set())

// ─── 拖拽排序状态（班次内拖拽） ─────────────────────────
let dragSrcGi = -1
let dragSrcSi = -1
const dragOverIndex = ref(-1)

const onShiftDragStart = (gi, si, e) => {
  dragSrcGi = gi; dragSrcSi = si
  e.dataTransfer.effectAllowed = 'move'
}
const onShiftDragOver = (gi, si) => { dragOverIndex.value = gi * 100 + si }
const onShiftDrop = (gi, si) => {
  if (dragSrcGi < 0 || (dragSrcGi === gi && dragSrcSi === si)) return
  if (dragSrcGi !== gi) return // 只允许同组内拖拽
  const arr = [...schedules.value[gi].shifts]
  const [moved] = arr.splice(dragSrcSi, 1)
  arr.splice(si, 0, moved)
  schedules.value[gi].shifts = arr
  expandedShifts.value = new Set()
  dragSrcGi = -1; dragSrcSi = -1; dragOverIndex.value = -1
}
const onShiftDragEnd = () => { dragSrcGi = -1; dragSrcSi = -1; dragOverIndex.value = -1 }

// ─── 设备列表 & 设备-时间安排组关联 ──────────────────────
const allDevices = ref([])
// scheduleDeviceMap: scheduleID → deviceID[]（标题栏计数）
const scheduleDeviceMap = ref({})
// deviceScheduleMap: deviceID → scheduleID（快速判断）
const deviceScheduleMap = ref({})

const loadDevices = async () => {
  try {
    if (!window.go?.main?.App?.GetAllDevices) return
    const result = await window.go.main.App.GetAllDevices()
    allDevices.value = (result || []).map(d => ({
      id: d.id,
      device_name: d.device_name,
      device_code: d.device_code,
      schedule_id: d.schedule_id ?? null,
      cycle_time: d.cycle_time ?? null
    }))
    rebuildDeviceMaps()
  } catch (e) { console.error('加载设备列表失败:', e) }
}

// refreshDeviceCT 重新从数据库读取所有设备 CT（用于独立配置卡片刷新）
// Refresh device CT values from DB (for the standalone CT config card).
// DB から全設備の CT を再読み込み（独立 CT 設定カード用）。
const refreshDeviceCT = async () => {
  await loadDevices()
  showToast('已刷新设备节拍数据', 'success')
}

const rebuildDeviceMaps = () => {
  const sdMap = {}
  const dsMap = {}
  for (const d of allDevices.value) {
    dsMap[d.id] = d.schedule_id
    if (d.schedule_id) {
      if (!sdMap[d.schedule_id]) sdMap[d.schedule_id] = []
      sdMap[d.schedule_id].push(d.id)
    }
  }
  scheduleDeviceMap.value = sdMap
  deviceScheduleMap.value = dsMap
}

const deviceAssignedToSchedule = (deviceID, scheduleID) => {
  return deviceScheduleMap.value[deviceID] === scheduleID
}

const toggleDeviceSchedule = async (deviceID, scheduleID) => {
  try {
    const current = deviceScheduleMap.value[deviceID]
    const newSchedID = (current === scheduleID) ? 0 : scheduleID
    await window.go.main.App.SetDeviceSchedule(deviceID, newSchedID)
    const dev = allDevices.value.find(d => d.id === deviceID)
    if (dev) dev.schedule_id = newSchedID > 0 ? newSchedID : null
    rebuildDeviceMaps()
    showToast(newSchedID > 0 ? '设备已关联到此时间安排' : '设备关联已解除', 'success')
  } catch (e) {
    showToast('操作失败: ' + e.message, 'error')
  }
}

const saveDeviceCT = async (dev, event) => {
  try {
    const val = parseFloat(event.target.value)
    const ct = isNaN(val) || val <= 0 ? 0 : val
    await window.go.main.App.SetDeviceCycleTime(dev.id, ct)
    dev.cycle_time = ct > 0 ? ct : null
    showToast(ct > 0 ? `${dev.device_name} CT已设为 ${ct}s` : `${dev.device_name} CT已恢复为全局默认`, 'success')
  } catch (e) {
    showToast('CT保存失败: ' + e.message, 'error')
  }
}

// clearDeviceCT 将设备 CT 清除为 NULL，恢复使用全局默认值
// Clear device CT to NULL so it falls back to the global default.
// 設備の CT を NULL にクリアし、グローバルデフォルトに戻す。
const clearDeviceCT = async (dev) => {
  try {
    await window.go.main.App.SetDeviceCycleTime(dev.id, 0)
    dev.cycle_time = null
    showToast(`${dev.device_name} CT已恢复为全局默认（${config.value.productionCoefficient || 100}s）`, 'success')
  } catch (e) {
    showToast('CT清除失败: ' + e.message, 'error')
  }
}

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

// fmtMins 将分钟数格式化为"X小时Y分钟"
const fmtMins = (mins) => {
  if (!mins || mins <= 0) return '—'
  const h = Math.floor(mins / 60)
  const m = mins % 60
  if (h === 0) return `${m} 分钟`
  return m === 0 ? `${h} 小时` : `${h} 小时 ${m} 分钟`
}

// dailyWorkSummary 从所有时间安排的所有活动班次自动汇总净工时
const dailyWorkSummary = computed(() => {
  const allShifts = schedules.value.flatMap(g => g.shifts).filter(s => s.is_active)
  const rows = allShifts.map(s => {
    const shiftMins = spanMins(s.start_hour, s.start_min, s.end_hour, s.end_min)
    const breakMins = (s.breaks || []).reduce((sum, b) => sum + calculateDuration(b), 0)
    const netMins = Math.max(0, shiftMins - breakMins)
    return {
      name: s.name || '未命名',
      window: `${padTime(s.start_hour, s.start_min)}–${padTime(s.end_hour, s.end_min)}`,
      shiftMins, breakMins, netMins
    }
  })
  const totalNet = rows.reduce((sum, r) => sum + r.netMins, 0)
  return { rows, totalNet }
})

// 计算单个休息时间段的时长
// CN: 使用 spanMins 计算休息时长，支持跨零点（如 23:00→01:00 = 120 分钟）
// EN: Use spanMins to compute break duration; cross-midnight (23:00→01:00 = 120 min) is handled.
// JP: spanMins を使い休憩時間を計算。深夜跨ぎ（23:00→01:00 = 120 分）に対応。
const calculateDuration = (breakTime) => {
  return spanMins(breakTime.start_hour, breakTime.start_min, breakTime.end_hour, breakTime.end_min)
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

// 保存所有配置
// 每日工时由班次配置自动计算并写回，不再单独保存
const saveAllConfig = async () => {
  try {
    await saveProductionCoefficient()
    await saveShifts()
    showToast('所有配置保存成功！', 'success')
  } catch (e) {
    console.error('❌ 保存配置失败:', e)
    showToast('保存配置失败: ' + e.message, 'error')
  }
}

// ─── 班次时间配置方法 ────────────────────────────────────

// padTime 时间补零格式化
const padTime = (h, m) =>
  `${String(h ?? 0).padStart(2, '0')}:${String(m ?? 0).padStart(2, '0')}`

// spanMins 计算两个时刻之间的分钟数，支持跨零点（如 22:00→06:00 = 480 分钟）
const spanMins = (startH, startM, endH, endM) => {
  const s = startH * 60 + startM
  let e = endH * 60 + endM
  if (e <= s) e += 24 * 60   // 跨零点：结束时刻在次日
  return e - s
}

// shiftDurationLabel 计算班次有效时长文字（不减休息）
const shiftDurationLabel = (shift) => {
  const total = spanMins(shift.start_hour, shift.start_min, shift.end_hour, shift.end_min)
  if (total <= 0) return '—'
  const h = Math.floor(total / 60)
  const m = total % 60
  if (h === 0) return `${m} 分钟`
  return m === 0 ? `${h} 小时` : `${h} 小时 ${m} 分钟`
}

// applyTime 将 "HH:MM" 字符串解析后写回目标对象的 start_hour/start_min 或 end_hour/end_min
const applyTime = (val, target, field) => {
  if (!val) return
  const [hStr, mStr] = val.split(':')
  const h = parseInt(hStr, 10)
  const m = parseInt(mStr, 10)
  if (isNaN(h) || isNaN(m)) return
  target[`${field}_hour`] = Math.min(23, Math.max(0, h))
  target[`${field}_min`]  = Math.min(59, Math.max(0, m))
}

// toggleSchedule 展开/收起时间安排组
const toggleSchedule = (gi) => {
  if (expandedSchedules.value.has(gi)) expandedSchedules.value.delete(gi)
  else expandedSchedules.value.add(gi)
  expandedSchedules.value = new Set(expandedSchedules.value)
}

// toggleShift 展开/收起班次（key = gi*1000+si）
const toggleShift = (gi, si) => {
  const key = gi * 1000 + si
  if (expandedShifts.value.has(key)) expandedShifts.value.delete(key)
  else expandedShifts.value.add(key)
  expandedShifts.value = new Set(expandedShifts.value)
}

// addSchedule 新增一个时间安排组
const addSchedule = () => {
  const gi = schedules.value.length
  schedules.value.push({ id: 0, name: `时间安排${gi + 1}`, sort_order: gi, is_active: true, shifts: [] })
  expandedSchedules.value = new Set([...expandedSchedules.value, gi])
}

// deleteSchedule 删除时间安排组
const deleteSchedule = (gi) => {
  const sched = schedules.value[gi]
  confirmDialog.value = {
    show: true, type: 'warning', title: '确认删除时间安排',
    message: `确定要删除时间安排"${sched.name}"吗？该安排下的所有班次和休息段也将删除。`,
    details: [`包含班次：${sched.shifts.length} 个`],
    confirmText: '确认删除', cancelText: '取消',
    onConfirm: () => {
      schedules.value.splice(gi, 1)
      expandedSchedules.value = new Set()
      expandedShifts.value = new Set()
      confirmDialog.value.show = false
      showToast('已删除，记得保存', 'success')
    },
    onCancel: () => { confirmDialog.value.show = false }
  }
}

// addShiftToSchedule 向指定时间安排组添加一个班次
const addShiftToSchedule = (gi) => {
  const sched = schedules.value[gi]
  const si = sched.shifts.length
  sched.shifts.push({
    id: 0, schedule_id: sched.id, name: `班次${si + 1}`,
    start_hour: 7, start_min: 40, end_hour: 16, end_min: 20,
    is_active: true, sort_order: si, breaks: []
  })
  expandedShifts.value = new Set([...expandedShifts.value, gi * 1000 + si])
}

// deleteShiftFromSchedule 删除某安排组内的班次
const deleteShiftFromSchedule = (gi, si) => {
  const shift = schedules.value[gi].shifts[si]
  confirmDialog.value = {
    show: true, type: 'warning', title: '确认删除班次',
    message: `确定要删除班次"${shift.name}"吗？`,
    details: [`时间：${padTime(shift.start_hour, shift.start_min)} - ${padTime(shift.end_hour, shift.end_min)}`],
    confirmText: '确认删除', cancelText: '取消',
    onConfirm: () => {
      schedules.value[gi].shifts.splice(si, 1)
      expandedShifts.value = new Set()
      confirmDialog.value.show = false
      showToast('已删除，记得保存', 'success')
    },
    onCancel: () => { confirmDialog.value.show = false }
  }
}

// addBreakToShift 向班次添加休息段
const addBreakToShift = (gi, si) => {
  const shift = schedules.value[gi].shifts[si]
  shift.breaks.push({
    id: 0, shift_id: shift.id, name: `休息${shift.breaks.length + 1}`,
    start_hour: 12, start_min: 0, end_hour: 13, end_min: 0
  })
}

// deleteBreakFromShift 删除休息段
const deleteBreakFromShift = (gi, si, bi) => {
  schedules.value[gi].shifts[si].breaks.splice(bi, 1)
}

// loadSchedules 从后端加载时间安排组配置
const loadSchedules = async () => {
  try {
    if (!window.go?.main?.App?.GetShiftSchedules) return
    const result = await window.go.main.App.GetShiftSchedules()
    schedules.value = (result || []).map(g => ({
      id: g.id || 0,
      name: g.name || '',
      sort_order: g.sort_order ?? 0,
      is_active: g.is_active !== false,
      shifts: (g.shifts || []).map(s => ({
        id: s.id || 0,
        schedule_id: s.schedule_id || g.id || 0,
        name: s.name || '',
        start_hour: s.start_hour ?? 7,
        start_min: s.start_min ?? 40,
        end_hour: s.end_hour ?? 16,
        end_min: s.end_min ?? 20,
        is_active: s.is_active !== false,
        sort_order: s.sort_order ?? 0,
        breaks: (s.breaks || []).map(b => ({
          id: b.id || 0, shift_id: b.shift_id || s.id || 0, name: b.name || '',
          start_hour: b.start_hour ?? 0, start_min: b.start_min ?? 0,
          end_hour: b.end_hour ?? 0, end_min: b.end_min ?? 0
        }))
      }))
    }))
    console.log('✅ 加载时间安排配置成功:', schedules.value.length, '个组')
  } catch (e) {
    console.error('❌ 加载时间安排配置失败:', e)
    showToast('加载配置失败: ' + e.message, 'error')
  }
}

// saveSchedules 保存全部时间安排组配置
const saveSchedules = async () => {
  for (let gi = 0; gi < schedules.value.length; gi++) {
    const g = schedules.value[gi]
    if (!g.name?.trim()) { showToast(`第${gi+1}个时间安排名称不能为空`, 'warning'); return }
    for (let si = 0; si < g.shifts.length; si++) {
      const s = g.shifts[si]
      if (!s.name?.trim()) { showToast(`时间安排"${g.name}" - 第${si+1}个班次名称不能为空`, 'warning'); return }
      if (spanMins(s.start_hour, s.start_min, s.end_hour, s.end_min) <= 0) {
        showToast(`时间安排"${g.name}" - 班次"${s.name}"：起止时间不能相同`, 'warning'); return
      }
      for (let bi = 0; bi < s.breaks.length; bi++) {
        const b = s.breaks[bi]
        if (!b.name?.trim()) { showToast(`班次"${s.name}" - 第${bi+1}个休息段名称不能为空`, 'warning'); return }
        if (spanMins(b.start_hour, b.start_min, b.end_hour, b.end_min) <= 0) {
          showToast(`班次"${s.name}" - 休息"${b.name}"：起止时间不能相同`, 'warning'); return
        }
      }
    }
  }
  try {
    await window.go.main.App.SaveShiftSchedules(schedules.value)
    const totalNet = dailyWorkSummary.value.totalNet
    if (totalNet > 0 && window.go?.main?.App?.SetDailyWorkMinutes) {
      await window.go.main.App.SetDailyWorkMinutes(totalNet)
      config.value.dailyWorkMinutes = totalNet
    }
    showToast('时间安排配置保存成功！', 'success')
    await loadSchedules()
    await loadDevices()
  } catch (e) {
    console.error('❌ 保存时间安排配置失败:', e)
    showToast('保存失败: ' + e.message, 'error')
  }
}

// saveShifts 兼容别处调用（实际委托给 saveSchedules）
const saveShifts = saveSchedules

onMounted(() => {
  loadConfig()
  loadSchedules()
  loadDevices()
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

/* 单字段 HH:MM 时间输入框 */
.time-combined {
  width: 110px !important;
  text-align: center;
  font-size: 15px;
  font-weight: 500;
  /* 隐藏浏览器原生时间选择器的图标（部分浏览器） */
  color-scheme: dark;
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

/* ─── 逻辑日工时汇总表 ──────────────────────────────────── */
.daily-work-empty {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  margin: 4px 0 8px;
  background: rgba(84, 110, 122, 0.08);
  border: 1px dashed rgba(84, 110, 122, 0.3);
  border-radius: 8px;
  color: rgba(255,255,255,0.4);
  font-size: 13px;
}

.daily-work-table {
  margin: 4px 0 8px;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  overflow: hidden;
  font-size: 13px;
}

.dwt-header {
  display: flex;
  padding: 10px 16px;
  background: rgba(84, 110, 122, 0.12);
  border-bottom: 1px solid rgba(255,255,255,0.08);
  font-weight: 600;
  color: rgba(255,255,255,0.5);
  font-size: 12px;
  letter-spacing: 0.5px;
}

.dwt-row {
  display: flex;
  padding: 11px 16px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.75);
  transition: background 0.15s;
}

.dwt-row:hover {
  background: rgba(84, 110, 122, 0.08);
}

.dwt-total {
  display: flex;
  padding: 12px 16px;
  background: rgba(94, 139, 126, 0.1);
  border-top: 1px solid rgba(94, 139, 126, 0.2);
  font-weight: 600;
  color: rgba(255,255,255,0.85);
}

.dwt-col {
  display: flex;
  align-items: center;
}

.dwt-col.name   { flex: 1.2; min-width: 0; }
.dwt-col.window { flex: 1.8; min-width: 0; color: rgba(255,255,255,0.55); }
.dwt-col.dur    { flex: 1; }
.dwt-col.brk    { flex: 1; color: #c49090; }
.dwt-col.net    { flex: 1.4; }
.dwt-col.net.highlight { color: #7ea896; }
.dwt-col.net.bold { font-size: 14px; }

.net-mins {
  font-size: 12px;
  color: rgba(255,255,255,0.4);
  margin-left: 6px;
  font-weight: 400;
}

/* ─── 班次时间配置 ─────────────────────────────────────── */
.schedules-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.schedule-block {
  border: 1px solid rgba(0, 200, 150, 0.2);
  border-radius: 12px;
  overflow: hidden;
  background: rgba(15, 25, 40, 0.4);
}

.schedule-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  cursor: pointer;
  background: rgba(0, 200, 150, 0.06);
  transition: background 0.15s;
}
.schedule-header:hover { background: rgba(0, 200, 150, 0.1); }

.schedule-body {
  padding: 16px;
  border-top: 1px solid rgba(255,255,255,0.06);
}

.sched-badge { background: rgba(0, 200, 150, 0.2); color: rgba(0, 200, 150, 0.9); }

.sched-fields { margin-bottom: 12px; }

.sched-shifts-section {
  margin-top: 16px;
  border-top: 1px dashed rgba(255,255,255,0.08);
  padding-top: 12px;
}

.inner-shifts { gap: 8px; }

.inner-shift-block {
  margin-left: 4px;
  background: rgba(30, 40, 60, 0.35) !important;
}

.shifts-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.shift-block {
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 10px;
  overflow: hidden;
  background: rgba(20, 30, 50, 0.3);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.shift-block.drag-over {
  border-color: rgba(0, 200, 150, 0.6);
  box-shadow: 0 0 0 2px rgba(0, 200, 150, 0.2);
}

/* 拖拽手柄 */
.shift-drag-handle {
  color: rgba(140, 160, 185, 0.5);
  cursor: grab;
  padding: 4px 6px;
  font-size: 14px;
  transition: color 0.2s;
}
.shift-drag-handle:hover {
  color: rgba(0, 200, 150, 0.9);
}
.shift-drag-handle:active {
  cursor: grabbing;
}

/* 设备数量标签 */
.device-count-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(0, 150, 200, 0.18);
  color: #60bfff;
  border: 1px solid rgba(0, 150, 200, 0.3);
  border-radius: 10px;
  padding: 2px 8px;
  font-size: 11px;
}

.shift-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  cursor: pointer;
  transition: background 0.2s;
  background: rgba(20, 30, 50, 0.5);
  user-select: none;
}

.shift-header:hover {
  background: rgba(84, 110, 122, 0.15);
}

.shift-badge {
  width: 32px;
  height: 32px;
  background: rgba(84, 110, 122, 0.4);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  color: #90a4ae;
  flex-shrink: 0;
}

.shift-meta {
  display: flex;
  align-items: center;
  gap: 14px;
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}

.shift-name-display {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  white-space: nowrap;
}

.shift-window-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: rgba(255,255,255,0.6);
  background: rgba(84, 110, 122, 0.15);
  padding: 4px 10px;
  border-radius: 20px;
  border: 1px solid rgba(84, 110, 122, 0.2);
  white-space: nowrap;
}

.shift-dur-tag {
  font-size: 12px;
  color: #7ea896;
  margin-left: 4px;
}

.shift-status-tag {
  font-size: 12px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 20px;
}

.shift-status-tag.active {
  background: rgba(94, 139, 126, 0.2);
  color: #7ea896;
  border: 1px solid rgba(94, 139, 126, 0.3);
}

.shift-status-tag.inactive {
  background: rgba(142, 110, 110, 0.15);
  color: #c49090;
  border: 1px solid rgba(142, 110, 110, 0.2);
}

.shift-header-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.break-count-tag {
  font-size: 13px;
  color: rgba(255,255,255,0.5);
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 5px;
}

.shift-expand-icon {
  color: rgba(255,255,255,0.4);
  font-size: 13px;
  transition: transform 0.25s;
  flex-shrink: 0;
}

.shift-expand-icon.expanded {
  transform: rotate(180deg);
}

/* 展开内容 */
.shift-body {
  padding: 20px;
  border-top: 1px solid rgba(255,255,255,0.08);
  background: rgba(20, 30, 50, 0.2);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.shift-fields-row {
  display: flex;
  align-items: flex-end;
  gap: 20px;
  flex-wrap: wrap;
}

.shift-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.shift-field label {
  font-size: 12px;
  color: rgba(255,255,255,0.6);
  font-weight: 500;
}

.shift-name-input {
  width: 160px !important;
}

.shift-field-toggle {
  justify-content: flex-end;
}

.toggle-btn {
  padding: 9px 18px;
  border-radius: 8px;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 7px;
  transition: all 0.2s;
}

.toggle-btn.on {
  background: rgba(94, 139, 126, 0.3);
  color: #7ea896;
  border: 1px solid rgba(94, 139, 126, 0.4);
}

.toggle-btn.off {
  background: rgba(100, 100, 120, 0.2);
  color: rgba(255,255,255,0.4);
  border: 1px solid rgba(255,255,255,0.1);
}

.shift-duration-display {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 9px 14px;
  background: rgba(84, 110, 122, 0.15);
  border: 1px solid rgba(84, 110, 122, 0.2);
  border-radius: 8px;
  color: #90a4ae;
  font-size: 13px;
  font-weight: 500;
  align-self: flex-end;
  white-space: nowrap;
}

/* 班内休息时间段 */
.shift-breaks-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.shift-breaks-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: rgba(255,255,255,0.7);
  font-weight: 500;
}

.shift-breaks-header span {
  display: flex;
  align-items: center;
  gap: 7px;
}

.action-btn.small {
  padding: 6px 14px;
  font-size: 13px;
}

.breaks-empty {
  font-size: 13px;
  color: rgba(255,255,255,0.35);
  padding: 14px 20px;
  background: rgba(255,255,255,0.03);
  border-radius: 6px;
  border: 1px dashed rgba(255,255,255,0.1);
}

/* 底部保存栏 */
.shifts-footer {
  margin-top: 20px;
  padding: 18px 20px;
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}

/* ─── 设备关联区 ───────────────────────────────────────── */
.shift-devices-section {
  padding: 14px 20px;
  border-top: 1px solid rgba(255,255,255,0.06);
}

.shift-devices-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  font-size: 13px;
  color: #a0b8cc;
}

.section-hint {
  font-size: 11px;
  color: rgba(140,160,185,0.6);
}

.device-assign-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.device-assign-item {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid rgba(255,255,255,0.12);
  background: rgba(30, 40, 60, 0.5);
  cursor: pointer;
  font-size: 12px;
  color: #8baabb;
  transition: all 0.18s;
  user-select: none;
}

.device-assign-item:hover {
  border-color: rgba(0,200,150,0.45);
  color: #c0dde8;
}

.device-assign-item.assigned {
  border-color: rgba(0,200,150,0.6);
  background: rgba(0,200,150,0.1);
  color: #00e0a0;
}

.device-assign-left {
  display: flex;
  align-items: center;
  gap: 7px;
  cursor: pointer;
  flex: 1;
  min-width: 0;
}

.device-assign-name {
  font-weight: 500;
}

.device-assign-code {
  font-size: 11px;
  color: rgba(140,160,185,0.7);
}

.device-ct-inline {
  display: flex;
  align-items: center;
  gap: 3px;
  margin-left: auto;
  flex-shrink: 0;
}

.ct-label {
  font-size: 10px;
  color: rgba(140,160,185,0.6);
  font-weight: 600;
}

.ct-input-small {
  width: 52px;
  padding: 2px 5px;
  background: rgba(20, 30, 50, 0.6);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 4px;
  color: #c0dde8;
  font-size: 11px;
  text-align: right;
  outline: none;
}

.ct-input-small:focus {
  border-color: rgba(0,200,150,0.5);
}

.ct-input-small::placeholder {
  color: rgba(140,160,185,0.4);
  font-style: italic;
}

.ct-unit {
  font-size: 10px;
  color: rgba(140,160,185,0.5);
}

/* ─── 设备独立CT配置表格 ──────────────────────────────── */
.device-ct-table {
  width: 100%;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid rgba(255,255,255,0.08);
}

.device-ct-header {
  display: flex;
  align-items: center;
  padding: 8px 14px;
  background: rgba(255,255,255,0.05);
  font-size: 11px;
  font-weight: 600;
  color: rgba(140,160,185,0.7);
  letter-spacing: 0.5px;
  text-transform: uppercase;
  border-bottom: 1px solid rgba(255,255,255,0.07);
}

.device-ct-row {
  display: flex;
  align-items: center;
  padding: 10px 14px;
  border-bottom: 1px solid rgba(255,255,255,0.04);
  transition: background 0.15s;
}
.device-ct-row:last-child { border-bottom: none; }
.device-ct-row:hover { background: rgba(255,255,255,0.03); }
.device-ct-row.ct-overridden { background: rgba(0,200,150,0.04); }

.dct-col { display: flex; align-items: center; }
.dct-name  { flex: 2; font-size: 13px; font-weight: 500; color: #c0dde8; }
.dct-code  { flex: 1.2; font-size: 12px; color: rgba(140,160,185,0.65); }
.dct-sched { flex: 1.5; }
.dct-ct    { flex: 2; }
.dct-eff   { flex: 1.8; }

.sched-tag {
  display: inline-block;
  padding: 2px 8px;
  background: rgba(100,160,220,0.15);
  border: 1px solid rgba(100,160,220,0.25);
  border-radius: 12px;
  font-size: 11px;
  color: rgba(150,200,240,0.9);
}
.no-sched-tag {
  font-size: 11px;
  color: rgba(140,160,185,0.4);
  font-style: italic;
}

.ct-edit-row {
  display: flex;
  align-items: center;
  gap: 4px;
}

.ct-input-full {
  width: 80px;
  padding: 4px 8px;
  background: rgba(20, 30, 50, 0.7);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 5px;
  color: #c0dde8;
  font-size: 13px;
  text-align: right;
  outline: none;
  transition: border-color 0.2s;
}
.ct-input-full:focus { border-color: rgba(0,200,150,0.5); }
.ct-input-full::placeholder { color: rgba(140,160,185,0.35); font-style: italic; }

.btn-clear-ct {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px; height: 20px;
  border: none;
  background: rgba(220,80,80,0.15);
  border-radius: 50%;
  color: rgba(220,120,120,0.9);
  cursor: pointer;
  font-size: 10px;
  transition: background 0.15s;
}
.btn-clear-ct:hover { background: rgba(220,80,80,0.3); }

.eff-ct-val { font-size: 12px; font-weight: 600; }
.eff-custom { color: rgba(0,220,150,0.9); }
.eff-global { color: rgba(140,160,185,0.5); }

/* ─── 响应式 ────────────────────────────────────────── */
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

