import { Platform } from "react-native";
import { emitAuthSessionExpired } from "./authSession";
import { deepCamelToSnake, deepSnakeToCamel } from "./caseConverter";
import { secureStorage } from "./secureStorage";
import type {
  AddDeviceRequest,
  AuditEvent,
  AuthResponse,
  Bill,
  BillCategory,
  ChatMessage,
  CreateChatMessageForm,
  CreateBillForm,
  CreateCategoryForm,
  CreateItemForm,
  CreateItemsForm,
  CreateNoteCategoryForm,
  CreateNoteForm,
  CreatePollForm,
  CreateScheduleForm,
  CreateTaskForm,
  HAState,
  Home,
  HomeMembership,
  HomeNotification,
  Note,
  NoteCategory,
  Notification,
  OCRResult,
  Poll,
  Room,
  ShoppingCategory,
  ShoppingItem,
  SmartDevice,
  Task,
  TaskAssignment,
  TaskSchedule,
  UpdateChatMessageForm,
  UpdateDeviceRequest,
  UpdateNoteCategoryForm,
  UpdateNoteForm,
  User,
} from "./types";

export const isDevMode = process.env.EXPO_PUBLIC_NODE_ENV !== "prod";
const API_BASE_URL = isDevMode ? process.env.EXPO_PUBLIC_API_URL_DEV : process.env.EXPO_PUBLIC_API_URL;
const API_PREFIX = "/api";

type ApiRequestOptions = {
  params?: Record<string, unknown>;
  headers?: Record<string, string>;
  body?: unknown;
  rawBody?: BodyInit;
  method?: string;
};

type ApiResponse<T> = {
  data: T;
  status: number;
};

type ApiErrorResponse = {
  status: number;
  data: unknown;
};

export type ApiError = Error & {
  status?: number;
  response?: ApiErrorResponse;
};

function buildUrl(path: string, params?: Record<string, unknown>) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  const url = new URL(`${API_PREFIX}${normalizedPath}`, API_BASE_URL || "http://localhost");

  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null) continue;
      if (Array.isArray(value)) {
        for (const item of value) url.searchParams.append(key, String(item));
        continue;
      }
      url.searchParams.set(key, String(value));
    }
  }

  if (API_BASE_URL) {
    return url.toString();
  }

  return `${url.pathname}${url.search}`;
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value) && !(value instanceof FormData);
}

async function getAuthHeaders() {
  const token = await secureStorage.getItem("auth_token");
  return {
    token,
    headers: token ? { Authorization: `Bearer ${token}` } : null,
  };
}

async function request<T>(path: string, options: ApiRequestOptions = {}): Promise<ApiResponse<T>> {
  const { token, headers: authHeaders } = await getAuthHeaders();
  const headers: Record<string, string> = {
    ...(authHeaders ?? {}),
    ...(options.headers ?? {}),
  };

  let body: BodyInit | undefined = options.rawBody;

  if (!body && options.body !== undefined) {
    if (options.body instanceof FormData || options.body instanceof URLSearchParams) {
      body = options.body;
      if (options.body instanceof FormData) {
        delete headers["Content-Type"];
      }
    } else if (typeof options.body === "string") {
      body = options.body;
    } else if (isPlainObject(options.body) || Array.isArray(options.body)) {
      body = JSON.stringify(deepCamelToSnake(options.body));
      if (!headers["Content-Type"]) {
        headers["Content-Type"] = "application/json";
      }
    } else {
      body = String(options.body);
    }
  }

  const response = await fetch(buildUrl(path, options.params), {
    method: options.method ?? "GET",
    headers,
    body,
    credentials: Platform.OS === "web" ? "include" : "same-origin",
  });

  const text = await response.text();
  const contentType = response.headers.get("content-type") ?? "";
  const data = text
    ? contentType.includes("application/json")
      ? deepSnakeToCamel(JSON.parse(text))
      : (text as unknown)
    : (null as unknown);

  if (!response.ok) {
    if (response.status === 401 && token) {
      await secureStorage.removeItem("auth_token");
      await secureStorage.removeItem("user");
      emitAuthSessionExpired();
    }

    const error = new Error(`Request failed with status ${response.status}`) as ApiError;
    error.status = response.status;
    error.response = {
      status: response.status,
      data,
    };
    throw error;
  }

  return { data: data as T, status: response.status };
}

