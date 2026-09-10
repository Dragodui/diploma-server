import { useRouter } from "expo-router";
import { ArrowLeft, MessageCircle, Users } from "lucide-react-native";
import { useCallback, useEffect, useState } from "react";
import { ActivityIndicator, Image, RefreshControl, ScrollView, Text, TouchableOpacity, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { chatApi } from "@/lib/api";
import type { ChatConversation } from "@/lib/types";
import { useRealtimeRefresh } from "@/lib/useRealtimeRefresh";
import { useResponsiveLayout } from "@/lib/useResponsiveLayout";
import { useHome } from "@/stores/homeStore";
import { useI18n } from "@/stores/i18nStore";
import { useTheme } from "@/stores/themeStore";

export default function ChatListScreen() {
  const insets = useSafeAreaInsets();
  const router = useRouter();
  const { theme } = useTheme();
  const { t, language } = useI18n();
  const { home } = useHome();
  const { horizontalPadding, contentMaxWidth, isDesktop } = useResponsiveLayout();

  const [conversations, setConversations] = useState<ChatConversation[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const loadConversations = useCallback(async () => {
    if (!home) return;
    try {
      const data = await chatApi.getConversations(home.id);
      setConversations(data);
    } catch (error) {
      console.error("Failed to load conversations:", error);
    } finally {
      setIsLoading(false);
    }
  }, [home]);

  useEffect(() => {
    loadConversations();
  }, [loadConversations]);

  useRealtimeRefresh(["CHAT"], loadConversations);

  const onRefresh = async () => {
    setRefreshing(true);
    await loadConversations();
    setRefreshing(false);
  };

  const goBack = () => {
    if (router.canGoBack()) {
      router.back();
      return;
    }
    router.replace("/(tabs)/home");
  };

  // Same-day messages show the time, older ones the date.
  const formatStamp = (dateStr: string) => {
    const date = new Date(dateStr);
    const isToday = date.toDateString() === new Date().toDateString();
    return isToday
      ? date.toLocaleTimeString(language, { hour: "2-digit", minute: "2-digit", hour12: false })
      : date.toLocaleDateString(language);
  };

  const previewOf = (conversation: ChatConversation) => {
    const last = conversation.lastMessage;
    if (!last) return "";
    const author = last.creator?.name || last.creator?.username || "";
    const body = last.content.trim() || (last.imageUrl ? "🖼" : "");
    // Only the shared chat needs the author prefix; in a direct chat it's obvious.
    return conversation.peerId == null && author ? `${author}: ${body}` : body;
  };

  return (
    <View className="flex-1" style={{ backgroundColor: theme.background }}>
      <View
        className="flex-row items-center gap-3 pb-4"
        style={{ paddingTop: insets.top + 16, paddingHorizontal: horizontalPadding }}
      >
        <TouchableOpacity
          className="w-12 h-12 rounded-2xl justify-center items-center"
          style={{ backgroundColor: theme.surface }}
          onPress={goBack}
        >
          <ArrowLeft size={22} color={theme.text} />
        </TouchableOpacity>
        <Text className="flex-1 text-2xl font-manrope-bold" style={{ color: theme.text }}>
          {t.chat.title}
        </Text>
      </View>

      {isLoading ? (
        <View className="flex-1 justify-center items-center">
          <ActivityIndicator color={theme.text} />
        </View>
      ) : (
        <ScrollView
          className="flex-1"
          contentContainerStyle={{
            paddingHorizontal: horizontalPadding,
            paddingBottom: isDesktop ? 48 : 120,
            width: "100%",
            maxWidth: contentMaxWidth,
            alignSelf: "center",
          }}
          refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={theme.text} />}
          showsVerticalScrollIndicator={false}
        >
          <View className="gap-2">
            {conversations.map((conversation) => {
              const isHomeChat = conversation.peerId == null;
              const title = isHomeChat
                ? home?.name || t.chat.title
                : conversation.peer?.name || conversation.peer?.username || "";
              const preview = previewOf(conversation);

              return (
                <TouchableOpacity
                  key={isHomeChat ? "home" : conversation.peerId}
                  className="flex-row items-center gap-3 p-3 rounded-2xl"
                  style={{ backgroundColor: theme.surface }}
                  activeOpacity={0.8}
                  onPress={() => router.push(`/chat/${isHomeChat ? "home" : conversation.peerId}` as never)}
                >
                  <View
                    className="w-12 h-12 rounded-full justify-center items-center overflow-hidden"
                    style={{ backgroundColor: isHomeChat ? theme.accent.purple : theme.background }}
                  >
                    {isHomeChat ? (
                      <Users size={22} color="#1C1C1E" />
                    ) : conversation.peer?.avatar ? (
                      <Image source={{ uri: conversation.peer.avatar }} className="w-full h-full" />
                    ) : (
                      <Text className="text-xs font-manrope-bold" style={{ color: theme.text }}>
                        {(conversation.peer?.name || conversation.peer?.username || "?").slice(0, 2).toUpperCase()}
                      </Text>
                    )}
                  </View>

                  <View className="flex-1">
                    <Text className="font-manrope-bold text-base" numberOfLines={1} style={{ color: theme.text }}>
                      {title}
                    </Text>
                    {preview ? (
                      <Text className="font-manrope text-sm" numberOfLines={1} style={{ color: theme.textSecondary }}>
                        {preview}
                      </Text>
                    ) : (
                      <Text className="font-manrope text-sm italic" style={{ color: theme.textSecondary }}>
                        {t.chat.noMessages}
                      </Text>
                    )}
                  </View>

                  <View className="items-end gap-1">
                    {conversation.lastMessage && (
                      <Text className="font-manrope text-[11px]" style={{ color: theme.textSecondary }}>
                        {formatStamp(conversation.lastMessage.createdAt)}
                      </Text>
                    )}
                    {conversation.unreadCount > 0 && (
                      <View
                        className="min-w-5 h-5 rounded-full justify-center items-center px-1.5"
                        style={{ backgroundColor: theme.accent.pink }}
                      >
                        <Text className="text-[10px] font-manrope-bold text-white">
                          {conversation.unreadCount > 99 ? "99+" : conversation.unreadCount}
                        </Text>
                      </View>
                    )}
                  </View>
                </TouchableOpacity>
              );
            })}
          </View>

          {conversations.length === 0 && (
            <View className="items-center py-20 px-6">
              <View
                className="w-16 h-16 rounded-full justify-center items-center mb-4"
                style={{ backgroundColor: theme.surface }}
              >
                <MessageCircle size={32} color={theme.textSecondary} />
              </View>
              <Text className="text-xl font-manrope-bold mb-2 text-center" style={{ color: theme.text }}>
                {t.chat.noMessages}
              </Text>
              <Text className="text-sm font-manrope text-center leading-5" style={{ color: theme.textSecondary }}>
                {t.chat.noMessagesDescription}
              </Text>
            </View>
          )}
        </ScrollView>
      )}
    </View>
  );
}
