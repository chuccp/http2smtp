/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// API Response types
interface ApiResponse<T> {
  code: number
  data: T
  msg: string
}

interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// Entity types
interface SMTPConfig {
  id: number
  name: string
  host: string
  port: number
  username: string
  password: string
  from: string
  ssl: boolean
  createTime: string
}

interface MailConfig {
  id: number
  name: string
  email: string
  createTime: string
}

interface TokenConfig {
  id: number
  token: string
  name: string
  smtpId: number
  receiveEmailIds: string
  enable: boolean
  createTime: string
}

interface ScheduleConfig {
  id: number
  name: string
  tokenId: number
  cron: string
  url: string
  method: 'GET' | 'POST'
  headers: string
  body: string
  useTemplate: boolean
  enable: boolean
  createTime: string
}

interface LogEntry {
  id: number
  token: string
  subject: string
  content: string
  status: string
  result: string
  createTime: string
}

interface SetInfo {
  webPort: number
  apiPort: number
  dbType: 'sqlite' | 'mysql'
  dbHost: string
  dbPort: number
  dbName: string
  dbUser: string
  dbPass: string
  dbCharset: string
  dbFile: string
  adminUser: string
  adminPass: string
}

interface SystemInfo {
  initialized: boolean
  version: string
}