export const api = {
  get: <T>(path: string, config?: { params?: Record<string, unknown> }) =>
    request<T>(path, { method: "GET", params: config?.params }),
  post: <T>(path: string, body?: unknown, config?: { headers?: Record<string, string> }) =>
    request<T>(path, { method: "POST", body, headers: config?.headers }),
  put: <T>(path: string, body?: unknown, config?: { headers?: Record<string, string> }) =>
    request<T>(path, { method: "PUT", body, headers: config?.headers }),
  patch: <T>(path: string, body?: unknown, config?: { headers?: Record<string, string> }) =>
    request<T>(path, { method: "PATCH", body, headers: config?.headers }),
  delete: <T>(path: string, config?: { headers?: Record<string, string> }) =>
    request<T>(path, { method: "DELETE", headers: config?.headers }),
};

export function getApiErrorStatus(error: unknown): number | undefined {
  if (typeof error !== "object" || error === null) return undefined;
  if ("status" in error && typeof (error as { status?: unknown }).status === "number") {
    return (error as { status: number }).status;
  }
  if ("response" in error && typeof (error as { response?: { status?: unknown } }).response?.status === "number") {
    return (error as { response: { status: number } }).response.status;
  }
  return undefined;
}

export function getApiErrorMessage(error: unknown, fallback: string): string {
  if (typeof error !== "object" || error === null) return fallback;
  const responseData = (error as { response?: { data?: unknown } }).response?.data;
  if (typeof responseData === "object" && responseData !== null && "error" in responseData) {
    const message = (responseData as { error?: unknown }).error;
    if (typeof message === "string" && message) return message;
  }
  return fallback;
}

// ============ Auth API ============
export const authApi = {
  register: async (email: string, password: string, name: string, username: string) => {
    const response = await api.post("/auth/register", { email, password, name, username });
    return response.data;
  },

  login: async (email: string, password: string): Promise<AuthResponse> => {
    const response = await api.post<AuthResponse>("/auth/login", { email, password });
    if (response.data.token) {
      await secureStorage.setItem("auth_token", response.data.token);
      await secureStorage.setItem("user", JSON.stringify(response.data.user));
    }
    return response.data;
  },

  verifyEmail: async (token: string) => {
    const response = await api.get("/auth/verify", { params: { token } });
    return response.data;
  },

  regenerateVerify: async (email: string) => {
    const response = await api.get("/auth/verify/regenerate", { params: { email } });
    return response.data;
  },

  forgotPassword: async (email: string) => {
    const params = new URLSearchParams();
    params.append("email", email);
    const response = await api.post("/auth/forgot", params, {
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    });
    return response.data;
  },

  resetPassword: async (token: string, password: string) => {
    const params = new URLSearchParams();
    params.append("token", token);
    params.append("password", password);
    const response = await api.post("/auth/reset", params, {
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    });
    return response.data;
  },

  changePassword: async (currentPassword: string, newPassword: string): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>("/auth/change-password", {
      currentPassword,
      newPassword,
    });
    return { message: response.data.message };
  },

  logout: async () => {
    try {
      await api.post("/auth/logout");
    } finally {
      await secureStorage.removeItem("auth_token");
      await secureStorage.removeItem("user");
    }
  },

  googleSignIn: async (accessToken: string): Promise<AuthResponse> => {
    const response = await api.post<AuthResponse>("/auth/google/mobile", { accessToken });
    if (response.data.token) {
      await secureStorage.setItem("auth_token", response.data.token);
      await secureStorage.setItem("user", JSON.stringify(response.data.user));
    }
    return response.data;
  },
};

