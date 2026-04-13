<template>
  <el-dialog
    v-model="open"
    :title="t('token.sendTestMail')"
    width="500px"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="{{ t('mail.emailAddress') }}" prop="toEmail">
        <el-input v-model="form.toEmail" placeholder="recipient@example.com" />
      </el-form-item>
      <el-form-item label="{{ t('token.subject') }}" prop="subject">
        <el-input v-model="form.subject" :placeholder="t('smtp.testEmailSubject')" />
      </el-form-item>
      <el-form-item label="{{ t('log.content') }}" prop="content">
        <el-input
          v-model="form.content"
          type="textarea"
          rows="4"
          :placeholder="t('smtp.testEmailContent')"
        />
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
import { sendMailByToken } from '@/api/token'
import { ElMessage } from 'element-plus'

interface Props {
  open: boolean
  tokenId: number
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = ref({
  toEmail: '',
  subject: t('smtp.testEmailSubject'),
  content: t('smtp.testEmailContent')
})

const rules = computed<FormRules<any>>(() => ({
  toEmail: [
    { required: true, message: t('mail.pleaseEnterValidEmail'), trigger: 'blur' },
    { type: 'email', message: t('mail.pleaseEnterValidEmail'), trigger: 'blur' }
  ],
  subject: [{ required: true, message: t('token.subject'), trigger: 'blur' }],
  content: [{ required: true, message: t('log.content'), trigger: 'blur' }]
}))

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        await sendMailByToken({
          tokenId: props.tokenId,
          toEmail: form.value.toEmail,
          subject: form.value.subject,
          content: form.value.content
        })
        ElMessage.success(t('smtp.testEmailSent'))
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
