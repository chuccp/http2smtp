<template>
  <el-dialog
    v-model="open"
    :title="t('smtp.sendTestMail')"
    width="500px"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item :label="t('mail.recipientName')" prop="toEmail">
        <el-select v-model="form.toEmail" :placeholder="t('smtp.pleaseSelectRecipient')" filterable clearable>
          <el-option
            v-for="item in mailList"
            :key="item.id"
            :label="`${item.name} <${item.mail}>`"
            :value="item.mail"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('token.subject')" prop="subject">
        <el-input v-model="form.subject" :placeholder="t('smtp.testEmailSubject')" />
      </el-form-item>
      <el-form-item :label="t('log.content')" prop="content">
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="4"
          :placeholder="t('smtp.testEmailContent')"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="open = false">
        {{ t('common.cancel') }}
      </el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        {{ t('smtp.sendTestMail') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
import { sendTestMail } from '@/api/smtp'
import { getMails } from '@/api/mail'
import { ElMessage } from 'element-plus'

interface Props {
  open: boolean
  smtpId: number
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const mailList = ref<MailConfig[]>([])

const form = ref({
  toEmail: '',
  subject: t('smtp.testEmailSubject'),
  content: t('smtp.testEmailContent')
})

const rules = computed<FormRules<any>>(() => ({
  toEmail: [{ required: true, message: t('smtp.pleaseSelectRecipient'), trigger: 'change' }],
  subject: [{ required: true, message: t('token.subject'), trigger: 'blur' }],
  content: [{ required: true, message: t('log.content'), trigger: 'blur' }]
}))

const loadMails = async () => {
  const res = await getMails(1, 1000)
  if (res.code === 0 || res.code === 200) {
    mailList.value = res.data.list
  }
}

onMounted(() => {
  if (props.open) {
    loadMails()
  }
})

watch(() => props.open, (newVal) => {
  if (newVal) {
    loadMails()
  }
})

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        await sendTestMail({
          SMTPId: props.smtpId,
          recipients: [form.value.toEmail],
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
