import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Rules from '../views/Rules.vue'
import Posts from '../views/Posts.vue'
import Stats from '../views/Stats.vue'

const routes = [
  { path: '/', name: 'Home', component: Home },
  { path: '/rules', name: 'Rules', component: Rules },
  { path: '/posts', name: 'Posts', component: Posts },
  { path: '/stats', name: 'Stats', component: Stats }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
