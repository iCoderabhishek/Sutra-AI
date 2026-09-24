import axios, { AxiosError } from "axios";

export function normalizeError(err: unknown): {
  message: string;
  status?: number;
  code?: string;
} {
  if (axios.isAxiosError(err)) {
    const status = err.response?.status;
    const responseData = err.response?.data as any;
    
    // Prefer the backend's explicit error message if present
    const serverMessage = responseData?.error || responseData?.message;

    if (status === 400) {
      return {
        message: serverMessage || "Please check the highlighted fields.",
        status,
      };
    }
    if (status === 401) {
      return {
        message: "Your session expired. Please sign in again.",
        status,
      };
    }
    if (status === 403) {
      return {
        message: serverMessage || "Only the workspace owner can do this.",
        status,
      };
    }
    if (status === 404) {
      return {
        message: "We couldn't find that. It may have been removed.",
        status,
      };
    }
    if (status === 413) {
      if (responseData?.used && responseData?.limit) {
        return {
          message: `Workspace quota exceeded.`,
          status,
        };
      }
      return {
        message: "That exceeds the size limit.",
        status,
      };
    }
    if (status === 415) {
      return {
        message: "That file type isn't supported.",
        status,
      };
    }
    if (status === 429) {
      return {
        message: "Too many requests. Please wait a moment and try again.",
        status,
      };
    }

    return {
      message: serverMessage || "Something went wrong. Please try again.",
      status,
    };
  }

  if (err instanceof Error) {
    return { message: err.message };
  }

  return { message: "Something went wrong. Please try again." };
}
