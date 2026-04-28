import request from './request'

export function getSMTPServers(page: number = 1, pageSize: number = 10, filters?: { adminOnly?: boolean; name?: string; host?: string; username?: string }): Promise<ApiResponse<PageResponse<SMTPConfig>>> {
  const params: Record<string, any> = { page, pageSize }
  if (filters) {
    if (filters.adminOnly) params.adminOnly = true
    if (filters.name) params.name = filters.name
    if (filters.host) params.host = filters.host
    if (filters.username) params.username = filters.username
  }
  return request.get('/smtp', { params })
}

export function getSMTP(id: number): Promise<ApiResponse<SMTPConfig>> {
  return request.get(`/smtp/${id}`)
}

export function createSMTP(server: Partial<SMTPConfig>): Promise<ApiResponse<any>> {
  return request.post('/smtp', server)
}

export function updateSMTP(server: SMTPConfig): Promise<ApiResponse<any>> {
  return request.put('/smtp', server)
}

export function deleteSMTP(id: number): Promise<ApiResponse<any>> {
  return request.delete(`/smtp/${id}`)
}

export function testSMTPConnection(server: Partial<SMTPConfig>): Promise<ApiResponse<any>> {
  return request.post('/test', server)
}

export function sendTestMail(data: { SMTPId: number; recipients: string[]; subject: string; content: string }): Promise<ApiResponse<any>> {
  return request.post('/sendMailBySMTP', data)
}
