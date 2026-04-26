export default {
  // Common
  common: {
    confirm: '確認',
    cancel: 'キャンセル',
    save: '保存',
    add: '追加',
    edit: '編集',
    delete: '削除',
    search: '検索',
    reset: 'リセット',
    action: '操作',
    operations: '操作',
    create: '作成',
    update: '更新',
    close: '閉じる',
    yes: 'はい',
    no: 'いいえ',
    loading: '読み込み中...',
    success: '成功',
    failed: '失敗',
    copied: 'クリップボードにコピーしました',
    enable: '有効',
    disable: '無効',
    status: 'ステータス',
    name: '名前',
    createTime: '作成日時',
    description: '説明'
  },

  // Auth
  auth: {
    login: 'ログイン',
    logout: 'ログアウト',
    username: 'ユーザー名',
    password: 'パスワード',
    rememberMe: 'ログイン状態を保持',
    loginSuccess: 'ログインに成功しました',
    loginFailed: 'ログインに失敗しました',
    pleaseEnterUsername: 'ユーザー名を入力してください',
    pleaseEnterPassword: 'パスワードを入力してください'
  },

  // Setup
  setup: {
    systemSetup: 'システム初期化',
    setupWizard: 'インストールウィザード',
    databaseConfig: 'データベース設定',
    databaseType: 'データベースの種類',
    sqlite: 'SQLite',
    mysql: 'MySQL',
    sqliteFile: 'データベースファイルパス',
    mysqlHost: 'MySQL ホスト',
    mysqlPort: 'MySQL ポート',
    mysqlDatabase: 'データベース名',
    mysqlUser: 'ユーザー名',
    mysqlPassword: 'パスワード',
    mysqlCharset: '文字セット',
    testConnection: '接続テスト',
    connectionSuccess: '接続に成功しました',
    connectionFailed: '接続に失敗しました',
    portConfig: 'ポート設定',
    webPort: '管理ポート',
    apiPort: 'API ポート',
    adminAccount: '管理者アカウント',
    adminUsername: '管理者ユーザー名',
    adminPassword: '管理者パスワード',
    confirmPassword: 'パスワードの確認',
    passwordsNotMatch: 'パスワードが一致しません',
    initialize: 'システムを初期化',
    initializing: '初期化中...',
    initializationSuccess: 'システムの初期化に成功しました',
    alreadyInitialized: 'システムは既に初期化されています。リダイレクト中...'
  },

  // Dashboard
  dashboard: {
    dashboard: 'ダッシュボード',
    systemInfo: 'システム情報',
    quickActions: 'クイック操作',
    statistics: '統計情報',
    version: 'バージョン',
    initialized: '初期化済み',
    notInitialized: '未初期化',
    smtpServers: 'SMTP サーバー',
    recipients: '受信者',
    tokens: 'API トークン',
    scheduledTasks: 'スケジュールタスク',
    goToSMTP: 'SMTP 管理',
    goToMail: '受信者管理',
    goToTokens: 'トークン管理',
    goToSchedule: 'スケジュール管理'
  },

  // SMTP
  smtp: {
    smtpManagement: 'SMTP 管理',
    addSMTP: 'SMTP サーバーを追加',
    editSMTP: 'SMTP サーバーを編集',
    smtpName: 'サーバー名',
    host: 'ホストアドレス',
    port: 'ポート',
    username: 'ユーザー名',
    password: 'パスワード',
    fromAddress: '送信者アドレス',
    ssl: 'SSL を使用',
    testConnection: '接続テスト',
    sendTestMail: 'テストメール送信',
    deleteConfirm: 'この SMTP サーバーを削除してもよろしいですか？',
    connectionSuccess: '接続テストに成功しました',
    connectionFailed: '接続テストに失敗しました',
    pleaseSelectRecipient: '受信者を 1 人選択してください',
    testEmailSubject: 'テストメール',
    testEmailContent: 'これは HTTP2SMTP からのテストメールです。',
    testEmailSent: 'テストメールの送信に成功しました'
  },

  // Mail
  mail: {
    mailManagement: '受信者管理',
    addRecipient: '受信者を追加',
    editRecipient: '受信者を編集',
    recipientName: '受信者名',
    emailAddress: 'メールアドレス',
    deleteConfirm: 'この受信者を削除してもよろしいですか？',
    pleaseEnterValidEmail: '有効なメールアドレスを入力してください'
  },

  // Token
  token: {
    tokenManagement: 'API トークン管理',
    generateToken: 'トークン生成',
    editToken: 'トークンを編集',
    tokenName: 'トークン名',
    tokenValue: 'トークン',
    subject: '件名',
    associatedSMTP: 'SMTP サーバー',
    allowedRecipients: '許可された受信者',
    status: 'ステータス',
    inUse: '使用中',
    userDisabled: 'ユーザーが無効化',
    adminDisabled: '管理者が無効化',
    copyToken: 'トークンをコピー',
    sendTestMail: 'テストメール送信',
    deleteConfirm: 'このトークンを削除してもよろしいですか？',
    pleaseSelectSMTP: 'SMTP サーバーを選択してください',
    pleaseSelectRecipients: '受信者を少なくとも 1 人選択してください',
    tokenGenerated: 'トークンが生成されました。すぐにコピーして保存してください：{token}'
  },

  // Schedule
  schedule: {
    scheduleManagement: 'スケジュール管理',
    addSchedule: 'スケジュールを追加',
    editSchedule: 'スケジュールを編集',
    taskName: 'タスク名',
    associatedToken: 'API トークン',
    cronExpression: 'Cron 式',
    requestUrl: 'リクエスト URL',
    requestMethod: 'リクエストメソッド',
    requestHeaders: 'リクエストヘッダー',
    requestBody: 'リクエストボディ',
    useTemplate: 'テンプレートを使用',
    template: 'テンプレート内容',
    status: 'ステータス',
    enabled: '有効',
    disabled: '無効',
    triggerNow: '今すぐ実行',
    deleteConfirm: 'このスケジュールを削除してもよろしいですか？',
    triggered: 'タスクの実行に成功しました',
    testSchedule: 'テスト',
    testSuccess: 'テストメールの送信に成功しました',
    cronHelp: 'Cron 式の形式: 秒 分 時 日 月 週',
    exampleCron: '0 0 9 * * * = 毎日午前 9 時'
  },

  // Log
  log: {
    systemLogs: 'システムログ',
    searchKeyword: 'キーワードで検索...',
    token: 'トークン',
    subject: '件名',
    content: '内容',
    result: '結果',
    logDetail: 'ログ詳細',
    status: 'ステータス',
    createTime: '作成日時',
    smtpServer: 'SMTP サーバー',
    recipients: '受信者',
    attachments: '添付ファイル',
    download: 'ダウンロード',
    noResults: 'ログが見つかりません',
    clear: 'クリア',
    success: '成功',
    error: 'エラー',
    warning: '警告'
  },

  // Settings
  settings: {
    systemSettings: 'システム設定',
    generalSettings: '一般設定',
    databaseConfig: 'データベース設定',
    adminAccount: '管理者アカウント',
    managementPort: '管理ポート',
    apiPort: 'API ポート',
    databaseType: 'データベースの種類',
    sqlite: 'SQLite',
    mysql: 'MySQL',
    sqliteFile: 'データベースファイル',
    mysqlHost: 'MySQL ホスト',
    mysqlPort: 'MySQL ポート',
    mysqlDatabase: 'データベース名',
    mysqlUser: 'MySQL ユーザー名',
    mysqlPassword: 'MySQL パスワード',
    mysqlCharset: 'MySQL 文字セット',
    adminUsername: '管理者ユーザー名',
    newPassword: '新しいパスワード',
    confirmPassword: 'パスワードの確認',
    passwordsNotMatch: 'パスワードが一致しません',
    saveSettings: '設定を保存',
    restartSystem: 'システムを再起動',
    restartConfirm: 'システムを再起動してもよろしいですか？現在のすべての接続が切断されます。',
    restarting: '再起動中...',
    settingsSaved: '設定が保存されました',
    systemRestarted: 'システムの再起動に成功しました',
    pleaseWait: 'システムの再起動をお待ちください...'
  },

  // Layout
  layout: {
    sidebar: {
      dashboard: 'ダッシュボード',
      smtp: 'SMTP',
      mail: '受信者',
      tokens: 'API トークン',
      schedule: 'スケジュール',
      logs: 'ログ',
      settings: '設定',
      users: 'ユーザー管理'
    },
    tabs: {
      closeOthers: '他を閉じる',
      closeAll: 'すべて閉じる'
    },
    logoutConfirm: 'ログアウトしてもよろしいですか？'
  },

  // User
  user: {
    userManagement: 'ユーザー管理',
    addUser: 'ユーザーを追加',
    editUser: 'ユーザーを編集',
    username: 'ユーザー名',
    isAdmin: '管理者',
    admin: '管理者',
    normalUser: '一般ユーザー',
    passwordPlaceholder: '変更しない場合は空欄',
    deleteConfirm: 'このユーザーを削除してもよろしいですか？'
  }
}
