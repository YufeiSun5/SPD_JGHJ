<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container large">
      <div class="modal-header">
        <h3>
          <i class="fas fa-layer-group"></i>
          班组管理
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <!-- 新增班组表单 -->
        <div class="add-section">
          <h4><i class="fas fa-plus-circle"></i> 快速新增班组</h4>
          <div class="form-row">
            <div class="form-group">
              <input 
                v-model="newTeam.team_name" 
                type="text" 
                placeholder="班组名称 *" 
              />
            </div>
            <div class="form-group">
              <input 
                v-model="newTeam.leader_name" 
                type="text" 
                placeholder="班组长（可选）" 
              />
            </div>
            <button class="add-btn" @click="addTeam">
              <i class="fas fa-plus"></i> 添加
            </button>
          </div>
        </div>

        <!-- 班组列表 -->
        <div class="team-list">
          <div class="list-header">
            <h4><i class="fas fa-list"></i> 班组列表</h4>
            <span class="badge">{{ teams.length }} 个班组</span>
          </div>
          
          <div v-if="teams.length === 0" class="empty-state">
            <i class="fas fa-inbox"></i>
            <p>暂无班组数据，请先添加班组</p>
          </div>
          
          <div v-else class="team-grid">
            <div v-for="team in teams" :key="team.id" class="team-card">
              <div v-if="editingId === team.id" class="team-card-edit">
                <div class="edit-form">
                  <div class="edit-field">
                    <label>班组名称</label>
                    <input 
                      v-model="editForm.team_name"
                      type="text"
                      class="edit-input"
                    />
                  </div>
                  <div class="edit-field">
                    <label>班组长</label>
                    <input 
                      v-model="editForm.leader_name"
                      type="text"
                      class="edit-input"
                    />
                  </div>
                  <div class="edit-field">
                    <label>状态</label>
                    <div class="custom-select-inline">
                      <select v-model="editForm.status" class="edit-select">
                        <option :value="1">启用</option>
                        <option :value="0">禁用</option>
                      </select>
                    </div>
                  </div>
                </div>
                <div class="edit-actions">
                  <button class="btn-save" @click="saveEdit(team)">
                    <i class="fas fa-check"></i> 保存
                  </button>
                  <button class="btn-cancel" @click="cancelEdit">
                    <i class="fas fa-times"></i> 取消
                  </button>
                </div>
              </div>
              
              <div v-else class="team-card-view">
                <div class="team-info">
                  <div class="team-icon">
                    <i class="fas fa-users"></i>
                  </div>
                  <div class="team-details">
                    <div class="team-name">{{ team.team_name }}</div>
                    <div class="team-leader">
                      <i class="fas fa-user"></i>
                      {{ team.leader_name || '未指定班组长' }}
                    </div>
                  </div>
                  <span :class="['status-badge', team.status === 1 ? 'active' : 'inactive']">
                    {{ team.status === 1 ? '启用' : '禁用' }}
                  </span>
                </div>
                <div class="team-actions">
                  <button class="card-btn edit" @click="startEdit(team)" title="编辑">
                    <i class="fas fa-edit"></i>
                  </button>
                  <button class="card-btn delete" @click="deleteTeam(team)" title="删除">
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="modal-btn confirm" @click="$emit('close')">
          <i class="fas fa-check"></i> 完成
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'

const props = defineProps({
  mode: String,
  team: Object
})

const emit = defineEmits(['close', 'save', 'refresh'])

const teams = ref([])
const editingId = ref(null)
const newTeam = reactive({
  team_name: '',
  leader_name: ''
})
const editForm = reactive({
  id: null,
  team_name: '',
  leader_name: '',
  status: 1
})

// 加载班组列表
const loadTeams = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllTeams(null) // 加载所有班组
      teams.value = result || []
    }
  } catch (e) {
    console.error('加载班组失败:', e)
  }
}

// 新增班组
const addTeam = async () => {
  if (!newTeam.team_name) {
    alert('请输入班组名称')
    return
  }

  try {
    if (window.go && window.go.main && window.go.main.App) {
      await window.go.main.App.CreateTeam(
        newTeam.team_name,
        newTeam.leader_name || null
      )
      newTeam.team_name = ''
      newTeam.leader_name = ''
      await loadTeams()
      emit('refresh')
    }
  } catch (e) {
    console.error('新增班组失败:', e)
    alert('新增失败: ' + e)
  }
}

// 开始编辑
const startEdit = (team) => {
  editingId.value = team.id
  editForm.id = team.id
  editForm.team_name = team.team_name
  editForm.leader_name = team.leader_name || ''
  editForm.status = team.status
}

// 取消编辑
const cancelEdit = () => {
  editingId.value = null
}

// 保存编辑
const saveEdit = async (team) => {
  if (!editForm.team_name) {
    alert('班组名称不能为空')
    return
  }

  try {
    if (window.go && window.go.main && window.go.main.App) {
      await window.go.main.App.UpdateTeam(
        editForm.id,
        editForm.team_name,
        editForm.leader_name || null,
        editForm.status
      )
      editingId.value = null
      await loadTeams()
      emit('refresh')
    }
  } catch (e) {
    console.error('更新班组失败:', e)
    alert('更新失败: ' + e)
  }
}

