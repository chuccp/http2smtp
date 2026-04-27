import request from './request'

export function checkInitStatus(): Promise<ApiResponse<SystemInfo>> {
  return request.get('/set').then(res => {
    return {
      code: res.code,
      msg: res.msg,
      data: {
        initialized: res.data?.hasInit ?? false,
        dbInitialized: res.data?.hasDbInit ?? false,
        hasAdmin: res.data?.hasAdmin ?? false,
        isDocker: res.data?.isDocker ?? false,
        version: ''
      }
    }
  })
}

export function getDefaultSettings(): Promise<ApiResponse<SetInfo>> {
  return request.get('/readSet').then(res => {
    const d = res.data
    return {
      code: res.code,
      msg: res.msg,
      data: {
        webPort: d?.manage?.port ?? 12566,
        apiPort: d?.api?.port ?? 12567,
        dbType: d?.dbType ?? 'sqlite',
        dbHost: d?.mysql?.host ?? '127.0.0.1',
        dbPort: d?.mysql?.port ?? 3306,
        dbName: d?.mysql?.dbname ?? '',
        dbUser: d?.mysql?.username ?? '',
        dbPass: d?.mysql?.password ?? '',
        dbCharset: d?.mysql?.charset ?? 'utf8mb4',
        dbFile: d?.sqlite?.filename ?? 'data.db',
        adminUser: d?.manage?.username ?? '',
        adminPass: d?.manage?.password ?? ''
      }
    }
  })
}

export function testConnection(settings: SetInfo): Promise<ApiResponse<any>> {
  return request.put('/reSet', {
    dbType: settings.dbType,
    sqlite: { filename: settings.dbFile },
    mysql: {
      host: settings.dbHost,
      port: settings.dbPort,
      dbname: settings.dbName,
      username: settings.dbUser,
      password: settings.dbPass,
      charset: settings.dbCharset
    },
    manage: {
      port: settings.webPort,
      username: settings.adminUser,
      password: settings.adminPass
    },
    api: { port: settings.apiPort }
  })
}

// Step 1: Initialize database connection
export function initDatabase(settings: SetInfo): Promise<ApiResponse<any>> {
  return request.put('/dbInit', {
    dbType: settings.dbType,
    sqlite: { filename: settings.dbFile },
    mysql: {
      host: settings.dbHost,
      port: settings.dbPort,
      dbname: settings.dbName,
      username: settings.dbUser,
      password: settings.dbPass,
      charset: settings.dbCharset
    },
    manage: {
      port: settings.webPort
    },
    api: { port: settings.apiPort }
  })
}

// Step 2: Initialize admin account
export function initAdmin(username: string, password: string): Promise<ApiResponse<any>> {
  return request.put('/adminInit', {
    username,
    password
  })
}

// Step 2 alternative: Skip admin creation (admin already exists)
export function skipAdmin(): Promise<ApiResponse<any>> {
  return request.put('/adminSkip', {})
}

// Check if an admin user already exists
export function checkAdminExists(): Promise<ApiResponse<{ hasAdmin: boolean; adminName: string }>> {
  return request.get('/adminExists')
}
