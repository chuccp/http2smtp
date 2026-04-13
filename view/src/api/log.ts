import request from './request'

export function getLogs(page: number = 1, pageSize: number = 10, searchKey: string = ''): Promise<ApiResponse<PageResponse<LogEntry>>> {
  return request.get('/log', {
    params: {
      page,
      pageSize,
      searchKey
    }
  })
}

export function getLog(id: number): Promise<ApiResponse<LogEntry>> {
  return request.get(`/log/${id}`)
}
