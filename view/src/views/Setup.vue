<template>
  <div class="setup-container">
    <el-card class="setup-card">
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
      <h1 class="title">{{ t('setup.systemSetup') }}</h1>
      <p class="subtitle">{{ t('setup.setupWizard') }}</p>

      <div v-if="pageLoading" class="page-loading">
        <el-icon class="loading-icon" :size="32"><Loading /></el-icon>
      </div>

      <template v-else>
        <!-- Steps indicator -->
        <el-steps :active="currentStep" finish-status="success" simple class="setup-steps">
          <el-step :title="t('setup.step1Database')" />
          <el-step :title="t('setup.step2Admin')" />
        </el-steps>

        <!-- Step 1: Database Configuration -->
        <el-form
          v-show="currentStep === 0"
          ref="dbFormRef"
          :model="setupForm"
          :rules="dbRules"
          label-width="140px"
        >
          <el-divider content-position="left">{{ t('setup.databaseConfig') }}</el-divider>

          <el-form-item :label="t('setup.databaseType')" prop="dbType">
            <el-radio-group v-model="setupForm.dbType">
              <el-radio value="sqlite">{{ t('setup.sqlite') }}</el-radio>
              <el-radio value="mysql">{{ t('setup.mysql') }}</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item v-if="setupForm.dbType === 'sqlite'" :label="t('setup.sqliteFile')" prop="dbFile">
            <el-input v-model="setupForm.dbFile" :placeholder="'data.db'" />
          </el-form-item>

          <template v-if="setupForm.dbType === 'mysql'">
            <el-form-item :label="t('setup.mysqlHost')" prop="dbHost">
              <el-input v-model="setupForm.dbHost" placeholder="127.0.0.1" />
            </el-form-item>
            <el-form-item :label="t('setup.mysqlPort')" prop="dbPort">
              <el-input-number v-model="setupForm.dbPort" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item :label="t('setup.mysqlDatabase')" prop="dbName">
              <el-input v-model="setupForm.dbName" />
            </el-form-item>
            <el-form-item :label="t('setup.mysqlUser')" prop="dbUser">
              <el-input v-model="setupForm.dbUser" />
            </el-form-item>
            <el-form-item :label="t('setup.mysqlPassword')" prop="dbPass">
              <el-input v-model="setupForm.dbPass" type="password" show-password />
            </el-form-item>
            <el-form-item :label="t('setup.mysqlCharset')" prop="dbCharset">
              <el-input v-model="setupForm.dbCharset" placeholder="utf8mb4" />
            </el-form-item>
            <el-form-item>
              <el-button @click="handleTestConnection" :loading="testing">
                {{ t('setup.testConnection') }}
              </el-button>
            </el-form-item>
          </template>

          <template v-if="!isDocker">
            <el-divider content-position="left">{{ t('setup.portConfig') }}</el-divider>

            <el-form-item :label="t('setup.webPort')" prop="webPort">
              <el-input-number v-model="setupForm.webPort" :min="1" :max="65535" />
            </el-form-item>

            <el-form-item :label="t('setup.apiPort')" prop="apiPort">
              <el-input-number v-model="setupForm.apiPort" :min="1" :max="65535" />
            </el-form-item>
          </template>

          <el-form-item>
            <el-button
              type="primary"
              :loading="dbInitializing"
              @click="handleDbInit"
            >
              {{ t('setup.nextStep') }}
            </el-button>
          </el-form-item>
        </el-form>

        <!-- Step 2: Admin Account -->
        <el-form
          v-show="currentStep === 1"
          ref="adminFormRef"
          :model="adminForm"
          :rules="adminRules"
          label-width="140px"
        >
          <el-divider content-position="left">{{ t('setup.adminAccount') }}</el-divider>

          <!-- Show when admin already exists -->
          <template v-if="adminExists">
            <el-alert
              :title="t('setup.adminAlreadyExists')"
              type="warning"
              :closable="false"
              show-icon
              class="admin-exists-alert"
            />

            <!-- Show current admin username -->
            <el-form-item :label="t('setup.currentAdmin')">
              <span class="admin-name-display">{{ existingAdminName }}</span>
            </el-form-item>

            <el-form-item :label="t('setup.adminPassword')" prop="adminPass">
              <el-input v-model="adminForm.adminPass" type="password" show-password :placeholder="t('setup.newPasswordPlaceholder')" />
            </el-form-item>

            <el-form-item :label="t('setup.confirmPassword')" prop="confirmPassword">
              <el-input v-model="adminForm.confirmPassword" type="password" show-password />
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="adminInitializing"
                @click="handleAdminInit"
              >
                {{ t('setup.resetAdminConfirm') }}
              </el-button>
              <el-button @click="handleSkipAdmin" :loading="skipLoading">
                {{ t('setup.skipAdmin') }}
              </el-button>
            </el-form-item>
          </template>

          <!-- Normal admin creation form (no existing admin) -->
          <template v-else>
            <el-form-item :label="t('setup.adminUsername')" prop="adminUser">
              <el-input v-model="adminForm.adminUser" />
            </el-form-item>

            <el-form-item :label="t('setup.adminPassword')" prop="adminPass">
              <el-input v-model="adminForm.adminPass" type="password" show-password />
            </el-form-item>

            <el-form-item :label="t('setup.confirmPassword')" prop="confirmPassword">
              <el-input v-model="adminForm.confirmPassword" type="password" show-password />
            </el-form-item>

            <el-form-item>
              <el-button @click="currentStep = 0">
                {{ t('setup.prevStep') }}
              </el-button>
              <el-button
                type="primary"
                :loading="adminInitializing"
                @click="handleAdminInit"
              >
                {{ t('setup.initialize') }}
              </el-button>
            </el-form-item>
          </template>
        </el-form>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { checkInitStatus, getDefaultSettings, testConnection, initDatabase, initAdmin, skipAdmin, checkAdminExists } from '@/api/setup'
