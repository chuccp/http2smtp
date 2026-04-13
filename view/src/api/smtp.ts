import request from './request'

export function getSMTPServers(page: number = 1, pageSize: number = 10): Promise<ApiResponse<PageResponse<SMTPConfig>>> {
  return request.get('/smtp', {
    params: {
      page,
      pageSize
    }
  })
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

export function sendTestMail(data: { smtpId: number; toEmail: string; subject: string; content: string }): Promise<ApiResponse<any>> {
  return request.post('/sendMailBySMTP', data)
}
