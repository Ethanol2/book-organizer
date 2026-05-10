import axios from 'axios';

const api = axios.create({
    baseURL: '',
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json',
    }
});

export default api

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // If 401 Unauthorized and we haven't tried to refresh yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
        // This call automatically sends the HttpOnly refresh_token cookie
        await axios.post('/api/auth/refresh', {}, { withCredentials: true });
        
        // Retry the original request now that cookies are updated
        return api(originalRequest);
      } catch (refreshError) {
        // Refresh failed (token truly expired/invalid)
        window.location.reload();
        return Promise.reject(refreshError);
      }
    }
    return Promise.reject(error);
  }
);
