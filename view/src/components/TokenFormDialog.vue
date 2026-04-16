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
      <el-form-item :label="t('token.tokenValue')" prop="token">
        <el-input v-model="form.token" :placeholder="t('token.tokenValue')">
          <template #suffix>
            <el-button link size="small" @click="form.token = generateRandomString(32)">
              {{ t('token.generateToken') }}
            </el-button>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item :label="t('token.associatedSMTP')" prop="SMTPId">
        <SmtpSelector v-model="form.SMTPId" />
      </el-form-item>
      <el-form-item :label="t('token.allowedRecipients')" prop="receiveEmailIds">
        <el-select v-model="selectedRecipientIds" multiple :placeholder="t('token.pleaseSelectRecipients')" filterable clearable>
          <el-option
            v-for="item in mailList"
            :key="item.id"
            :label="`${item.name} (${item.mail})`"
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')" prop="state">
        <el-select v-model="form.state" :placeholder="t('common.status')">
          <el-option :label="t('token.inUse')" :value="0" />
          <el-option :label="t('token.userDisabled')" :value="1" />
          <el-option :label="t('token.adminDisabled')" :value="2" />
        </el-select>
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
import { getMails } from '@/api/mail'
import { ElMessage } from 'element-plus'
import { generateRandomString } from '@/utils/crypto'
import SmtpSelector from '@/components/SmtpSelector.vue'

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
const mailList = ref<MailConfig[]>([])
const selectedRecipientIds = ref<number[]>([])

const defaultForm: Partial<TokenConfig> = {
  name: '',
  token: generateRandomString(32),
  SMTPId: 0,
  receiveEmailIds: '',
  state: 0
}

const form = ref<Partial<TokenConfig>>({ ...defaultForm })

const rules = computed<FormRules<Partial<TokenConfig>>>(() => ({
  name: [{ required: true, message: t('token.tokenName'), trigger: 'blur' }],
  token: [{ required: true, message: t('token.tokenValue'), trigger: 'blur' }],
  SMTPId: [{ required: true, message: t('token.pleaseSelectSMTP'), trigger: 'change' }]
}))

const loadOptions = async () => {
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
    form.value = { ...defaultForm, token: generateRandomString(32) }
    selectedRecipientIds.value = []
  }
}, { immediate: true })

watch(() => props.open, (newVal) => {
  if (newVal) {
    loadOptions()
    if (!props.editing) {
      form.value = { ...defaultForm, token: generateRandomString(32) }
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
