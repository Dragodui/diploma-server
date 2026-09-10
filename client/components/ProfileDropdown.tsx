import { useRouter } from "expo-router";
import {
  Bell,
  ChartColumn,
  DoorOpen,
  Home as HomeIcon,
  MessageCircle,
  Notebook,
  Settings,
  Tv,
  User,
  Users,
  Wifi,
} from "lucide-react-native";
import { useCallback, useEffect, useRef, useState } from "react";
import { Animated, Image, Pressable, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import Colors from "@/constants/colors";
import { chatApi, notificationApi } from "@/lib/api";
import { useAuth } from "@/stores/authStore";
import { useHome } from "@/stores/homeStore";
import { useI18n } from "@/stores/i18nStore";
import { useTheme } from "@/stores/themeStore";

const CLOSED_SIZE = 48;
const OPEN_WIDTH = 224;
const FULL_RADIUS = 30;

export default function ProfileDropdown() {
  const { theme } = useTheme();
  const { t } = useI18n();
  const { isAuthenticated, isLoading: authLoading, user } = useAuth();
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const { home, isLoading: homeLoading } = useHome();
  const [unreadCount, setUnreadCount] = useState(0);
  const [chatUnreadCount, setChatUnreadCount] = useState(0);
  const slideAnim = useRef(new Animated.Value(0)).current;
  const [menuHeight, setMenuHeight] = useState(0);

  const loadNotifications = useCallback(async () => {
    if (!home || !user) {
      return;
    }
    try {
      const [userNotifs, homeNotifs] = await Promise.all([
        notificationApi.getUserNotifications(home.id).catch(() => []),
        notificationApi.getHomeNotifications(home.id).catch(() => []),
      ]);

      const allNotifs = [...(userNotifs || []), ...(homeNotifs || [])];
      setUnreadCount(allNotifs.filter((n) => !n.read).length);

      const chatUnread = await chatApi.getUnreadCount(home.id).catch(() => 0);
      setChatUnreadCount(chatUnread);
    } catch (error) {
      console.error(`error while load notifications: ${error}`);
    }
  }, [home, user]);

  useEffect(() => {
    if (authLoading) return;

    if (!isAuthenticated) {
      router.replace("/login");
      return;
    }

    loadNotifications();
  }, [authLoading, isAuthenticated, loadNotifications, router]);

  const toggleMenu = (open: boolean) => {
    if (open) {
      setIsOpen(true);
      Animated.timing(slideAnim, {
        toValue: 1,
        duration: 320,
        useNativeDriver: false,
      }).start();
    } else {
      Animated.timing(slideAnim, {
        toValue: 0,
        duration: 220,
        useNativeDriver: false,
      }).start(() => setIsOpen(false));
    }
  };

  const navigateTo = (path: string) => {
    toggleMenu(false);
    setTimeout(() => {
      router.push(path as any);
    }, 100);
  };

  const width = slideAnim.interpolate({
    inputRange: [0, 0.5, 1],
    outputRange: [CLOSED_SIZE, OPEN_WIDTH, OPEN_WIDTH],
  });

  const height = slideAnim.interpolate({
    inputRange: [0, 0.5, 1],
    outputRange: [CLOSED_SIZE, CLOSED_SIZE, Math.max(menuHeight, CLOSED_SIZE)],
  });

  const contentOpacity = slideAnim.interpolate({
    inputRange: [0, 0.6, 1],
    outputRange: [0, 0, 1],
  });

  const contentTranslateY = slideAnim.interpolate({
    inputRange: [0.5, 1],
    outputRange: [-8, 0],
    extrapolate: "clamp",
  });

  const buttonBottomRadius = slideAnim.interpolate({
    inputRange: [0, 0.3, 1],
    outputRange: [FULL_RADIUS, 0, 0],
  });

  const dropdownTopRightRadius = slideAnim.interpolate({
    inputRange: [0, 0.3, 1],
    outputRange: [20, 0, 0],
  });

  const getInitials = () => {
    if (!user) return "?";
    const name = user.name || user.username || "";
    return name.slice(0, 2).toUpperCase();
  };

  const menuItems = [
    { icon: MessageCircle, label: t.chat.title, path: "/chat", badge: chatUnreadCount },
    { icon: User, label: t.tabs.profile || "Profile", path: "/(tabs)/profile" },
    { icon: Notebook, label: t.tabs.notes || "Notes", path: "/(tabs)/notes" },
    { icon: ChartColumn, label: t.tabs.polls || "Polls", path: "/(tabs)/polls" },
    { icon: DoorOpen, label: t.rooms.title || "Rooms", path: "/rooms" },
    { icon: HomeIcon, label: t.profile.homeSettings, path: "/home-settings" },
    { icon: Settings, label: "Settings", path: "/settings" },
    { icon: Users, label: "Members", path: "/members" },
    { icon: Wifi, label: "Smart Home", path: "/smarthome" },
  ];

  return (
    <View style={styles.wrapper}>
      <Animated.View
        style={[
          styles.buttonRow,
          {
            backgroundColor: theme.surface,
            borderBottomLeftRadius: buttonBottomRadius,
            borderBottomRightRadius: buttonBottomRadius,
          },
        ]}
      >
        <TouchableOpacity
          onPress={() => toggleMenu(!isOpen)}
          className="flex gap-4 flex-row px-2 py-2"
          activeOpacity={0.8}
        >
          <TouchableOpacity
            onPress={() => router.push("/notifications")}
            activeOpacity={0.7}
            className="w-12 h-12 rounded-full justify-center items-center"
            style={{ backgroundColor: theme.surface }}
          >
            <Bell size={22} color={theme.text} />
            {unreadCount > 0 && (
              <View
                className="absolute -top-[-3] -right-[-4] min-w-5 min-h-5 rounded-full justify-center items-center px-1"
                style={{ backgroundColor: theme.accent.pink }}
              >
                <Text className="text-[11px] font-manrope-bold text-white">
                  {unreadCount > 99 ? "99+" : unreadCount}
                </Text>
              </View>
            )}
          </TouchableOpacity>
          <View
            className="w-12 h-12 rounded-full justify-center items-center overflow-hidden border-[2px]"
            style={{ borderColor: Colors.accentPurple }}
          >
            {user?.avatar ? (
              <Image source={{ uri: user.avatar }} className="w-full h-full" />
            ) : (
              <Text className="text-sm font-manrope-bold" style={{ color: theme.text }}>
                {getInitials()}
              </Text>
            )}
          </View>
        </TouchableOpacity>
      </Animated.View>

      <View
        style={styles.measureHidden}
        pointerEvents="none"
        onLayout={(e) => {
          const h = e.nativeEvent.layout.height;
          if (h > 0 && Math.abs(h - menuHeight) > 0.5) setMenuHeight(h + 20);
        }}
      >
        <MenuContent items={menuItems} theme={theme} onPress={() => {}} />
      </View>

      {isOpen && <Pressable style={styles.overlay} onPress={() => toggleMenu(false)} />}

      {isOpen && (
        <Animated.View
          style={[
            styles.dropdown,
            {
              backgroundColor: theme.surface,
              width,
              height,
              borderTopRightRadius: dropdownTopRightRadius,
            },
          ]}
        >
          <Animated.View
            style={{
              opacity: contentOpacity,
              transform: [{ translateY: contentTranslateY }],
              width: OPEN_WIDTH - 20,
            }}
          >
            <MenuContent items={menuItems} theme={theme} onPress={navigateTo} />
          </Animated.View>
        </Animated.View>
      )}
    </View>
  );
}

function MenuContent({
  items,
  theme,
  onPress,
}: {
  items: { icon: any; label: string; path: string; badge?: number }[];
  theme: any;
  onPress: (path: string) => void;
}) {
  return (
    <>
      {items.map(({ icon: Icon, label, path, badge }, i) => (
        <TouchableOpacity
          key={path}
          onPress={() => onPress(path)}
          className={`flex-row items-center gap-3 p-2.5 rounded-2xl${i > 0 ? " mt-0.5" : ""}`}
        >
          <Icon size={18} color={theme.text} />
          <Text className="flex-1 text-sm font-manrope-semibold" style={{ color: theme.text }}>
            {label}
          </Text>
          {badge != null && badge > 0 && (
            <View
              className="min-w-5 h-5 rounded-full justify-center items-center px-1.5"
              style={{ backgroundColor: theme.accent.pink }}
            >
              <Text className="text-[10px] font-manrope-bold text-white">{badge > 99 ? "99+" : badge}</Text>
            </View>
          )}
        </TouchableOpacity>
      ))}
    </>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    position: "relative",
    alignItems: "flex-end",
    width: 48,
    top: -20,
    left: 5,
    height: 48,
    zIndex: 50,
  },
  buttonRow: {
    borderTopLeftRadius: FULL_RADIUS,
    borderTopRightRadius: FULL_RADIUS,
    overflow: "hidden",
    zIndex: 2,
  },
  dropdown: {
    position: "absolute",
    top: CLOSED_SIZE,
    right: 0,
    zIndex: 1,
    padding: 10,
    borderTopLeftRadius: 20,
    borderBottomLeftRadius: 24,
    borderBottomRightRadius: 24,
    overflow: "hidden",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 12 },
    shadowOpacity: 0.14,
    shadowRadius: 24,
    elevation: 12,
  },
  overlay: {
    position: "absolute",
    top: -2000,
    left: -2000,
    width: 4000,
    height: 4000,
    zIndex: 0,
  },
  measureHidden: {
    position: "absolute",
    opacity: 0,
    left: -9999,
    top: -9999,
    width: OPEN_WIDTH - 20,
  },
});
