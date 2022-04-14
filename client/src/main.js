import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap'
import 'vue-select/dist/vue-select.css'
import axios from 'axios'
import _ from 'lodash'

import App from './App.vue'
import router from './router'

const app = createApp(App)

_.extend(app.config.globalProperties, {
  $http: axios,
  $api: axios.create({
    baseURL: import.meta.env.VITE_API,
  }),
})

app.use(createPinia())
app.use(router)

app.mount('#app')
