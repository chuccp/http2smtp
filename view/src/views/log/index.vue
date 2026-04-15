<template>
  <div class="log-container app-container">
    <div class="search-bar">
      <el-form :inline="true" class="demo-form-inline">
        <el-form-item>
          <el-input
            v-model="searchKey"
            :placeholder="t('log.searchKeyword')"
            clearable
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="handleReset">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-container table-card-responsive">
      <el-table
        :data="tableData"
        border
        striped
        v-loading="loading"
        @row-click="handleRowClick"
      >
        <el-table-column prop="token" :label="t('log.token')" width="120">
          <template #default="{ row }">
            {{ row.token.slice(0, 12) }}...
          </template>
        </el-table-column>
        <el-table-column prop="subject" :label="t('log.subject')" min-width="150" />
        <el-table-column prop="status" :label="t('log.status')" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : row.status === 'error' ? 'danger' : 'warning'">
              {{ t('log.' + row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" :label="t('log.createTime')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="100" align="center" class="full-width-on-mobile">
          <template #default="{ row }">
            <el-button size="small" @click.stop="handleDetail(row)">
              {{ t('log.logDetail') }}
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

    <!-- Log Detail Dialog -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="t('log.logDetail')"
      width="600px"
    >
      <el-descriptions :column="1" border v-if="currentDetail">
        <el-descriptions-item :label="t('log.token')">
          {{ currentDetail.token }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('log.subject')">
          {{ currentDetail.subject }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('log.content')">
          <pre>{{ currentDetail.content }}</pre>
        </el-descriptions-item>
        <el-descriptions-item :label="t('log.result')">
          {{ currentDetail.result }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('log.status')">
          <el-tag :type="currentDetail.status === 'success' ? 'success' : 'danger'">
            {{ currentDetail.status }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('log.createTime')">
          {{ formatTime(currentDetail.createTime) }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatTime } from '@/utils/time'
import { getLogs } from '@/api/log'

const { t } = useI18n()

const loading = ref(false)
const tableData = ref<LogEntry[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchKey = ref('')
const detailDialogVisible = ref(false)
const currentDetail = ref<LogEntry | null>(null)

const loadData = async () => {
  loading.value = true
  try {
    const res = await getLogs(page.value, pageSize.value, searchKey.value)
    if (res.code === 0 || res.code === 200) {
      tableData.value = res.data.list
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadData()
}

const handleReset = () => {
  searchKey.value = ''
  page.value = 1
  loadData()
}

const handleRowClick = (row: LogEntry) => {
  handleDetail(row)
}

const handleDetail = (row: LogEntry) => {
  currentDetail.value = row
  detailDialogVisible.value = true
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.log-container {
  background: #f0f2f5;
}

pre {
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  font-family: inherit;
}
</style>
