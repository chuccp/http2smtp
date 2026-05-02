import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { md5, generateRandomString } from '@/utils/crypto'
import { login, getSystemInfo } from '@/api/auth'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('http2smtp-token') || '')
  const username = ref<string>(localStorage.getItem('http2smtp-username') || '')
  const isAdmin = ref<boolean>(localStorage.getItem('http2smtp-isAdmin') === 'true')
  const rememberUsername = ref<boolean>(localStorage.getItem('http2smtp-remember') === 'true')

  const isLoggedIn = computed(() => {
    return token.value.length > 0
  })

  const getUsername = computed(() => {
    return username.value
  })

  const getIsAdmin = computed(() => {
    return isAdmin.value
  })

  async function loginAction(user: string, pass: string, remember: boolean): Promise<boolean> {
    try {
      const res = await login(user, pass)
      if (res.code === 0 || res.code === 200) {
        token.value = res.data?.user_token ?? user
        username.value = user
        isAdmin.value = res.data?.isAdmin ?? false
        localStorage.setItem('http2smtp-token', res.data?.user_token ?? user)
        localStorage.setItem('http2smtp-username', user)
        localStorage.setItem('http2smtp-isAdmin', String(isAdmin.value))
        localStorage.setItem('http2smtp-remember', String(remember))
        rememberUsername.value = remember
        return true
      }
      return false
    } catch (e) {
      console.error('Login error', e)
      return false
    }
  }

  function logoutAction() {
    token.value = ''
    username.value = ''
    isAdmin.value = false
    localStorage.removeItem('http2smtp-token')
    localStorage.removeItem('http2smtp-username')
    localStorage.removeItem('http2smtp-isAdmin')
    router.push('/')
  }

  async function checkSystemInit(): Promise<boolean> {
    try {
      const res = await getSystemInfo()
      if (res.code === 0 || res.code === 200) {
        return res.data.initialized
      }
      return false
    } catch (e) {
      console.error('Check init error', e)
      return false
    }
  }

  return {
    token,
    username,
    isAdmin,
    rememberUsername,
    isLoggedIn,
    getUsername,
    getIsAdmin,
    loginAction,
    logoutAction,
    checkSystemInit
  }
})
