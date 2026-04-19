import request from './request'

export function getTokens(page: number = 1, pageSize: number = 10): Promise<ApiResponse<PageResponse<TokenConfig>>> {
  return request.get('/token', {
    params: {
      page,
      pageSize
    }
  })
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
