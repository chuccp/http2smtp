<template>
  <div class="dashboard-container app-container">
    <div class="stats-row">
      <div class="stat-card smtp-card" @click="goTo('/smtp')">
        <div class="stat-icon">
          <el-icon :size="28"><Setting /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-number">{{ counts.smtp }}</div>
          <div class="stat-label">{{ t('dashboard.smtpServers') }}</div>
        </div>
      </div>
      <div class="stat-card mail-card" @click="goTo('/mail')">
        <div class="stat-icon">
          <el-icon :size="28"><Message /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-number">{{ counts.mail }}</div>
          <div class="stat-label">{{ t('dashboard.recipients') }}</div>
        </div>
      </div>
      <div class="stat-card token-card" @click="goTo('/tokens')">
        <div class="stat-icon">
          <el-icon :size="28"><Key /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-number">{{ counts.token }}</div>
          <div class="stat-label">{{ t('dashboard.tokens') }}</div>
        </div>
      </div>
      <div class="stat-card schedule-card" @click="goTo('/schedule')">
        <div class="stat-icon">
          <el-icon :size="28"><Clock /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-number">{{ counts.schedule }}</div>
          <div class="stat-label">{{ t('dashboard.scheduledTasks') }}</div>
        </div>
      </div>
    </div>

    <div class="quick-section">
      <div class="section-title">{{ t('dashboard.quickActions') }}</div>
      <div class="quick-grid">
        <div class="quick-item" @click="goTo('/smtp')">
          <el-icon :size="32" color="#409EFF"><Setting /></el-icon>
          <span>{{ t('dashboard.goToSMTP') }}</span>
        </div>
        <div class="quick-item" @click="goTo('/mail')">
          <el-icon :size="32" color="#67C23A"><Message /></el-icon>
          <span>{{ t('dashboard.goToMail') }}</span>
        </div>
        <div class="quick-item" @click="goTo('/tokens')">
          <el-icon :size="32" color="#E6A23C"><Key /></el-icon>
          <span>{{ t('dashboard.goToTokens') }}</span>
        </div>
        <div class="quick-item" @click="goTo('/schedule')">
          <el-icon :size="32" color="#F56C6C"><Clock /></el-icon>
          <span>{{ t('dashboard.goToSchedule') }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Setting, Message, Key, Clock } from '@element-plus/icons-vue'

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
  // uses global app-container styles
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: all 0.3s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.smtp-card .stat-icon {
  background: rgba(64, 158, 255, 0.1);
  color: #409EFF;
}
.mail-card .stat-icon {
  background: rgba(103, 194, 58, 0.1);
  color: #67C23A;
}
.token-card .stat-icon {
  background: rgba(230, 162, 60, 0.1);
  color: #E6A23C;
}
.schedule-card .stat-icon {
  background: rgba(245, 108, 108, 0.1);
  color: #F56C6C;
}

.stat-info {
  .stat-number {
    font-size: 28px;
    font-weight: 700;
    color: #303133;
    line-height: 1.2;
  }
  .stat-label {
    font-size: 13px;
    color: #909399;
    margin-top: 2px;
  }
}

.quick-section {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 20px;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.quick-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 24px 16px;
  border-radius: 8px;
  background: #f5f7fa;
  cursor: pointer;
  transition: all 0.3s;

  span {
    font-size: 13px;
    color: #606266;
    font-weight: 500;
  }

  &:hover {
    background: #ecf5ff;
    transform: translateY(-2px);

    span {
      color: #409EFF;
    }
  }
}

@media (max-width: 992px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
  .quick-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 576px) {
  .stats-row {
    grid-template-columns: 1fr;
  }
  .quick-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
