<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3>
          <i class="fas fa-exchange-alt"></i>
          员工调动
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="staff-info">
          <div class="info-row">
            <label><i class="fas fa-user"></i> 员工姓名：</label>
            <span class="value">{{ staff.name }}</span>
          </div>
          <div class="info-row">
            <label><i class="fas fa-id-card"></i> 工号：</label>
            <span class="value">{{ staff.staff_code }}</span>
          </div>
          <div class="info-row">
            <label><i class="fas fa-users"></i> 当前班组：</label>
            <span class="value">
              {{ staff.current_team ? staff.current_team.team_name : '未分配' }}
            </span>
          </div>
        </div>

        <div class="transfer-arrow">
          <i class="fas fa-arrow-down"></i>
        </div>

        <div class="form-group">
          <label>调动至班组 <span class="required">*</span></label>
          <div class="custom-select">
            <select v-model="form.new_team_id">
              <option :value="null" disabled>请选择目标班组</option>
              <option v-for="team in filteredTeams" :key="team.id" :value="team.id">
                {{ team.team_name }}
              </option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label>操作人</label>
          <input v-model="form.operator_name" type="text" placeholder="请输入操作人姓名（可选）" />
          <span class="hint">记录本次调动的操作人</span>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="modal-btn cancel" @click="$emit('close')">
          <i class="fas fa-times"></i> 取消
        </button>
        <button class="modal-btn confirm" @click="handleSave">
          <i class="fas fa-check"></i> 确认调动
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed } from 'vue'

const props = defineProps({
  staff: {
    type: Object,
    required: true
  },
  teams: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'save'])

const form = reactive({
  staff_id: props.staff.id,
  new_team_id: null,
  operator_name: ''
})

// 过滤掉当前所在的班组
const filteredTeams = computed(() => {
  return props.teams.filter(team => team.id !== props.staff.current_team_id)
})

const handleSave = () => {
  if (!form.new_team_id) {
    alert('请选择目标班组')
    return
  }

  if (form.new_team_id === props.staff.current_team_id) {
    alert('目标班组不能与当前班组相同')
    return
  }
  
  emit('save', { ...form })
}
</script>

<style scoped>
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

.modal-container {
  background: rgba(20, 30, 48, 0.98);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  animation: slideUp 0.3s;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, rgba(243, 156, 18, 0.1) 0%, rgba(230, 126, 34, 0.1) 100%);
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

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  background: rgba(10, 14, 39, 0.5);
}

.staff-info {
  background: rgba(255,255,255,0.05);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
}

.info-row {
  display: flex;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.info-row:last-child {
  border-bottom: none;
}

.info-row label {
  font-size: 13px;
  color: rgba(255,255,255,0.6);
  min-width: 120px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-row .value {
  font-size: 14px;
  color: rgba(255,255,255,0.9);
  font-weight: 500;
}

.transfer-arrow {
  text-align: center;
  font-size: 24px;
  color: #f39c12;
  margin: 20px 0;
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(10px); }
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 13px;
  color: rgba(255,255,255,0.8);
  font-weight: 500;
}

.required {
  color: #e74c3c;
}

.form-group input {
  width: 100%;
  padding: 10px 12px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  transition: all 0.2s;
  font-family: inherit;
}

.form-group input:focus {
  outline: none;
  border-color: #f39c12;
  background: rgba(255,255,255,0.08);
}

.custom-select {
  position: relative;
  width: 100%;
}

.custom-select::after {
  content: '\f078';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255,255,255,0.5);
  pointer-events: none;
  font-size: 11px;
}

.custom-select select {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  width: 100%;
  padding: 10px 36px 10px 12px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  transition: all 0.2s;
  font-family: inherit;
  cursor: pointer;
}

.custom-select select:focus {
  outline: none;
  border-color: #f39c12;
  background: rgba(255,255,255,0.08);
}

.custom-select select option {
  background: #1a1f3a;
  color: #fff;
  padding: 10px;
}

.hint {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: rgba(255,255,255,0.5);
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
  background: linear-gradient(135deg, #f39c12 0%, #e67e22 100%);
  color: #fff;
}

.modal-btn.confirm:hover {
  box-shadow: 0 4px 12px rgba(243, 156, 18, 0.4);
  transform: translateY(-2px);
}
</style>

