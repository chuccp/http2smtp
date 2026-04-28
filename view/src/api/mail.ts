import request from './request'

export function getMails(page: number = 1, pageSize: number = 10, filters?: { name?: string; mail?: string; adminOnly?: boolean }): Promise<ApiResponse<PageResponse<MailConfig>>> {
  const params: Record<string, any> = { page, pageSize }
  if (filters) {
    if (filters.name) params.name = filters.name
    if (filters.mail) params.mail = filters.mail
    if (filters.adminOnly) params.adminOnly = true
  }
  return request.get('/mail', { params })
}

export function getMail(id: number): Promise<ApiResponse<MailConfig>> {
  return request.get(`/mail/${id}`)
}

export function createMail(mail: Partial<MailConfig>): Promise<ApiResponse<any>> {
  return request.post('/mail', mail)
}

export function updateMail(mail: MailConfig): Promise<ApiResponse<any>> {
  return request.put('/mail', mail)
}

export function deleteMail(id: number): Promise<ApiResponse<any>> {
  return request.delete(`/mail/${id}`)
}
