import request from './request'

export function getSchedules(page: number = 1, pageSize: number = 10): Promise<ApiResponse<PageResponse<ScheduleConfig>>> {
  return request.get('/schedule', {
    params: {
      page,
      pageSize
    }
  })
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
