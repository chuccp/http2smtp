import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RouteLocationNormalized } from 'vue-router'

interface Tab {
  path: string
  fullPath: string
  name: string
  title: string
  closable: boolean
}

export const useTabsStore = defineStore('tabs', () => {
  // 固定的首页tab
  const homeTab: Tab = {
    path: '/dashboard',
    fullPath: '/dashboard',
    name: 'Dashboard',
    title: 'Dashboard',
    closable: false
  }

  // 打开的tab列表
  const visitedTabs = ref<Tab[]>([homeTab])

  // 当前激活的tab完整路径
  const activeTabPath = ref('/dashboard')

  // 添加tab
  const addTab = (route: RouteLocationNormalized | { path: string; fullPath?: string; name?: string; meta?: { title?: string } }) => {
    const fullPath = route.fullPath || route.path
    const path = route.path

    // 首页不重复添加
    if (path === '/dashboard') {
      activeTabPath.value = fullPath
      return
    }

    // 用path作为key判断是否已存在
    const index = visitedTabs.value.findIndex(tab => tab.path === path)
    if (index > -1) {
      // 已存在的tab不更新fullPath，保留原来的参数
    } else {
      visitedTabs.value.push({
        path: path,
        fullPath: fullPath,
        name: String(route.name || ''),
        title: (route.meta?.title as string) || String(route.name) || 'Untitled',
        closable: true
      })
    }
    activeTabPath.value = fullPath
  }

  // 关闭tab
  const closeTab = (path: string): string => {
    const index = visitedTabs.value.findIndex(tab => tab.path === path)
    if (index > -1) {
      visitedTabs.value.splice(index, 1)

      // 如果关闭的是当前激活的tab，切换到前一个或后一个
      if (activeTabPath.value === path || visitedTabs.value.every(tab => tab.path !== activeTabPath.value.replace(/\?.*$/, ''))) {
        const nextTab = visitedTabs.value[index] || visitedTabs.value[index - 1]
        if (nextTab) {
          activeTabPath.value = nextTab.fullPath
          return nextTab.fullPath
        }
      }
    }
    return activeTabPath.value
  }

  // 关闭其他tab
  const closeOtherTabs = (path: string) => {
    visitedTabs.value = visitedTabs.value.filter(tab => !tab.closable || tab.path === path)
    const tab = visitedTabs.value.find(tab => tab.path === path)
    if (tab) {
      activeTabPath.value = tab.fullPath
    }
  }

  // 关闭所有tab（保留首页）
  const closeAllTabs = (): string => {
    visitedTabs.value = [homeTab]
    activeTabPath.value = '/dashboard'
    return '/dashboard'
  }

  // 关闭右侧tab
  const closeRightTabs = (path: string) => {
    const index = visitedTabs.value.findIndex(tab => tab.path === path)
    if (index > -1) {
      visitedTabs.value = visitedTabs.value.filter((tab, i) => !tab.closable || i <= index)
      const tab = visitedTabs.value.find(tab => tab.path === activeTabPath.value.replace(/\?.*$/, ''))
      if (!tab) {
        const currentTab = visitedTabs.value.find(tab => tab.path === path)
        if (currentTab) {
          activeTabPath.value = currentTab.fullPath
        }
      }
    }
  }

  // 设置当前激活tab
  const setActiveTab = (fullPath: string) => {
    activeTabPath.value = fullPath
  }

  return {
    visitedTabs,
    activeTabPath,
    addTab,
    closeTab,
    closeOtherTabs,
    closeAllTabs,
    closeRightTabs,
    setActiveTab
  }
})
