<template>
  <div class="setup-container">
    <el-card class="setup-card">
      <h1 class="title">{{ t('setup.systemSetup') }}</h1>
      <p class="subtitle">{{ t('setup.setupWizard') }}</p>

      <el-form
        ref="setupFormRef"
        :model="setupForm"
        :rules="rules"
        label-width="140px"
      >
        <!-- Database Configuration -->
        <el-divider content-position="left">{{ t('setup.databaseConfig') }}</el-divider>

        <el-form-item :label="t('setup.databaseType')" prop="dbType">
          <el-radio-group v-model="setupForm.dbType">
            <el-radio label="sqlite">{{ t('setup.sqlite') }}</el-radio>
            <el-radio label="mysql">{{ t('setup.mysql') }}</el-radio>
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
            <el-input v-model="setupForm.dbPass" type="password" />
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

        <!-- Port Configuration -->
        <el-divider content-position="left">{{ t('setup.portConfig') }}</el-divider>

        <el-form-item :label="t('setup.webPort')" prop="webPort">
          <el-input-number v-model="setupForm.webPort" :min="1" :max="65535" />
        </el-form-item>

        <el-form-item :label="t('setup.apiPort')" prop="apiPort">
          <el-input-number v-model="setupForm.apiPort" :min="1" :max="65535" />
        </el-form-item>

        <!-- Admin Account -->
        <el-divider content-position="left">{{ t('setup.adminAccount') }}</el-divider>

        <el-form-item :label="t('setup.adminUsername')" prop="adminUser">
          <el-input v-model="setupForm.adminUser" />
        </el-form-item>

        <el-form-item :label="t('setup.adminPassword')" prop="adminPass">
          <el-input v-model="setupForm.adminPass" type="password" />
        </el-form-item>

        <el-form-item :label="t('setup.confirmPassword')" prop="confirmPassword">
          <el-input v-model="confirmPassword" type="password" />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="initializing"
            @click="handleInitialize"
          >
            {{ t('setup.initialize') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { checkInitStatus, getDefaultSettings, testConnection, initializeSystem } from '@/api/setup'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const router = useRouter()
const setupFormRef = ref<FormInstance>()
const testing = ref(false)
const initializing = ref(false)
const confirmPassword = ref('')

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

const rules = reactive<FormRules<SetInfo>>({
  adminUser: [{ required: true, message: t('setup.adminUsername'), trigger: 'blur' }],
  adminPass: [{ required: true, message: t('setup.adminPassword'), trigger: 'blur' }],
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

onMounted(async () => {
  try {
    const res = await checkInitStatus()
    if (res.data?.initialized) {
      ElMessage.info(t('setup.alreadyInitialized'))
      router.push('/login')
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
})

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

const handleInitialize = async () => {
  if (setupForm.adminPass !== confirmPassword.value) {
    ElMessage.error(t('setup.passwordsNotMatch'))
    return
  }

  initializing.value = true
  try {
    await initializeSystem(setupForm)
    ElMessage.success(t('setup.initializationSuccess'))
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  } catch (e) {
    console.error(e)
  } finally {
    initializing.value = false
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
