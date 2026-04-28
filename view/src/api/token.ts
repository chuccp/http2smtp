import request from './request'

export function getTokens(page: number = 1, pageSize: number = 10, filters?: { name?: string }): Promise<ApiResponse<PageResponse<TokenConfig>>> {
  const params: Record<string, any> = { page, pageSize }
  if (filters) {
    if (filters.name) params.name = filters.name
  }
  return request.get('/token', { params })
}

export function getToken(id: number): Promise<ApiResponse<TokenConfig>> {
  return request.get(`/token/${id}`)
}

export function createToken(token: Partial<TokenConfig>): Promise<ApiResponse<any>> {
  return request.post('/token', token)
}

export function updateToken(token: TokenConfig): Promise<ApiResponse<any>> {
  return request.put('/token', token)
}

export function deleteToken(id: number): Promise<ApiResponse<any>> {
  return request.delete(`/token/${id}`)
}

export function sendMailByToken(data: { tokenId: number; recipients: string[]; subject: string; content: string }): Promise<ApiResponse<any>> {
  return request.post('/sendMailByTokenId', data)
}
