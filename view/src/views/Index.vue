<template>
  <div class="loading-container">
    <el-icon class="loading-icon" :size="48">
      <Loading />
    </el-icon>
    <p>{{ t('common.loading') }}</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { getSystemInfo } from '@/api/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const router = useRouter()
const authStore = useAuthStore()

onMounted(async () => {
  try {
    const res = await getSystemInfo()
    if (res.code === 0 || res.code === 200) {
      const systemInfo = res.data
      if (!systemInfo.initialized) {
        router.push('/setup')
        return
      }
      if (systemInfo.hasLogin === false) {
        authStore.logoutAction()
        router.push('/login')
        return
      }
      if (authStore.isLoggedIn) {
        router.push('/dashboard')
        return
      }
      router.push('/login')
    }
  } catch (e) {
    console.error('Check system status error', e)
    // Fallback: go to login on error
    router.push('/login')
  }
})
</script>

<style scoped lang="scss">
.loading-container {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;

  .loading-icon {
    animation: spin 1.5s linear infinite;
    margin-bottom: 16px;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
}
</style>
