<template>
  <el-dialog
    v-model="open"
    :title="editing ? t('token.editToken') : t('token.generateToken')"
    width="550px"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
    >
      <el-form-item :label="t('token.tokenName')" prop="name">
        <el-input v-model="form.name" :placeholder="t('token.tokenName')" />
      </el-form-item>
      <el-form-item :label="t('token.associatedSMTP')" prop="smtpId">
        <el-select v-model="form.smtpId" :placeholder="t('token.pleaseSelectSMTP')" filterable clearable>
          <el-option
            v-for="item in smtpList"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('token.allowedRecipients')" prop="receiveEmailIds">
        <el-select v-model="selectedRecipientIds" multiple :placeholder="t('token.pleaseSelectRecipients')" filterable clearable>
          <el-option
            v-for="item in mailList"
            :key="item.id"
            :label="`${item.name} (${item.email})`"
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')" prop="enable">
        <el-switch v-model="form.enable" />
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
import { createToken, updateToken } from '@/api/token'
import { getSMTPServers } from '@/api/smtp'
import { getMails } from '@/api/mail'
import { ElMessage } from 'element-plus'

interface Props {
  open: boolean
  editing: TokenConfig | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const smtpList = ref<SMTPConfig[]>([])
const mailList = ref<MailConfig[]>([])
const selectedRecipientIds = ref<number[]>([])

const defaultForm: Partial<TokenConfig> = {
  name: '',
  smtpId: 0,
  receiveEmailIds: '',
  enable: true
}

const form = ref<Partial<TokenConfig>>({ ...defaultForm })

const rules = computed<FormRules<Partial<TokenConfig>>>(() => ({
  name: [{ required: true, message: t('token.tokenName'), trigger: 'blur' }],
  smtpId: [{ required: true, message: t('token.pleaseSelectSMTP'), trigger: 'change' }]
}))

const loadOptions = async () => {
  const smtpRes = await getSMTPServers(1, 1000)
  if (smtpRes.code === 0 || smtpRes.code === 200) {
    smtpList.value = smtpRes.data.list
  }

  const mailRes = await getMails(1, 1000)
  if (mailRes.code === 0 || mailRes.code === 200) {
    mailList.value = mailRes.data.list
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
    // Parse receiveEmailIds from comma-separated string
    if (props.editing.receiveEmailIds) {
      selectedRecipientIds.value = props.editing.receiveEmailIds
        .split(',')
        .filter(Boolean)
        .map(id => parseInt(id, 10))
    } else {
      selectedRecipientIds.value = []
    }
  } else {
    form.value = { ...defaultForm }
    selectedRecipientIds.value = []
  }
}, { immediate: true })

watch(() => props.open, (newVal) => {
  if (newVal) {
    loadOptions()
    if (!props.editing) {
      form.value = { ...defaultForm }
      selectedRecipientIds.value = []
    }
  }
})

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      if (selectedRecipientIds.value.length === 0) {
        ElMessage.warning(t('token.pleaseSelectRecipients'))
        return
      }
      submitting.value = true
      try {
        form.value.receiveEmailIds = selectedRecipientIds.value.join(',')
        if (props.editing) {
          await updateToken(form.value as TokenConfig)
        } else {
          await createToken(form.value)
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
