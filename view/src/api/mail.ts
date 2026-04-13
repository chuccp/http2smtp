import request from './request'

export function getMails(page: number = 1, pageSize: number = 10): Promise<ApiResponse<PageResponse<MailConfig>>> {
  return request.get('/mail', {
    params: {
      page,
      pageSize
    }
  })
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
