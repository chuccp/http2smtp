import request from './request'

export function checkInitStatus(): Promise<ApiResponse<SystemInfo>> {
  return request.get('/set')
}

export function getDefaultSettings(): Promise<ApiResponse<SetInfo>> {
  return request.get('/readSet')
}

export function testConnection(settings: SetInfo): Promise<ApiResponse<any>> {
  return request.put('/reSet', settings)
}

export function initializeSystem(settings: SetInfo): Promise<ApiResponse<any>> {
  return request.put('/reSet', settings)
}
