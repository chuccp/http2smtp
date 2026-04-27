import request from './request'

export function getSystemInfo(): Promise<ApiResponse<any>> {
  return request.get('/set').then(res => {
    return {
      code: res.code,
      msg: res.msg,
      data: {
        initialized: res.data?.hasInit ?? false,
        dbInitialized: res.data?.hasDbInit ?? false,
        hasAdmin: res.data?.hasAdmin ?? false,
        hasLogin: res.data?.hasLogin,
        isDocker: res.data?.isDocker
      }
    }
  })
}

export function login(username: string, password: string): Promise<ApiResponse<any>> {
  return request.post('/signIn', {
    username,
    password
  })
}

export function logout(): Promise<ApiResponse<any>> {
  return request.post('/logout')
}
