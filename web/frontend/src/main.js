import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import ruLocale from 'element-plus/es/locale/lang/ru'

import App from './App.vue'
import router from './router'
import store from './store'

const app = createApp(App)

// Регистрируем все иконки Element Plus
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(store)
app.use(router)
app.use(ElementPlus, {
  locale: ruLocale,
})

app.mount('#app')