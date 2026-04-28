<template>
  <div class="smtp-container app-container">
    <div class="toolbar">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        {{ t('smtp.addSMTP') }}
      </el-button>
      <div class="filter-bar">
        <el-input v-model="filterName" :placeholder="t('smtp.smtpName')" :prefix-icon="Search" clearable size="default" style="width: 150px" @clear="loadData" @keyup.enter="loadData" />
        <el-input v-model="filterHost" :placeholder="t('smtp.host')" clearable size="default" style="width: 150px" @clear="loadData" @keyup.enter="loadData" />
        <el-input v-model="filterUsername" :placeholder="t('smtp.username')" clearable size="default" style="width: 150px" @clear="loadData" @keyup.enter="loadData" />
        <el-checkbox v-if="authStore.getIsAdmin" v-model="adminOnly" :label="t('smtp.adminCreated')" @change="loadData" />
        <el-button size="default" type="primary" plain @click="loadData">{{ t('common.search') }}</el-button>
        <el-button size="default" @click="resetFilters">{{ t('common.reset') }}</el-button>
      </div>
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
        <el-table-column prop="mail" :label="t('smtp.fromAddress')" />
        <el-table-column prop="username" :label="t('smtp.username')" />
        <el-table-column v-if="authStore.getIsAdmin" prop="userName" :label="t('common.creator')" width="120" />
        <el-table-column prop="createTime" :label="t('common.createTime')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.operations')" width="260" class="full-width-on-mobile">
          <template #default="{ row }">
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
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { formatTime } from '@/utils/time'
import { useAuthStore } from '@/store/auth'

const { t } = useI18n()
const authStore = useAuthStore()
import { getSMTPServers, createSMTP, updateSMTP, deleteSMTP } from '@/api/smtp'
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
const filterName = ref('')
const filterHost = ref('')
const filterUsername = ref('')
const adminOnly = ref(false)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getSMTPServers(page.value, pageSize.value, {
      adminOnly: adminOnly.value,
      name: filterName.value,
      host: filterHost.value,
      username: filterUsername.value
    })
    if (res.code === 0 || res.code === 200) {
      tableData.value = res.data.list
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filterName.value = ''
  filterHost.value = ''
  filterUsername.value = ''
  adminOnly.value = false
  page.value = 1
  loadData()
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
  // uses global app-container styles
}
</style>
