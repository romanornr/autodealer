import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Dashboard,
      meta: {
        title: 'Dashboard',
      },
    },
    {
      path: '/trade',
      name: 'trade',
      component: () => import('../views/Trade.vue'),
      meta: {
        title: 'Trade',
      },
    },
    {
      path: '/deposit',
      name: 'deposit',
      component: () => import('../views/Deposit.vue'),
      meta: {
        title: 'Deposit',
      },
    },
    {
      path: '/transfer',
      name: 'transfer',
      component: () => import('../views/Transfer.vue'),
      meta: {
        title: 'Transfer capital',
      },
    },
    {
      path: '/bank/transfer',
      name: 'bank-transfer',
      component: () => import('../views/BankTransfer.vue'),
      meta: {
        title: 'International Bank Transfer',
      },
    },
    {
      path: '/move',
      name: 'move',
      component: () => import('../views/Move.vue'),
      meta: {
        title: 'FTX MOVE Term Structure',
      },
    },
  ],
})

export default router
