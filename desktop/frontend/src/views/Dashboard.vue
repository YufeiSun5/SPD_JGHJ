<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <div class="dashboard-title">
        <i class="fas fa-chart-line"></i>
        实时监控驾驶舱
      </div>
      <div class="dashboard-subtitle">Real-time Monitoring Dashboard</div>
    </div>
    
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-label">
          <i class="fas fa-database"></i>
          测点总数
        </div>
        <div class="stat-value">{{ tags.length }}</div>
      </div>
      <div class="stat-card green">
        <div class="stat-label">
          <i class="fas fa-microchip"></i>
          Logic队列
        </div>
        <div class="stat-value">{{ channels.logic_chan || 0 }}</div>
      </div>
      <div class="stat-card orange">
        <div class="stat-label">
          <i class="fas fa-sync-alt"></i>
          Change队列
        </div>
        <div class="stat-value">{{ channels.change_chan || 0 }}</div>
      </div>
      <div class="stat-card blue">
        <div class="stat-label">
          <i class="fas fa-clock"></i>
          Cycle队列
        </div>
        <div class="stat-value">{{ channels.cycle_chan || 0 }}</div>
      </div>
    </div>
    
    <div class="tag-grid">
      <div v-for="tag in tags" :key="tag.var_name" 
           :class="['tag-card', tag.alarm_state ? 'alarm-' + tag.alarm_state.toLowerCase() : '']">
        <div class="tag-name">
          <i class="fas fa-tag"></i>
          {{ tag.display_name || tag.var_name }}
        </div>
        <div class="tag-value">
          {{ tag.value }}
          <span class="tag-unit">{{ tag.unit }}</span>
          <span v-if="tag.alarm_state" class="alarm-badge">
            <i class="fas fa-exclamation-triangle"></i> {{ tag.alarm_state }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'

const tagsRaw = ref([])
const channels = ref({})
let refreshTimer = null

// 固定顺序的标签列表 - 按照 var_id 排序
const tags = computed(() => {
  return [...tagsRaw.value].sort((a, b) => {
    // 首先按照 var_id 排序（如果存在）
    if (a.var_id !== undefined && b.var_id !== undefined) {
      return a.var_id - b.var_id
    }
    // 如果没有 var_id，按照 var_name 字母顺序排序
    const nameA = a.var_name || a.display_name || ''
    const nameB = b.var_name || b.display_name || ''
    return nameA.localeCompare(nameB)
  })
})

const refresh = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      tagsRaw.value = await window.go.main.App.GetRealtimeData()
      channels.value = await window.go.main.App.GetChannelStats()
    }
  } catch (e) {
    console.error('刷新失败:', e)
  }
}

onMounted(() => {
  refresh()
  refreshTimer = setInterval(refresh, 1000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.dashboard {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.dashboard-header {
  padding: 30px 40px 20px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
}

.dashboard-title {
  font-size: 28px;
  font-weight: bold;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.dashboard-title i {
  color: #667eea;
}

.dashboard-subtitle {
  font-size: 14px;
  color: rgba(255,255,255,0.6);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  padding: 0 40px 20px;
}

.stat-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 25px;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.3);
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 100px;
  height: 100px;
  background: rgba(255,255,255,0.1);
  border-radius: 50%;
  transform: translate(30%, -30%);
}

.stat-card.green { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
.stat-card.orange { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.stat-card.blue { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }

.stat-label { 
  font-size: 14px; 
  opacity: 0.9; 
  margin-bottom: 10px; 
  position: relative; 
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-label i {
  font-size: 20px;
}

.stat-value { 
  font-size: 42px; 
  font-weight: bold; 
  position: relative; 
  z-index: 1;
}

.tag-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
  padding: 20px 40px 40px;
  flex: 1;
  overflow-y: auto;
}

.tag-card {
  background: rgba(20, 30, 48, 0.6);
  backdrop-filter: blur(10px);
  padding: 24px;
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  transition: all 0.3s;
}

.tag-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
  border-color: rgba(102, 126, 234, 0.5);
}

.tag-card.alarm-hh { border-color: #e74c3c; box-shadow: 0 4px 20px rgba(231, 76, 60, 0.3); }
.tag-card.alarm-h { border-color: #e67e22; box-shadow: 0 4px 20px rgba(230, 126, 34, 0.3); }
.tag-card.alarm-l { border-color: #f39c12; box-shadow: 0 4px 20px rgba(243, 156, 18, 0.3); }
.tag-card.alarm-ll { border-color: #e74c3c; box-shadow: 0 4px 20px rgba(231, 76, 60, 0.3); }

.tag-name { 
  font-size: 13px; 
  color: rgba(255,255,255,0.5); 
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-name i {
  font-size: 14px;
}

.tag-value { 
  font-size: 32px; 
  font-weight: bold; 
  color: #fff;
}

.tag-unit { 
  font-size: 14px; 
  color: rgba(255,255,255,0.5); 
  margin-left: 8px;
}

.alarm-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 11px;
  margin-left: 12px;
  background: #e74c3c;
  color: #fff;
  font-weight: bold;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}
</style>



