<template>
  <div class="schedule-container app-container">
    <div class="toolbar">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        {{ t('schedule.addSchedule') }}
      </el-button>
      <div class="filter-bar">
        <el-input v-model="filterName" :placeholder="t('schedule.taskName')" :prefix-icon="Search" clearable size="default" style="width: 170px" @clear="loadData" @keyup.enter="loadData" />
        <el-checkbox v-if="authStore.getIsAdmin" v-model="adminOnly" :label="t('schedule.adminCreated')" @change="loadData" />
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
        <el-table-column prop="name" :label="t('schedule.taskName')" />
        <el-table-column prop="cron" :label="t('schedule.cronExpression')" width="140" />
        <el-table-column prop="isUse" :label="t('common.status')" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isUse ? 'success' : 'danger'">
              {{ row.isUse ? t('schedule.enabled') : t('schedule.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="authStore.getIsAdmin" prop="userName" :label="t('common.creator')" width="120" />
        <el-table-column prop="createTime" :label="t('common.createTime')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.operations')" width="280" class="full-width-on-mobile">
          <template #default="{ row }">
            <el-button size="small" @click="handleTrigger(row)">
              {{ t('schedule.triggerNow') }}
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
    <ScheduleFormDialog
      v-model:open="formDialogVisible"
      :editing="editingItem"
      @success="handleDialogSuccess"
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
import { getSchedules, createSchedule, updateSchedule, deleteSchedule, triggerSendMail } from '@/api/schedule'
import ScheduleFormDialog from '@/components/ScheduleFormDialog.vue'

const loading = ref(false)
const tableData = ref<ScheduleConfig[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const formDialogVisible = ref(false)
const editingItem = ref<ScheduleConfig | null>(null)
const filterName = ref('')
const adminOnly = ref(false)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getSchedules(page.value, pageSize.value, {
      name: filterName.value,
      adminOnly: adminOnly.value
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
  adminOnly.value = false
  page.value = 1
  loadData()
}

const handleAdd = () => {
  editingItem.value = null
  formDialogVisible.value = true
}

const handleEdit = (row: ScheduleConfig) => {
  editingItem.value = row
  formDialogVisible.value = true
}

const handleDialogSuccess = () => {
  formDialogVisible.value = false
  loadData()
}

const handleTrigger = async (row: ScheduleConfig) => {
  try {
    await triggerSendMail(row.id)
    ElMessage.success(t('schedule.triggered'))
  } catch (e) {
    console.error(e)
  }
}

const handleDelete = (row: ScheduleConfig) => {
  ElMessageBox.confirm(
    t('schedule.deleteConfirm'),
    t('common.confirm'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }
  ).then(async () => {
    await deleteSchedule(row.id)
    ElMessage.success(t('common.success'))
    loadData()
  }).catch(() => {})
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.schedule-container {
  // uses global app-container styles
}
</style>
