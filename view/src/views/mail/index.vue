<template>
  <div class="mail-container app-container">
    <div class="toolbar">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        {{ t('mail.addRecipient') }}
      </el-button>
    </div>

    <div class="table-container table-card-responsive">
      <el-table
        :data="tableData"
        border
        striped
        v-loading="loading"
      >
        <el-table-column prop="name" :label="t('mail.recipientName')" />
        <el-table-column prop="email" :label="t('mail.emailAddress')" />
        <el-table-column prop="createTime" :label="t('common.createTime')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.operations')" width="140" class="full-width-on-mobile">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">
              {{ t('common.edit') }}
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="loadData"
      @current-change="loadData"
    />

    <!-- Add/Edit Dialog -->
    <MailFormDialog
      v-model:open="formDialogVisible"
      :editing="editingItem"
      @success="handleDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { formatTime } from '@/utils/time'

const { t } = useI18n()
import { getMails, createMail, updateMail, deleteMail } from '@/api/mail'
import MailFormDialog from '@/components/MailFormDialog.vue'

const loading = ref(false)
const tableData = ref<MailConfig[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const formDialogVisible = ref(false)
const editingItem = ref<MailConfig | null>(null)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getMails(page.value, pageSize.value)
    if (res.code === 0 || res.code === 200) {
      tableData.value = res.data.list
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  editingItem.value = null
  formDialogVisible.value = true
}

const handleEdit = (row: MailConfig) => {
  editingItem.value = row
  formDialogVisible.value = true
}

const handleDialogSuccess = () => {
  formDialogVisible.value = false
  loadData()
}

const handleDelete = (row: MailConfig) => {
  ElMessageBox.confirm(
    t('mail.deleteConfirm'),
    t('common.confirm'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }
  ).then(async () => {
    await deleteMail(row.id)
    ElMessage.success(t('common.success'))
    loadData()
  }).catch(() => {})
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.mail-container {
  background: #f0f2f5;
}
</style>
