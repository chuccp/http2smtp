<template>
  <div class="dashboard-container app-container">
    <div class="row">
      <el-col :xs="24" :sm="12" :md="12" :lg="12">
        <el-card class="stat-card gradient-2">
          <div class="stat-content">
            <div class="stat-title">{{ t('dashboard.quickActions') }}</div>
            <div class="actions">
              <el-button size="small" @click="goTo('/smtp')">
                {{ t('dashboard.goToSMTP') }}
              </el-button>
              <el-button size="small" @click="goTo('/mail')">
                {{ t('dashboard.goToMail') }}
              </el-button>
              <el-button size="small" @click="goTo('/tokens')">
                {{ t('dashboard.goToTokens') }}
              </el-button>
              <el-button size="small" @click="goTo('/schedule')">
                {{ t('dashboard.goToSchedule') }}
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="24" :lg="12">
        <el-card class="stat-card gradient-3">
          <div class="stat-content">
            <div class="stat-title">{{ t('dashboard.statistics') }}</div>
            <div class="stats-grid">
              <div class="stat-item">
                <div class="stat-number">{{ counts.smtp }}</div>
                <div class="stat-label">{{ t('dashboard.smtpServers') }}</div>
              </div>
              <div class="stat-item">
                <div class="stat-number">{{ counts.mail }}</div>
                <div class="stat-label">{{ t('dashboard.recipients') }}</div>
              </div>
              <div class="stat-item">
                <div class="stat-number">{{ counts.token }}</div>
                <div class="stat-label">{{ t('dashboard.tokens') }}</div>
              </div>
              <div class="stat-item">
                <div class="stat-number">{{ counts.schedule }}</div>
                <div class="stat-label">{{ t('dashboard.scheduledTasks') }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
import { getSMTPServers } from '@/api/smtp'
import { getMails } from '@/api/mail'
import { getTokens } from '@/api/token'
import { getSchedules } from '@/api/schedule'

const router = useRouter()

const counts = ref({
  smtp: 0,
  mail: 0,
  token: 0,
  schedule: 0
})

const goTo = (path: string) => {
  router.push(path)
}

onMounted(async () => {
  try {
    const smtpRes = await getSMTPServers(1, 1000)
    counts.value.smtp = smtpRes.data?.total || 0

    const mailRes = await getMails(1, 1000)
    counts.value.mail = mailRes.data?.total || 0

    const tokenRes = await getTokens(1, 1000)
    counts.value.token = tokenRes.data?.total || 0

    const scheduleRes = await getSchedules(1, 1000)
    counts.value.schedule = scheduleRes.data?.total || 0
  } catch (e) {
    console.error('Failed to load counts', e)
  }
})
</script>

<style scoped lang="scss">
.dashboard-container {
  background: #f0f2f5;
}

.row {
  display: flex;
  flex-wrap: wrap;
  margin: 0 -10px;
}

.el-col {
  padding: 10px;
}

.stat-card {
  border: none;
  color: white;
  background: transparent !important;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);

  :deep(.el-card__body) {
    padding: 20px;
  }

  &.gradient-2 {
    background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%) !important;
  }

  &.gradient-3 {
    background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%) !important;
  }

  .stat-content {
    .stat-title {
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 16px;
      opacity: 0.95;
    }

    .stat-item {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;

      .label {
        opacity: 0.85;
      }

      .value {
        font-weight: 500;
      }
    }

    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;

      .el-button {
        background: rgba(255, 255, 255, 0.2);
        border-color: rgba(255, 255, 255, 0.4);
        color: white;

        &:hover {
          background: rgba(255, 255, 255, 0.3);
          color: white;
        }
      }
    }

    .stats-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 12px;

      .stat-item {
        text-align: center;
        background: rgba(255, 255, 255, 0.15);
        padding: 12px 8px;
        border-radius: 8px;

        .stat-number {
          font-size: 28px;
          font-weight: bold;
          line-height: 1.2;
        }

        .stat-label {
          font-size: 12px;
          opacity: 0.85;
          margin-top: 4px;
        }
      }
    }
  }
}

@media (max-width: 768px) {
  .stat-card .stat-content .actions {
    flex-direction: column;
  }
}
</style>
