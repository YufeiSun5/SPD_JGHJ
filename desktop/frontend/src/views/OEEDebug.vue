<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <div class="page-title">
          <i class="fas fa-calculator"></i>
          OEE 数据调试
        </div>
        <div class="page-subtitle">OEE Debug &amp; Quality Analysis</div>
      </div>
      <div class="header-actions">
        <button class="action-btn secondary" @click="loadAll" :disabled="loading">
          <i class="fas fa-sync-alt" :class="{ 'fa-spin': loading }"></i>
          {{ loading ? '查询中...' : '刷新全部' }}
        </button>
      </div>
    </div>

    <!-- 顶部汇总卡片 -->
    <div class="summary-cards">
      <!-- 今日OEE汇总 -->
      <div class="summary-card" v-for="row in todaySummaryRows" :key="row.device_name">
        <div class="card-device">{{ row.device_name }}</div>
        <div class="card-metrics">
          <div class="metric">
            <span class="metric-label">总产量</span>
            <span class="metric-value primary">{{ row.total_products }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">良品</span>
            <span class="metric-value ok">{{ row.ok_qty }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">NG</span>
            <span class="metric-value" :class="row.ng_qty > 0 ? 'warn' : ''">{{ row.ng_qty }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">良品率</span>
            <span class="metric-value" :class="row.quality_pct < 90 ? 'warn' : 'ok'">{{ row.quality_pct }}%</span>
          </div>
          <div class="metric">
            <span class="metric-label">时间稼动率</span>
            <span class="metric-value" :class="row.availability_pct > 100 ? 'warn' : ''">{{ row.availability_pct }}%</span>
          </div>
          <div class="metric">
            <span class="metric-label">性能稼动率</span>
            <span class="metric-value" :class="row.performance_pct > 110 ? 'warn' : ''">{{ row.performance_pct }}%</span>
          </div>
          <div class="metric">
            <span class="metric-label">OEE</span>
            <span class="metric-value primary">{{ row.oee_pct }}%</span>
          </div>
        </div>
      </div>
      <div v-if="todaySummaryRows.length === 0 && !loading" class="summary-card empty">
        <i class="fas fa-inbox"></i> 暂无今日OEE数据
      </div>
    </div>

    <!-- 本月良品率 -->
    <div class="section-card">
      <div class="section-header">
        <span><i class="fas fa-calendar-alt"></i> 本月良品率（工单汇总）</span>
      </div>
      <div class="section-body">
        <div v-if="monthlyQuality.length === 0" class="empty-hint">暂无本月工单数据</div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>设备</th>
              <th>总产量</th>
              <th>良品数</th>
              <th>NG数</th>
              <th>良品率</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in monthlyQuality" :key="item.device_id">
              <td>{{ item.device_name }}</td>
              <td>{{ item.total_qty }}</td>
              <td class="ok">{{ item.ok_qty }}</td>
              <td :class="item.ng_qty > 0 ? 'warn' : ''">{{ item.ng_qty }}</td>
              <td :class="item.quality_rate < 90 ? 'warn' : 'ok'">{{ item.quality_rate }}%</td>
            </tr>
            <tr class="summary-row">
              <td>合计</td>
              <td>{{ monthlyTotalQty }}</td>
              <td class="ok">{{ monthlyTotalOk }}</td>
              <td :class="monthlyTotalNg > 0 ? 'warn' : ''">{{ monthlyTotalNg }}</td>
              <td :class="monthlyQualityRate < 90 ? 'warn' : 'ok'">{{ monthlyQualityRate }}%</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 今日良品率（运行记录） -->
    <div class="section-card">
      <div class="section-header">
        <span><i class="fas fa-calendar-day"></i> 今日良品率（运行记录）</span>
      </div>
      <div class="section-body">
        <div v-if="dailyQuality.length === 0" class="empty-hint">暂无今日运行记录</div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>设备</th>
              <th>总产量</th>
              <th>良品数</th>
              <th>NG数</th>
              <th>良品率</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in dailyQuality" :key="item.device_id">
              <td>{{ item.device_name }}</td>
              <td>{{ item.total_qty }}</td>
              <td class="ok">{{ item.ok_qty }}</td>
              <td :class="item.ng_qty > 0 ? 'warn' : ''">{{ item.ng_qty }}</td>
              <td :class="item.quality_rate < 90 ? 'warn' : 'ok'">{{ item.quality_rate }}%</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 今日OEE逐小时明细 -->
    <div class="section-card">
      <div class="section-header">
        <span><i class="fas fa-table"></i> 今日OEE逐小时明细</span>
        <span v-if="oeeError" class="error-hint">{{ oeeError }}</span>
      </div>
      <div class="section-body">
        <div v-if="oeeLoading" class="empty-hint">查询中...</div>
        <div v-else-if="oeeRows.length === 0" class="empty-hint">暂无数据</div>
        <table v-else class="data-table oee-table">
          <thead>
            <tr>
              <th>时段</th>
              <th>设备</th>
              <th>t_run(s)</th>
              <th>t_plan(s)</th>
              <th>总产量</th>
              <th>良品</th>
              <th>NG</th>
              <th>时间稼动率</th>
              <th>性能稼动率</th>
              <th>良品率</th>
              <th>OEE</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, i) in oeeRows"
              :key="i"
              :class="{ 'summary-row': row.time_period === '合计' || row.time_period === '=== 全天合计 ===' }"
            >
              <td>{{ row.time_period }}</td>
              <td>{{ row.device_name }}</td>
              <td>{{ row.t_run }}</td>
              <td>{{ row.t_plan }}</td>
              <td :class="row.total_products > 0 ? 'ok' : ''">{{ row.total_products }}</td>
              <td>{{ row.ok_qty }}</td>
              <td :class="row.ng_qty > 0 ? 'warn' : ''">{{ row.ng_qty }}</td>
              <td :class="row.availability_pct > 100 ? 'warn' : ''">{{ row.availability_pct }}%</td>
              <td :class="row.performance_pct > 110 ? 'warn' : ''">{{ row.performance_pct }}%</td>
              <td>{{ row.quality_pct }}%</td>
              <td>{{ row.oee_pct }}%</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const loading = ref(false)
const oeeLoading = ref(false)
const oeeError = ref('')
const oeeRows = ref([])
const monthlyQuality = ref([])
const dailyQuality = ref([])

// 今日合计行（过滤出合计行）
const todaySummaryRows = computed(() => {
  return oeeRows.value.filter(r =>
    r.time_period === '合计' || (r.time_period && r.time_period.includes('合计'))
  )
})

// 本月良品率汇总计算
const monthlyTotalQty = computed(() => monthlyQuality.value.reduce((s, i) => s + (i.total_qty || 0), 0))
const monthlyTotalOk = computed(() => monthlyQuality.value.reduce((s, i) => s + (i.ok_qty || 0), 0))
const monthlyTotalNg = computed(() => monthlyQuality.value.reduce((s, i) => s + (i.ng_qty || 0), 0))
const monthlyQualityRate = computed(() => {
  const total = monthlyTotalOk.value + monthlyTotalNg.value
  if (total === 0) return '100.0'
  return ((monthlyTotalOk.value / total) * 100).toFixed(1)
})

const loadOEE = async () => {
  oeeLoading.value = true
  oeeError.value = ''
  try {
    if (window.go?.main?.App?.DebugOEEDirect) {
      const result = await window.go.main.App.DebugOEEDirect()
      oeeRows.value = result?.rows || []
    }
  } catch (e) {
    oeeError.value = '查询失败: ' + e
  } finally {
    oeeLoading.value = false
  }
}

const loadMonthlyQuality = async () => {
  try {
    if (window.go?.main?.App?.GetMonthlyQualityByOrder) {
      monthlyQuality.value = await window.go.main.App.GetMonthlyQualityByOrder() || []
    }
  } catch (e) {
    console.error('加载本月良品率失败:', e)
  }
}

const loadDailyQuality = async () => {
  try {
    if (window.go?.main?.App?.GetDailyQualityByRun) {
      dailyQuality.value = await window.go.main.App.GetDailyQualityByRun() || []
    }
  } catch (e) {
    console.error('加载今日良品率失败:', e)
  }
}

const loadAll = async () => {
  loading.value = true
  await Promise.all([loadOEE(), loadMonthlyQuality(), loadDailyQuality()])
  loading.value = false
}

onMounted(() => {
  loadAll()
})
</script>

<style scoped>
.page-container {
  padding: 32px 40px;
  height: 100vh;
  overflow-y: auto;
  background: transparent;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 26px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title i { color: #00aaff; }

.page-subtitle {
  font-size: 13px;
  color: rgba(255,255,255,0.4);
  margin-top: 4px;
}

.header-actions { display: flex; gap: 10px; }

.action-btn {
  padding: 9px 18px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}

.action-btn.secondary {
  background: rgba(0,170,255,0.15);
  color: #00aaff;
  border: 1px solid rgba(0,170,255,0.3);
}

.action-btn.secondary:hover:not(:disabled) {
  background: rgba(0,170,255,0.25);
}

.action-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* 顶部汇总卡片 */
.summary-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.summary-card {
  flex: 1;
  min-width: 300px;
  background: rgba(0,170,255,0.06);
  border: 1px solid rgba(0,170,255,0.2);
  border-radius: 12px;
  padding: 16px 20px;
}

.summary-card.empty {
  color: rgba(255,255,255,0.3);
  text-align: center;
  padding: 30px;
  font-size: 14px;
}

.card-device {
  font-size: 15px;
  font-weight: 600;
  color: #00aaff;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(0,170,255,0.15);
}

.card-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 20px;
}

.metric {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.metric-label {
  font-size: 11px;
  color: rgba(255,255,255,0.4);
}

.metric-value {
  font-size: 18px;
  font-weight: 700;
  color: rgba(255,255,255,0.85);
}

/* Section 卡片 */
.section-card {
  background: rgba(20,30,50,0.6);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  margin-bottom: 20px;
  overflow: hidden;
}

.section-header {
  padding: 14px 20px;
  background: rgba(0,0,0,0.2);
  border-bottom: 1px solid rgba(255,255,255,0.06);
  font-size: 14px;
  font-weight: 600;
  color: rgba(255,255,255,0.8);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.section-header i { color: #00aaff; margin-right: 6px; }

.error-hint { color: #ff6b6b; font-size: 12px; font-weight: 400; }

.section-body { padding: 16px 20px; overflow-x: auto; }

.empty-hint {
  color: rgba(255,255,255,0.3);
  text-align: center;
  padding: 20px;
  font-size: 13px;
}

/* 表格 */
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.data-table th {
  padding: 8px 12px;
  text-align: left;
  color: rgba(255,255,255,0.5);
  font-weight: 500;
  border-bottom: 1px solid rgba(255,255,255,0.08);
  white-space: nowrap;
}

.data-table td {
  padding: 8px 12px;
  color: rgba(255,255,255,0.75);
  border-bottom: 1px solid rgba(255,255,255,0.04);
  white-space: nowrap;
}

.data-table tr:hover td { background: rgba(255,255,255,0.03); }

.data-table tr.summary-row td {
  background: rgba(0,170,255,0.08);
  font-weight: 700;
  color: rgba(255,255,255,0.95);
  border-top: 1px solid rgba(0,170,255,0.2);
}

.oee-table { min-width: 900px; }

/* 颜色 */
.ok { color: #4ade80 !important; }
.warn { color: #ff6b6b !important; font-weight: 600; }
.primary { color: #00aaff !important; }
</style>
