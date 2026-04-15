<template>
  <el-dialog
    v-model="open"
    :title="editing ? t('schedule.editSchedule') : t('schedule.addSchedule')"
    width="600px"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
    >
      <el-form-item :label="t('schedule.taskName')" prop="name">
        <el-input v-model="form.name" :placeholder="t('schedule.taskName')" />
      </el-form-item>
      <el-form-item :label="t('schedule.associatedToken')" prop="tokenId">
        <el-select v-model="form.tokenId" :placeholder="t('schedule.associatedToken')" filterable clearable>
          <el-option
            v-for="item in tokenList"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('schedule.cronExpression')" prop="cron">
        <el-input v-model="form.cron" placeholder="* * * * *" />
        <div class="help-text">
          {{ t('schedule.cronHelp') }}<br>
          {{ t('schedule.exampleCron') }}
        </div>
      </el-form-item>
      <el-form-item :label="t('schedule.requestUrl')" prop="url">
        <el-input v-model="form.url" placeholder="https://example.com/api/data" />
      </el-form-item>
      <el-form-item :label="t('schedule.requestMethod')" prop="method">
        <el-radio-group v-model="form.method">
          <el-radio value="GET">GET</el-radio>
          <el-radio value="POST">POST</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item :label="t('schedule.requestHeaders')" prop="headers">
        <el-input
          v-model="form.headers"
          type="textarea"
          rows="3"
          placeholder='{"Authorization": "Bearer token"}'
        />
      </el-form-item>
      <el-form-item :label="t('schedule.requestBody')" prop="body">
        <el-input
          v-model="form.body"
          type="textarea"
          rows="4"
          placeholder="{}"
        />
      </el-form-item>
      <el-form-item :label="t('schedule.useTemplate')" prop="useTemplate">
        <el-switch v-model="form.useTemplate" />
      </el-form-item>
      <el-form-item :label="t('common.status')" prop="isUse">
        <el-switch v-model="form.isUse" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="open = false">
        {{ t('common.cancel') }}
      </el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        {{ t('common.save') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
import { createSchedule, updateSchedule } from '@/api/schedule'
import { getTokens } from '@/api/token'
import { ElMessage } from 'element-plus'

interface Props {
  open: boolean
  editing: ScheduleConfig | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const tokenList = ref<TokenConfig[]>([])

const defaultForm: Partial<ScheduleConfig> = {
  name: '',
  tokenId: 0,
  cron: '',
  url: '',
  method: 'GET',
  headers: '',
  body: '',
  useTemplate: false,
  isUse: true
}

const form = ref<Partial<ScheduleConfig>>({ ...defaultForm })

const rules = computed<FormRules<Partial<ScheduleConfig>>>(() => ({
  name: [{ required: true, message: t('schedule.taskName'), trigger: 'blur' }],
  tokenId: [{ required: true, message: t('schedule.associatedToken'), trigger: 'change' }],
  cron: [{ required: true, message: t('schedule.cronExpression'), trigger: 'blur' }],
  url: [{ required: true, message: t('schedule.requestUrl'), trigger: 'blur' }]
}))

const loadOptions = async () => {
  const tokenRes = await getTokens(1, 1000)
  if (tokenRes.code === 0 || tokenRes.code === 200) {
    tokenList.value = tokenRes.data.list.filter(t => t.isUse)
  }
}

onMounted(() => {
  if (props.open) {
    loadOptions()
  }
})

watch(() => props.editing, () => {
  if (props.editing) {
    form.value = { ...props.editing }
  } else {
    form.value = { ...defaultForm }
  }
}, { immediate: true })

watch(() => props.open, (newVal) => {
  if (newVal) {
    loadOptions()
    if (!props.editing) {
      form.value = { ...defaultForm }
    }
  }
})

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (props.editing) {
          await updateSchedule(form.value as ScheduleConfig)
        } else {
          await createSchedule(form.value)
        }
        ElMessage.success(t('common.success'))
        emit('success')
        emit('update:open', false)
      } finally {
        submitting.value = false
      }
    }
  })
}

const open = computed({
  get: () => props.open,
  set: (val) => emit('update:open', val)
})
</script>

<style scoped>
.help-text {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.5;
}
</style>
