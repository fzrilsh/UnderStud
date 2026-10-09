import axios from 'axios'

/**
 * Axios instance untuk semua call API.
 * - baseURL '/api/v1': dev = Vite proxy, docker = nginx proxy (selalu same-origin).
 * - withCredentials: cookie auth_token httpOnly ikut otomatis — tidak ada token di JS.
 */
export const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
})

// Response interceptor: 401 (dan bukan dari endpoint auth login/register)
// berarti cookie invalid/expired → clear auth state & redirect ke /login.
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    const url: string = error.config?.url ?? ''
    const isAuthEndpoint = url.startsWith('/auth/login') || url.startsWith('/auth/register')
    if (status === 401 && !isAuthEndpoint) {
      window.dispatchEvent(new CustomEvent('auth:unauthorized'))
    }
    return Promise.reject(error)
  },
)
