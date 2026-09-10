// User types
export interface User {
  id: number;
  email: string;
  emailVerified: boolean;
  name: string;
  username?: string;
  profilePublic?: boolean;
  avatar?: string;
  createdAt: string;
}

// Home/Household types
export interface Home {
  id: number;
  name: string;
  inviteCode: string;
  currency: string;
  createdAt: string;
  memberships?: HomeMembership[];
}

export interface HomeMembership {
  id: number;
  homeId: number;
  userId: number;
  role: "admin" | "member";
  status: "pending" | "approved";
  joinedAt: string;
  createdAt: string;
  user?: User;
  home?: Home;
}

export interface AuditEvent {
  id: number;
  homeId?: number;
  actorUserId?: number;
  eventType: string;
  entityType: string;
  entityId?: number;
  metadata?: Record<string, unknown>;
  ip?: string;
  userAgent?: string;
  createdAt: string;
  home?: Home;
  actorUser?: User;
}

// Room types
export interface Room {
  id: number;
  homeId: number;
  name: string;
  icon?: string;
  color: string;
  createdAt: string;
  tasks?: Task[];
}

// Task types
export interface Task {
  id: number;
  homeId: number;
  roomId?: number;
  name: string;
  description: string;
  scheduleType: string;
  dueDate?: string;
  reminderMinutes: number;
  createdAt: string;
  room?: Room;
  assignments?: TaskAssignment[];
  schedule?: TaskSchedule;
}

export interface TaskSchedule {
  id: number;
  taskId: number;
  recurrenceType: "daily" | "weekly" | "monthly";
  rotationUserIds: string;
  currentRotationIndex: number;
  nextRunDate: string;
  isActive: boolean;
  createdAt: string;
  task?: Task;
}

export interface CreateScheduleForm {
  taskId: number;
  homeId: number;
  recurrenceType: "daily" | "weekly" | "monthly";
  userIds: number[];
}

export interface TaskAssignment {
  id: number;
  taskId: number;
  userId: number;
  status: "assigned" | "completed";
  assignedDate: string;
  completeDate?: string;
  reminderSentAt?: string;
  createdAt: string;
  task?: Task;
  user?: User;
}

// Bill types
export interface BillSplit {
  id: number;
  billId: number;
  userId: number;
  amount: number;
  paid: boolean;
  user?: User;
}

export interface Bill {
  id: number;
  public: boolean;
  homeId: number;
  type: string;
  billCategoryId?: number;
  billCategory?: BillCategory;
  isPayed: boolean;
  paymentDate?: string;
  totalAmount: number;
  periodStart: string;
  periodEnd: string;
  uploadedBy: number;
  description?: string;
  receiptImage?: string;
  ocrData?: Record<string, any>;
  createdAt: string;
  user?: User;
  splits?: BillSplit[];
}

// Shopping types
export interface ShoppingCategory {
  id: number;
  homeId: number;
  name: string;
  icon?: string;
  color: string;
  createdAt: string;
  items?: ShoppingItem[];
}

export interface ShoppingItem {
  id: number;
  categoryId: number;
  name: string;
  addedBy: number;
  isBought: boolean;
  image?: string;
  link?: string;
  boughtDate?: string;
  createdAt: string;
  user?: User;
}

export interface BillCategory {
  id: number;
  homeId: number;
  createdBy: number;
  public: boolean;
  name: string;
  icon?: string;
  color: string;
  createdAt: string;
}

// Poll types
export interface Poll {
  id: number;
  homeId: number;
  question: string;
  type: "public" | "anonymous";
  status: "open" | "closed";
  allowRevote: boolean;
  endsAt?: string;
  createdAt: string;
  options?: PollOption[];
}

export interface PollOption {
  id: number;
  pollId: number;
  title: string;
  createdAt: string;
  votes?: Vote[];
  voteCount?: number;
}

export interface Vote {
  id: number;
  userId: number;
  optionId: number;
  user?: User;
}

// Notification types
export interface Notification {
  id: number;
  from?: number;
  to: number;
  homeId: number;
  description: string;
  read: boolean;
  createdAt: string;
  userFrom?: User;
}

export interface HomeNotification {
  id: number;
  from?: number;
  homeId: number;
  description: string;
  read: boolean;
  createdAt: string;
  userFrom?: User;
}

