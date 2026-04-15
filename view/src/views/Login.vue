<template>
  <div class="login-container">
    <el-card class="login-card">
      <h1 class="title">HTTP2SMTP</h1>
      <p class="subtitle">HTTP to SMTP Gateway</p>
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        class="login-form"
        @submit.prevent="handleLogin"
      >
        <el-form-item>
          <el-input
            v-model="loginForm.username"
            :placeholder="t('auth.username')"
            prefix-icon="User"
            size="large"
            autocomplete="username"
          />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="loginForm.password"
            type="password"
            show-password
            :placeholder="t('auth.password')"
            prefix-icon="Lock"
            size="large"
            autocomplete="current-password"
          />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="loginForm.rememberMe">
            {{ t('auth.rememberMe') }}
          </el-checkbox>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="login-button"
            @click="handleLogin"
          >
            {{ t('auth.login') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const loginForm = ref({
  username: authStore.rememberUsername ? authStore.username : '',
  password: '',
  rememberMe: authStore.rememberUsername
})

const handleLogin = async () => {
  if (!loginForm.value.username) {
    ElMessage.warning(t('auth.pleaseEnterUsername'))
    return
  }
  if (!loginForm.value.password) {
    ElMessage.warning(t('auth.pleaseEnterPassword'))
    return
  }

  loading.value = true
  try {
    const success = await authStore.loginAction(
      loginForm.value.username,
      loginForm.value.password,
      loginForm.value.rememberMe
    )
    if (success) {
      ElMessage.success(t('auth.loginSuccess'))
      router.push('/dashboard')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);

  .title {
    text-align: center;
    margin: 0 0 8px;
    font-size: 28px;
    color: #303133;
  }

  .subtitle {
    text-align: center;
    margin: 0 0 32px;
    color: #909399;
    font-size: 14px;
  }

  .login-button {
    width: 100%;
  }
}

@media (max-width: 576px) {
  .login-card {
    width: 90%;
  }
}
</style>
