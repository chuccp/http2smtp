<template>
  <div class="token-container app-container">
    <div class="toolbar">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        {{ t('token.generateToken') }}
      </el-button>
    </div>

    <div class="table-container table-card-responsive">
      <el-table
        :data="tableData"
        border
        striped
        v-loading="loading"
      >
        <el-table-column prop="name" label="{{ t('token.tokenName') }}" />
        <el-table-column prop="token" label="{{ t('token.tokenValue') }}" min-width="120">
          <template #default="{ row }">
            <div class="token-cell">
              <span>{{ row.token.slice(0, 8) }}...{{ row.token.slice(-8) }}</span>
              <el-button size="small" link @click="copyToken(row.token)">
                <el-icon><CopyDocument /></el-icon>
                {{ t('token.copyToken') }}
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="enable" label="{{ t('common.status') }}" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enable ? 'success' : 'danger'">
              {{ row.enable ? t('token.enabled') : t('token.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="{{ t('common.createTime') }}" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="{{ t('common.operations') }}" width="200" class="full-width-on-mobile">
          <template #default="{ row }">
            <el-button size="small" @click="handleSendTest(row)">
              {{ t('token.sendTestMail') }}
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
    <TokenFormDialog
      v-model:open="formDialogVisible"
      :editing="editingItem"
      @success="handleDialogSuccess"
    />

    <!-- Send Test Dialog -->
    <SendMailByTokenDialog
      v-model:open="sendDialogVisible"
      :token-id="currentTokenId"
      @success="handleSendSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, CopyDocument } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { formatTime } from '@/utils/time'

const { t } = useI18n()
import { getTokens, createToken, updateToken, deleteToken } from '@/api/token'
import TokenFormDialog from '@/components/TokenFormDialog.vue'
import SendMailByTokenDialog from '@/components/SendMailByTokenDialog.vue'

const loading = ref(false)
const tableData = ref<TokenConfig[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const formDialogVisible = ref(false)
const sendDialogVisible = ref(false)
const editingItem = ref<TokenConfig | null>(null)
const currentTokenId = ref(0)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getTokens(page.value, pageSize.value)
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

const handleEdit = (row: TokenConfig) => {
  editingItem.value = row
  formDialogVisible.value = true
}

const copyToken = async (token: string) => {
  await navigator.clipboard.writeText(token)
  ElMessage.success(t('common.copied'))
}

const handleDialogSuccess = () => {
  formDialogVisible.value = false
  loadData()
}

const handleDelete = (row: TokenConfig) => {
  ElMessageBox.confirm(
    t('token.deleteConfirm'),
    t('common.confirm'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }
  ).then(async () => {
    await deleteToken(row.id)
    ElMessage.success(t('common.success'))
    loadData()
  }).catch(() => {})
}

const handleSendTest = (row: TokenConfig) => {
  currentTokenId.value = row.id
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
.token-container {
  background: #f0f2f5;
}

.token-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