// ============ User API ============
export const userApi = {
  getMe: async (): Promise<User> => {
    const response = await api.post<{ status: boolean; user: User }>("/user");
    return response.data.user;
  },

  update: async (data: {
    name?: string;
    username?: string;
    avatar?: string;
    profilePublic?: boolean;
  }): Promise<{ message: string }> => {
    const formData = new FormData();
    if (data.name) formData.append("name", data.name);
    if (data.username) formData.append("username", data.username);
    if (data.avatar) formData.append("avatar", data.avatar);
    if (typeof data.profilePublic === "boolean") formData.append("profile_public", String(data.profilePublic));

    const response = await api.patch<{ status: boolean; message: string }>("/user", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
    return { message: response.data.message };
  },
};

// ============ Home API ============
export const homeApi = {
  create: async (name: string): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>("/homes/create", { name });
    return { message: response.data.message };
  },

  getUserHomes: async (): Promise<Home[]> => {
    const response = await api.get<{ homes: Home[] }>("/homes/list");
    return response.data.homes ?? [];
  },

  getUserHome: async (): Promise<Home> => {
    const response = await api.get<{ home: Home }>("/homes/my");
    return response.data.home;
  },

  getById: async (homeId: number): Promise<Home> => {
    const response = await api.get<{ home: Home }>(`/homes/${homeId}`);
    return response.data.home;
  },

  join: async (code: string): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>("/homes/join", { code });
    return { message: response.data.message };
  },

  leave: async (homeId: number): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/leave`);
    return { message: response.data.message };
  },

  delete: async (homeId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}`);
    return { message: response.data.message };
  },

  removeMember: async (homeId: number, userId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}/members/${userId}`);
    return { message: response.data.message };
  },

  getMembers: async (homeId: number): Promise<HomeMembership[]> => {
    const response = await api.get<{ members: HomeMembership[] }>(`/homes/${homeId}/members`);
    return response.data.members || [];
  },

  regenerateInviteCode: async (homeId: number): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/regenerate_code`);
    return { message: response.data.message };
  },

  getPendingMembers: async (homeId: number): Promise<HomeMembership[]> => {
    const response = await api.get<{ members: HomeMembership[] }>(`/homes/${homeId}/pending-members`);
    return response.data.members || [];
  },

  approveMember: async (homeId: number, userId: number): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/members/${userId}/approve`);
    return { message: response.data.message };
  },

  rejectMember: async (homeId: number, userId: number): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/members/${userId}/reject`);
    return { message: response.data.message };
  },

  updateMemberRole: async (homeId: number, userId: number, role: string): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(`/homes/${homeId}/members/${userId}/role`, {
      role,
    });
    return { message: response.data.message };
  },

  // Patches name and/or currency in one call - pass only the fields that changed.
  update: async (homeId: number, fields: { name?: string; currency?: string }): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(`/homes/${homeId}`, fields);
    return { message: response.data.message };
  },
};

// ============ Room API ============
export const roomApi = {
  create: async (
    homeId: number,
    data: { name: string; icon?: string; color?: string },
  ): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/rooms`, {
      ...data,
      homeId,
    });
    return { message: response.data.message };
  },

  getByHomeId: async (homeId: number): Promise<Room[]> => {
    const response = await api.get<{ status: boolean; rooms: Room[] }>(`/homes/${homeId}/rooms`);
    return response.data.rooms || [];
  },

  getById: async (homeId: number, roomId: number): Promise<Room> => {
    const response = await api.get<{ status: boolean; room: Room }>(`/homes/${homeId}/rooms/${roomId}`);
    return response.data.room;
  },

  update: async (
    homeId: number,
    roomId: number,
    data: { name?: string; icon?: string; color?: string },
  ): Promise<{ message: string }> => {
    const response = await api.put<{ status: boolean; message: string }>(`/homes/${homeId}/rooms/${roomId}`, data);
    return { message: response.data.message };
  },

  delete: async (homeId: number, roomId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}/rooms/${roomId}`);
    return { message: response.data.message };
  },
};

