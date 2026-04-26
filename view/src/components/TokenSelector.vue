<template>
  <div class="token-selector">
    <div class="selector-trigger" @click="dialogVisible = true">
      <template v-if="selectedToken">
        <el-tag size="small" closable @close.stop="handleClear" :type="selectedToken.state === 0 ? 'success' : 'info'" class="selector-tag">
          {{ selectedToken.name }}
        </el-tag>
      </template>
      <span v-else class="selector-placeholder">{{ placeholder }}</span>
      <el-button size="small" class="add-btn">
        <el-icon><Plus /></el-icon>
        {{ selectedToken ? '更换' : '选择' }}
      </el-button>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="title"
      width="600px"
    >
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
        <el-table-column prop="name" :label="t('token.tokenName')" />
        <el-table-column prop="token" label="Token" width="200">
          <template #default="{ row }">
            <span class="token-text">{{ row.token ? row.token.substring(0, 8) + '...' : '' }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.state === 0 ? 'success' : 'info'" size="small">
              {{ row.state === 0 ? t('token.inUse') : row.state === 1 ? t('token.userDisabled') : t('token.adminDisabled') }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <div class="page-info">
          {{ (currentPage - 1) * pageSize + 1 }}-{{ Math.min(currentPage * pageSize, tokenList.length) }} / 共 {{ tokenList.length }} 条
        </div>
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="tokenList.length"
          layout="prev, pager, next"
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
import { getTokens } from '@/api/token'

const { t } = useI18n()

interface Props {
  modelValue: number
  title?: string
  placeholder?: string
  onlyActive?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '选择 API 令牌',
  placeholder: '请选择 API 令牌',
  onlyActive: true
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
}>()

const dialogVisible = ref(false)
const loading = ref(false)
const tokenList = ref<TokenConfig[]>([])
const selectedId = ref<number>(0)
const tempSelectedId = ref<number>(0)
const currentPage = ref(1)
const pageSize = 10

const paginatedList = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return tokenList.value.slice(start, start + pageSize)
})

const selectedToken = computed(() => {
  return tokenList.value.find(t => t.id === selectedId.value) || null
})

const loadTokenList = async () => {
  loading.value = true
  try {
    const res = await getTokens(1, 1000)
    if (res.code === 0 || res.code === 200) {
      tokenList.value = props.onlyActive
        ? res.data.list.filter((t: TokenConfig) => t.state === 0)
        : res.data.list
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

watch(() => props.modelValue, (val) => {
  selectedId.value = val
  tempSelectedId.value = val
})

watch(dialogVisible, (val) => {
  if (val) {
    currentPage.value = 1
    loadTokenList()
  }
})

onMounted(() => {
  selectedId.value = props.modelValue
  tempSelectedId.value = props.modelValue
})
</script>

<style scoped>
.token-selector {
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
.token-text {
  font-family: monospace;
  font-size: 13px;
}
</style>
