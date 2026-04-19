<template>
  <div class="recipient-selector">
    <div class="selector-trigger" @click="dialogVisible = true">
      <template v-if="selectedRecipients.length > 0">
        <el-tag
          v-for="item in selectedRecipients"
          :key="item.id"
          closable
          @close.stop="handleRemove(item.id)"
          size="small"
          class="selector-tag"
        >
          {{ item.name }} ({{ item.mail }})
        </el-tag>
      </template>
      <span v-else class="selector-placeholder">{{ placeholder }}</span>
      <el-button size="small" class="add-btn">
        <el-icon><Plus /></el-icon>
        添加
      </el-button>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="title"
      width="600px"
    >
      <div v-if="selectedRecipients.length > 0" class="dialog-selected-tags">
        <el-tag
          v-for="item in selectedRecipients"
          :key="item.id"
          closable
          @close="handleRemove(item.id)"
          size="small"
          type="info"
        >
          {{ item.name }} ({{ item.mail }})
        </el-tag>
      </div>
      <el-table
        ref="tableRef"
        :data="paginatedList"
        border
        stripe
        v-loading="loading"
        @selection-change="handleSelectionChange"
        :row-key="row => row.id"
        max-height="400"
      >
        <el-table-column type="selection" width="50" align="center" reserve-selection />
        <el-table-column prop="name" :label="t('mail.recipientName')" />
        <el-table-column prop="mail" :label="t('mail.emailAddress')" />
      </el-table>

      <div class="pagination-wrapper">
        <div class="page-info">
          {{ (currentPage - 1) * pageSize + 1 }}-{{ Math.min(currentPage * pageSize, mailList.length) }} / 共 {{ mailList.length }} 条
        </div>
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="mailList.length"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" @click="handleConfirm" :disabled="selectedIds.length === 0">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { getMails } from '@/api/mail'

const { t } = useI18n()

interface Recipient {
  id: number
  name: string
  mail: string
}

interface Props {
  modelValue: number[]
  title?: string
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '选择收件人',
  placeholder: '请选择收件人'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: number[]): void
}>()

const dialogVisible = ref(false)
const loading = ref(false)
const mailList = ref<MailConfig[]>([])
const tableRef = ref()
const currentPage = ref(1)
const pageSize = 10

const paginatedList = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return mailList.value.slice(start, start + pageSize)
})

const selectedRecipients = ref<Recipient[]>([])
const selectedIds = ref<number[]>([])

const loadMailList = async () => {
  loading.value = true
  try {
    const res = await getMails(1, 1000)
    if (res.code === 0 || res.code === 200) {
      mailList.value = res.data.list
      await syncTableSelection()
    }
  } finally {
    loading.value = false
  }
}

const handleSelectionChange = (selection: MailConfig[]) => {
  // Merge: keep non-page items, add page selections
  const pageIds = new Set(paginatedList.value.map(r => r.id))
  const keep = selectedRecipients.value.filter(r => !pageIds.has(r.id))
  selectedRecipients.value = [...keep, ...selection.map(s => ({ id: s.id, name: s.name, mail: s.mail }))]
  selectedIds.value = selectedRecipients.value.map(r => r.id)
  emit('update:modelValue', selectedIds.value)
}

const handleRemove = (id: number) => {
  selectedIds.value = selectedIds.value.filter(v => v !== id)
  selectedRecipients.value = selectedRecipients.value.filter(r => r.id !== id)
  if (tableRef.value) {
    const row = mailList.value.find(m => m.id === id)
    if (row) {
      tableRef.value.toggleRowSelection(row, false)
    }
  }
  emit('update:modelValue', selectedIds.value)
}

const handlePageChange = async () => {
  await nextTick()
  if (tableRef.value) {
    const idSet = new Set(selectedIds.value)
    tableRef.value.clearSelection()
    paginatedList.value.forEach(row => {
      if (idSet.has(row.id)) {
        tableRef.value.toggleRowSelection(row, true)
      }
    })
  }
}

const handleConfirm = () => {
  emit('update:modelValue', selectedIds.value)
  dialogVisible.value = false
}

const syncTableSelection = async () => {
  await nextTick()
  if (tableRef.value && selectedIds.value.length > 0) {
    const idSet = new Set(selectedIds.value)
    tableRef.value.clearSelection()
    paginatedList.value.forEach(row => {
      if (idSet.has(row.id)) {
        tableRef.value.toggleRowSelection(row, true)
      }
    })
  }
}

watch(dialogVisible, async (val) => {
  if (val) {
    currentPage.value = 1
    if (mailList.value.length === 0) {
      await loadMailList()
    } else {
      await syncTableSelection()
    }
  }
})

watch(() => props.modelValue, (val) => {
  selectedIds.value = [...val]
  if (mailList.value.length > 0) {
    selectedRecipients.value = mailList.value
      .filter(m => val.includes(m.id))
      .map(m => ({ id: m.id, name: m.name, mail: m.mail }))
  }
})
</script>

<style scoped>
.recipient-selector {
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
.dialog-selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 8px 0 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  margin-bottom: 12px;
}
.dialog-selected-tags .el-tag {
  margin: 0;
}
</style>
