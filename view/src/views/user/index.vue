<template>
  <div class="user-container app-container">
    <div class="toolbar">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        {{ t('user.addUser') }}
      </el-button>
    </div>

    <div class="table-container table-card-responsive">
      <el-table
        :data="tableData"
        border
        stripe
        v-loading="loading"
      >
        <el-table-column prop="name" :label="t('user.username')" />
        <el-table-column prop="isAdmin" :label="t('user.isAdmin')" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isAdmin ? 'danger' : 'info'" size="small">
              {{ row.isAdmin ? t('user.admin') : t('user.normalUser') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="isUse" :label="t('common.status')" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isUse ? 'success' : 'warning'" size="small">
              {{ row.isUse ? t('common.enable') : t('common.disable') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" :label="t('common.createTime')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.operations')" width="240">
          <template #default="{ row }">
            <el-button size="small" @click="handleChangePassword(row)">
              {{ t('user.changePassword') }}
            </el-button>
            <el-button size="small" @click="handleEdit(row)" :disabled="row.isAdmin">
              {{ t('common.edit') }}
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" :disabled="row.isAdmin">
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
    <UserFormDialog
      v-model:open="formDialogVisible"
      :editing="editingItem"
      @success="handleDialogSuccess"
    />

    <!-- Change Password Dialog -->
    <ChangePasswordDialog
      v-model:open="passwordDialogVisible"
      :user="passwordUser"
      @success="handlePasswordSuccess"
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
import { getUsers, deleteUser } from '@/api/user'
import UserFormDialog from '@/components/UserFormDialog.vue'
import ChangePasswordDialog from '@/components/ChangePasswordDialog.vue'

const loading = ref(false)
const tableData = ref<UserConfig[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const formDialogVisible = ref(false)
const passwordDialogVisible = ref(false)
const editingItem = ref<UserConfig | null>(null)
const passwordUser = ref<UserConfig | null>(null)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getUsers(page.value, pageSize.value)
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

const handleEdit = (row: UserConfig) => {
  if (row.isAdmin) return
  editingItem.value = row
  formDialogVisible.value = true
}

const handleDialogSuccess = () => {
  formDialogVisible.value = false
  loadData()
}

const handleChangePassword = (row: UserConfig) => {
  passwordUser.value = row
  passwordDialogVisible.value = true
}

const handlePasswordSuccess = () => {
  passwordDialogVisible.value = false
}

const handleDelete = (row: UserConfig) => {
  if (row.isAdmin) return
  ElMessageBox.confirm(
    t('user.deleteConfirm'),
    t('common.confirm'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }
  ).then(async () => {
    await deleteUser(row.id)
    ElMessage.success(t('common.success'))
    loadData()
  }).catch(() => {})
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.user-container {
  // uses global app-container styles
}
</style>
