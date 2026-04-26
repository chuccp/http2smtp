<template>
  <el-dialog
    v-model="open"
    :title="editing ? t('user.editUser') : t('user.addUser')"
    width="500px"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item :label="t('user.username')" prop="name">
        <el-input v-model="form.name" :placeholder="t('user.username')" />
      </el-form-item>
      <el-form-item :label="t('auth.password')" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          :placeholder="editing ? t('user.passwordPlaceholder') : t('auth.password')"
        />
      </el-form-item>
      <el-form-item :label="t('user.isAdmin')" prop="isAdmin">
        <el-switch v-model="form.isAdmin" />
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
import { ref, computed, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { createUser, updateUser } from '@/api/user'
import { ElMessage } from 'element-plus'

const { t } = useI18n()

interface Props {
  open: boolean
  editing: UserConfig | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const defaultForm = {
  name: '',
  password: '',
  isAdmin: false,
  isUse: true
}

const form = ref({ ...defaultForm })

const rules = computed<FormRules>(() => ({
  name: [{ required: true, message: t('user.username'), trigger: 'blur' }],
  password: [{
    required: !props.editing,
    message: t('auth.pleaseEnterPassword'),
    trigger: 'blur'
  }]
}))

watch(() => props.editing, () => {
  if (props.editing) {
    form.value = {
      name: props.editing.name,
      password: '',
      isAdmin: props.editing.isAdmin,
      isUse: props.editing.isUse
    }
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
          const payload: any = {
            id: props.editing.id,
            name: form.value.name,
            isAdmin: form.value.isAdmin,
            isUse: form.value.isUse
          }
          if (form.value.password) {
            payload.password = form.value.password
          }
          await updateUser(payload)
        } else {
          await createUser(form.value)
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
