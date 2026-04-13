import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import i18n from '@/locales'

const { t } = i18n.global

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Index',
    component: () => import('@/views/Index.vue'),
    meta: { titleKey: '', requiresAuth: false }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { titleKey: 'auth.login', requiresAuth: false }
  },
  {
    path: '/setup',
    name: 'Setup',
    component: () => import('@/views/Setup.vue'),
    meta: { titleKey: 'setup.systemSetup', requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { titleKey: 'dashboard.dashboard', requiresAuth: true }
      },
      {
        path: 'smtp',
        name: 'SMTPManagement',
        component: () => import('@/views/smtp/index.vue'),
        meta: { titleKey: 'smtp.smtpManagement', requiresAuth: true }
      },
      {
        path: 'mail',
        name: 'MailManagement',
        component: () => import('@/views/mail/index.vue'),
        meta: { titleKey: 'mail.mailManagement', requiresAuth: true }
      },
      {
        path: 'tokens',
        name: 'TokenManagement',
        component: () => import('@/views/token/index.vue'),
        meta: { titleKey: 'token.tokenManagement', requiresAuth: true }
      },
      {
        path: 'schedule',
        name: 'ScheduleManagement',
        component: () => import('@/views/schedule/index.vue'),
        meta: { titleKey: 'schedule.scheduleManagement', requiresAuth: true }
      },
      {
        path: 'logs',
        name: 'SystemLogs',
        component: () => import('@/views/log/index.vue'),
        meta: { titleKey: 'log.systemLogs', requiresAuth: true }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/index.vue'),
        meta: { titleKey: 'settings.systemSettings', requiresAuth: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // Set page title
  if (to.meta?.titleKey) {
    document.title = `${t(to.meta.titleKey as string)} - HTTP2SMTP`
  } else {
    document.title = 'HTTP2SMTP - HTTP to SMTP Gateway'
  }

  const requiresAuth = to.meta.requiresAuth as boolean
  if (requiresAuth) {
    if (!authStore.isLoggedIn) {
      ElMessage.error(t('auth.pleaseEnterUsername'))
      next('/')
      return
    }
  }

  next()
})

export default router
