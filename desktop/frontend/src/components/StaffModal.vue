<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3>
          <i class="fas fa-user-edit"></i>
          {{ mode === 'add' ? '新增员工' : '编辑员工' }}
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>工号 <span class="required">*</span></label>
          <input 
            v-model="form.staff_code" 
            type="text" 
            placeholder="请输入员工工号" 
            :disabled="mode === 'edit'"
          />
          <span class="hint">员工唯一标识，创建后不可修改</span>
        </div>
        
        <div class="form-group">
          <label>姓名 <span class="required">*</span></label>
          <input v-model="form.name" type="text" placeholder="请输入员工姓名" />
        </div>
        
        <div class="form-group" v-if="mode === 'edit'">
          <label>当前班组</label>
          <div class="readonly-field">
            <i class="fas fa-users"></i>
            <span>{{ getCurrentTeamName() }}</span>
          </div>
          <span class="hint">班组分配请通过"班组管理"或"一键换班"操作</span>
        </div>
        
        <div class="form-group" v-if="mode === 'edit'">
          <label>在职状态</label>
          <div class="radio-group">
            <label class="radio-label">
              <input type="radio" :value="1" v-model="form.is_active" />
              <span class="radio-text">
                <i class="fas fa-check-circle"></i> 在职
              </span>
            </label>
            <label class="radio-label">
              <input type="radio" :value="0" v-model="form.is_active" />
              <span class="radio-text">
                <i class="fas fa-times-circle"></i> 离职
              </span>
            </label>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="modal-btn cancel" @click="$emit('close')">
          <i class="fas fa-times"></i> 取消
        </button>
        <button class="modal-btn confirm" @click="handleSave">
          <i class="fas fa-save"></i> 保存
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'

const props = defineProps({
  mode: {
    type: String,
    required: true
  },
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

const form = reactive({ ...props.staff })

const getCurrentTeamName = () => {
  if (!form.current_team_id) return '未分配班组'
  const team = props.teams.find(t => t.id === form.current_team_id)
  return team ? team.team_name : '未分配班组'
}

const handleSave = () => {
  if (!form.staff_code || !form.name) {
    alert('请填写工号和姓名')
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

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
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

/* 表单 */
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
  border-color: #667eea;
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
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}

.custom-select select option {
  background: #1a1f3a;
  color: #fff;
  padding: 10px;
}

.readonly-field {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: rgba(255,255,255,0.7);
  font-size: 14px;
}

.readonly-field i {
  color: #667eea;
}

.form-group input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.hint {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: rgba(255,255,255,0.5);
}

.radio-group {
  display: flex;
  gap: 20px;
}

.radio-label {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 10px 16px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  transition: all 0.2s;
}

.radio-label:hover {
  border-color: rgba(102, 126, 234, 0.5);
  background: rgba(255,255,255,0.08);
}

.radio-label input[type="radio"] {
  margin-right: 8px;
  width: auto;
  padding: 0;
}

.radio-text {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: rgba(255,255,255,0.8);
}
</style>