// ============ Task API ============
export const taskApi = {
  create: async (homeId: number, data: CreateTaskForm): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/tasks`, data);
    return { message: response.data.message };
  },

  getByHomeId: async (homeId: number): Promise<Task[]> => {
    const response = await api.get<{ status: boolean; tasks: Task[] }>(`/homes/${homeId}/tasks`);
    return response.data.tasks || [];
  },

  getById: async (homeId: number, taskId: number): Promise<Task> => {
    const response = await api.get<{ status: boolean; task: Task }>(`/homes/${homeId}/tasks/${taskId}`);
    return response.data.task;
  },

  update: async (
    homeId: number,
    taskId: number,
    data: { name?: string; description?: string; roomId?: number; dueDate?: string; reminderMinutes?: number },
  ): Promise<{ message: string }> => {
    const response = await api.put<{ status: boolean; message: string }>(`/homes/${homeId}/tasks/${taskId}`, data);
    return { message: response.data.message };
  },

  delete: async (homeId: number, taskId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}/tasks/${taskId}`);
    return { message: response.data.message };
  },

  assignUser: async (homeId: number, taskId: number, userId: number, date: string): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/tasks/${taskId}/assign`, {
      taskId,
      homeId,
      userId,
      date,
    });
    return { message: response.data.message };
  },

  reassignRoom: async (homeId: number, taskId: number, roomId: number): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(
      `/homes/${homeId}/tasks/${taskId}/reassign-room`,
      {
        taskId,
        roomId,
      },
    );
    return { message: response.data.message };
  },

  markCompleted: async (homeId: number, taskId: number, assignmentId: number): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(
      `/homes/${homeId}/tasks/${taskId}/mark-completed`,
      {
        assignmentId,
      },
    );
    return { message: response.data.message };
  },

  markUncompleted: async (homeId: number, taskId: number, assignmentId: number): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(
      `/homes/${homeId}/tasks/${taskId}/mark-uncompleted`,
      {
        assignmentId,
      },
    );
    return { message: response.data.message };
  },

  completeTask: async (homeId: number, taskId: number): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(`/homes/${homeId}/tasks/${taskId}/complete`);
    return { message: response.data.message };
  },

  deleteAssignment: async (homeId: number, taskId: number, assignmentId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(
      `/homes/${homeId}/tasks/${taskId}/assignments/${assignmentId}`,
    );
    return { message: response.data.message };
  },

  getUserAssignments: async (homeId: number, userId: number): Promise<TaskAssignment[]> => {
    const response = await api.get<{ status: boolean; assignments: TaskAssignment[] }>(
      `/homes/${homeId}/users/${userId}/assignments`,
    );
    return response.data.assignments || [];
  },

  getClosestAssignment: async (homeId: number, userId: number): Promise<TaskAssignment | null> => {
    const response = await api.get<{ status: boolean; assignment: TaskAssignment | null }>(
      `/homes/${homeId}/users/${userId}/assignments/closest`,
    );
    return response.data.assignment;
  },
};

// ============ Task Schedule API ============
export const taskScheduleApi = {
  create: async (homeId: number, data: CreateScheduleForm): Promise<{ schedule: TaskSchedule }> => {
    const response = await api.post<{ status: boolean; schedule: TaskSchedule }>(
      `/homes/${homeId}/tasks/schedules`,
      data,
    );
    return { schedule: response.data.schedule };
  },

  getByHomeId: async (homeId: number): Promise<TaskSchedule[]> => {
    const response = await api.get<{ status: boolean; schedules: TaskSchedule[] }>(`/homes/${homeId}/tasks/schedules`);
    return response.data.schedules || [];
  },

  getByTaskId: async (homeId: number, taskId: number): Promise<TaskSchedule | null> => {
    const response = await api.get<{ status: boolean; schedule: TaskSchedule | null }>(
      `/homes/${homeId}/tasks/${taskId}/schedule`,
    );
    return response.data.schedule;
  },

  delete: async (homeId: number, scheduleId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(
      `/homes/${homeId}/tasks/schedules/${scheduleId}`,
    );
    return { message: response.data.message };
  },
};

// ============ Bill API ============
export const billApi = {
  create: async (homeId: number, data: CreateBillForm & { billCategoryId?: number }): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/bills`, data);
    return { message: response.data.message };
  },

  getByHomeId: async (homeId: number, categoryId?: number): Promise<Bill[]> => {
    const params = categoryId != null ? { category_id: categoryId } : undefined;
    const response = await api.get<{ status: boolean; bills: Bill[] }>(`/homes/${homeId}/bills`, { params });
    return response.data.bills || [];
  },

  getPrivate: async (homeId: number, categoryId?: number): Promise<Bill[]> => {
    const params = categoryId != null ? { category_id: categoryId } : undefined;
    const response = await api.get<{ status: boolean; bills: Bill[] }>(`/homes/${homeId}/bills/private`, { params });
    return response.data.bills || [];
  },

  getById: async (homeId: number, billId: number): Promise<Bill> => {
    const response = await api.get<{ status: boolean; bill: Bill }>(`/homes/${homeId}/bills/${billId}`);
    return response.data.bill;
  },

  update: async (
    homeId: number,
    billId: number,
    data: {
      type?: string;
      public?: boolean;
      billCategoryId?: number;
      description?: string;
      receiptImage?: string;
      totalAmount?: number;
      periodStart?: string;
      periodEnd?: string;
    },
  ): Promise<{ message: string }> => {
    const response = await api.put<{ status: boolean; message: string }>(`/homes/${homeId}/bills/${billId}`, data);
    return { message: response.data.message };
  },

  delete: async (homeId: number, billId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}/bills/${billId}`);
    return { message: response.data.message };
  },

  markPayed: async (homeId: number, billId: number): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(`/homes/${homeId}/bills/${billId}`);
    return { message: response.data.message };
  },

  updateSplits: async (
    homeId: number,
    billId: number,
    splits: { userId: number; amount: number }[],
  ): Promise<{ message: string }> => {
    const response = await api.put<{ status: boolean; message: string }>(`/homes/${homeId}/bills/${billId}/splits`, {
      splits,
    });
    return { message: response.data.message };
  },

  markSplitPaid: async (homeId: number, billId: number, splitId: number): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(
      `/homes/${homeId}/bills/${billId}/splits/${splitId}/paid`,
    );
    return { message: response.data.message };
  },
};

export const billCategoryApi = {
  create: async (
    homeId: number,
    data: { name: string; icon?: string; color?: string; public?: boolean },
  ): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/bill_categories`, data);
    return { message: response.data.message };
  },

  getAll: async (homeId: number, scope: "home" | "private" = "home"): Promise<BillCategory[]> => {
    const params = scope === "private" ? { private: true } : undefined;
    const response = await api.get<{ status: boolean; categories: BillCategory[] }>(
      `/homes/${homeId}/bill_categories`,
      {
        params,
      },
    );
    return response.data.categories || [];
  },

  delete: async (homeId: number, categoryId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(
      `/homes/${homeId}/bill_categories/${categoryId}`,
    );
    return { message: response.data.message };
  },

  update: async (
    homeId: number,
    categoryId: number,
    data: { name?: string; icon?: string; color?: string },
  ): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(
      `/homes/${homeId}/bill_categories/${categoryId}`,
      data,
    );
    return { message: response.data.message };
  },
};

