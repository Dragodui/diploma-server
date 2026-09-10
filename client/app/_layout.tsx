import "../global.css";

import {
  Manrope_200ExtraLight,
  Manrope_300Light,
  Manrope_400Regular,
  Manrope_500Medium,
  Manrope_600SemiBold,
  Manrope_700Bold,
  Manrope_800ExtraBold,
  useFonts,
} from "@expo-google-fonts/manrope";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Stack, usePathname, useRouter } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { StatusBar } from "expo-status-bar";
import { useEffect, useState } from "react";
import { GestureHandlerRootView } from "react-native-gesture-handler";
import { InstallPrompt } from "@/components/InstallPrompt";
import { NotificationBridge } from "@/components/NotificationBridge";
import { AlertProvider } from "@/components/ui/alert";
import { ToastProvider } from "@/components/ui/toast";
import { wsManager } from "@/lib/websocket";
import { useAuthStore } from "@/stores/authStore";
import { useHomeStore } from "@/stores/homeStore";
import { useI18nStore } from "@/stores/i18nStore";
import { useTheme, useThemeStore } from "@/stores/themeStore";

SplashScreen.preventAutoHideAsync();

const queryClient = new QueryClient();

function StoreInitializer({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(false);

  useEffect(() => {
    async function init() {
      // Init independent stores in parallel
      await Promise.all([
        useI18nStore.getState().init(),
        useThemeStore.getState().init(),
        useAuthStore.getState().init(),
      ]);
      // These depend on auth being loaded
      wsManager.init();
      useHomeStore.getState().init();
      setReady(true);
    }
    init();
  }, []);

  if (!ready) return null;
  return <>{children}</>;
}

function RootLayoutNav() {
  const { theme } = useTheme();

  return (
    <>
      <StatusBar style={theme.isDark ? "light" : "dark"} />
      <Stack screenOptions={{ headerBackTitle: "Back" }}>
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen name="login" options={{ headerShown: false }} />
        <Stack.Screen name="register" options={{ headerShown: false }} />
        <Stack.Screen name="verify" options={{ headerShown: false }} />
        <Stack.Screen name="forgot-password" options={{ headerShown: false }} />
        <Stack.Screen name="reset-password" options={{ headerShown: false }} />
        <Stack.Screen name="rooms/index" options={{ headerShown: false }} />
        <Stack.Screen name="rooms/[id]" options={{ headerShown: false }} />
        <Stack.Screen name="tasks/[id]" options={{ headerShown: false }} />
        <Stack.Screen name="settings" options={{ headerShown: false }} />
        <Stack.Screen name="home-settings" options={{ headerShown: false }} />
        <Stack.Screen name="notifications" options={{ headerShown: false }} />
        <Stack.Screen name="security" options={{ headerShown: false }} />
        <Stack.Screen name="members" options={{ headerShown: false }} />
        <Stack.Screen name="member-profile" options={{ headerShown: false }} />
        <Stack.Screen name="smarthome/index" options={{ headerShown: false }} />
        <Stack.Screen name="chat/index" options={{ headerShown: false }} />
        <Stack.Screen name="chat/[peer]" options={{ headerShown: false }} />
      </Stack>
    </>
  );
}

function AppContent() {
  const { theme } = useTheme();

  return (
    <GestureHandlerRootView className={`flex-1 ${theme.isDark ? "bg-background-dark" : "bg-background"}`}>
      <ToastProvider>
        <AlertProvider>
          <AuthGate />
          <NotificationBridge />
          <RootLayoutNav />
          <InstallPrompt />
        </AlertProvider>
      </ToastProvider>
    </GestureHandlerRootView>
  );
}

const AUTH_ROUTES = new Set(["/login", "/register", "/verify", "/forgot-password", "/reset-password"]);

function AuthGate() {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated, isLoading } = useAuthStore();

  useEffect(() => {
    if (isLoading) return;

    const isAuthRoute = AUTH_ROUTES.has(pathname);

    if (!isAuthenticated && !isAuthRoute) {
      router.replace("/login");
      return;
    }

    if (isAuthenticated && isAuthRoute) {
      router.replace("/(tabs)/home");
    }
  }, [isAuthenticated, isLoading, pathname, router]);

  return null;
}

export default function RootLayout() {
  const [fontsLoaded] = useFonts({
    Manrope_200ExtraLight,
    Manrope_300Light,
    Manrope_400Regular,
    Manrope_500Medium,
    Manrope_600SemiBold,
    Manrope_700Bold,
    Manrope_800ExtraBold,
  });

  useEffect(() => {
    if (fontsLoaded) {
      SplashScreen.hideAsync();
    }
  }, [fontsLoaded]);

  if (!fontsLoaded) {
    return null;
  }

  return (
    <QueryClientProvider client={queryClient}>
      <StoreInitializer>
        <AppContent />
      </StoreInitializer>
    </QueryClientProvider>
  );
}
