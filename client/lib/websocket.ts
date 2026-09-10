import { Platform } from "react-native";
import { useAuthStore } from "@/stores/authStore";
import { isDevMode } from "./api";

const rawUrl = (isDevMode ? process.env.EXPO_PUBLIC_API_URL_DEV : process.env.EXPO_PUBLIC_API_URL) || "";
const WS_URL = `${rawUrl.replace(/^http/, "ws")}/ws`;

export type EventModule =
  | "BILL_CATEGORY"
  | "BILL"
  | "HOME"
  | "NOTIFICATION"
  | "HOME_NOTIFICATION"
  | "POLL"
  | "ROOM"
  | "SHOPPING_CATEGORY"
  | "SHOPPING_ITEM"
  | "TASK"
  | "USER"
  | "NOTE"
  | "NOTE_CATEGORY"
  | "CHAT";

export type EventAction =
  | "CREATED"
  | "UPDATED"
  | "DELETED"
  | "MARKED_PAYED"
  | "CLOSED"
  | "VOTED"
  | "UNVOTED"
  | "MEMBER_JOINED"
  | "MEMBER_LEFT"
  | "MEMBER_REMOVED"
  | "ASSIGNED"
  | "COMPLETED"
  | "UNCOMPLETED"
  | "MARK_READ";

export interface RealTimeEvent {
  module: EventModule;
  action: EventAction;
  data: any;
}

type EventCallback = (event: RealTimeEvent) => void;

class WebSocketManager {
  private ws: WebSocket | null = null;
  private subscribers = new Map<string, Set<EventCallback>>();
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private initialized = false;
  private disconnecting = false;
  private authUnsubscribe: (() => void) | null = null;

  init() {
    if (this.initialized) return;
    this.initialized = true;

    // React to auth state changes
    this.authUnsubscribe = useAuthStore.subscribe(
      (state) => ({ isAuthenticated: state.isAuthenticated, token: state.token }),
      ({ isAuthenticated, token }) => {
        if (isAuthenticated && (token || Platform.OS === "web")) {
          this.connect(token);
        } else {
          this.disconnect();
        }
      },
      { equalityFn: (a, b) => a.isAuthenticated === b.isAuthenticated && a.token === b.token },
    );

    // Connect immediately if already authenticated
    const { isAuthenticated, token } = useAuthStore.getState();
    if (isAuthenticated && (token || Platform.OS === "web")) {
      this.connect(token);
    }
  }

  dispose() {
    this.authUnsubscribe?.();
    this.authUnsubscribe = null;
    this.initialized = false;
    this.disconnect();
  }

  private connect(token: string | null) {
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      return;
    }

    this.disconnecting = false;

    try {
      const ws = new WebSocket(WS_URL);

      ws.onopen = () => {
        if (token) {
          ws.send(JSON.stringify({ token }));
        }
        if (__DEV__) console.log("[WS] Connected");
      };

      ws.onmessage = (e) => {
        try {
          const event: RealTimeEvent = JSON.parse(e.data);
          const moduleSubscribers = this.subscribers.get(event.module);
          if (moduleSubscribers) {
            moduleSubscribers.forEach((cb) => cb(event));
          }
        } catch (err) {
          console.error("[WS] Failed to parse message:", err);
        }
      };

      ws.onclose = () => {
        if (__DEV__) console.log("[WS] Disconnected");
        this.ws = null;
        // Don't reconnect if disconnect was intentional
        if (this.disconnecting) return;
        this.reconnectTimeout = setTimeout(() => {
          const { isAuthenticated, token } = useAuthStore.getState();
          if (isAuthenticated && (token || Platform.OS === "web")) {
            this.connect(token);
          }
        }, 3000);
      };

      ws.onerror = (err) => {
        console.error("[WS] Error:", err);
      };

      this.ws = ws;
    } catch (err) {
      console.error("[WS] Connection failed:", err);
    }
  }

  private disconnect() {
    this.disconnecting = true;
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  subscribe(modules: EventModule[], callback: EventCallback): () => void {
    modules.forEach((module) => {
      if (!this.subscribers.has(module)) {
        this.subscribers.set(module, new Set());
      }
      this.subscribers.get(module)!.add(callback);
    });

    return () => {
      modules.forEach((module) => {
        this.subscribers.get(module)?.delete(callback);
      });
    };
  }
}

export const wsManager = new WebSocketManager();