// ============ Shopping API ============
export const shoppingApi = {
  createCategory: async (homeId: number, data: CreateCategoryForm): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/shopping/categories`, data);
    return { message: response.data.message };
  },

  getCategories: async (homeId: number): Promise<ShoppingCategory[]> => {
    const response = await api.get<{ status: boolean; categories: ShoppingCategory[] }>(
      `/homes/${homeId}/shopping/categories/all`,
    );
    return response.data.categories || [];
  },

  getCategoryById: async (homeId: number, categoryId: number): Promise<ShoppingCategory> => {
    const response = await api.get<{ status: boolean; category: ShoppingCategory }>(
      `/homes/${homeId}/shopping/categories/${categoryId}`,
    );
    return response.data.category;
  },

  getCategoryItems: async (homeId: number, categoryId: number): Promise<ShoppingItem[]> => {
    const response = await api.get<{ status: boolean; items: ShoppingItem[] }>(
      `/homes/${homeId}/shopping/categories/${categoryId}/items`,
    );
    return response.data.items || [];
  },
  deleteCategory: async (homeId: number, categoryId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(
      `/homes/${homeId}/shopping/categories/${categoryId}`,
    );
    return { message: response.data.message };
  },

  editCategory: async (
    homeId: number,
    categoryId: number,
    data: { name?: string; icon?: string; color?: string },
  ): Promise<{ message: string }> => {
    const response = await api.put<{ status: boolean; message: string }>(
      `/homes/${homeId}/shopping/categories/${categoryId}`,
      data,
    );
    return { message: response.data.message };
  },

  createItem: async (homeId: number, data: CreateItemForm): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/shopping/items`, data);
    return { message: response.data.message };
  },

  createItems: async (homeId: number, data: CreateItemsForm): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/shopping/items`, data);
    return { message: response.data.message };
  },

  getItemById: async (homeId: number, itemId: number): Promise<ShoppingItem> => {
    const response = await api.get<{ status: boolean; item: ShoppingItem }>(
      `/homes/${homeId}/shopping/items/${itemId}`,
    );
    return response.data.item;
  },

  deleteItem: async (homeId: number, itemId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(
      `/homes/${homeId}/shopping/items/${itemId}`,
    );
    return { message: response.data.message };
  },

  editItem: async (
    homeId: number,
    itemId: number,
    data: { name?: string; image?: string; link?: string; isBought?: boolean; boughtDate?: string },
  ): Promise<{ message: string }> => {
    const response = await api.put<{ status: boolean; message: string }>(
      `/homes/${homeId}/shopping/items/${itemId}`,
      data,
    );
    return { message: response.data.message };
  },

  markBought: async (homeId: number, itemId: number): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(`/homes/${homeId}/shopping/items/${itemId}`);
    return { message: response.data.message };
  },
};

// ============ Poll API ============
export const pollApi = {
  create: async (homeId: number, data: CreatePollForm): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/polls`, data);
    return { message: response.data.message };
  },

  getByHomeId: async (homeId: number): Promise<Poll[]> => {
    const response = await api.get<{ status: boolean; polls: Poll[] }>(`/homes/${homeId}/polls`);
    return response.data.polls || [];
  },

  getById: async (homeId: number, pollId: number): Promise<Poll> => {
    const response = await api.get<{ status: boolean; poll: Poll }>(`/homes/${homeId}/polls/${pollId}`);
    return response.data.poll;
  },

  close: async (homeId: number, pollId: number): Promise<{ message: string }> => {
    const response = await api.patch<{ status: boolean; message: string }>(`/homes/${homeId}/polls/${pollId}/close`);
    return { message: response.data.message };
  },

  delete: async (homeId: number, pollId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}/polls/${pollId}`);
    return { message: response.data.message };
  },

  vote: async (homeId: number, pollId: number, optionId: number): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/polls/${pollId}/vote`, {
      optionId,
    });
    return { message: response.data.message };
  },

  unvote: async (homeId: number, pollId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}/polls/${pollId}/vote`);
    return { message: response.data.message };
  },
};

// ============ Notification API ============
export const notificationApi = {
  // Scope to a single home via homeId, so switching homes doesn't show notifications from others.
  getUserNotifications: async (homeId?: number): Promise<Notification[]> => {
    const response = await api.get<{ status: boolean; notifications: Notification[] }>("/homes/notifications", {
      params: homeId ? { home_id: homeId } : undefined,
    });
    return response.data.notifications || [];
  },

  markAsRead: async (notificationId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/notifications/${notificationId}`);
    return { message: response.data.message };
  },

  getHomeNotifications: async (homeId: number): Promise<HomeNotification[]> => {
    const response = await api.get<{ status: boolean; notifications: HomeNotification[] }>(
      `/homes/${homeId}/notifications`,
    );
    return response.data.notifications || [];
  },

  markHomeNotificationAsRead: async (homeId: number, notificationId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(
      `/homes/${homeId}/notifications/${notificationId}`,
    );
    return { message: response.data.message };
  },

  subscribeToPushNotifications: async (subscription: PushSubscription): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(
      "/notifications/subscribe",
      subscription.toJSON(),
    );
    return { message: response.data.message };
  },
};

