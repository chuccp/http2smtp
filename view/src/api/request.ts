import axios, { type AxiosRequestConfig, type InternalAxiosRequestConfig, type AxiosResponse } from 'axios'

// Override the AxiosInstance type to reflect that our response interceptor
// unwraps response.data, so callers receive the raw API payload, not AxiosResponse
declare module 'axios' {
  interface AxiosInstance {
    request<T = any>(config: AxiosRequestConfig): Promise<T>
    get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
    delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
    post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
    put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
    patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  }
}
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || (import.meta.env.DEV ? 'http://127.0.0.1:12566/api' : '/api'),
  timeout: 10000,
  withCredentials: true
})

// 处理未登录状态，跳转到登录页
let isRedirecting = false
const handleUnauthorized = () => {
  if (isRedirecting) return
  isRedirecting = true
  localStorage.removeItem('http2smtp-token')
  localStorage.removeItem('http2smtp-username')
  ElMessage.error('登录已过期，请重新登录')
  window.location.href = '/login'
}

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('http2smtp-token')
    if (token) {
      config.headers.Authorization = token
    }
    return config
  },
  (error: any) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    // 如果是 blob 类型响应，直接返回
    if (response.config.responseType === 'blob') {
      return response.data
    }

    const res = response.data
    if (res.code === 0 || res.code === 200) {
      return res
    } else {
      ElMessage.error(res.msg || res.message || '请求失败')
      return Promise.reject(new Error(res.msg || res.message || '请求失败'))
    }
  },
  (error: any) => {
    // 处理 HTTP 错误状态码
    if (error.response) {
      const status = error.response.status
      const data = error.response.data

      // 尝试获取后端返回的错误消息
      const errorMsg = data?.msg || data?.message || `请求失败(${status})`

      if (status === 401) {
        handleUnauthorized()
      } else if (status >= 500) {
        ElMessage.error(errorMsg)
      } else {
        ElMessage.error(errorMsg)
      }
    } else {
      ElMessage.error(error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

export default request