import { useI18n } from 'vue-i18n'
import { ChatDotRound, Loading } from '@element-plus/icons-vue'

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
const dbFormRef = ref<FormInstance>()
const adminFormRef = ref<FormInstance>()
const testing = ref(false)
const dbInitializing = ref(false)
const adminInitializing = ref(false)
const skipLoading = ref(false)
const currentStep = ref(0)
const adminExists = ref(false)
const showResetForm = ref(false)
const isDocker = ref(false)
const pageLoading = ref(true)
const existingAdminName = ref('')

const setupForm = reactive<SetInfo>({
  dbType: 'sqlite',
  dbFile: 'data.db',
  dbHost: '127.0.0.1',
  dbPort: 3306,
  dbName: '',
  dbUser: '',
  dbPass: '',
  dbCharset: 'utf8mb4',
  webPort: 12566,
  apiPort: 12567,
  adminUser: '',
  adminPass: ''
})

const adminForm = reactive({
  adminUser: '',
  adminPass: '',
  confirmPassword: ''
})

const dbRules = reactive<FormRules<SetInfo>>({
  dbHost: [
    {
      required: setupForm.dbType === 'mysql',
      message: t('setup.mysqlHost'),
      trigger: 'blur'
    }
  ],
  dbName: [
    {
      required: setupForm.dbType === 'mysql',
      message: t('setup.mysqlDatabase'),
      trigger: 'blur'
    }
  ]
})

const adminRules = computed<FormRules>(() => ({
  adminUser: adminExists.value ? [] : [{ required: true, message: () => t('setup.adminUsernameRequired'), trigger: 'blur' }],
  adminPass: [{ required: true, message: () => t('setup.adminPasswordRequired'), trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: () => t('setup.confirmPasswordRequired'), trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: (error?: Error) => void) => {
        if (value !== adminForm.adminPass) {
          callback(new Error(t('setup.passwordsNotMatch')))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}))

onMounted(async () => {
  try {
    const res = await checkInitStatus()
    if (res.data?.initialized) {
      ElMessage.info(t('setup.alreadyInitialized'))
      router.push('/login')
      return
    }
    isDocker.value = res.data?.isDocker ?? false
    // If database is already initialized, go to step 2
    if (res.data?.dbInitialized) {
      currentStep.value = 1
      await checkExistingAdmin()
    }
  } catch (e) {
    console.error(e)
  }

  // Load default settings
  try {
    const res = await getDefaultSettings()
    if (res.code === 0 || res.code === 200) {
      Object.assign(setupForm, res.data)
    }
  } catch (e) {
    console.error(e)
  }

  pageLoading.value = false
})

const checkExistingAdmin = async () => {
  try {
    const res = await checkAdminExists()
    if (res.code === 0 || res.code === 200) {
      adminExists.value = res.data?.hasAdmin ?? false
      existingAdminName.value = res.data?.adminName ?? ''
      adminForm.adminUser = existingAdminName.value
    }
  } catch (e) {
    console.error(e)
  }
}

const handleTestConnection = async () => {
  testing.value = true
  try {
    await testConnection(setupForm)
    ElMessage.success(t('setup.connectionSuccess'))
  } catch (e) {
    ElMessage.error(t('setup.connectionFailed'))
  } finally {
    testing.value = false
  }
}

const handleDbInit = async () => {
  const valid = await dbFormRef.value?.validate().catch(() => false)
  if (!valid) return

  dbInitializing.value = true
  try {
    await initDatabase(setupForm)
    ElMessage.success(t('setup.dbInitSuccess'))
    currentStep.value = 1
    await checkExistingAdmin()
  } catch (e) {
    console.error(e)
  } finally {
    dbInitializing.value = false
  }
}

const handleAdminInit = async () => {
  const valid = await adminFormRef.value?.validate().catch(() => false)
  if (!valid) return

  if (adminForm.adminPass !== adminForm.confirmPassword) {
    ElMessage.error(t('setup.passwordsNotMatch'))
    return
  }

  adminInitializing.value = true
  try {
    await initAdmin(adminForm.adminUser, adminForm.adminPass)
    ElMessage.success(t('setup.initializationSuccess'))
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  } catch (e) {
    console.error(e)
  } finally {
    adminInitializing.value = false
  }
}

const handleSkipAdmin = async () => {
  skipLoading.value = true
  try {
    await skipAdmin()
    ElMessage.success(t('setup.initializationSuccess'))
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  } catch (e) {
    ElMessage.error(t('setup.skipFailed'))
    console.error(e)
  } finally {
    skipLoading.value = false
  }
}
</script>

<style scoped lang="scss">
.setup-container {
  width: 100%;
  min-height: 100vh;
  padding: 40px 20px;
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

.page-loading {
  display: flex;
  justify-content: center;
  padding: 40px 0;

  .loading-icon {
    animation: spin 1.5s linear infinite;
    color: #909399;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
}

.setup-steps {
  margin-bottom: 24px;
}

.admin-exists-alert {
  margin-bottom: 20px;
}

.admin-name-display {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.setup-card {
  width: 600px;
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
}

@media (max-width: 768px) {
  .setup-card {
    width: 100%;
  }
}
</style>