// API Response types
export interface ApiResponse<T> {
  data?: T;
  message?: string;
  error?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Form types
export interface LoginForm {
  email: string;
  password: string;
}

export interface RegisterForm {
  email: string;
  password: string;
  name: string;
}

export interface CreateHomeForm {
  name: string;
}

export interface JoinHomeForm {
  code: string;
}

export interface CreateTaskForm {
  name: string;
  description: string;
  scheduleType: string;
  dueDate?: string;
  reminderMinutes?: number;
  homeId: number;
  roomId?: number;
  assignUserIds?: number[];
}

export interface CreateBillForm {
  type: string;
  public?: boolean;
  isRegular?: boolean;
  recurrenceType?: "daily" | "weekly" | "monthly";
  recurrenceDay?: number;
  description?: string;
  receiptImage?: string;
  totalAmount: number;
  periodStart: string;
  periodEnd: string;
  ocrData?: Record<string, any>;
  splits?: { userId: number; amount: number }[];
}

export interface CreatePollForm {
  question: string;
  type: "public" | "anonymous";
  options: { title: string }[];
  allowRevote: boolean;
  endsAt?: string;
}

export interface CreateCategoryForm {
  name: string;
  icon?: string;
  color?: string;
}

export interface CreateItemForm {
  categoryId: number;
  name: string;
  image?: string;
  link?: string;
}

export interface CreateItemPayload {
  name: string;
  image?: string;
  link?: string;
}

export interface CreateItemsForm {
  categoryId: number;
  items: CreateItemPayload[];
}

// Smart Home Types
export interface SmartDevice {
  id: number;
  homeId: number;
  roomId?: number;
  entityId: string;
  name: string;
  type: string;
  icon?: string;
  createdAt: string;
  room?: Room;
}

export interface HomeAssistantConfig {
  id: number;
  homeId: number;
  url: string;
  isActive: boolean;
  createdAt: string;
}

export interface HAState {
  entityId: string;
  state: string;
  attributes: Record<string, any>;
  lastChanged: string;
  lastUpdated: string;
}

export interface ConnectHARequest {
  url: string;
  token: string;
}

export interface AddDeviceRequest {
  entityId: string;
  name: string;
  roomId?: number;
  icon?: string;
}

export interface UpdateDeviceRequest {
  name?: string;
  roomId?: number;
  icon?: string;
}

export interface ControlDeviceRequest {
  service: string;
  data?: Record<string, any>;
}

export interface OCRResult {
  vendor?: string;
  date?: string;
  total: number;
  items?: {
    name: string;
    quantity: number;
    price: number;
  }[];
  rawText?: string;
  confidence?: number;
}

// Note types
export interface NoteCategory {
  id: number;
  homeId: number;
  createdBy: number;
  name: string;
  icon?: string;
  color: string;
  createdAt: string;
  creator?: User;
}

export interface Note {
  id: number;
  homeId: number;
  noteCategoryId?: number;
  createdBy: number;
  title: string;
  content: string;
  createdAt: string;
  updatedAt: string;
  creator?: User;
  noteCategory?: NoteCategory;

  mentionedUsers?: User[];
  mentionedTasks?: Task[];
  mentionedBills?: Bill[];
  mentionedShoppingItems?: ShoppingItem[];
  mentionedNoteCategories?: NoteCategory[];
  mentionedBillCategories?: BillCategory[];
  mentionedShoppingCategories?: ShoppingCategory[];
}

export interface CreateNoteForm {
  noteCategoryId?: number | null;
  title: string;
  content: string;
  mentionedUserIds?: number[];
  mentionedTaskIds?: number[];
  mentionedBillIds?: number[];
  mentionedShoppingItemIds?: number[];
  mentionedNoteCategoryIds?: number[];
  mentionedBillCategoryIds?: number[];
  mentionedShoppingCategoryIds?: number[];
}

export interface UpdateNoteForm {
  noteCategoryId?: number | null;
  title?: string;
  content?: string;
  mentionedUserIds?: number[];
  mentionedTaskIds?: number[];
  mentionedBillIds?: number[];
  mentionedShoppingItemIds?: number[];
  mentionedNoteCategoryIds?: number[];
  mentionedBillCategoryIds?: number[];
  mentionedShoppingCategoryIds?: number[];
}

// Chat types
export interface ChatMessageRead {
  id: number;
  messageId: number;
  userId: number;
  readAt: string;
  user?: User;
}

export interface ChatMessage {
  id: number;
  homeId: number;
  createdBy: number;
  content: string;
  imageUrl?: string | null;
  editedAt?: string | null;
  createdAt: string;
  mentionsAll: boolean;
  creator?: User;
  reads?: ChatMessageRead[];

  mentionedUsers?: User[];
  mentionedTasks?: Task[];
  mentionedBills?: Bill[];
  mentionedShoppingItems?: ShoppingItem[];
  mentionedNoteCategories?: NoteCategory[];
  mentionedBillCategories?: BillCategory[];
  mentionedShoppingCategories?: ShoppingCategory[];
}

export interface CreateChatMessageForm {
  content: string;
  imageUrl?: string | null;
  mentionsAll?: boolean;
  mentionedUserIds?: number[];
  mentionedTaskIds?: number[];
  mentionedBillIds?: number[];
  mentionedShoppingItemIds?: number[];
  mentionedNoteCategoryIds?: number[];
  mentionedBillCategoryIds?: number[];
  mentionedShoppingCategoryIds?: number[];
}

export interface UpdateChatMessageForm {
  content?: string;
  imageUrl?: string | null;
  mentionsAll?: boolean;
  mentionedUserIds?: number[];
  mentionedTaskIds?: number[];
  mentionedBillIds?: number[];
  mentionedShoppingItemIds?: number[];
  mentionedNoteCategoryIds?: number[];
  mentionedBillCategoryIds?: number[];
  mentionedShoppingCategoryIds?: number[];
}

export interface CreateNoteCategoryForm {
  name: string;
  icon?: string;
  color?: string;
}

export interface UpdateNoteCategoryForm {
  name?: string;
  icon?: string;
  color?: string;
}
