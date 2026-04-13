import request from './request'

export function getSystemInfo(): Promise<ApiResponse<any>> {
  return request.get('/set').then(res => {
    return {
      code: res.code,
      msg: res.msg,
      data: {
        initialized: res.data?.hasInit ?? false,
        hasLogin: res.data?.hasLogin,
        isDocker: res.data?.isDocker
      }
    }
  })
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
