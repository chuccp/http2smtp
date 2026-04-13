<template>
  <el-dialog
    v-model="open"
    :title="editing ? t('smtp.editSMTP') : t('smtp.addSMTP')"
    width="500px"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="{{ t('smtp.smtpName') }}" prop="name">
        <el-input v-model="form.name" placeholder="{{ t('smtp.smtpName') }}" />
      </el-form-item>
      <el-form-item label="{{ t('smtp.host') }}" prop="host">
        <el-input v-model="form.host" placeholder="smtp.gmail.com" />
      </el-form-item>
      <el-form-item label="{{ t('smtp.port') }}" prop="port">
        <el-input-number v-model="form.port" :min="1" :max="65535" />
      </el-form-item>
      <el-form-item label="{{ t('smtp.fromAddress') }}" prop="from">
        <el-input v-model="form.from" placeholder="sender@example.com" />
      </el-form-item>
      <el-form-item label="{{ t('smtp.username') }}" prop="username">
        <el-input v-model="form.username" />
      </el-form-item>
      <el-form-item label="{{ t('smtp.password') }}" prop="password">
        <el-input v-model="form.password" type="password" show-password />
      </el-form-item>
      <el-form-item label="{{ t('smtp.ssl') }}" prop="ssl">
        <el-switch v-model="form.ssl" />
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
import { ref, computed, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
import { createSMTP, updateSMTP } from '@/api/smtp'
import { ElMessage } from 'element-plus'

interface Props {
  open: boolean
  editing: SMTPConfig | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const defaultForm: Partial<SMTPConfig> = {
  name: '',
  host: '',
  port: 587,
  username: '',
  password: '',
  from: '',
  ssl: true
}

const form = ref<Partial<SMTPConfig>>({ ...defaultForm })

const rules = computed<FormRules<Partial<SMTPConfig>>>(() => ({
  name: [{ required: true, message: t('smtp.smtpName'), trigger: 'blur' }],
  host: [{ required: true, message: t('smtp.host'), trigger: 'blur' }],
  port: [{ required: true, message: t('smtp.port'), trigger: 'blur' }],
  from: [{ required: true, message: t('smtp.fromAddress'), trigger: 'blur' }],
  username: [{ required: true, message: t('smtp.username'), trigger: 'blur' }],
  password: [{ required: true, message: t('smtp.password'), trigger: 'blur' }]
}))

watch(() => props.editing, () => {
  if (props.editing) {
    form.value = { ...props.editing }
  } else {
    form.value = { ...defaultForm }
  }
}, { immediate: true })

watch(() => props.open, (newVal) => {
  if (newVal && !props.editing) {
    form.value = { ...defaultForm }
  }
})

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (props.editing) {
          await updateSMTP(form.value as SMTPConfig)
        } else {
          await createSMTP(form.value)
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
