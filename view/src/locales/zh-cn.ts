export default {
  // Common
  common: {
    confirm: '确认',
    cancel: '取消',
    save: '保存',
    add: '添加',
    edit: '编辑',
    delete: '删除',
    search: '搜索',
    reset: '重置',
    action: '操作',
    operations: '操作',
    create: '创建',
    update: '更新',
    close: '关闭',
    yes: '是',
    no: '否',
    loading: '加载中...',
    success: '成功',
    failed: '失败',
    copied: '已复制到剪贴板',
    enable: '启用',
    disable: '禁用',
    status: '状态',
    name: '名称',
    createTime: '创建时间',
    description: '描述'
  },

  // Auth
  auth: {
    login: '登录',
    logout: '退出登录',
    username: '用户名',
    password: '密码',
    rememberMe: '记住我',
    loginSuccess: '登录成功',
    loginFailed: '登录失败',
    pleaseEnterUsername: '请输入用户名',
    pleaseEnterPassword: '请输入密码'
  },

  // Setup
  setup: {
    systemSetup: '系统初始化',
    setupWizard: '安装向导',
    databaseConfig: '数据库配置',
    databaseType: '数据库类型',
    sqlite: 'SQLite',
    mysql: 'MySQL',
    sqliteFile: '数据库文件路径',
    mysqlHost: 'MySQL 主机',
    mysqlPort: 'MySQL 端口',
    mysqlDatabase: '数据库名',
    mysqlUser: '用户名',
    mysqlPassword: '密码',
    mysqlCharset: '字符集',
    testConnection: '测试连接',
    connectionSuccess: '连接成功',
    connectionFailed: '连接失败',
    portConfig: '端口配置',
    webPort: '管理端口',
    apiPort: 'API 端口',
    adminAccount: '管理员账号',
    adminUsername: '管理员用户名',
    adminPassword: '管理员密码',
    confirmPassword: '确认密码',
    passwordsNotMatch: '两次密码输入不一致',
    initialize: '初始化系统',
    initializing: '初始化中...',
    initializationSuccess: '系统初始化成功',
    alreadyInitialized: '系统已初始化，正在跳转...'
  },

  // Dashboard
  dashboard: {
    dashboard: '仪表盘',
    systemInfo: '系统信息',
    quickActions: '快捷操作',
    statistics: '统计信息',
    version: '版本',
    initialized: '已初始化',
    notInitialized: '未初始化',
    smtpServers: 'SMTP 服务器',
    recipients: '收件人',
    tokens: 'API 令牌',
    scheduledTasks: '定时任务',
    goToSMTP: '管理 SMTP',
    goToMail: '管理收件人',
    goToTokens: '管理令牌',
    goToSchedule: '管理计划'
  },

  // SMTP
  smtp: {
    smtpManagement: 'SMTP 管理',
    addSMTP: '添加 SMTP 服务器',
    editSMTP: '编辑 SMTP 服务器',
    smtpName: '服务器名称',
    host: '主机地址',
    port: '端口',
    username: '用户名',
    password: '密码',
    fromAddress: '发件人地址',
    ssl: '使用 SSL',
    testConnection: '测试连接',
    sendTestMail: '发送测试邮件',
    deleteConfirm: '确定要删除这个 SMTP 服务器吗？',
    connectionSuccess: '连接测试成功',
    connectionFailed: '连接测试失败',
    pleaseSelectRecipient: '请选择一个收件人',
    testEmailSubject: '测试邮件',
    testEmailContent: '这是来自 HTTP2SMTP 的测试邮件。',
    testEmailSent: '测试邮件发送成功'
  },

  // Mail
  mail: {
    mailManagement: '收件人管理',
    addRecipient: '添加收件人',
    editRecipient: '编辑收件人',
    recipientName: '收件人名称',
    emailAddress: '邮箱地址',
    deleteConfirm: '确定要删除这个收件人吗？',
    pleaseEnterValidEmail: '请输入有效的邮箱地址'
  },

  // Token
  token: {
    tokenManagement: 'API 令牌管理',
    generateToken: '生成令牌',
    editToken: '编辑令牌',
    tokenName: '令牌名称',
    tokenValue: '令牌',
    subject: '主题',
    associatedSMTP: 'SMTP 服务器',
    allowedRecipients: '允许的收件人',
    status: '状态',
    enabled: '启用',
    disabled: '禁用',
    copyToken: '复制令牌',
    sendTestMail: '发送测试邮件',
    deleteConfirm: '确定要删除这个令牌吗？',
    pleaseSelectSMTP: '请选择 SMTP 服务器',
    pleaseSelectRecipients: '请至少选择一个收件人',
    tokenGenerated: '令牌已生成，请立即复制保存：{token}'
  },

  // Schedule
  schedule: {
    scheduleManagement: '定时任务管理',
    addSchedule: '添加定时任务',
    editSchedule: '编辑定时任务',
    taskName: '任务名称',
    associatedToken: 'API 令牌',
    cronExpression: 'Cron 表达式',
    requestUrl: '请求 URL',
    requestMethod: '请求方法',
    requestHeaders: '请求头 (JSON 格式)',
    requestBody: '请求体',
    useTemplate: '使用模板',
    status: '状态',
    enabled: '启用',
    disabled: '禁用',
    triggerNow: '立即触发',
    deleteConfirm: '确定要删除这个定时任务吗？',
    triggered: '任务触发成功',
    cronHelp: 'Cron 表达式格式：分 时 日 月 周',
    exampleCron: '0 9 * * * = 每天早上 9 点'
  },

  // Log
  log: {
    systemLogs: '系统日志',
    searchKeyword: '搜索关键词...',
    token: '令牌',
    subject: '主题',
    content: '内容',
    result: '结果',
    logDetail: '日志详情',
    status: '状态',
    createTime: '创建时间',
    smtpServer: 'SMTP 服务器',
    recipients: '收件人',
    attachments: '附件',
    download: '下载',
    noResults: '未找到日志',
    clear: '清除',
    success: '成功',
    error: '错误',
    warning: '警告'
  },

  // Settings
  settings: {
    systemSettings: '系统设置',
    generalSettings: '通用设置',
    databaseConfig: '数据库配置',
    adminAccount: '管理员账号',
    managementPort: '管理端口',
    apiPort: 'API 端口',
    databaseType: '数据库类型',
    sqlite: 'SQLite',
    mysql: 'MySQL',
    sqliteFile: '数据库文件',
    mysqlHost: 'MySQL 主机',
    mysqlPort: 'MySQL 端口',
    mysqlDatabase: '数据库名',
    mysqlUser: 'MySQL 用户名',
    mysqlPassword: 'MySQL 密码',
    mysqlCharset: 'MySQL 字符集',
    adminUsername: '管理员用户名',
    newPassword: '新密码',
    confirmPassword: '确认密码',
    passwordsNotMatch: '两次密码输入不一致',
    saveSettings: '保存设置',
    restartSystem: '重启系统',
    restartConfirm: '确定要重启系统吗？这会断开所有当前连接。',
    restarting: '重启中...',
    settingsSaved: '设置保存成功',
    systemRestarted: '系统重启成功',
    pleaseWait: '请等待系统重启...'
  },

  // Layout
  layout: {
    sidebar: {
      dashboard: '仪表盘',
      smtp: 'SMTP',
      mail: '收件人',
      tokens: 'API 令牌',
      schedule: '定时任务',
      logs: '日志',
      settings: '设置'
    },
    tabs: {
      closeOthers: '关闭其他',
      closeAll: '关闭全部'
    },
    logoutConfirm: '确定要退出登录吗？'
  }
}