export const auditApi = {
  getMyEvents: async (limit?: number): Promise<AuditEvent[]> => {
    const response = await api.get<{ status: boolean; events: AuditEvent[] }>("/audit/my", {
      params: limit ? { limit } : undefined,
    });
    return response.data.events || [];
  },

  getHomeEvents: async (homeId: number, limit?: number): Promise<AuditEvent[]> => {
    const response = await api.get<{ status: boolean; events: AuditEvent[] }>(`/homes/${homeId}/audit`, {
      params: limit ? { limit } : undefined,
    });
    return response.data.events || [];
  },
};

// ============ Smart Home API ============
export const smarthomeApi = {
  connect: async (homeId: number, url: string, token: string): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/smarthome/connect`, {
      url,
      token,
    });
    return { message: response.data.message };
  },

  disconnect: async (homeId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}/smarthome/disconnect`);
    return { message: response.data.message };
  },

  getStatus: async (
    homeId: number,
  ): Promise<{ connected: boolean; url?: string; isActive?: boolean; createdAt?: string; error?: string }> => {
    const response = await api.get<{
      connected: boolean;
      url?: string;
      isActive?: boolean;
      createdAt?: string;
      error?: string;
    }>(`/homes/${homeId}/smarthome/status`);
    return response.data;
  },

  discover: async (homeId: number): Promise<HAState[]> => {
    const response = await api.get<{ status: boolean; devices: HAState[] }>(`/homes/${homeId}/smarthome/discover`);
    return response.data.devices || [];
  },

  addDevice: async (homeId: number, data: AddDeviceRequest): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(`/homes/${homeId}/smarthome/devices`, data);
    return { message: response.data.message };
  },

  getDevices: async (homeId: number): Promise<SmartDevice[]> => {
    const response = await api.get<{ status: boolean; devices: SmartDevice[] }>(`/homes/${homeId}/smarthome/devices`);
    return response.data.devices || [];
  },

  getDevice: async (
    homeId: number,
    deviceId: number,
  ): Promise<{ device: SmartDevice; state?: HAState; error?: string }> => {
    const response = await api.get<{ status: boolean; device: SmartDevice; state?: HAState; error?: string }>(
      `/homes/${homeId}/smarthome/devices/${deviceId}`,
    );
    return { device: response.data.device, state: response.data.state, error: response.data.error };
  },

  updateDevice: async (homeId: number, deviceId: number, data: UpdateDeviceRequest): Promise<{ message: string }> => {
    const response = await api.put<{ status: boolean; message: string }>(
      `/homes/${homeId}/smarthome/devices/${deviceId}`,
      data,
    );
    return { message: response.data.message };
  },

  deleteDevice: async (homeId: number, deviceId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(
      `/homes/${homeId}/smarthome/devices/${deviceId}`,
    );
    return { message: response.data.message };
  },

  controlDevice: async (
    homeId: number,
    deviceId: number,
    service: string,
    data?: Record<string, any>,
  ): Promise<{ message: string }> => {
    const response = await api.post<{ status: boolean; message: string }>(
      `/homes/${homeId}/smarthome/devices/${deviceId}/control`,
      {
        service,
        data,
      },
    );
    return { message: response.data.message };
  },

  getAllStates: async (homeId: number): Promise<HAState[]> => {
    const response = await api.get<{ status: boolean; states: HAState[] }>(`/homes/${homeId}/smarthome/states`);
    return response.data.states || [];
  },

  getDevicesByRoom: async (homeId: number, roomId: number): Promise<SmartDevice[]> => {
    const response = await api.get<{ status: boolean; devices: SmartDevice[] }>(
      `/homes/${homeId}/rooms/${roomId}/devices`,
    );
    return response.data.devices || [];
  },
};

