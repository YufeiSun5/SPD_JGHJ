<template>
  <slot></slot>
</template>

<script setup>
import { onMounted } from 'vue'
import { useMessage } from 'naive-ui'

onMounted(() => {
  // 在 n-message-provider 内部设置全局 message
  window.$message = useMessage()
})
</script>

























