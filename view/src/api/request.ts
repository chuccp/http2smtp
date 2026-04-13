import axios, { type InternalAxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
  withCredentials: true
})

// 处理未登录状态，跳转到登录页
const handleUnauthorized = () => {
  localStorage.removeItem('http2smtp-token')
  localStorage.removeItem('http2smtp-username')
  router.push('/')
  ElMessage.error('登录已过期，请重新登录')
}

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
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
