import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

/**
 * 后端统一响应格式
 * 对应后端 pkg/response/response.go 的 Response 结构
 */
interface ApiResponse<T = any> {
    /** 状态码：0表示成功，其他表示错误 */
    code: number
    /** 响应消息 */
    message: string
    /** 响应数据 */
    data: T
}

/**
 * 创建axios实例
 * 基础配置：基础URL、超时时间、请求头
 */
const service: AxiosInstance = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || '',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

/**
 * 请求拦截器
 * 功能：自动添加JWT令牌到请求头
 */
service.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem('access_token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => {
        console.error('请求错误:', error)
        return Promise.reject(error)
    }
)

/**
 * 响应拦截器
 * 功能：
 * 1. 统一处理后端响应格式（code: 0表示成功）
 * 2. 自动提取data字段
 * 3. 处理业务错误和网络错误
 * 4. Token失效自动跳转登录
 */
service.interceptors.response.use(
    (response) => {
        const res: ApiResponse = response.data

        // 开发环境打印API响应（生产环境不打印）
        if (import.meta.env.DEV) {
            console.log('API响应:', {
                url: response.config.url,
                code: res.code,
                message: res.message
            })
        }

        // 成功响应：直接返回data字段
        if (res.code === 0) {
            return res.data
        }

        // 业务错误处理
        const errorMsg = res.message || '请求失败'
        ElMessage.error(errorMsg)

        // Token失效错误码：清除本地存储并跳转登录
        // 对应后端 pkg/errors/code.go 的错误码定义
        if ([1002, 4001, 4002, 4003].includes(res.code)) {
            localStorage.clear()
            window.location.href = '/login'
        }

        return Promise.reject(new Error(errorMsg))
    },
    (error) => {
        // 网络错误或HTTP状态码错误
        console.error('响应错误:', {
            message: error.message,
            status: error.response?.status,
            url: error.config?.url
        })

        const errorMessage = error.response?.data?.message || error.message || '网络错误'
        ElMessage.error(errorMessage)
        return Promise.reject(error)
    }
)

/**
 * 导出的request函数
 * @param config Axios请求配置
 * @returns Promise<T> 直接返回data字段的类型
 */
export function request<T = any>(config: AxiosRequestConfig): Promise<T> {
    return service.request(config) as Promise<T>
}

export default service
