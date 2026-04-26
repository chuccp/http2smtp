<template>
  <div class="login-container">
    <el-card class="login-card">
      <div class="card-header">
        <div></div>
        <el-dropdown trigger="click" @command="handleLangChange">
          <div class="lang-switcher">
            <el-icon><ChatDotRound /></el-icon>
            <span class="lang-label">{{ langLabels[currentLang] || 'EN' }}</span>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="zh-cn" :disabled="currentLang === 'zh-cn'">中文</el-dropdown-item>
              <el-dropdown-item command="zh-tw" :disabled="currentLang === 'zh-tw'">繁體</el-dropdown-item>
              <el-dropdown-item command="ja" :disabled="currentLang === 'ja'">日本語</el-dropdown-item>
              <el-dropdown-item command="en" :disabled="currentLang === 'en'">English</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
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
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { ChatDotRound } from '@element-plus/icons-vue'

const { t, locale } = useI18n()

const currentLang = computed(() => locale.value)
const langLabels: Record<string, string> = {
  'zh-cn': '中文',
  'zh-tw': '繁體',
  ja: '日本語',
  en: 'EN'
}

const handleLangChange = (command: string) => {
  locale.value = command
  localStorage.setItem('http2smtp-lang', command)
}

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

.card-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
}

.lang-switcher {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 4px 10px;
  border-radius: 6px;
  color: #909399;
  transition: all 0.2s;
}

.lang-switcher:hover {
  color: #409EFF;
  background: #f0f2f5;
}

.lang-label {
  font-size: 13px;
  font-weight: 500;
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
