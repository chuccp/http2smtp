<template>
  <el-dialog
    v-model="open"
    :title="t('user.changePassword')"
    width="450px"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item :label="t('user.username')">
        <el-input :model-value="user?.name" disabled />
      </el-form-item>
      <el-form-item :label="t('user.newPassword')" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          :placeholder="t('user.newPassword')"
        />
      </el-form-item>
      <el-form-item :label="t('settings.confirmPassword')" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          show-password
          :placeholder="t('settings.confirmPassword')"
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
import { updateUser } from '@/api/user'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const authStore = useAuthStore()

interface Props {
  open: boolean
  user: UserConfig | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = ref({
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== form.value.password) {
    callback(new Error(t('settings.passwordsNotMatch')))
  } else {
    callback()
  }
}

const rules = computed<FormRules>(() => ({
  password: [{ required: true, message: t('user.newPassword'), trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: t('settings.confirmPassword'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}))

watch(() => props.open, (newVal) => {
  if (newVal) {
    form.value = { password: '', confirmPassword: '' }
  }
})

const handleSubmit = async () => {
  if (!formRef.value || !props.user) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        await updateUser({
          id: props.user!.id,
          name: props.user!.name,
          password: form.value.password,
          isAdmin: props.user!.isAdmin,
          isUse: props.user!.isUse
        })
        ElMessage.success(t('common.success'))
        emit('update:open', false)
        if (props.user!.isAdmin) {
          authStore.logoutAction()
        } else {
          emit('success')
        }
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
