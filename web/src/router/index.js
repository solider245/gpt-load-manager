import { createRouter, createWebHashHistory } from 'vue-router'
import Servers from '@/views/Servers.vue'
import ServerDetail from '@/views/ServerDetail.vue'
import ServerForm from '@/views/ServerForm.vue'

const routes = [
  { path: '/', redirect: '/servers' },
  { path: '/servers', name: 'Servers', component: Servers },
  { path: '/servers/add', name: 'AddServer', component: ServerForm },
  { path: '/servers/:id', name: 'ServerDetail', component: ServerDetail },
  { path: '/servers/:id/edit', name: 'EditServer', component: ServerForm },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
