import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// 导入全局样式
import './assets/styles/global.css'
import './assets/styles/common.css'
import './assets/styles/fontawesome.css'
import './assets/styles/naive-theme.css'

// Naive UI
import {
  create,
  NDataTable,
  NButton,
  NInput,
  NDatePicker,
  NSelect,
  NTag,
  NPagination,
  NSpace,
  NCheckbox,
  NCheckboxGroup,
  NSwitch,
  NInputNumber,
  NMessageProvider,
  NConfigProvider,
  zhCN,
  dateZhCN
} from 'naive-ui'

const naive = create({
  components: [
    NDataTable, 
    NButton, 
    NInput, 
    NDatePicker, 
    NSelect, 
    NTag, 
    NPagination, 
    NSpace,
    NCheckbox,
    NCheckboxGroup,
    NSwitch,
    NInputNumber,
    NMessageProvider,
    NConfigProvider
  ]
})

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(naive)

app.mount('#app')

