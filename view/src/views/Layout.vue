<template>
  <div class="layout-container">
    <!-- 遮罩层（移动端） -->
    <div
      v-if="isMobile && !isCollapse"
      class="sidebar-overlay"
      @click="closeSidebar"
    ></div>

    <!-- 左侧导航 -->
    <div class="sidebar" :class="{
      'is-collapse': isCollapse,
      'is-mobile': isMobile,
      'is-hidden': isMobile && isCollapse
    }">
      <div class="logo">
        <span v-show="!isCollapse">HTTP2SMTP</span>
        <span v-show="isCollapse">H</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        :collapse="isMobile ? false : isCollapse"
        :router="true"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        @select="handleMenuSelect"
      >
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>{{ t('layout.sidebar.dashboard') }}</span>
        </el-menu-item>

        <el-menu-item index="/smtp">
          <el-icon><Setting /></el-icon>
          <span>{{ t('layout.sidebar.smtp') }}</span>
        </el-menu-item>

        <el-menu-item index="/mail">
          <el-icon><Message /></el-icon>
          <span>{{ t('layout.sidebar.mail') }}</span>
        </el-menu-item>

        <el-menu-item index="/tokens">
          <el-icon><Key /></el-icon>
          <span>{{ t('layout.sidebar.tokens') }}</span>
        </el-menu-item>

        <el-menu-item index="/schedule">
          <el-icon><Clock /></el-icon>
          <span>{{ t('layout.sidebar.schedule') }}</span>
        </el-menu-item>

        <el-menu-item index="/logs">
          <el-icon><Document /></el-icon>
          <span>{{ t('layout.sidebar.logs') }}</span>
        </el-menu-item>

        <el-menu-item index="/settings">
          <el-icon><Tools /></el-icon>
          <span>{{ t('layout.sidebar.settings') }}</span>
        </el-menu-item>
      </el-menu>
    </div>

    <!-- 右侧内容 -->
    <div class="main-container">
      <!-- 顶部导航 -->
      <div class="header">
        <div class="header-left">
          <!-- 汉堡菜单按钮（移动端显示） -->
          <el-icon v-if="isMobile" class="hamburger-btn" @click="toggleSidebar">
            <Menu />
          </el-icon>
          <!-- 折叠按钮（桌面端显示） -->
          <el-icon v-else class="collapse-btn" @click="toggleCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
              {{ item.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown trigger="click">
            <div class="user-info">
              <el-avatar :size="32">
                <el-icon><UserFilled /></el-icon>
              </el-avatar>
              <span class="username">{{ username }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item divided @click="logout">
                  {{ t('auth.logout') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- Tab标签栏 -->
      <div class="tabs-container">
        <div class="tabs-wrapper">
          <div
            v-for="tab in tabsStore.visitedTabs"
            :key="tab.path"
            class="tab-item"
            :class="{ 'is-active': tab.path === route.path }"
            @click="handleTabClick(tab)"
          >
            <span class="tab-title">{{ tab.title }}</span>
            <el-icon
              v-if="tab.closable"
              class="tab-close"
              @click.stop="handleTabClose(tab)"
            >
              <Close />
            </el-icon>
          </div>
        </div>
        <div class="tabs-actions">
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="tabs-dropdown">
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="closeOther">
                  {{ t('layout.tabs.closeOthers') }}
                </el-dropdown-item>
                <el-dropdown-item command="closeAll">
                  {{ t('layout.tabs.closeAll') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="content">
        <router-view v-slot="{ Component }">
          <component :is="Component" :key="componentKey" />
        </router-view>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { useTabsStore } from '@/store/tabs'
import { useI18n } from 'vue-i18n'
import {
  HomeFilled,
  Setting,
  Message,
  Key,
  Clock,
  Document,
  Tools,
  Fold,
  Expand,
  UserFilled,
  Close,
  ArrowDown,
  Menu
} from '@element-plus/icons-vue'

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const tabsStore = useTabsStore()

const isCollapse = ref(false)
const componentKey = ref(0)
const isMobile = ref(false)

// 断点检测
const checkMobile = () => {
  const width = window.innerWidth
  isMobile.value = width < 768
  // 小屏幕下自动收起侧边栏
  if (isMobile.value) {
    isCollapse.value = true
  } else {
    // 大屏幕下默认展开
    isCollapse.value = false
  }
}

// 监听窗口大小变化
let resizeTimer: any = null
const handleResize = () => {
  clearTimeout(resizeTimer)
  resizeTimer = setTimeout(() => {
    checkMobile()
  }, 100)
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

const activeMenu = computed(() => route.path)
const username = computed(() => authStore.username || 'Admin')

const breadcrumbs = computed(() => {
  const matched = route.matched.filter(item => item.meta && item.meta.titleKey)
  return matched.map(item => ({
    path: item.path,
    title: t(item.meta.titleKey as string)
  }))
})

// 监听路由变化，添加tab
watch(
  () => route.fullPath,
  (newFullPath, oldFullPath) => {
    tabsStore.addTab(route)
    // 只有点击菜单或跳转新页面时刷新，切换tab不刷新（在handleTabClick中处理）
    if (!oldFullPath || route.path !== oldFullPath.replace(/\?.*$/, '')) {
      componentKey.value++
    }
  },
  { immediate: true }
)

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

// 移动端切换侧边栏
const toggleSidebar = () => {
  isCollapse.value = !isCollapse.value
}

// 关闭侧边栏（点击遮罩层）
const closeSidebar = () => {
  isCollapse.value = true
}

// 点击菜单项后自动关闭侧边栏（移动端）
const handleMenuSelect = () => {
  if (isMobile.value) {
    // 延迟关闭，让路由跳转有动画时间
    setTimeout(() => {
      isCollapse.value = true
    }, 150)
  }
}

const logout = () => {
  authStore.logoutAction()
  router.push('/login')
}

// 点击tab切换 - 使用fullPath跳转以保留参数
const handleTabClick = (tab: any) => {
  if (tab.path !== route.path) {
    tabsStore.setActiveTab(tab.fullPath)
    router.push(tab.fullPath)
    // 切换tab时刷新组件
    componentKey.value++
  }
}

// 关闭tab
const handleTabClose = (tab: any) => {
  const activeFullPath = tabsStore.closeTab(tab.path)
  if (route.path === tab.path) {
    router.push(activeFullPath)
  }
}

// 下拉菜单操作
const handleCommand = (command: string) => {
  const currentPath = route.path
  switch (command) {
    case 'closeOther':
      tabsStore.closeOtherTabs(currentPath)
      break
    case 'closeAll':
      const path = tabsStore.closeAllTabs()
      router.push(path)
      break
  }
}
</script>

<style scoped lang="scss">
.layout-container {
  display: flex;
  height: 100vh;
  width: 100%;
}

/* 遮罩层 */
.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 999;
}

.sidebar {
  width: 210px;
  background-color: #304156;
  transition: all 0.3s;
  overflow: hidden;
  flex-shrink: 0;
}

.sidebar.is-collapse {
  width: 64px;
}

/* 移动端侧边栏样式 */
.sidebar.is-mobile {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  z-index: 1000;
}

.sidebar.is-mobile.is-collapse {
  width: 210px;
  transform: translateX(0);
}

.sidebar.is-mobile.is-hidden {
  transform: translateX(-100%);
  width: 210px;
}

.logo {
  height: 50px;
  line-height: 50px;
  text-align: center;
  color: #fff;
  font-size: 16px;
  font-weight: bold;
  background-color: #263445;
  overflow: hidden;
  white-space: nowrap;
}

.sidebar-menu {
  border-right: none;
  height: calc(100% - 50px);

  :deep(.el-scrollbar) {
    height: 100%;
    .el-scrollbar__wrap {
      height: 100%;
    }
    .el-scrollbar__view {
      height: 100%;
    }
  }
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 210px;
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.header {
  height: 50px;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.collapse-btn,
.hamburger-btn {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
}

.collapse-btn:hover,
.hamburger-btn:hover {
  color: #409EFF;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.username {
  color: #606266;
}

/* Tab样式 */
.tabs-container {
  display: flex;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #d8dce5;
  padding: 0 10px;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.04);
}

.tabs-wrapper {
  display: flex;
  align-items: center;
  flex: 1;
  overflow-x: auto;
  padding: 5px 0;
}

.tabs-wrapper::-webkit-scrollbar {
  height: 4px;
}

.tabs-wrapper::-webkit-scrollbar-thumb {
  background: #ccc;
  border-radius: 2px;
}

.tab-item {
  display: flex;
  align-items: center;
  height: 28px;
  line-height: 28px;
  padding: 0 12px;
  margin-right: 5px;
  border: 1px solid #d8dce5;
  border-radius: 3px;
  background: #fff;
  cursor: pointer;
  font-size: 12px;
  color: #495060;
  white-space: nowrap;
  transition: all 0.3s;
}

.tab-item:hover {
  color: #409EFF;
}

.tab-item.is-active {
  background: #409EFF;
  border-color: #409EFF;
  color: #fff;
}

.tab-item.is-active .tab-close {
  color: #fff;
}

.tab-title {
  margin-right: 5px;
}

.tab-close {
  font-size: 12px;
  border-radius: 50%;
  transition: all 0.3s;
}

.tab-close:hover {
  background: rgba(0, 0, 0, 0.1);
}

.tab-item.is-active .tab-close:hover {
  background: rgba(255, 255, 255, 0.3);
}

.tabs-actions {
  padding: 0 10px;
  border-left: 1px solid #d8dce5;
}

.tabs-dropdown {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #495060;
  padding: 0 5px;
}

.tabs-dropdown:hover {
  color: #409EFF;
}

.content {
  flex: 1;
  padding: 20px;
  background: #f0f2f5;
  overflow: auto;
}

/* 响应式样式 */
@media (max-width: 768px) {
  .header {
    padding: 0 15px;
  }

  .username {
    display: none;
  }

  .tabs-actions {
    display: none;
  }

  .content {
    padding: 15px;
  }
}

@media (max-width: 576px) {
  .content {
    padding: 10px;
  }

  .tab-item {
    padding: 0 8px;
    margin-right: 3px;
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
