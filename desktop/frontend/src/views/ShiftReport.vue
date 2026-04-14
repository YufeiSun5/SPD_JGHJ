<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-clipboard-check"></i>
          班次生产追溯
        </div>
        <div class="page-subtitle">Shift Production History · シフト生産履歴</div>
      </div>
      <div class="header-actions">
        <button class="action-btn refresh" @click="loadSnapshots">
          <i class="fas fa-sync-alt"></i> 刷新
        </button>
        <button class="action-btn success" @click="exportCSV" :disabled="!snapshots.length">
          <i class="fas fa-file-csv"></i> 导出 CSV
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div v-if="snapshots.length" class="stats-grid">
      <div class="stat-card card-records">
        <div class="stat-icon"><i class="fas fa-list-alt"></i></div>
        <div class="stat-info">
          <div class="stat-value">{{ snapshots.length }}</div>
          <div class="stat-label">班次记录数</div>
        </div>
      </div>
      <div class="stat-card card-production">
        <div class="stat-icon"><i class="fas fa-boxes"></i></div>
        <div class="stat-info">
          <div class="stat-value">{{ summary.totalQty }}</div>
          <div class="stat-label">总产量（件）</div>
        </div>
      </div>
      <div class="stat-card card-quality">
        <div class="stat-icon"><i class="fas fa-check-circle"></i></div>
        <div class="stat-info">
          <div class="stat-value">{{ summary.qualityPct }}%</div>
          <div class="stat-label">整体良品率</div>
        </div>
      </div>
      <div class="stat-card card-oee">
        <div class="stat-icon"><i class="fas fa-tachometer-alt"></i></div>
        <div class="stat-info">
          <div class="stat-value">{{ summary.avgOee }}%</div>
          <div class="stat-label">平均 OEE</div>
        </div>
      </div>
    </div>

    <!-- 筛选 + 数据合并卡片 -->
    <div class="card">
      <div class="card-header">
        <h3>
          <i class="fas fa-table"></i> 班次快照列表
          <span class="badge-count">{{ snapshots.length }} 条</span>
        </h3>
      </div>
      <!-- 筛选条件 -->
      <div class="filter-row">
        <!-- 第一行：时间筛选 -->
        <div class="filter-line">
          <div class="filter-group">
            <label><i class="fas fa-clock"></i> 快捷选择：</label>
            <div class="quick-btns">
              <button v-for="q in quickDates" :key="q.key"
                :class="['quick-btn', activeQuick === q.key ? 'active' : '']"
                @click="applyQuick(q)">{{ q.label }}</button>
            </div>
          </div>
          <div class="filter-group">
            <label><i class="fas fa-calendar-alt"></i> 开始时间：</label>
            <input type="datetime-local" v-model="startDate" class="filter-input dt-input" @change="activeQuick = ''; loadSnapshots()" />
          </div>
          <div class="filter-group">
            <label><i class="fas fa-calendar-alt"></i> 结束时间：</label>
            <input type="datetime-local" v-model="endDate" class="filter-input dt-input" @change="activeQuick = ''; loadSnapshots()" />
          </div>
        </div>
        <!-- 第二行：条件筛选 -->
        <div class="filter-line">
          <div class="filter-group">
            <label><i class="fas fa-desktop"></i> 设备：</label>
            <div class="custom-select">
              <select v-model="filterDeviceID" @change="loadSnapshots">
                <option :value="null">全部设备</option>
                <option v-for="d in devices" :key="d.id" :value="d.id">{{ d.device_name }}</option>
              </select>
            </div>
          </div>
          <div class="filter-group">
            <label><i class="fas fa-users"></i> 班组：</label>
            <div class="custom-select">
              <select v-model="filterTeamID" @change="loadSnapshots">
                <option :value="null">全部班组</option>
                <option v-for="t in teams" :key="t.id" :value="t.id">{{ t.team_name }}</option>
              </select>
            </div>
          </div>
          <div class="filter-group">
            <label><i class="fas fa-user"></i> 人员：</label>
            <div class="custom-select">
              <select v-model="filterStaffName" @change="loadSnapshots">
                <option value="">全部人员</option>
                <option v-for="s in staffList" :key="s.id" :value="s.name">
                  {{ s.name }}{{ s.staff_code ? '（' + s.staff_code + '）' : '' }}
                </option>
              </select>
            </div>
          </div>
          <button class="action-btn primary" @click="loadSnapshots">
            <i class="fas fa-search"></i> 查询
          </button>
        </div>
      </div>
      <div class="table-container" v-if="snapshots.length">
        <table class="data-table">
          <thead>
            <tr>
              <th class="th-group-a" colspan="4">基本信息</th>
              <th class="th-group-b" colspan="5">节拍 / 产量</th>
              <th class="th-group-c" colspan="3">时间分析</th>
              <th class="th-group-d" colspan="4">OEE 指标</th>
              <th class="th-group-e" colspan="2">班组 / 人员</th>
              <th class="th-group-op">操作</th>
            </tr>
            <tr>
              <th class="col-date">日期</th>
              <th class="col-shift">班次</th>
              <th class="col-time">时间段</th>
              <th class="col-dev">设备</th>
              <th class="col-num">CT<br><small>秒/件</small></th>
              <th class="col-num">理论产量<br><small>件</small></th>
              <th class="col-num">总产量<br><small>件</small></th>
              <th class="col-num ok-h">良品</th>
              <th class="col-num ng-h">不良品</th>
              <th class="col-num">理论工时</th>
              <th class="col-num">运行时间</th>
              <th class="col-num">空闲时间</th>
              <th class="col-pct">时间<br>稼动率</th>
              <th class="col-pct">性能<br>稼动率</th>
              <th class="col-pct">良品率</th>
              <th class="col-pct">OEE</th>
              <th class="col-team">班组</th>
              <th class="col-staff">人员</th>
              <th class="col-op">详情</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in snapshots" :key="s.id" class="data-row">
              <td class="col-date center">{{ fmtDate(s.snapshot_date) }}</td>
              <td class="col-shift center"><span :class="['shift-badge', shiftClass(s.shift_name)]">{{ s.shift_name }}</span></td>
              <td class="col-time center mono">{{ fmtTime(s.shift_start) }}~{{ fmtTime(s.shift_end) }}</td>
              <td class="col-dev">{{ s.device_name }}</td>
              <td class="col-num center">{{ s.cycle_time > 0 ? s.cycle_time : '-' }}</td>
              <td class="col-num center theory">{{ theoryQty(s) }}</td>
              <td class="col-num center">{{ s.total_qty }}</td>
              <td class="col-num center ok">{{ s.ok_qty }}</td>
              <td class="col-num center ng">{{ s.ng_qty }}</td>
              <td class="col-num center">{{ fmtSec(s.plan_work_sec) }}</td>
              <td class="col-num center">{{ fmtSec(s.device_run_sec) }}</td>
              <td class="col-num center dim">{{ fmtSec(s.device_idle_sec) }}</td>
              <td class="col-pct center">{{ s.availability_pct.toFixed(1) }}%</td>
              <td class="col-pct center">{{ s.performance_pct.toFixed(1) }}%</td>
              <td class="col-pct center">{{ s.quality_pct.toFixed(1) }}%</td>
              <td class="col-pct center oee-cell" :class="oeeClass(s.oee_pct)">{{ s.oee_pct.toFixed(1) }}%</td>
              <td class="col-team center">{{ s.team_name || '-' }}</td>
              <td class="col-staff" :title="fmtStaffFull(s.staff_snapshot)">{{ fmtStaff(s.staff_snapshot) }}</td>
              <td class="col-op center"><button class="detail-btn" @click="openDetail(s)"><i class="fas fa-eye"></i> 详情</button></td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="summary-row">
              <td colspan="4" class="sum-label-cell">汇总 / 平均（{{ snapshots.length }} 条）</td>
              <td class="center">—</td>
              <td class="center">{{ summary.totalTheory }}</td>
              <td class="center">{{ summary.totalQty }}</td>
              <td class="center ok">{{ summary.okQty }}</td>
              <td class="center ng">{{ summary.ngQty }}</td>
              <td class="center">{{ fmtSec(summary.totalPlanSec) }}</td>
              <td class="center">{{ fmtSec(summary.totalRunSec) }}</td>
              <td class="center">{{ fmtSec(summary.totalIdleSec) }}</td>
              <td class="center">{{ summary.avgAvail }}%</td>
              <td class="center">{{ summary.avgPerf }}%</td>
              <td class="center">{{ summary.qualityPct }}%</td>
              <td class="center oee-cell" :class="oeeClass(summary.avgOee)">{{ summary.avgOee }}%</td>
              <td colspan="2" class="center dim">—</td>
              <td class="col-op"></td>
            </tr>
          </tfoot>
        </table>
      </div>
      <div v-else-if="loaded" class="empty-state-card">
        <i class="fas fa-inbox"></i>
        <p>所选条件下无快照数据</p>
        <span>班次结束后系统自动生成快照；重启后新完成的班次会追加新记录，已有记录不会重复</span>
      </div>
    </div>

    <!-- 班次快照详情弹窗 -->
    <div v-if="showDetailModal && detailRecord" class="modal-overlay" @click.self="closeDetail">
      <div class="modal-container detail-modal">
        <div class="modal-header">
          <h3>
            <i class="fas fa-clipboard-list"></i>
            班次快照详情
            <span class="detail-title-badge">{{ fmtDate(detailRecord.snapshot_date) }} · {{ detailRecord.shift_name }}</span>
          </h3>
          <button class="modal-close" @click="closeDetail"><i class="fas fa-times"></i></button>
        </div>
        <div class="modal-body">
          <!-- 基本信息 -->
          <div class="detail-section">
            <div class="detail-section-title"><i class="fas fa-info-circle"></i> 基本信息</div>
            <div class="detail-grid">
              <div class="detail-item"><span class="di-label">日期</span><span class="di-val">{{ fmtDate(detailRecord.snapshot_date) }}</span></div>
              <div class="detail-item"><span class="di-label">班次</span><span class="di-val"><span :class="['shift-badge', shiftClass(detailRecord.shift_name)]">{{ detailRecord.shift_name }}</span></span></div>
              <div class="detail-item"><span class="di-label">设备</span><span class="di-val">{{ detailRecord.device_name }}</span></div>
              <div class="detail-item"><span class="di-label">时间段</span><span class="di-val mono">{{ fmtTime(detailRecord.shift_start) }} ~ {{ fmtTime(detailRecord.shift_end) }}</span></div>
              <div class="detail-item"><span class="di-label">班组</span><span class="di-val">{{ detailRecord.team_name || '-' }}</span></div>
              <div class="detail-item"><span class="di-label">节拍 (CT)</span><span class="di-val">{{ detailRecord.cycle_time > 0 ? detailRecord.cycle_time + ' 秒/件' : '-' }}</span></div>
            </div>
          </div>
          <!-- 产量统计 -->
          <div class="detail-section">
            <div class="detail-section-title"><i class="fas fa-boxes"></i> 产量统计</div>
            <div class="detail-grid">
              <div class="detail-item"><span class="di-label">理论产量</span><span class="di-val theory">{{ theoryQty(detailRecord) }} 件</span></div>
              <div class="detail-item"><span class="di-label">总产量</span><span class="di-val">{{ detailRecord.total_qty }} 件</span></div>
              <div class="detail-item"><span class="di-label">良品数</span><span class="di-val ok">{{ detailRecord.ok_qty }} 件</span></div>
              <div class="detail-item"><span class="di-label">不良品数</span><span class="di-val ng">{{ detailRecord.ng_qty }} 件</span></div>
            </div>
          </div>
          <!-- 时间分析 -->
          <div class="detail-section">
            <div class="detail-section-title"><i class="fas fa-clock"></i> 时间分析</div>
            <div class="detail-grid">
              <div class="detail-item"><span class="di-label">理论工时</span><span class="di-val">{{ fmtSec(detailRecord.plan_work_sec) }}</span></div>
              <div class="detail-item"><span class="di-label">运行时间</span><span class="di-val">{{ fmtSec(detailRecord.device_run_sec) }}</span></div>
              <div class="detail-item"><span class="di-label">空闲时间</span><span class="di-val dim">{{ fmtSec(detailRecord.device_idle_sec) }}</span></div>
            </div>
          </div>
          <!-- OEE 指标 -->
          <div class="detail-section">
            <div class="detail-section-title"><i class="fas fa-tachometer-alt"></i> OEE 指标</div>
            <div class="detail-grid">
              <div class="detail-item"><span class="di-label">时间稼动率</span><span class="di-val">{{ detailRecord.availability_pct.toFixed(2) }}%</span></div>
              <div class="detail-item"><span class="di-label">性能稼动率</span><span class="di-val">{{ detailRecord.performance_pct.toFixed(2) }}%</span></div>
              <div class="detail-item"><span class="di-label">良品率</span><span class="di-val">{{ detailRecord.quality_pct.toFixed(2) }}%</span></div>
              <div class="detail-item oee-highlight"><span class="di-label">OEE</span><span class="di-val oee-cell" :class="oeeClass(detailRecord.oee_pct)">{{ detailRecord.oee_pct.toFixed(2) }}%</span></div>
            </div>
          </div>
          <!-- 当班人员 -->
          <div class="detail-section" v-if="detailRecord.staff_snapshot">
            <div class="detail-section-title"><i class="fas fa-users"></i> 当班人员</div>
            <div class="staff-detail-list">
              <template v-if="parseStaffList(detailRecord.staff_snapshot).length">
                <div v-for="st in parseStaffList(detailRecord.staff_snapshot)" :key="st.id"
                  :class="['staff-detail-item', st._rankClass]">
                  <div class="staff-avatar"><i class="fas fa-user-circle"></i></div>
                  <div class="staff-info">
                    <div class="staff-name">{{ st.name }}<span class="staff-code-tag">{{ st.code ? '（' + st.code + '）' : '' }}</span></div>
                    <div class="staff-meta">
                      <span v-if="st.team_name"><i class="fas fa-users"></i> {{ st.team_name }}</span>
                      <span v-if="st.work_sec > 0" class="meta-sep">·</span>
                      <span v-if="st.work_sec > 0"><i class="fas fa-clock"></i> {{ fmtSec(st.work_sec) }}</span>
                    </div>
                    <!-- 在班时长占比进度条 -->
                    <div class="staff-pct-row" v-if="st.pct > 0">
                      <div class="pct-bar"><div class="pct-fill" :style="{ width: Math.min(st.pct, 100) + '%' }"></div></div>
                      <span class="pct-label">{{ st.pct }}%</span>
                    </div>
                  </div>
                  <div class="staff-hours">
                    <div class="sh-time" v-if="st.work_sec > 0">{{ fmtSec(st.work_sec) }}</div>
                    <div class="sh-pct" v-if="st.pct > 0">占班次 {{ st.pct }}%</div>
                    <div class="sh-time dim" v-else>—</div>
                  </div>
                </div>
              </template>
              <div v-else class="no-staff">无人员记录</div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <div class="footer-spacer"></div>
          <button class="modal-btn cancel" @click="closeDetail"><i class="fas fa-times"></i> 关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

