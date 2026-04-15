<template>
  <div class="schedule-container app-container">
    <div class="toolbar">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        {{ t('schedule.addSchedule') }}
      </el-button>
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
        <el-table-column prop="createTime" :label="t('common.createTime')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.operations')" width="200" class="full-width-on-mobile">
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
import { Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { formatTime } from '@/utils/time'

const { t } = useI18n()
import { getSchedules, createSchedule, updateSchedule, deleteSchedule, triggerSendMail } from '@/api/schedule'
import ScheduleFormDialog from '@/components/ScheduleFormDialog.vue'

const loading = ref(false)
const tableData = ref<ScheduleConfig[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const formDialogVisible = ref(false)
const editingItem = ref<ScheduleConfig | null>(null)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getSchedules(page.value, pageSize.value)
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
  background: #f0f2f5;
}
</style>
