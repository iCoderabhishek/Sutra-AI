import axios from "axios";

const rawBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "") || "http://localhost:8080";
export const API_BASE_URL = rawBaseUrl.endsWith("/api/v1") ? rawBaseUrl : `${rawBaseUrl}/api/v1`;

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    // 401 unauthenticated interceptor
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      // If we are not already on the landing page or an invite link, redirect
      if (
        typeof window !== "undefined" &&
        window.location.pathname !== "/" &&
        !window.location.pathname.startsWith("/invite")
      ) {
        try {
          await axios.post(`${API_BASE_URL}/auth/logout`, {}, { withCredentials: true });
        } catch (err) {
          console.error("Logout failed during 401 redirect", err);
        }
        window.location.href = "/";
      }
    }
    return Promise.reject(error);
  }
);
