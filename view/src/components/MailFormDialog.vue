<template>
  <el-dialog
    v-model="open"
    :title="editing ? t('mail.editRecipient') : t('mail.addRecipient')"
    width="500px"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item :label="t('mail.recipientName')" prop="name">
        <el-input v-model="form.name" :placeholder="t('mail.recipientName')" />
      </el-form-item>
      <el-form-item :label="t('mail.emailAddress')" prop="email">
        <el-input v-model="form.email" placeholder="recipient@example.com" />
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
import { createMail, updateMail } from '@/api/mail'
import { ElMessage } from 'element-plus'

interface Props {
  open: boolean
  editing: MailConfig | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const defaultForm: Partial<MailConfig> = {
  name: '',
  email: ''
}

const form = ref<Partial<MailConfig>>({ ...defaultForm })

const rules = computed<FormRules<Partial<MailConfig>>>(() => ({
  name: [{ required: true, message: t('mail.recipientName'), trigger: 'blur' }],
  email: [
    { required: true, message: t('mail.emailAddress'), trigger: 'blur' },
    { type: 'email', message: t('mail.pleaseEnterValidEmail'), trigger: 'blur' }
  ]
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
          await updateMail(form.value as MailConfig)
        } else {
          await createMail(form.value)
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