// 删除班组
const deleteTeam = async (team) => {
  if (confirm(`确定要删除班组 "${team.team_name}" 吗？`)) {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        await window.go.main.App.DeleteTeam(team.id)
        await loadTeams()
        emit('refresh')
      }
    } catch (e) {
      console.error('删除班组失败:', e)
      alert('删除失败: ' + e)
    }
  }
}

onMounted(() => {
  loadTeams()
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background: rgba(30, 40, 60, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.1);
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-container.large {
  max-width: 800px;
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
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
  border: none;
  background: none;
  color: rgba(255,255,255,0.6);
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
  transition: all 0.2s;
}

.modal-close:hover {
  color: #fff;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: flex-end;
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

.modal-btn.confirm {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.modal-btn.confirm:hover {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  transform: translateY(-2px);
}

.modal-body {
  padding: 0;
}

.add-section {
  padding: 24px;
  background: rgba(102, 126, 234, 0.05);
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.add-section h4 {
  font-size: 14px;
  color: rgba(255,255,255,0.9);
  margin: 0 0 16px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 12px;
}

.add-btn {
  padding: 10px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}

.add-btn:hover {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  transform: translateY(-2px);
}

.team-list {
  padding: 24px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.list-header h4 {
  font-size: 14px;
  color: rgba(255,255,255,0.9);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.badge {
  padding: 4px 12px;
  background: rgba(102, 126, 234, 0.2);
  border: 1px solid #667eea;
  border-radius: 12px;
  font-size: 12px;
  color: #667eea;
  font-weight: 600;
}

.empty-state {
  text-align: center;
  padding: 60px 40px;
  color: rgba(255,255,255,0.5);
}

.empty-state i {
  font-size: 56px;
  display: block;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-state p {
  font-size: 14px;
  margin: 0;
}

.team-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
  max-height: 400px;
  overflow-y: auto;
  padding-right: 8px;
}

.team-grid::-webkit-scrollbar {
  width: 6px;
}

.team-grid::-webkit-scrollbar-track {
  background: rgba(255,255,255,0.05);
  border-radius: 3px;
}

.team-grid::-webkit-scrollbar-thumb {
  background: rgba(102, 126, 234, 0.5);
  border-radius: 3px;
}

.team-grid::-webkit-scrollbar-thumb:hover {
  background: rgba(102, 126, 234, 0.7);
}

.team-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  padding: 16px;
  transition: all 0.3s;
}

.team-card:hover {
  background: rgba(255,255,255,0.05);
  border-color: rgba(102, 126, 234, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}

.team-card-view {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.team-info {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.team-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #fff;
  flex-shrink: 0;
}

.team-details {
  flex: 1;
  min-width: 0;
}

.team-name {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.team-leader {
  font-size: 12px;
  color: rgba(255,255,255,0.6);
  display: flex;
  align-items: center;
  gap: 6px;
}

.team-actions {
  display: flex;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid rgba(255,255,255,0.05);
}

.card-btn {
  flex: 1;
  padding: 8px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  background: rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.card-btn:hover {
  transform: translateY(-2px);
}

.card-btn.edit:hover {
  background: #667eea;
  color: #fff;
}

.card-btn.delete:hover {
  background: #e74c3c;
  color: #fff;
}

.team-card-edit {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.edit-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.edit-field label {
  font-size: 12px;
  color: rgba(255,255,255,0.7);
  font-weight: 500;
}

.edit-input,
.edit-select {
  width: 100%;
  padding: 8px 12px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.2);
  border-radius: 6px;
  color: #fff;
  font-size: 13px;
}

.edit-input:focus,
.edit-select:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}

.custom-select-inline {
  position: relative;
}

.custom-select-inline::after {
  content: '\f078';
  font-family: 'Font Awesome 5 Free';
  font-weight: 900;
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255,255,255,0.5);
  pointer-events: none;
  font-size: 10px;
}

.custom-select-inline select {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
}

.edit-actions {
  display: flex;
  gap: 8px;
}

.btn-save,
.btn-cancel {
  flex: 1;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-weight: 500;
}

.btn-save {
  background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
  color: #fff;
}

.btn-save:hover {
  box-shadow: 0 4px 12px rgba(46, 204, 113, 0.4);
  transform: translateY(-2px);
}

.btn-cancel {
  background: rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.8);
}

.btn-cancel:hover {
  background: rgba(255,255,255,0.15);
}

.status-badge {
  padding: 4px 10px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
  display: inline-block;
  flex-shrink: 0;
}

.status-badge.active {
  background: rgba(46, 204, 113, 0.2);
  color: #2ecc71;
  border: 1px solid #2ecc71;
}

.status-badge.inactive {
  background: rgba(231, 76, 60, 0.2);
  color: #e74c3c;
  border: 1px solid #e74c3c;
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
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255,255,255,0.08);
}

.form-group input::placeholder {
  color: rgba(255,255,255,0.4);
}
</style>