// ============ Note API ============
export const noteApi = {
  create: async (homeId: number, data: CreateNoteForm): Promise<Note> => {
    const response = await api.post<{ status: boolean; data: Note }>(`/homes/${homeId}/notes`, data);
    return response.data.data;
  },

  getByHomeId: async (homeId: number, categoryId?: number): Promise<Note[]> => {
    const params = categoryId ? { category_id: categoryId } : undefined;
    const response = await api.get<{ status: boolean; data: Note[] }>(`/homes/${homeId}/notes`, { params });
    return response.data.data || [];
  },

  getById: async (homeId: number, noteId: number): Promise<Note> => {
    const response = await api.get<{ status: boolean; data: Note }>(`/homes/${homeId}/notes/${noteId}`);
    return response.data.data;
  },

  update: async (homeId: number, noteId: number, data: UpdateNoteForm): Promise<Note> => {
    const response = await api.put<{ status: boolean; data: Note }>(`/homes/${homeId}/notes/${noteId}`, data);
    return response.data.data;
  },

  delete: async (homeId: number, noteId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(`/homes/${homeId}/notes/${noteId}`);
    return { message: response.data.message };
  },

  createCategory: async (homeId: number, data: CreateNoteCategoryForm): Promise<NoteCategory> => {
    const response = await api.post<{ status: boolean; data: NoteCategory }>(`/homes/${homeId}/note_categories`, data);
    return response.data.data;
  },

  getCategoriesByHomeId: async (homeId: number): Promise<NoteCategory[]> => {
    const response = await api.get<{ status: boolean; data: NoteCategory[] }>(`/homes/${homeId}/note_categories`);
    return response.data.data || [];
  },

  updateCategory: async (homeId: number, categoryId: number, data: UpdateNoteCategoryForm): Promise<NoteCategory> => {
    const response = await api.put<{ status: boolean; data: NoteCategory }>(
      `/homes/${homeId}/note_categories/${categoryId}`,
      data,
    );
    return response.data.data;
  },

  deleteCategory: async (homeId: number, categoryId: number): Promise<{ message: string }> => {
    const response = await api.delete<{ status: boolean; message: string }>(
      `/homes/${homeId}/note_categories/${categoryId}`,
    );
    return { message: response.data.message };
  },
};

