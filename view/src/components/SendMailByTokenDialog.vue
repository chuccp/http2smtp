<template>
  <el-dialog
    v-model="open"
    :title="t('token.sendTestMail')"
    width="500px"
  >
    <div v-loading="infoLoading" class="token-info">
      <div class="info-row" v-if="smtpInfo">
        <span class="info-label">{{ t('token.associatedSMTP') }}:</span>
        <span class="info-value">{{ smtpInfo.name }} &lt;{{ smtpInfo.mail }}&gt;</span>
      </div>
      <div class="info-row" v-if="recipientList.length > 0">
        <span class="info-label">{{ t('token.allowedRecipients') }}:</span>
        <el-tag
          v-for="item in recipientList"
          :key="item.id"
          size="small"
          type="info"
          class="recipient-tag"
        >
          {{ item.name }} ({{ item.mail }})
        </el-tag>
      </div>
    </div>
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item :label="t('token.subject')" prop="subject">
        <el-input v-model="form.subject" :placeholder="t('smtp.testEmailSubject')" />
      </el-form-item>
      <el-form-item :label="t('log.content')" prop="content">
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
        {{ t('token.sendTestMail') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
import { sendMailByToken, getToken } from '@/api/token'
import { getSMTP } from '@/api/smtp'
import { getMails } from '@/api/mail'
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
const infoLoading = ref(false)
const smtpInfo = ref<SMTPConfig | null>(null)
const recipientList = ref<MailConfig[]>([])

const form = ref({
  subject: t('smtp.testEmailSubject'),
  content: t('smtp.testEmailContent')
})

const rules = computed<FormRules<any>>(() => ({
  subject: [{ required: true, message: t('token.subject'), trigger: 'blur' }],
  content: [{ required: true, message: t('log.content'), trigger: 'blur' }]
}))

const loadTokenInfo = async () => {
  infoLoading.value = true
  try {
    const res = await getToken(props.tokenId)
    if (res.code === 0 || res.code === 200) {
      const token = res.data
      // Load SMTP info
      if (token.SMTPId) {
        const smtpRes = await getSMTP(token.SMTPId)
        if (smtpRes.code === 0 || smtpRes.code === 200) {
          smtpInfo.value = smtpRes.data
        }
      }
      // Load recipient info
      if (token.receiveEmailIds) {
        const ids = token.receiveEmailIds.split(',').filter(Boolean).map(Number)
        const mailRes = await getMails(1, 1000)
        if (mailRes.code === 0 || mailRes.code === 200) {
          recipientList.value = mailRes.data.list.filter(m => ids.includes(m.id))
        }
      } else {
        recipientList.value = []
      }
    }
  } finally {
    infoLoading.value = false
  }
}

watch(() => props.open, (val) => {
  if (val && props.tokenId) {
    loadTokenInfo()
  } else if (!val) {
    smtpInfo.value = null
    recipientList.value = []
  }
})

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        await sendMailByToken({
          tokenId: props.tokenId,
          recipients: recipientList.value.map(r => r.mail),
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

<style scoped>
.token-info {
  margin-bottom: 16px;
  padding: 12px;
  background-color: var(--el-fill-color-light);
  border-radius: 4px;
}
.info-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.info-row:last-child {
  margin-bottom: 0;
}
.info-label {
  font-weight: 600;
  font-size: 13px;
  color: var(--el-text-color-regular);
  flex-shrink: 0;
}
.info-value {
  font-size: 13px;
}
.recipient-tag {
  margin: 0;
}
</style>
