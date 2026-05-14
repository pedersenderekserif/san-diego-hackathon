import { createRouter, createWebHistory } from 'vue-router'
import LandingPage from '../pages/LandingPage.vue'
import SearchPage from '../pages/SearchPage.vue'
import ExplorerPage from '../pages/ExplorerPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: LandingPage, meta: { title: 'Home' } },
    { path: '/explore', component: ExplorerPage, meta: { title: 'Explore' } },
    { path: '/search', component: SearchPage, meta: { title: 'Search' } }
  ]
})

router.afterEach((to) => {
  document.title = to.meta.title
    ? `${to.meta.title} — Employer Network Explorer`
    : 'Employer Network Explorer'
})
