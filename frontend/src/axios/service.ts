import axios, { AxiosError } from 'axios'
import { defaultRequestInterceptors, defaultResponseInterceptors } from './config'

import { AxiosInstance, RequestConfig } from './types'
import { ElMessage } from 'element-plus'
import { REQUEST_TIMEOUT } from '@/constants'
import { useUserStoreWithOut } from '@/store/modules/user'

export const PATH_URL = import.meta.env.VITE_API_BASE_PATH

const abortControllerMap: Map<string, Set<AbortController>> = new Map()

const axiosInstance: AxiosInstance = axios.create({
  timeout: REQUEST_TIMEOUT,
  baseURL: PATH_URL
})

axiosInstance.interceptors.response.use(undefined, (error: AxiosError) => {
  // Cancelled navigation and user actions should not show request errors.
  if (axios.isCancel(error)) return Promise.reject(error)
  if (error.response?.data) {
    const responseData = error.response.data as { message: string; data: string }
    // 显示后端返回的具体错误消息
    const errorMsg =
      responseData.data && responseData.data !== ''
        ? responseData.message + ':' + responseData.data
        : responseData.message
    ElMessage.error(errorMsg || 'Request failed')
    if (error.response?.status === 401) {
      const userStore = useUserStoreWithOut()
      userStore.logout()
    }
  } else {
    // 如果没有响应数据，显示网络错误或其他错误
    ElMessage.error(error.message || 'NetworkError')
  }
  return Promise.reject(error)
})

axiosInstance.interceptors.request.use(defaultRequestInterceptors)
axiosInstance.interceptors.response.use(defaultResponseInterceptors)

const service = {
  request: async (config: RequestConfig) => {
    if (config.interceptors?.requestInterceptors) {
      config = config.interceptors.requestInterceptors(config as any)
    }

    // Track each concurrent request under its original URL.
    const url = config.url || ''
    const controller = new AbortController()
    const controllers = abortControllerMap.get(url) ?? new Set<AbortController>()
    controllers.add(controller)
    abortControllerMap.set(url, controllers)

    const signal = config.signal
    const abort = () => controller.abort()
    if (signal?.aborted) abort()
    else signal?.addEventListener?.('abort', abort)

    try {
      return await axiosInstance.request({ ...config, signal: controller.signal })
    } finally {
      // Release tracking even when a request fails or is cancelled.
      signal?.removeEventListener?.('abort', abort)
      controllers.delete(controller)
      if (controllers.size === 0) abortControllerMap.delete(url)
    }
  },
  cancelRequest: (url: string | string[]) => {
    const urlList = Array.isArray(url) ? url : [url]
    for (const _url of urlList) {
      for (const controller of abortControllerMap.get(_url) ?? []) controller.abort()
    }
  },
  cancelAllRequest() {
    for (const controllers of abortControllerMap.values()) {
      for (const controller of controllers) controller.abort()
    }
  }
}

export default service
