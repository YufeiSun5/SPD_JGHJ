<template>
  <div class="custom-titlebar">
    <div class="titlebar-left">
      <img src="/assets/images/icon.ico" alt="Icon" class="titlebar-icon-img">
      <span>NihonSpindle</span>
    </div>
    <div class="window-controls">
      <button class="window-btn" @click="$emit('minimize')" title="最小化">
        <i class="fas fa-window-minimize"></i>
      </button>
      <button class="window-btn" @click="$emit('maximize')" title="最大化">
        <i class="far fa-window-maximize"></i>
      </button>
      <button class="window-btn" @click="$emit('fullscreen')" title="全屏">
        <i class="fas fa-expand"></i>
      </button>
      <button class="window-btn close" @click="$emit('close')" title="关闭">
        <i class="fas fa-times"></i>
      </button>
    </div>
  </div>
</template>

<script setup>
defineEmits(['minimize', 'maximize', 'fullscreen', 'close'])
</script>

<style scoped>
.custom-titlebar {
  height: 32px;
  background: rgba(10, 14, 39, 0.98);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  user-select: none;
  transition: all 0.3s;
  -webkit-app-region: drag;
  --wails-draggable: drag;
}

.titlebar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: rgba(255,255,255,0.8);
  flex: 1;
  -webkit-app-region: drag;
  --wails-draggable: drag;
}

.titlebar-icon-img {
  width: 16px;
  height: 16px;
  object-fit: contain;
}

.window-controls {
  display: flex;
  -webkit-app-region: no-drag;
  --wails-draggable: no-drag;
}

.window-btn {
  width: 46px;
  height: 32px;
  border: none;
  background: transparent;
  color: rgba(255,255,255,0.7);
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  -webkit-app-region: no-drag;
  --wails-draggable: no-drag;
}

.window-btn:hover {
  background: rgba(255,255,255,0.1);
  color: #fff;
}

.window-btn.close:hover {
  background: #e81123;
  color: #fff;
}
</style>



