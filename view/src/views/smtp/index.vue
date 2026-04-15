<template>
  <div class="smtp-container app-container">
    <div class="toolbar">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        {{ t('smtp.addSMTP') }}
      </el-button>
    </div>

    <div class="table-container table-card-responsive">
      <el-table
        :data="tableData"
        border
        striped
        v-loading="loading"
      >
        <el-table-column prop="name" :label="t('smtp.smtpName')" />
        <el-table-column prop="host" :label="t('smtp.host')" />
        <el-table-column prop="port" :label="t('smtp.port')" width="80" />
        <el-table-column prop="from" :label="t('smtp.fromAddress')" />
        <el-table-column prop="username" :label="t('smtp.username')" />
        <el-table-column prop="ssl" label="SSL" width="60" align="center">
          <template #default="{ row }">
            <el-tag :type="row.ssl ? 'success' : 'info'">
              {{ row.ssl ? t('common.yes') : t('common.no') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" :label="t('common.createTime')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.operations')" width="260" class="full-width-on-mobile">
          <template #default="{ row }">
            <el-button size="small" @click="handleTestConnection(row)">
              {{ t('smtp.testConnection') }}
            </el-button>
            <el-button size="small" @click="handleSendTest(row)">
              {{ t('smtp.sendTestMail') }}
            </el-button>
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
    <SMTPFormDialog
      v-model:open="formDialogVisible"
      :editing="editingItem"
      @success="handleDialogSuccess"
    />

    <!-- Send Test Dialog -->
    <SendMailDialog
      v-model:open="sendDialogVisible"
      :smtp-id="currentSMTPId"
      @success="handleSendSuccess"
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
import { getSMTPServers, createSMTP, updateSMTP, deleteSMTP, testSMTPConnection } from '@/api/smtp'
import SMTPFormDialog from '@/components/SMTPFormDialog.vue'
import SendMailDialog from '@/components/SendMailDialog.vue'

const loading = ref(false)
const tableData = ref<SMTPConfig[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const formDialogVisible = ref(false)
const sendDialogVisible = ref(false)
const editingItem = ref<SMTPConfig | null>(null)
const currentSMTPId = ref(0)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getSMTPServers(page.value, pageSize.value)
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

const handleEdit = (row: SMTPConfig) => {
  editingItem.value = row
  formDialogVisible.value = true
}

const handleDialogSuccess = () => {
  formDialogVisible.value = false
  loadData()
}

const handleDelete = (row: SMTPConfig) => {
  ElMessageBox.confirm(
    t('smtp.deleteConfirm'),
    t('common.confirm'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }
  ).then(async () => {
    await deleteSMTP(row.id)
    ElMessage.success(t('common.success'))
    loadData()
  }).catch(() => {})
}

const handleTestConnection = async (row: SMTPConfig) => {
  try {
    await testSMTPConnection(row)
    ElMessage.success(t('smtp.connectionSuccess'))
  } catch (e) {
    ElMessage.error(t('smtp.connectionFailed'))
  }
}

const handleSendTest = (row: SMTPConfig) => {
  currentSMTPId.value = row.id
  sendDialogVisible.value = true
}

const handleSendSuccess = () => {
  sendDialogVisible.value = false
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.smtp-container {
  background: #f0f2f5;
}
</style>
