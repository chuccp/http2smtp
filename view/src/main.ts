import { createApp } from 'vue'
import pinia from './store'
import router from './router'
import i18n from './locales'
import App from './App.vue'
import './styles/index.scss'
import 'element-plus/dist/index.css'

const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(i18n)

app.mount('#app')
