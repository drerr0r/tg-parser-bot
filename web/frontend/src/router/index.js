import { createRouter, createWebHistory } from 'vue-router'
import store from '../store'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/rules',
    name: 'Rules',
    component: () => import('../views/Rules.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/posts',
    name: 'Posts',
    component: () => import('../views/Posts.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/stats',
    name: 'Stats',
    component: () => import('../views/Stats.vue'),
    meta: { requiresAuth: true }
  },
  {
  path: '/logs',
  name: 'Logs',
  component: () => import('../views/Logs.vue'),
  meta: { requiresAuth: true }
}
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Навигационный guard
router.beforeEach(async (to, from, next) => {
  // Проверяем аутентификацию при первом посещении
  if (!store.state.isAuthenticated) {
    await store.dispatch('checkAuth')
  }

  const isAuthenticated = store.state.isAuthenticated

  if (to.meta.requiresAuth && !isAuthenticated) {
    // Перенаправляем на страницу входа если требуется авторизация
    next('/login')
  } else if (to.meta.requiresGuest && isAuthenticated) {
    // Если уже авторизован, не пускаем на страницу входа
    next('/')
  } else {
    next()
  }
})

export default router