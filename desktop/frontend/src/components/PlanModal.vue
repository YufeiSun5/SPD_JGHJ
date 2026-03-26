<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3>
          <i class="fas fa-clipboard-list"></i>
          {{ mode === 'add' ? '新增计划' : '编辑计划' }}
        </h3>
        <button class="modal-close" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>工单号 <span class="required">*</span></label>
          <input 
            v-model="form.order_no" 
            type="text" 
            placeholder="请输入工单号" 
            :disabled="mode === 'edit'"
          />
        </div>
        
        <div class="form-group">
          <label>产品型号 <span class="required">*</span></label>
          <input v-model="form.product_code" type="text" placeholder="请输入产品型号" />
        </div>
        
        <div class="form-group">
          <label>计划数量 <span class="required">*</span></label>
          <input v-model.number="form.plan_qty" type="number" min="0" placeholder="请输入数量" />
        </div>
        
        <div class="form-row">
          <div class="form-group">
            <label>目标网关</label>
            <div class="custom-select">
              <select v-model.number="selectedGatewayId" @change="onGatewayChange">
                <option :value="null">请选择网关</option>
                <option v-for="gateway in gateways" :key="gateway.id" :value="gateway.id">
                  {{ gateway.gw_name }}
                </option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label>目标设备</label>
            <div class="custom-select">
              <select v-model.number="form.target_device_id" :disabled="!selectedGatewayId">
                <option :value="null">请选择设备</option>
                <option v-for="device in filteredDevices" :key="device.id" :value="device.id">
                  {{ device.device_name }}
                </option>
              </select>
            </div>
          </div>
        </div>
        
        <div class="form-group" v-if="mode === 'edit'">
          <label>状态</label>
          <div class="custom-select">
            <select v-model.number="form.status">
              <option :value="0">待产</option>
              <option :value="1">生产中</option>
              <option :value="2">暂停</option>
              <option :value="3">完工</option>
              <option :value="4">关闭</option>
            </select>
          </div>
        </div>
        
        <div class="form-group" v-if="mode === 'edit'">
          <label>实际产出数量</label>
          <input v-model.number="form.actual_qty" type="number" min="0" readonly disabled />
        </div>
      </div>
      
      <div class="modal-footer">
        <button 
          v-if="mode === 'edit'" 
          class="modal-btn delete" 
          @click="handleDelete"
        >
          <i class="fas fa-trash"></i> 删除
        </button>
        <div class="footer-spacer"></div>
        <button class="modal-btn cancel" @click="$emit('close')">取消</button>
        <button class="modal-btn confirm" @click="handleSave">
          <i class="fas fa-save"></i> 保存
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted } from 'vue'

const props = defineProps({
  mode: {
    type: String,
    required: true
  },
  plan: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'save', 'delete'])

const form = reactive({ ...props.plan })
const gateways = ref([])
const devices = ref([])
const selectedGatewayId = ref(null)

// 根据选中的网关过滤设备
const filteredDevices = computed(() => {
  if (!selectedGatewayId.value) {
    return []
  }
  return devices.value.filter(d => d.gateway_id === selectedGatewayId.value)
})

// 网关切换时清空设备选择
const onGatewayChange = () => {
  form.target_device_id = null
}

// 加载网关列表
const loadGateways = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllGateways()
      gateways.value = result || []
      
      // 如果是新增模式且有网关，默认选中第一个
      if (props.mode === 'add' && gateways.value.length > 0 && !selectedGatewayId.value) {
        selectedGatewayId.value = gateways.value[0].id
      }
    }
  } catch (e) {
    console.error('加载网关失败:', e)
  }
}

// 加载设备列表
const loadDevices = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const result = await window.go.main.App.GetAllDevices()
      devices.value = result || []
      
      // 如果是编辑模式且已有设备ID，自动选中对应的网关
      if (props.mode === 'edit' && form.target_device_id) {
        const device = devices.value.find(d => d.id === form.target_device_id)
        if (device) {
          selectedGatewayId.value = device.gateway_id
        }
      }
      // 如果是新增模式且已选中网关但未选设备，默认选中第一个设备
      else if (props.mode === 'add' && selectedGatewayId.value && !form.target_device_id) {
        const firstDevice = devices.value.find(d => d.gateway_id === selectedGatewayId.value)
        if (firstDevice) {
          form.target_device_id = firstDevice.id
        }
      }
    }
  } catch (e) {
    console.error('加载设备失败:', e)
  }
}

const handleSave = () => {
  // 简单验证
  if (!form.order_no || !form.product_code || !form.plan_qty) {
    alert('请填写必填项（工单号、产品型号、计划数量）')
    return
  }
  
  if (form.plan_qty <= 0) {
    alert('计划数量必须大于0')
    return
  }
  
  emit('save', { ...form })
}

const handleDelete = () => {
  emit('delete', { ...form })
}

onMounted(async () => {
  await loadGateways()
  await loadDevices()
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
  max-width: 600px;
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
}

.modal-header h3 {
  font-size: 18px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
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
  align-items: center;
  gap: 10px;
}

.footer-spacer {
  flex: 1;
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
}

.modal-btn.delete {
  background: rgba(231, 76, 60, 0.2);
  border: 1px solid rgba(231, 76, 60, 0.4);
  color: #e74c3c;
}

.modal-btn.delete:hover {
  background: #e74c3c;
  color: #fff;
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.3);
}

/* 表单样式 */
.form-group {
  margin-bottom: 20px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 13px;
  color: rgba(255,255,255,0.8);
  font-weight: 500;
}

.form-group .required {
  color: #e74c3c;
}

.form-group input,
.form-group textarea {
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

.form-group input:focus,
.form-group textarea:focus {
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

.form-group textarea {
  resize: vertical;
}
</style>