// ============ Image Upload API ============
export const imageApi = {
  upload: async (formData: FormData): Promise<{ url: string }> => {
    const response = await api.post<{ status: boolean; message: string; url: string }>("/upload", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
    return { url: response.data.url };
  },
};

// ============ OCR API ============
export const ocrApi = {
  process: async (fileUri: string, fileName: string, mimeType: string, language?: string): Promise<OCRResult> => {
    const formData = new FormData();

    if (Platform.OS === "web") {
      const response = await fetch(fileUri);
      const blob = await response.blob();
      formData.append("file", blob, fileName);
    } else {
      // @ts-expect-error - React Native FormData accepts object with uri/name/type
      formData.append("file", { uri: fileUri, name: fileName, type: mimeType });
    }

    if (language) formData.append("language", language);
    const response = await api.post<OCRResult>("/ocr/process", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return response.data;
  },
};

// ============ Chat API ============
export const chatApi = {
  // Messages come back newest-first; pass beforeId to page further into history.
  getMessages: async (homeId: number, params?: { limit?: number; beforeId?: number }): Promise<ChatMessage[]> => {
    const response = await api.get<{ status: boolean; messages: ChatMessage[] }>(`/homes/${homeId}/chat`, {
      params: { limit: params?.limit, before_id: params?.beforeId },
    });
    return response.data.messages || [];
  },

  send: async (homeId: number, data: CreateChatMessageForm): Promise<ChatMessage> => {
    const response = await api.post<{ status: boolean; data: ChatMessage }>(`/homes/${homeId}/chat`, data);
    return response.data.data;
  },

  update: async (homeId: number, messageId: number, data: UpdateChatMessageForm): Promise<ChatMessage> => {
    const response = await api.put<{ status: boolean; data: ChatMessage }>(`/homes/${homeId}/chat/${messageId}`, data);
    return response.data.data;
  },

  delete: async (homeId: number, messageId: number): Promise<void> => {
    await api.delete(`/homes/${homeId}/chat/${messageId}`);
  },

  markRead: async (homeId: number, lastMessageId: number): Promise<void> => {
    await api.post(`/homes/${homeId}/chat/read`, { lastMessageId });
  },

  getUnreadCount: async (homeId: number): Promise<number> => {
    const response = await api.get<{ status: boolean; count: number }>(`/homes/${homeId}/chat/unread`);
    return response.data.count || 0;
  },
};

// Re-export types for convenience
export type {
  AddDeviceRequest,
  AuditEvent,
  Bill,
  BillSplit,
  ChatMessage,
  ChatMessageRead,
  ControlDeviceRequest,
  HAState,
  Home,
  HomeAssistantConfig,
  HomeMembership,
  HomeNotification,
  Note,
  NoteCategory,
  Notification,
  OCRResult,
  Poll,
  PollOption,
  Room,
  ShoppingCategory,
  ShoppingItem,
  SmartDevice,
  Task,
  TaskAssignment,
  TaskSchedule,
  UpdateDeviceRequest,
  User,
  Vote,
} from "./types";
