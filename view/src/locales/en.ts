export default {
  // Common
  common: {
    confirm: 'Confirm',
    cancel: 'Cancel',
    save: 'Save',
    add: 'Add',
    edit: 'Edit',
    delete: 'Delete',
    search: 'Search',
    reset: 'Reset',
    action: 'Action',
    operations: 'Operations',
    create: 'Create',
    update: 'Update',
    close: 'Close',
    yes: 'Yes',
    no: 'No',
    loading: 'Loading...',
    success: 'Success',
    failed: 'Failed',
    copied: 'Copied to clipboard',
    enable: 'Enable',
    disable: 'Disable',
    status: 'Status',
    name: 'Name',
    createTime: 'Create Time',
    description: 'Description'
  },

  // Auth
  auth: {
    login: 'Login',
    logout: 'Logout',
    username: 'Username',
    password: 'Password',
    rememberMe: 'Remember me',
    loginSuccess: 'Login successful',
    loginFailed: 'Login failed',
    pleaseEnterUsername: 'Please enter username',
    pleaseEnterPassword: 'Please enter password'
  },

  // Setup
  setup: {
    systemSetup: 'System Initialization',
    setupWizard: 'Setup Wizard',
    databaseConfig: 'Database Configuration',
    databaseType: 'Database Type',
    sqlite: 'SQLite',
    mysql: 'MySQL',
    sqliteFile: 'Database File Path',
    mysqlHost: 'MySQL Host',
    mysqlPort: 'MySQL Port',
    mysqlDatabase: 'Database Name',
    mysqlUser: 'Username',
    mysqlPassword: 'Password',
    mysqlCharset: 'Charset',
    testConnection: 'Test Connection',
    connectionSuccess: 'Connection successful',
    connectionFailed: 'Connection failed',
    portConfig: 'Port Configuration',
    webPort: 'Management Port',
    apiPort: 'API Port',
    adminAccount: 'Administrator Account',
    adminUsername: 'Admin Username',
    adminPassword: 'Admin Password',
    confirmPassword: 'Confirm Password',
    passwordsNotMatch: 'Passwords do not match',
    initialize: 'Initialize System',
    initializing: 'Initializing...',
    initializationSuccess: 'System initialized successfully',
    alreadyInitialized: 'System is already initialized, redirecting...'
  },

  // Dashboard
  dashboard: {
    dashboard: 'Dashboard',
    systemInfo: 'System Information',
    quickActions: 'Quick Actions',
    statistics: 'Statistics',
    version: 'Version',
    initialized: 'Initialized',
    notInitialized: 'Not Initialized',
    smtpServers: 'SMTP Servers',
    recipients: 'Recipients',
    tokens: 'API Tokens',
    scheduledTasks: 'Scheduled Tasks',
    goToSMTP: 'Manage SMTP',
    goToMail: 'Manage Recipients',
    goToTokens: 'Manage Tokens',
    goToSchedule: 'Manage Schedules'
  },

  // SMTP
  smtp: {
    smtpManagement: 'SMTP Management',
    addSMTP: 'Add SMTP Server',
    editSMTP: 'Edit SMTP Server',
    smtpName: 'Server Name',
    host: 'Host',
    port: 'Port',
    username: 'Username',
    password: 'Password',
    fromAddress: 'From Address',
    ssl: 'Use SSL',
    testConnection: 'Test Connection',
    sendTestMail: 'Send Test Email',
    deleteConfirm: 'Are you sure you want to delete this SMTP server?',
    connectionSuccess: 'Connection test successful',
    connectionFailed: 'Connection test failed',
    pleaseSelectRecipient: 'Please select a recipient',
    testEmailSubject: 'Test Email',
    testEmailContent: 'This is a test email from HTTP2SMTP.',
    testEmailSent: 'Test email sent successfully'
  },

  // Mail
  mail: {
    mailManagement: 'Recipient Management',
    addRecipient: 'Add Recipient',
    editRecipient: 'Edit Recipient',
    recipientName: 'Recipient Name',
    emailAddress: 'Email Address',
    deleteConfirm: 'Are you sure you want to delete this recipient?',
    pleaseEnterValidEmail: 'Please enter a valid email address'
  },

  // Token
  token: {
    tokenManagement: 'API Token Management',
    generateToken: 'Generate Token',
    editToken: 'Edit Token',
    tokenName: 'Token Name',
    tokenValue: 'Token',
    associatedSMTP: 'SMTP Server',
    allowedRecipients: 'Allowed Recipients',
    status: 'Status',
    enabled: 'Enabled',
    disabled: 'Disabled',
    copyToken: 'Copy Token',
    sendTestMail: 'Send Test Email',
    deleteConfirm: 'Are you sure you want to delete this token?',
    pleaseSelectSMTP: 'Please select an SMTP server',
    pleaseSelectRecipients: 'Please select at least one recipient',
    tokenGenerated: 'Token generated, please copy it now: {token}'
  },

  // Schedule
  schedule: {
    scheduleManagement: 'Scheduled Tasks',
    addSchedule: 'Add Scheduled Task',
    editSchedule: 'Edit Scheduled Task',
    taskName: 'Task Name',
    associatedToken: 'API Token',
    cronExpression: 'Cron Expression',
    requestUrl: 'Request URL',
    requestMethod: 'Request Method',
    requestHeaders: 'Request Headers (JSON format)',
    requestBody: 'Request Body',
    useTemplate: 'Use Template',
    status: 'Status',
    enabled: 'Enabled',
    disabled: 'Disabled',
    triggerNow: 'Trigger Now',
    deleteConfirm: 'Are you sure you want to delete this scheduled task?',
    triggered: 'Task triggered successfully',
    cronHelp: 'Cron expression: * * * * * = minute hour day month weekday',
    exampleCron: '0 9 * * * = 9 AM every day'
  },

  // Log
  log: {
    systemLogs: 'System Logs',
    searchKeyword: 'Search keyword...',
    token: 'Token',
    subject: 'Subject',
    content: 'Content',
    result: 'Result',
    logDetail: 'Log Detail',
    status: 'Status',
    createTime: 'Create Time',
    smtpServer: 'SMTP Server',
    recipients: 'Recipients',
    attachments: 'Attachments',
    download: 'Download',
    noResults: 'No logs found',
    clear: 'Clear',
    success: 'Success',
    error: 'Error',
    warning: 'Warning'
  },

  // Settings
  settings: {
    systemSettings: 'System Settings',
    generalSettings: 'General Settings',
    databaseConfig: 'Database Configuration',
    adminAccount: 'Admin Account',
    managementPort: 'Management Port',
    apiPort: 'API Port',
    databaseType: 'Database Type',
    sqlite: 'SQLite',
    mysql: 'MySQL',
    sqliteFile: 'Database File',
    mysqlHost: 'MySQL Host',
    mysqlPort: 'MySQL Port',
    mysqlDatabase: 'Database Name',
    mysqlUser: 'MySQL Username',
    mysqlPassword: 'MySQL Password',
    mysqlCharset: 'MySQL Charset',
    adminUsername: 'Admin Username',
    newPassword: 'New Password',
    confirmPassword: 'Confirm Password',
    passwordsNotMatch: 'Passwords do not match',
    saveSettings: 'Save Settings',
    restartSystem: 'Restart System',
    restartConfirm: 'Are you sure you want to restart the system? This will disconnect all current connections.',
    restarting: 'Restarting...',
    settingsSaved: 'Settings saved successfully',
    systemRestarted: 'System restarted successfully',
    pleaseWait: 'Please wait for the system to restart...'
  },

  // Layout
  layout: {
    sidebar: {
      dashboard: 'Dashboard',
      smtp: 'SMTP',
      mail: 'Recipients',
      tokens: 'API Tokens',
      schedule: 'Schedule',
      logs: 'Logs',
      settings: 'Settings'
    },
    tabs: {
      closeOthers: 'Close Others',
      closeAll: 'Close All'
    }
  }
}
