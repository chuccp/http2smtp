import request from './request'

export function getSystemInfo(): Promise<ApiResponse<SystemInfo>> {
  return request.get('/set')
}

export function login(username: string, nonce: string, response: string): Promise<ApiResponse<any>> {
  return request.post('/signIn', {
    username,
    nonce,
    response
  })
}

export function logout(): Promise<ApiResponse<any>> {
  return request.post('/logout')
}
