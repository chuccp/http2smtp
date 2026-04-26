import request from './request'

export function getUsers(page: number = 1, pageSize: number = 10): Promise<ApiResponse<PageResponse<UserConfig>>> {
  return request.get('/user', {
    params: {
      page,
      pageSize
    }
  })
}

export function getUser(id: number): Promise<ApiResponse<UserConfig>> {
  return request.get(`/user/${id}`)
}

export function createUser(user: Partial<UserConfig>): Promise<ApiResponse<any>> {
  return request.post('/user', user)
}

export function updateUser(user: Partial<UserConfig>): Promise<ApiResponse<any>> {
  return request.put('/user', user)
}

export function deleteUser(id: number): Promise<ApiResponse<any>> {
  return request.delete(`/user/${id}`)
}