// ── 筛选条件 ─────────────────────────────────────────────
// CN: 用本地时间（非 UTC）构造 datetime-local 格式字符串
// EN: Build datetime-local strings using local time (not UTC)
const localDT = (d = new Date(), h = d.getHours(), m = d.getMinutes()) => {
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(h)}:${pad(m)}`
}
const localDTDaysAgo = (n, h = 0, mi = 0) => {
  const d = new Date(); d.setDate(d.getDate() - n); return localDT(d, h, mi)
}

const startDate       = ref(localDTDaysAgo(6, 0, 0))
const endDate         = ref(localDT(new Date(), 23, 59))
const filterDeviceID  = ref(null)
const filterTeamID    = ref(null)
const filterStaffName = ref('')
const activeQuick     = ref('7d')

const quickDates = [
  { key: 'today', label: '今天',    start: () => localDTDaysAgo(0,0,0),  end: () => localDT(new Date(),23,59) },
  { key: 'yest',  label: '昨天',    start: () => localDTDaysAgo(1,0,0),  end: () => localDTDaysAgo(1,23,59) },
  { key: '7d',    label: '最近7天', start: () => localDTDaysAgo(6,0,0),  end: () => localDT(new Date(),23,59) },
  { key: '30d',   label: '最近30天',start: () => localDTDaysAgo(29,0,0), end: () => localDT(new Date(),23,59) },
]
const applyQuick = (q) => {
  startDate.value = q.start(); endDate.value = q.end()
  activeQuick.value = q.key; loadSnapshots()
}

// ── 数据 ─────────────────────────────────────────────────
const snapshots  = ref([])
const devices    = ref([])
const teams      = ref([])
const staffList  = ref([])
const loaded     = ref(false)

const loadSnapshots = async () => {
  try {
    const result = await window.go.main.App.GetShiftSnapshots(
      startDate.value, endDate.value,
      filterDeviceID.value, filterTeamID.value,
      filterStaffName.value.trim()
    )
    snapshots.value = result || []
    loaded.value = true
  } catch (e) { console.error('查询快照失败:', e) }
}

const loadFilters = async () => {
  try {
    const [devResult, teamResult, staffResult] = await Promise.all([
      window.go.main.App.GetAllDevices(),
      window.go.main.App.GetAllTeams(null),
      window.go.main.App.GetAllStaff(null, null),
    ])
    devices.value   = devResult  || []
    teams.value     = teamResult || []
    staffList.value = (staffResult || [])
      .filter(s => s.is_active !== 0)
      .sort((a, b) => a.name.localeCompare(b.name, 'zh'))
  } catch (e) { console.error('加载筛选项失败:', e) }
}

// ── 理论产量 ─────────────────────────────────────────────
// CN: 理论产量 = 理论工作秒数 / CT（四舍五入取整）
// EN: Theory qty = plan_work_sec / cycle_time (rounded)
// JP: 理論数量 = plan_work_sec ÷ CT（四捨五入）
const theoryQty = (s) => {
  if (!s.cycle_time || s.cycle_time <= 0 || !s.plan_work_sec) return '-'
  return Math.round(s.plan_work_sec / s.cycle_time)
}

// ── 汇总 ─────────────────────────────────────────────────
const summary = computed(() => {
  const rows = snapshots.value
  if (!rows.length) return {}
  const n   = rows.length
  const sum = (k) => rows.reduce((a, r) => a + (r[k] || 0), 0)
  const avg = (k) => (sum(k) / n).toFixed(1)

  const totalQty   = sum('total_qty')
  const okQty      = sum('ok_qty')
  const ngQty      = sum('ng_qty')
  const totalTheory = rows.reduce((a, s) => {
    if (s.cycle_time > 0 && s.plan_work_sec > 0)
      return a + Math.round(s.plan_work_sec / s.cycle_time)
    return a
  }, 0)
  const qualityPct = totalQty > 0 ? ((okQty / totalQty) * 100).toFixed(1) : '0.0'

  return {
    totalQty, okQty, ngQty, totalTheory, qualityPct,
    avgAvail:      avg('availability_pct'),
    avgPerf:       avg('performance_pct'),
    avgOee:        avg('oee_pct'),
    totalPlanSec:  sum('plan_work_sec'),
    totalRunSec:   sum('device_run_sec'),
    totalIdleSec:  sum('device_idle_sec'),
    totalFaultSec: sum('device_fault_sec'),
  }
})

// ── 导出 CSV ──────────────────────────────────────────────
// CN: 纯前端生成 CSV，包含完整 OEE 计算所需字段，编码 UTF-8 BOM 确保 Excel 识别中文
// EN: Client-side CSV generation; UTF-8 BOM ensures Excel opens Chinese characters correctly.
// JP: クライアント側で CSV 生成。UTF-8 BOM により Excel が中文を正しく認識する。
const exportCSV = () => {
  const headers = [
    '日期','班次','开始时间','结束时间','设备',
    'CT(秒/件)','理论产量','总产量','良品','不良品',
    '理论工时(s)','运行时间(s)','空闲时间(s)',
    '时间稼动率(%)','性能稼动率(%)','良品率(%)','OEE(%)',
    '班组','人员'
  ]
  const rows = snapshots.value.map(s => [
    fmtDate(s.snapshot_date),
    s.shift_name,
    fmtTime(s.shift_start),
    fmtTime(s.shift_end),
    s.device_name,
    s.cycle_time || '',
    theoryQty(s),
    s.total_qty,
    s.ok_qty,
    s.ng_qty,
    s.plan_work_sec,
    s.device_run_sec,
    s.device_idle_sec,
    s.availability_pct.toFixed(2),
    s.performance_pct.toFixed(2),
    s.quality_pct.toFixed(2),
    s.oee_pct.toFixed(2),
    s.team_name || '',
    fmtStaffFull(s.staff_snapshot)
  ])
  // 追加汇总行
  const sm = summary.value
  rows.push([
    '【汇总】','','','','',
    '',sm.totalTheory,sm.totalQty,sm.okQty,sm.ngQty,
    sm.totalPlanSec,sm.totalRunSec,sm.totalIdleSec,
    sm.avgAvail,sm.avgPerf,sm.qualityPct,sm.avgOee,
    '',`共 ${snapshots.value.length} 条`
  ])

  const csvContent = [headers, ...rows]
    .map(r => r.map(v => `"${String(v).replace(/"/g,'""')}"`).join(','))
    .join('\r\n')
  const bom = '\uFEFF'
  const blob = new Blob([bom + csvContent], { type: 'text/csv;charset=utf-8;' })
  const url  = URL.createObjectURL(blob)
  const a    = document.createElement('a')
  const dateStr = new Date().toISOString().slice(0,10)
  a.href     = url
  a.download = `班次追溯_${dateStr}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

// ── 工具函数 ─────────────────────────────────────────────
// CN: 解析时间字符串为本地时间（非 UTC）
// EN: Parse datetime string to local time (not UTC raw string)
const fmtTime = (isoStr) => {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  if (isNaN(d.getTime())) return ''
  const pad = n => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// 班次胶囊颜色：根据班次名称哈希到固定颜色组
const SHIFT_COLORS = ['shift-a', 'shift-b', 'shift-c', 'shift-d', 'shift-e']
const shiftColorMap = {}
const shiftClass = (name) => {
  if (!name) return ''
  if (!shiftColorMap[name]) {
    const idx = Object.keys(shiftColorMap).length % SHIFT_COLORS.length
    shiftColorMap[name] = SHIFT_COLORS[idx]
  }
  return shiftColorMap[name]
}

const fmtDate = (isoStr) => {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  if (isNaN(d.getTime())) return isoStr.slice(0, 10)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

const fmtSec = (sec) => {
  if (!sec || sec <= 0) return '-'
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  return h > 0 ? `${h}小时${String(m).padStart(2,'0')}分钟` : `${m}分钟`
}

const fmtStaff = (json) => {
  if (!json) return '-'
  try {
    const arr = JSON.parse(json)
    if (!arr.length) return '-'
    // 新格式（含 work_sec）：显示姓名 + 工时占比
    if (arr[0].work_sec !== undefined && arr[0].work_sec > 0) {
      return arr.map(s => {
        const h = Math.floor(s.work_sec / 3600)
        const m = Math.floor((s.work_sec % 3600) / 60)
        const timeStr = h > 0 ? `${h}小时${String(m).padStart(2,'0')}分钟` : `${m}分钟`
        return `${s.name}(${s.pct}%/${timeStr})`
      }).join('、')
    }
    // 旧格式降级：最多 3 名
    const names = arr.map(s => s.name)
    return names.length > 3 ? names.slice(0, 3).join('、') + `…共${names.length}人` : names.join('、')
  } catch { return '-' }
}

const fmtStaffFull = (json) => {
  if (!json) return ''
  try {
    const arr = JSON.parse(json)
    if (!arr.length) return ''
    // 新格式（含 work_sec）：展示姓名、工号、班组、时长、占比
    if (arr[0].work_sec !== undefined && arr[0].work_sec > 0) {
      return arr.map(s => {
        const h = Math.floor(s.work_sec / 3600)
        const m = Math.floor((s.work_sec % 3600) / 60)
        const timeStr = h > 0 ? `${h}小时${String(m).padStart(2,'0')}分钟` : `${m}分钟`
        const team = s.team_name ? `[${s.team_name}]` : ''
        return `${s.name}(${s.code || ''}) ${team} ${timeStr} (${s.pct}%)`
      }).join('  |  ')
    }
    // 旧格式降级
    return arr.map(s => `${s.name}(${s.code || ''})`).join('  ')
  } catch { return '' }
}

const oeeClass = (pct) => {
  const v = parseFloat(pct)
  if (v >= 85) return 'oee-good'
  if (v >= 60) return 'oee-mid'
  return 'oee-low'
}

// ── 详情弹窗 ─────────────────────────────────────────────
const detailRecord    = ref(null)
const showDetailModal = ref(false)
const openDetail  = (s) => { detailRecord.value = s; showDetailModal.value = true }
const closeDetail = () => { showDetailModal.value = false }

// CN: 解析 staff_snapshot JSON，用于弹窗中展示人员列表
// EN: Parse staff_snapshot JSON for detail modal display
const RANK_CLASSES = ['rank-gold', 'rank-silver', 'rank-bronze', 'rank-iron', 'rank-earth']
const parseStaffList = (json) => {
  if (!json) return []
  let list = []
  try { list = JSON.parse(json) || [] } catch { return [] }
  // 按 work_sec 降序，允许并列名次
  const sorted = [...list].sort((a, b) => (b.work_sec || 0) - (a.work_sec || 0))
  let rank = 1
  sorted.forEach((st, i) => {
    if (i > 0 && (st.work_sec || 0) < (sorted[i - 1].work_sec || 0)) rank = i + 1
    st._rank = rank
    st._rankClass = RANK_CLASSES[Math.min(rank - 1, RANK_CLASSES.length - 1)]
  })
  return sorted
}

onMounted(() => { loadFilters(); loadSnapshots() })
</script>

<style scoped>
/* ── 根容器 ──────────────────────────────────────────── */
.page-container {
  color: #ecf0f1;
  height: 100%;
  overflow-y: auto;
}

/* ── 页面头部 ────────────────────────────────────────── */
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

.page-title i { color: #667eea; }

.page-subtitle {
  font-size: 14px;
  color: rgba(255,255,255,0.5);
  margin-top: 4px;
}

.header-actions { display: flex; gap: 12px; }

/* ── 操作按钮 ────────────────────────────────────────── */
.action-btn {
  padding: 9px 20px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.action-btn.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  box-shadow: 0 4px 15px rgba(102,126,234,0.4);
}
.action-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102,126,234,0.6);
}

.action-btn.success {
  background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%);
  color: #fff;
  box-shadow: 0 4px 15px rgba(74,107,96,0.4);
}
.action-btn.success:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(74,107,96,0.6);
}
.action-btn.success:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.action-btn.refresh {
  background: rgba(84,110,122,0.2);
  border: 1px solid rgba(84,110,122,0.3);
  color: rgba(255,255,255,0.8);
}
.action-btn.refresh:hover {
  background: rgba(84,110,122,0.4);
  color: #fff;
  transform: translateY(-1px);
}

/* ── 统计卡片 ────────────────────────────────────────── */
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

.card-records    { background: linear-gradient(135deg, #556070 0%, #6e7e8e 100%); }
.card-production { background: linear-gradient(135deg, #5a7080 0%, #6e8e9e 100%); }
.card-quality    { background: linear-gradient(135deg, #4a6b60 0%, #5e8b7e 100%); }
.card-oee        { background: linear-gradient(135deg, #7a6550 0%, #9e7e5e 100%); }

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

.stat-info { flex: 1; }
.stat-value {
  font-size: 32px;
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

/* ── 卡片 ────────────────────────────────────────────── */
.card {
  background: rgba(20,30,48,0.6);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  overflow: hidden;
  margin-bottom: 24px;
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
  margin: 0;
}
.card-header h3 i { color: #546e7a; }


.badge-count {
  background: rgba(84,110,122,0.3);
  color: #fff;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
  margin-left: 8px;
}

/* ── 筛选区 ──────────────────────────────────────────── */
.filter-row {
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid rgba(255,255,255,0.06);
}

.filter-line {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
  padding: 12px 20px;
}

.filter-line + .filter-line {
  border-top: 1px solid rgba(255,255,255,0.05);
  background: rgba(0,0,0,0.06);
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
  display: flex;
  align-items: center;
  gap: 6px;
}
.filter-group label i { color: #546e7a; }

.custom-select {
  position: relative;
  display: inline-block;
  min-width: 130px;
}
.custom-select::after {
  content: '\f078';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255,255,255,0.5);
  pointer-events: none;
  font-size: 11px;
}
.custom-select select {
  appearance: none;
  -webkit-appearance: none;
  width: 100%;
  padding: 8px 28px 8px 14px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}
.custom-select select:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}
.custom-select select option { background: #1a1f3a; color: #fff; }

.filter-input {
  padding: 8px 14px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  outline: none;
  transition: all 0.2s;
}
.filter-input:focus { border-color: #667eea; }
.dt-input { min-width: 162px; color-scheme: dark; }

.quick-btns { display: flex; gap: 6px; }
.quick-btn {
  padding: 7px 14px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: rgba(255,255,255,0.7);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.quick-btn:hover  { background: rgba(102,126,234,0.15); border-color: rgba(102,126,234,0.4); color: #fff; }
.quick-btn.active { background: rgba(102,126,234,0.25); border-color: rgba(102,126,234,0.6); color: #fff; font-weight: 600; }

/* ── 表格 ────────────────────────────────────────────── */
.table-container { overflow-x: auto; }

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  white-space: nowrap;
}

/* 分组表头（第一行）*/
.data-table thead tr:first-child th {
  padding: 7px 10px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-align: center;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}
.th-group-a { background: rgba(60,100,180,0.25); color: #88aaee; border-right: 1px solid rgba(255,255,255,0.1); }
.th-group-b { background: rgba(0,160,120,0.20);  color: #60d0a0; border-right: 1px solid rgba(255,255,255,0.1); }
.th-group-c { background: rgba(100,60,180,0.20); color: #aa88ee; border-right: 1px solid rgba(255,255,255,0.1); }
.th-group-d { background: rgba(180,120,0,0.22);  color: #e0b840; border-right: 1px solid rgba(255,255,255,0.1); }
.th-group-e { background: rgba(40,120,160,0.22); color: #70c0d8; }

/* 字段表头（第二行）*/
.data-table thead tr:last-child th {
  background: rgba(15,25,50,0.98);
  color: rgba(190,210,235,0.85);
  padding: 10px 8px;
  text-align: center;
  font-weight: 600;
  font-size: 12px;
  border-bottom: 2px solid rgba(255,255,255,0.12);
  border-right: 1px solid rgba(255,255,255,0.07);
}
.data-table thead tr:last-child th:last-child { border-right: none; }

.ok-h { color: rgba(0,220,150,0.8) !important; }
.ng-h { color: rgba(255,100,120,0.8) !important; }

/* 数据行 */
.data-row {
  border-bottom: 1px solid rgba(255,255,255,0.05);
  transition: background 0.15s;
}
.data-row:hover { background: rgba(84,110,122,0.08); }
.data-row:nth-child(even) { background: rgba(255,255,255,0.02); }
.data-row:nth-child(even):hover { background: rgba(84,110,122,0.08); }

.data-table td {
  padding: 10px 8px;
  border-right: 1px solid rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
  vertical-align: middle;
}
.data-table td:last-child { border-right: none; }

/* 列宽 */
.col-date  { min-width: 90px; }
.col-shift { min-width: 62px; }
.col-time  { min-width: 96px; }
.col-dev   { min-width: 80px; }
.col-num   { min-width: 68px; }
.col-pct   { min-width: 64px; }
.col-team  { min-width: 68px; }
.col-staff { min-width: 150px; max-width: 220px; overflow: hidden; text-overflow: ellipsis; }

.center { text-align: center !important; }
.mono   { font-family: 'Courier New', monospace; font-size: 11px; }
.ok     { color: #2ecc71; font-weight: 600; }
.ng     { color: #e74c3c; font-weight: 500; }
.dim    { color: rgba(255,255,255,0.4); }
.theory { color: rgba(140,200,255,0.85); }

.shift-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}
/* 早班 — 蓝紫 */
.shift-badge.shift-a { background: rgba(102,126,234,0.2); border: 1px solid rgba(102,126,234,0.4); color: #a0b4f0; }
/* 中班 — 橙黄 */
.shift-badge.shift-b { background: rgba(243,156,18,0.15); border: 1px solid rgba(243,156,18,0.4); color: #f8c76a; }
/* 晚班 — 兆青 */
.shift-badge.shift-c { background: rgba(26,188,156,0.15); border: 1px solid rgba(26,188,156,0.4); color: #6ee0c8; }
/* 婿班 — 红 */
.shift-badge.shift-d { background: rgba(231,76,60,0.15);  border: 1px solid rgba(231,76,60,0.4);  color: #f09090; }
/* 备用 — 紫 */
.shift-badge.shift-e { background: rgba(155,89,182,0.15); border: 1px solid rgba(155,89,182,0.4); color: #c8a8e9; }

.oee-cell  { font-weight: 700; }
.oee-good  { color: #2ecc71; }
.oee-mid   { color: #f39c12; }
.oee-low   { color: #e74c3c; }

/* 汇总行 */
.summary-row {
  background: rgba(84,110,122,0.1) !important;
  border-top: 2px solid rgba(84,110,122,0.3);
}
.summary-row td {
  padding: 11px 8px;
  font-weight: 600;
  color: rgba(255,255,255,0.9) !important;
}
.sum-label-cell {
  color: rgba(255,255,255,0.7) !important;
  font-size: 12px;
  letter-spacing: 0.3px;
  text-align: left !important;
  padding-left: 16px !important;
}

/* ── 空状态 ──────────────────────────────────────────── */
.empty-state-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  background: rgba(20,30,48,0.6);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.4);
}
.empty-state-card i    { font-size: 48px; margin-bottom: 16px; opacity: 0.5; }
.empty-state-card p    { font-size: 16px; margin: 0 0 6px; }
.empty-state-card span { font-size: 13px; }

/* ── 操作列 / 详情按钮 ──────────────────────────────────── */
.th-group-op {
  background: rgba(84,110,122,0.22);
  color: #90aabb;
  text-align: center;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
  padding: 7px 10px;
}
.col-op { min-width: 72px; width: 72px; }

.detail-btn {
  padding: 4px 12px;
  background: rgba(102,126,234,0.15);
  border: 1px solid rgba(102,126,234,0.35);
  border-radius: 6px;
  color: #a0b4f0;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.detail-btn:hover {
  background: rgba(102,126,234,0.3);
  border-color: rgba(102,126,234,0.6);
  color: #fff;
  transform: translateY(-1px);
}

/* ── 弹窗 ──────────────────────────────────────────────── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.72);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.25s;
}
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

.modal-container {
  background: rgba(18,28,46,0.98);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 92%;
  max-width: 680px;
  max-height: 88vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 24px 64px rgba(0,0,0,0.55);
  animation: slideUp 0.28s;
}
@keyframes slideUp {
  from { opacity: 0; transform: translateY(28px); }
  to   { opacity: 1; transform: translateY(0); }
}

.modal-header {
  padding: 18px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}
.modal-header h3 {
  font-size: 17px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  flex-wrap: wrap;
}
.modal-header h3 i { color: #667eea; }

.detail-title-badge {
  font-size: 13px;
  font-weight: 500;
  color: rgba(255,255,255,0.55);
  background: rgba(255,255,255,0.07);
  padding: 3px 10px;
  border-radius: 8px;
}

.modal-close {
  width: 34px; height: 34px;
  border: none;
  background: rgba(255,255,255,0.08);
  border-radius: 8px;
  color: rgba(255,255,255,0.6);
  cursor: pointer;
  transition: all 0.2s;
  display: flex; align-items: center; justify-content: center;
  font-size: 15px;
  flex-shrink: 0;
}
.modal-close:hover { background: rgba(255,255,255,0.18); color: #fff; }

.modal-body {
  padding: 20px 24px;
  overflow-y: auto;
  flex: 1;
}
.modal-footer {
  padding: 14px 24px;
  border-top: 1px solid rgba(255,255,255,0.08);
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
.footer-spacer { flex: 1; }

.modal-btn {
  padding: 8px 22px;
  border-radius: 8px;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 7px;
}
.modal-btn.cancel {
  background: rgba(255,255,255,0.08);
  color: rgba(255,255,255,0.8);
  border: 1px solid rgba(255,255,255,0.15);
}
.modal-btn.cancel:hover { background: rgba(255,255,255,0.14); color: #fff; }

/* ── 详情内容区 ─────────────────────────────────────────── */
.detail-section {
  margin-bottom: 20px;
}
.detail-section:last-child { margin-bottom: 0; }

.detail-section-title {
  font-size: 13px;
  font-weight: 700;
  color: rgba(255,255,255,0.5);
  text-transform: uppercase;
  letter-spacing: 0.8px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 7px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255,255,255,0.07);
}
.detail-section-title i { color: #546e7a; }

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.detail-item {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 10px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.detail-item.oee-highlight {
  background: rgba(102,126,234,0.08);
  border-color: rgba(102,126,234,0.2);
}

.di-label {
  font-size: 11px;
  color: rgba(255,255,255,0.45);
  font-weight: 500;
  letter-spacing: 0.3px;
}
.di-val {
  font-size: 15px;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
}
.di-val.theory { color: rgba(140,200,255,0.9); }
.di-val.ok     { color: #2ecc71; }
.di-val.ng     { color: #e74c3c; }
.di-val.dim    { color: rgba(255,255,255,0.4); }

/* ── 人员列表 ───────────────────────────────────────────── */
.staff-detail-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.staff-detail-item {
  display: flex;
  align-items: center;
  gap: 14px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 10px;
  padding: 12px 16px;
  transition: background 0.15s;
}
.staff-detail-item:hover { background: rgba(255,255,255,0.07); }

/* 金色第一 */
.staff-detail-item.rank-gold {
  background: rgba(255,200,50,0.08);
  border-color: rgba(255,200,50,0.3);
}
.staff-detail-item.rank-gold .staff-avatar { color: #ffd700; }
.staff-detail-item.rank-gold .pct-fill { background: linear-gradient(90deg, #f6c700, #ffaa00); }
.staff-detail-item.rank-gold .pct-label { color: #ffd700; }
.staff-detail-item.rank-gold .sh-pct { color: #ffd700; }

/* 银色第二 */
.staff-detail-item.rank-silver {
  background: rgba(192,200,215,0.07);
  border-color: rgba(192,200,215,0.25);
}
.staff-detail-item.rank-silver .staff-avatar { color: #c0c8d7; }
.staff-detail-item.rank-silver .pct-fill { background: linear-gradient(90deg, #b0bec5, #cfd8dc); }
.staff-detail-item.rank-silver .pct-label { color: #c0c8d7; }
.staff-detail-item.rank-silver .sh-pct { color: #c0c8d7; }

/* 铜色第三 */
.staff-detail-item.rank-bronze {
  background: rgba(180,110,60,0.07);
  border-color: rgba(180,110,60,0.25);
}
.staff-detail-item.rank-bronze .staff-avatar { color: #cd7f32; }
.staff-detail-item.rank-bronze .pct-fill { background: linear-gradient(90deg, #cd7f32, #e8a060); }
.staff-detail-item.rank-bronze .pct-label { color: #cd7f32; }
.staff-detail-item.rank-bronze .sh-pct { color: #cd7f32; }

/* 铁色第四 */
.staff-detail-item.rank-iron {
  background: rgba(100,110,120,0.06);
  border-color: rgba(120,130,140,0.2);
}
.staff-detail-item.rank-iron .staff-avatar { color: #8898a8; }
.staff-detail-item.rank-iron .pct-fill { background: linear-gradient(90deg, #607080, #8898a8); }
.staff-detail-item.rank-iron .pct-label { color: #8898a8; }
.staff-detail-item.rank-iron .sh-pct { color: #8898a8; }

/* 土色第五及以后 */
.staff-detail-item.rank-earth {
  background: rgba(130,100,70,0.05);
  border-color: rgba(150,120,80,0.18);
}
.staff-detail-item.rank-earth .staff-avatar { color: #a08060; }
.staff-detail-item.rank-earth .pct-fill { background: linear-gradient(90deg, #7a6040, #a08060); }
.staff-detail-item.rank-earth .pct-label { color: #a08060; }
.staff-detail-item.rank-earth .sh-pct { color: #a08060; }

.staff-avatar {
  font-size: 28px;
  color: rgba(102,126,234,0.7);
  flex-shrink: 0;
}

.staff-info { flex: 1; }
.staff-name {
  font-size: 15px;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
}
.staff-code-tag {
  font-size: 12px;
  color: rgba(255,255,255,0.45);
  font-weight: 400;
  margin-left: 4px;
}
.staff-meta {
  font-size: 12px;
  color: rgba(255,255,255,0.45);
  margin-top: 3px;
  display: flex;
  align-items: center;
  gap: 5px;
  flex-wrap: wrap;
}
.meta-sep { opacity: 0.4; }

.staff-pct-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 7px;
}
.pct-bar {
  flex: 1;
  height: 5px;
  background: rgba(255,255,255,0.1);
  border-radius: 3px;
  overflow: hidden;
}
.pct-fill {
  height: 100%;
  background: linear-gradient(90deg, #667eea, #764ba2);
  border-radius: 3px;
  transition: width 0.4s ease;
}
.pct-label {
  font-size: 12px;
  font-weight: 600;
  color: #a0b4f0;
  min-width: 38px;
  text-align: right;
}

.staff-hours {
  text-align: right;
  flex-shrink: 0;
}
.sh-time {
  font-size: 14px;
  font-weight: 600;
  color: rgba(255,255,255,0.85);
}
.sh-pct {
  font-size: 12px;
  color: rgba(102,126,234,0.85);
  font-weight: 500;
}
.no-staff {
  font-size: 13px;
  color: rgba(255,255,255,0.3);
  padding: 12px 0;
  text-align: center;
}
</style>
