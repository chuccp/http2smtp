import request from './request'

export function getSchedules(page: number = 1, pageSize: number = 10, filters?: { name?: string; adminOnly?: boolean }): Promise<ApiResponse<PageResponse<ScheduleConfig>>> {
  const params: Record<string, any> = { page, pageSize }
  if (filters) {
    if (filters.name) params.name = filters.name
    if (filters.adminOnly) params.adminOnly = true
  }
  return request.get('/schedule', { params })
}

export function getSchedule(id: number): Promise<ApiResponse<ScheduleConfig>> {
  return request.get(`/schedule/${id}`)
}

export function createSchedule(schedule: Partial<ScheduleConfig>): Promise<ApiResponse<any>> {
  return request.post('/schedule', schedule)
}

export function updateSchedule(schedule: ScheduleConfig): Promise<ApiResponse<any>> {
  return request.put('/schedule', schedule)
}

export function deleteSchedule(id: number): Promise<ApiResponse<any>> {
  return request.delete(`/schedule/${id}`)
}

export function triggerSendMail(id: number): Promise<ApiResponse<any>> {
  return request.post(`/schedule/trigger/${id}`)
}

export function testSchedule(schedule: Partial<ScheduleConfig>): Promise<ApiResponse<any>> {
  return request.post('/sendMailBySchedule', schedule)
}
