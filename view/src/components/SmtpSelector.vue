<template>
  <div class="smtp-selector">
    <div class="selector-trigger" @click="dialogVisible = true">
      <template v-if="selectedSmtp">
        <el-tag size="small" closable @close.stop="handleClear" type="success" class="selector-tag">
          {{ selectedSmtp.name }} ({{ selectedSmtp.mail }})
        </el-tag>
      </template>
      <span v-else class="selector-placeholder">{{ placeholder }}</span>
      <el-button size="small" class="add-btn">
        <el-icon><Plus /></el-icon>
        {{ selectedSmtp ? '更换' : '选择' }}
      </el-button>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="title"
      width="600px"
    >
      <div v-if="authStore.getIsAdmin" style="margin-bottom: 12px;">
        <el-checkbox v-model="adminOnly" :label="t('smtp.adminCreated')" @change="loadSmtpList" />
      </div>

      <el-table
        :data="paginatedList"
        border
        stripe
        v-loading="loading"
        highlight-current-row
        :row-key="row => row.id"
        max-height="400"
      >
        <el-table-column label="" width="50" align="center">
          <template #default="{ row }">
            <el-radio v-model="tempSelectedId" :value="row.id">
              &nbsp;
            </el-radio>
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="t('smtp.smtpName')" />
        <el-table-column prop="host" :label="t('smtp.host')" />
        <el-table-column prop="mail" :label="t('smtp.fromAddress')" width="160" />
        <el-table-column v-if="authStore.getIsAdmin" prop="userName" :label="t('common.creator')" width="100" />
      </el-table>

      <div class="pagination-wrapper">
        <div class="page-info">
          {{ (currentPage - 1) * pageSize + 1 }}-{{ Math.min(currentPage * pageSize, smtpList.length) }} / 共 {{ smtpList.length }} 条
        </div>
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="smtpList.length"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" @click="handleConfirm" :disabled="!tempSelectedId">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { getSMTPServers } from '@/api/smtp'
import { useAuthStore } from '@/store/auth'

const { t } = useI18n()
const authStore = useAuthStore()

interface Props {
  modelValue: number
  title?: string
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '选择 SMTP 服务器',
  placeholder: '请选择 SMTP 服务器'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
}>()

const dialogVisible = ref(false)
const loading = ref(false)
const smtpList = ref<SMTPConfig[]>([])
const selectedId = ref<number>(0)
const tempSelectedId = ref<number>(0)
const currentPage = ref(1)
const pageSize = 10
const adminOnly = ref(false)

const paginatedList = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return smtpList.value.slice(start, start + pageSize)
})

const selectedSmtp = computed(() => {
  return smtpList.value.find(s => s.id === selectedId.value) || null
})

const loadSmtpList = async () => {
  loading.value = true
  try {
    const filters: { adminOnly?: boolean } = {}
    if (adminOnly.value) filters.adminOnly = true
    const res = await getSMTPServers(1, 1000, filters)
    if (res.code === 0 || res.code === 200) {
      smtpList.value = res.data.list
    }
  } finally {
    loading.value = false
  }
}

const handleClear = () => {
  emit('update:modelValue', 0)
}

const handleConfirm = () => {
  if (tempSelectedId.value) {
    emit('update:modelValue', tempSelectedId.value)
    dialogVisible.value = false
  }
}

const handlePageChange = () => {
  // Radio selection works across pages, no special handling needed
}

watch(() => props.modelValue, (val) => {
  selectedId.value = val
  tempSelectedId.value = val
})

watch(dialogVisible, (val) => {
  if (val) {
    currentPage.value = 1
    loadSmtpList()
  }
})

onMounted(() => {
  selectedId.value = props.modelValue
  tempSelectedId.value = props.modelValue
})
</script>

<style scoped>
.smtp-selector {
  width: 100%;
}
.selector-trigger {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  min-height: 32px;
  padding: 4px 8px;
  cursor: pointer;
  background-color: var(--el-bg-color);
}
.selector-placeholder {
  color: var(--el-text-color-placeholder);
  font-size: 14px;
}
.selector-tag {
  margin: 0;
}
.add-btn {
  flex-shrink: 0;
}
.pagination-wrapper {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.page-info {
  font-size: 13px;
  color: var(--el-text-color-regular);
}
</style>
