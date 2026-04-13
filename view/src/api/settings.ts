import request from './request'

export function getSettings(): Promise<ApiResponse<SetInfo>> {
  return request.get('/readSet')
}

export function updateSettings(settings: SetInfo): Promise<ApiResponse<any>> {
  return request.put('/reSet', settings)
}

export function restartSystem(): Promise<ApiResponse<any>> {
  return request.post('/reStart')
}
