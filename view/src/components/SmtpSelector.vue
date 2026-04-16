<template>
  <div>
    <el-input
      :model-value="selectedName"
      :placeholder="placeholder"
      readonly
      @click="dialogVisible = true"
      class="smtp-selector-input"
    >
      <template #suffix>
        <el-icon class="el-icon--right"><Search /></el-icon>
      </template>
    </el-input>

    <el-dialog
      v-model="dialogVisible"
      :title="title"
      width="600px"
    >
      <el-table
        :data="smtpList"
        border
        stripe
        v-loading="loading"
        highlight-current-row
        @current-change="handleSelect"
      >
        <el-table-column label="" width="50" align="center">
          <template #default="{ row }">
            <el-radio :model-value="selectedId" :value="row.id" @change="handleSelect(row)">
              &nbsp;
            </el-radio>
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="t('smtp.smtpName')" />
        <el-table-column prop="host" :label="t('smtp.host')" />
        <el-table-column prop="mail" :label="t('smtp.fromAddress')" width="160" />
      </el-table>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" @click="handleConfirm" :disabled="!selectedId">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { getSMTPServers } from '@/api/smtp'

const { t } = useI18n()

interface Props {
  modelValue: number
  title?: string
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '选择 SMTP 服务器',
  placeholder: '请选择 SMTP 服务器'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
}>()

const dialogVisible = ref(false)
const loading = ref(false)
const smtpList = ref<SMTPConfig[]>([])
const selectedId = ref<number>(0)
const tempSelectedId = ref<number>(0)

const selectedName = computed(() => {
  const item = smtpList.value.find(s => s.id === selectedId.value)
  return item ? item.name : ''
})

const loadSmtpList = async () => {
  loading.value = true
  try {
    const res = await getSMTPServers(1, 1000)
    if (res.code === 0 || res.code === 200) {
      smtpList.value = res.data.list
    }
  } finally {
    loading.value = false
  }
}

const handleSelect = (row: SMTPConfig) => {
  tempSelectedId.value = row.id
}

const handleConfirm = () => {
  emit('update:modelValue', tempSelectedId.value)
  dialogVisible.value = false
}

watch(() => props.modelValue, (val) => {
  selectedId.value = val
  tempSelectedId.value = val
})

watch(dialogVisible, (val) => {
  if (val) {
    loadSmtpList()
  }
})

onMounted(() => {
  selectedId.value = props.modelValue
  tempSelectedId.value = props.modelValue
})
</script>

<style scoped>
.smtp-selector-input {
  cursor: pointer;
}
</style>
