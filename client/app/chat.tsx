import { useRouter } from "expo-router";
import {
  ArrowLeft,
  Check,
  CheckCheck,
  CheckCircle,
  DollarSign,
  MessageCircle,
  Send,
  ShoppingBag,
  Tag,
  User as UserIcon,
  Users,
  X,
} from "lucide-react-native";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Image,
  KeyboardAvoidingView,
  Platform,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { useAlert } from "@/components/ui/alert";
import Modal from "@/components/ui/modal";
import { billApi, billCategoryApi, chatApi, homeApi, noteApi, shoppingApi, taskApi } from "@/lib/api";
import type {
  Bill,
  BillCategory,
  ChatMessage,
  NoteCategory,
  ShoppingCategory,
  ShoppingItem,
  Task,
  User,
} from "@/lib/types";
import { useRealtimeRefresh } from "@/lib/useRealtimeRefresh";
import { useResponsiveLayout } from "@/lib/useResponsiveLayout";
import { useAuth } from "@/stores/authStore";
import { useHome } from "@/stores/homeStore";
import { useI18n } from "@/stores/i18nStore";
import { useTheme } from "@/stores/themeStore";

const PAGE_SIZE = 50;

// Matches the same mention syntax notes use: @user:"Name", @task:Name, @Name, @all
const MENTION_REGEX = /@(user|task|bill|item|category):(?:"([^"]+)"|(\S+))|@(?:"([^"]+)"|([a-zA-Z0-9_-]+))/g;

type SuggestionStep = "category" | "users" | "tasks" | "bills" | "items" | "note_categories";

type Suggestion =
  | { key: string; name: string; kind: "all" }
  | { key: string; name: string; kind: "step"; step: SuggestionStep }
  | { key: string; name: string; kind: "value"; prefix: string };

export default function ChatScreen() {
  const insets = useSafeAreaInsets();
  const router = useRouter();
  const { theme } = useTheme();
  const { t, language } = useI18n();
  const { home } = useHome();
  const { user } = useAuth();
  const { alert } = useAlert();
  const { horizontalPadding, contentMaxWidth } = useResponsiveLayout();

  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [sending, setSending] = useState(false);
  const [draft, setDraft] = useState("");
  const [editingId, setEditingId] = useState<number | null>(null);

  // Mention source data
  const [members, setMembers] = useState<User[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [bills, setBills] = useState<Bill[]>([]);
  const [items, setItems] = useState<ShoppingItem[]>([]);
  const [noteCategories, setNoteCategories] = useState<NoteCategory[]>([]);
  const [billCategories, setBillCategories] = useState<BillCategory[]>([]);
  const [shoppingCategories, setShoppingCategories] = useState<ShoppingCategory[]>([]);

  // Mention autocomplete
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [suggestionStep, setSuggestionStep] = useState<SuggestionStep>("category");
  const [suggestionQuery, setSuggestionQuery] = useState("");

  // Message action / read-receipt sheets
  const [actionsMessage, setActionsMessage] = useState<ChatMessage | null>(null);

  const inputRef = useRef<TextInput>(null);

  const loadMessages = useCallback(async () => {
    if (!home) return;
    try {
      const data = await chatApi.getMessages(home.id, { limit: PAGE_SIZE });
      setMessages(data);
      setHasMore(data.length === PAGE_SIZE);
    } catch (error) {
      console.error("Failed to load chat messages:", error);
    } finally {
      setIsLoading(false);
    }
  }, [home]);

  const loadMentionSources = useCallback(async () => {
    if (!home) return;
    try {
      const [memberData, taskData, billData, noteCatData, billCatData, shoppingCatData] = await Promise.all([
        homeApi.getMembers(home.id).catch(() => []),
        taskApi.getByHomeId(home.id).catch(() => []),
        billApi.getByHomeId(home.id).catch(() => []),
        noteApi.getCategoriesByHomeId(home.id).catch(() => []),
        billCategoryApi.getAll(home.id).catch(() => []),
        shoppingApi.getCategories(home.id).catch(() => []),
      ]);

      setMembers(memberData.map((m) => m.user).filter(Boolean) as User[]);
      setTasks(taskData);
      setBills(billData);
      setNoteCategories(noteCatData);
      setBillCategories(billCatData);
      setShoppingCategories(shoppingCatData);

      // Shopping items live under categories, so collect them per category.
      const itemLists = await Promise.all(
        shoppingCatData.map((c) => shoppingApi.getCategoryItems(home.id, c.id).catch(() => [] as ShoppingItem[])),
      );
      setItems(itemLists.flat());
    } catch (error) {
      console.error("Failed to load mention sources:", error);
    }
  }, [home]);

  useEffect(() => {
    loadMessages();
    loadMentionSources();
  }, [loadMessages, loadMentionSources]);

  useRealtimeRefresh(["CHAT"], loadMessages);

  // Mark everything visible as read whenever the newest message changes.
  const newestId = messages.length > 0 ? messages[0].id : null;
  useEffect(() => {
    if (!home || newestId === null) return;
    chatApi.markRead(home.id, newestId).catch(() => {});
  }, [home, newestId]);

  const handleLoadMore = async () => {
    if (!home || loadingMore || !hasMore || messages.length === 0) return;
    setLoadingMore(true);
    try {
      const oldestId = messages[messages.length - 1].id;
      const older = await chatApi.getMessages(home.id, { limit: PAGE_SIZE, beforeId: oldestId });
      setMessages((prev) => [...prev, ...older]);
      setHasMore(older.length === PAGE_SIZE);
    } catch (error) {
      console.error("Failed to load older messages:", error);
    } finally {
      setLoadingMore(false);
    }
  };

  // Resolve the @mentions written in the text into the ID lists the API wants.
  const resolveMentions = (content: string) => {
    const userIds: number[] = [];
    const taskIds: number[] = [];
    const billIds: number[] = [];
    const itemIds: number[] = [];
    const noteCategoryIds: number[] = [];
    const billCategoryIds: number[] = [];
    const shoppingCategoryIds: number[] = [];
    let mentionsAll = false;

    const regex = new RegExp(MENTION_REGEX.source, "g");
    while (true) {
      const match = regex.exec(content);
      if (match === null) break;

      const prefixType = match[1];
      const name = match[2] || match[3] || match[4] || match[5];
      if (!name) continue;

      if (!prefixType && name.toLowerCase() === "all") {
        mentionsAll = true;
        continue;
      }

      if (prefixType === "user" || !prefixType) {
        const found = members.find((m) => m.username === name || m.name === name);
        if (found) {
          userIds.push(found.id);
          continue;
        }
      }
      if (prefixType === "task" || !prefixType) {
        const found = tasks.find((task) => task.name === name);
        if (found) {
          taskIds.push(found.id);
          continue;
        }
      }
      if (prefixType === "bill" || !prefixType) {
        const found = bills.find((b) => b.description === name);
        if (found) {
          billIds.push(found.id);
          continue;
        }
      }
      if (prefixType === "item" || !prefixType) {
        const found = items.find((i) => i.name === name);
        if (found) {
          itemIds.push(found.id);
          continue;
        }
      }
      if (prefixType === "category" || !prefixType) {
        const noteCat = noteCategories.find((c) => c.name === name);
        if (noteCat) {
          noteCategoryIds.push(noteCat.id);
          continue;
        }
        const billCat = billCategories.find((c) => c.name === name);
        if (billCat) {
          billCategoryIds.push(billCat.id);
          continue;
        }
        const shoppingCat = shoppingCategories.find((c) => c.name === name);
        if (shoppingCat) shoppingCategoryIds.push(shoppingCat.id);
      }
    }

    return {
      mentionsAll,
      mentionedUserIds: userIds,
      mentionedTaskIds: taskIds,
      mentionedBillIds: billIds,
      mentionedShoppingItemIds: itemIds,
      mentionedNoteCategoryIds: noteCategoryIds,
      mentionedBillCategoryIds: billCategoryIds,
      mentionedShoppingCategoryIds: shoppingCategoryIds,
    };
  };

  const handleDraftChange = (text: string) => {
    setDraft(text);

    const lastAtIndex = text.lastIndexOf("@");
    if (lastAtIndex === -1) {
      setShowSuggestions(false);
      return;
    }

    const afterAt = text.substring(lastAtIndex);
    const isTypedPrefix = /^@(user|task|bill|item|category):"/.test(afterAt);
    if (afterAt.includes(" ") && !isTypedPrefix) {
      setShowSuggestions(false);
      return;
    }

    const prefixMatch = afterAt.match(/^@(user|task|bill|item|category):(.*)$/);
    if (prefixMatch) {
      const [, kind, query] = prefixMatch;
      const stepByKind: Record<string, SuggestionStep> = {
        user: "users",
        task: "tasks",
        bill: "bills",
        item: "items",
        category: "note_categories",
      };
      setSuggestionStep(stepByKind[kind]);
      setSuggestionQuery(query.replace(/"/g, ""));
      setShowSuggestions(true);
      return;
    }

    setSuggestionStep("category");
    setSuggestionQuery(afterAt.substring(1).toLowerCase());
    setShowSuggestions(true);
  };

  const suggestions = useMemo<Suggestion[]>(() => {
    const query = suggestionQuery.toLowerCase();

    if (suggestionStep === "category") {
      const kinds: Suggestion[] = [
        { key: "all", name: t.chat.mentionAll, kind: "all" },
        { key: "users", name: "Users", kind: "step", step: "users" },
        { key: "tasks", name: "Tasks", kind: "step", step: "tasks" },
        { key: "bills", name: "Bills", kind: "step", step: "bills" },
        { key: "items", name: "Shopping Items", kind: "step", step: "items" },
        { key: "categories", name: "Categories", kind: "step", step: "note_categories" },
      ];
      return kinds.filter((k) => k.name.toLowerCase().includes(query));
    }

    if (suggestionStep === "users") {
      return members
        .filter((m) => (m.username || m.name || "").toLowerCase().includes(query))
        .map((m) => ({ key: `user-${m.id}`, name: m.username || m.name, kind: "value" as const, prefix: "user" }));
    }
    if (suggestionStep === "tasks") {
      return tasks
        .filter((task) => task.name.toLowerCase().includes(query))
        .map((task) => ({ key: `task-${task.id}`, name: task.name, kind: "value" as const, prefix: "task" }));
    }
    if (suggestionStep === "bills") {
      return bills
        .filter((b) => (b.description || "").toLowerCase().includes(query))
        .map((b) => ({ key: `bill-${b.id}`, name: b.description, kind: "value" as const, prefix: "bill" }));
    }
    if (suggestionStep === "items") {
      return items
        .filter((i) => i.name.toLowerCase().includes(query))
        .map((i) => ({ key: `item-${i.id}`, name: i.name, kind: "value" as const, prefix: "item" }));
    }

    return [...noteCategories, ...billCategories, ...shoppingCategories]
      .filter((c) => c.name.toLowerCase().includes(query))
      .map((c) => ({ key: `cat-${c.name}`, name: c.name, kind: "value" as const, prefix: "category" }));
  }, [
    suggestionStep,
    suggestionQuery,
    members,
    tasks,
    bills,
    items,
    noteCategories,
    billCategories,
    shoppingCategories,
    t.chat.mentionAll,
  ]);

  const applySuggestion = (suggestion: Suggestion) => {
    const lastAtIndex = draft.lastIndexOf("@");
    if (lastAtIndex === -1) return;
    const before = draft.substring(0, lastAtIndex);

    if (suggestion.kind === "step") {
      const prefixByStep: Record<SuggestionStep, string> = {
        category: "@",
        users: "@user:",
        tasks: "@task:",
        bills: "@bill:",
        items: "@item:",
        note_categories: "@category:",
      };
      setDraft(`${before}${prefixByStep[suggestion.step]}`);
      setSuggestionStep(suggestion.step);
      setSuggestionQuery("");
      inputRef.current?.focus();
      return;
    }

    if (suggestion.kind === "all") {
      setDraft(`${before}@all `);
      setShowSuggestions(false);
      return;
    }

    // Quote names that contain spaces so the parser keeps them in one token.
    const value = suggestion.name.includes(" ") ? `"${suggestion.name}"` : suggestion.name;
    setDraft(`${before}@${suggestion.prefix}:${value} `);
    setShowSuggestions(false);
  };

  const handleSend = async () => {
    if (!home || !draft.trim() || sending) return;
    setSending(true);
    const content = draft.trim();
    const mentions = resolveMentions(content);

    try {
      if (editingId !== null) {
        await chatApi.update(home.id, editingId, { content, ...mentions });
        setEditingId(null);
      } else {
        await chatApi.send(home.id, { content, ...mentions });
      }
      setDraft("");
      setShowSuggestions(false);
      await loadMessages();
    } catch (error) {
      console.error("Failed to send message:", error);
      alert(t.common.error, editingId !== null ? t.chat.failedToUpdate : t.chat.failedToSend);
    } finally {
      setSending(false);
    }
  };

  const handleDelete = (message: ChatMessage) => {
    setActionsMessage(null);
    alert(t.chat.deleteMessage, t.chat.deleteConfirm, [
      { text: t.common.cancel, style: "cancel" },
      {
        text: t.common.delete,
        style: "destructive",
        onPress: async () => {
          if (!home) return;
          try {
            await chatApi.delete(home.id, message.id);
            await loadMessages();
          } catch (error) {
            console.error("Failed to delete message:", error);
            alert(t.common.error, t.chat.failedToDelete);
          }
        },
      },
    ]);
  };

  const startEditing = (message: ChatMessage) => {
    setActionsMessage(null);
    setEditingId(message.id);
    setDraft(message.content);
    inputRef.current?.focus();
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleTimeString(language, { hour: "2-digit", minute: "2-digit", hour12: false });
  };

  const formatDateTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return `${date.toLocaleDateString(language)} ${formatTime(dateStr)}`;
  };

  // Renders message text with mention badges, mirroring how notes display them.
  const renderContent = (message: ChatMessage) => {
    const regex = new RegExp(MENTION_REGEX.source, "g");
    const parts: React.ReactNode[] = [];
    const isOwn = message.createdBy === user?.id;
    const baseColor = isOwn ? "#1C1C1E" : theme.text;
    let lastIndex = 0;
    let key = 0;

    while (true) {
      const match = regex.exec(message.content);
      if (match === null) break;

      if (match.index > lastIndex) {
        parts.push(
          <Text key={key++} className="font-manrope text-base" style={{ color: baseColor }}>
            {message.content.substring(lastIndex, match.index)}
          </Text>,
        );
      }

      const prefixType = match[1];
      const name = match[2] || match[3] || match[4] || match[5] || "";

      let badgeBg = "";
      let icon: React.ReactNode = null;
      let matched = false;

      if (!prefixType && name.toLowerCase() === "all") {
        matched = true;
        badgeBg = "bg-accent-pink/20 text-accent-pink";
        icon = <Users size={12} color={theme.accent.pink} />;
      } else if (
        prefixType === "user" ||
        (!prefixType && message.mentionedUsers?.some((u) => u.username === name || u.name === name))
      ) {
        matched = true;
        badgeBg = "bg-accent-purple/20 text-accent-purple";
        icon = <UserIcon size={12} color={theme.accent.purple} />;
      } else if (prefixType === "task" || (!prefixType && message.mentionedTasks?.some((task) => task.name === name))) {
        matched = true;
        badgeBg = "bg-accent-mint/20 text-accent-mint";
        icon = <CheckCircle size={12} color={theme.accent.mint} />;
      } else if (
        prefixType === "bill" ||
        (!prefixType && message.mentionedBills?.some((b) => b.description === name))
      ) {
        matched = true;
        badgeBg = "bg-accent-yellow/20 text-accent-yellow";
        icon = <DollarSign size={12} color={theme.accent.yellow} />;
      } else if (
        prefixType === "item" ||
        (!prefixType && message.mentionedShoppingItems?.some((i) => i.name === name))
      ) {
        matched = true;
        badgeBg = "bg-accent-cyan/20 text-accent-cyan";
        icon = <ShoppingBag size={12} color={theme.accent.cyan} />;
      } else if (
        prefixType === "category" ||
        (!prefixType &&
          (message.mentionedNoteCategories?.some((c) => c.name === name) ||
            message.mentionedBillCategories?.some((c) => c.name === name) ||
            message.mentionedShoppingCategories?.some((c) => c.name === name)))
      ) {
        matched = true;
        badgeBg = "bg-accent-pink/20 text-accent-pink";
        icon = <Tag size={12} color={theme.accent.pink} />;
      }

      if (matched) {
        parts.push(
          <Text
            key={key++}
            className={`font-manrope-semibold text-sm px-2 py-0.5 rounded-full overflow-hidden ${badgeBg}`}
          >
            {icon} {name}
          </Text>,
        );
      } else {
        parts.push(
          <Text key={key++} className="font-manrope text-base" style={{ color: baseColor }}>
            {match[0]}
          </Text>,
        );
      }
      lastIndex = regex.lastIndex;
    }

    if (lastIndex < message.content.length) {
      parts.push(
        <Text key={key++} className="font-manrope text-base" style={{ color: baseColor }}>
          {message.content.substring(lastIndex)}
        </Text>,
      );
    }

    return <Text className="leading-6">{parts}</Text>;
  };

  const renderMessage = ({ item }: { item: ChatMessage }) => {
    const isOwn = item.createdBy === user?.id;
    const readCount = item.reads?.length || 0;

    return (
      <View className={`mb-3 flex-row ${isOwn ? "justify-end" : "justify-start"}`}>
        <TouchableOpacity
          activeOpacity={0.9}
          onLongPress={() => isOwn && setActionsMessage(item)}
          className="max-w-[85%]"
        >
          {!isOwn && (
            <Text className="text-xs font-manrope-semibold mb-1 ml-1" style={{ color: theme.textSecondary }}>
              {item.creator?.name || item.creator?.username || ""}
            </Text>
          )}
          <View
            className="px-4 py-3 rounded-2xl"
            style={{ backgroundColor: isOwn ? theme.accent.cyan : theme.surface }}
          >
            {renderContent(item)}
            <View className="flex-row items-center justify-end gap-1.5 mt-1">
              {item.editedAt && (
                <Text className="text-[10px] font-manrope" style={{ color: isOwn ? "#1C1C1E99" : theme.textSecondary }}>
                  {t.chat.edited}
                </Text>
              )}
              <Text className="text-[10px] font-manrope" style={{ color: isOwn ? "#1C1C1E99" : theme.textSecondary }}>
                {formatTime(item.createdAt)}
              </Text>
              {/* One tick until somebody reads it, two once it's been read. */}
              {isOwn &&
                (readCount > 0 ? <CheckCheck size={13} color="#1C1C1E" /> : <Check size={13} color="#1C1C1E99" />)}
            </View>
          </View>
        </TouchableOpacity>
      </View>
    );
  };

  return (
    <View className="flex-1" style={{ backgroundColor: theme.background }}>
      {/* Header */}
      <View
        className="flex-row items-center gap-3 pb-4"
        style={{ paddingTop: insets.top + 16, paddingHorizontal: horizontalPadding }}
      >
        <TouchableOpacity
          className="w-12 h-12 rounded-2xl justify-center items-center"
          style={{ backgroundColor: theme.surface }}
          onPress={() => router.back()}
        >
          <ArrowLeft size={22} color={theme.text} />
        </TouchableOpacity>
        <View className="flex-1">
          <Text className="text-2xl font-manrope-bold" style={{ color: theme.text }}>
            {t.chat.title}
          </Text>
          {home?.name && (
            <Text className="text-sm font-manrope" numberOfLines={1} style={{ color: theme.textSecondary }}>
              {home.name}
            </Text>
          )}
        </View>
      </View>

      <KeyboardAvoidingView
        className="flex-1"
        behavior={Platform.OS === "ios" ? "padding" : undefined}
        keyboardVerticalOffset={Platform.OS === "ios" ? 8 : 0}
      >
        {isLoading ? (
          <View className="flex-1 justify-center items-center">
            <ActivityIndicator color={theme.text} />
          </View>
        ) : messages.length === 0 ? (
          <View className="flex-1 justify-center items-center px-10">
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
        ) : (
          <FlatList
            data={messages}
            renderItem={renderMessage}
            keyExtractor={(item) => String(item.id)}
            inverted
            contentContainerStyle={{
              paddingHorizontal: horizontalPadding,
              paddingBottom: 8,
              width: "100%",
              maxWidth: contentMaxWidth,
              alignSelf: "center",
            }}
            onEndReached={handleLoadMore}
            onEndReachedThreshold={0.4}
            ListFooterComponent={
              loadingMore ? <ActivityIndicator color={theme.textSecondary} className="my-4" /> : null
            }
            showsVerticalScrollIndicator={false}
          />
        )}

        {/* Mention suggestions */}
        {showSuggestions && suggestions.length > 0 && (
          <View className="mx-4 mb-2 rounded-2xl overflow-hidden max-h-48" style={{ backgroundColor: theme.surface }}>
            <FlatList
              data={suggestions}
              keyExtractor={(s) => s.key}
              keyboardShouldPersistTaps="handled"
              renderItem={({ item }) => (
                <TouchableOpacity className="px-4 py-3" onPress={() => applySuggestion(item)}>
                  <Text className="font-manrope-semibold text-sm" style={{ color: theme.text }}>
                    {item.name}
                  </Text>
                </TouchableOpacity>
              )}
            />
          </View>
        )}

        {/* Composer */}
        <View
          className="flex-row items-end gap-2 pt-2"
          style={{
            paddingHorizontal: horizontalPadding,
            paddingBottom: insets.bottom + 12,
            backgroundColor: theme.background,
          }}
        >
          {editingId !== null && (
            <TouchableOpacity
              className="w-12 h-12 rounded-2xl justify-center items-center"
              style={{ backgroundColor: theme.surface }}
              onPress={() => {
                setEditingId(null);
                setDraft("");
              }}
            >
              <X size={20} color={theme.textSecondary} />
            </TouchableOpacity>
          )}
          <TextInput
            ref={inputRef}
            className="flex-1 px-4 py-3 rounded-2xl font-manrope text-base max-h-32"
            style={{ backgroundColor: theme.surface, color: theme.text }}
            placeholder={t.chat.placeholder}
            placeholderTextColor={theme.textSecondary}
            value={draft}
            onChangeText={handleDraftChange}
            multiline
          />
          <TouchableOpacity
            className="w-12 h-12 rounded-2xl justify-center items-center"
            style={{ backgroundColor: draft.trim() ? theme.accent.cyan : theme.surface }}
            onPress={handleSend}
            disabled={!draft.trim() || sending}
          >
            {sending ? (
              <ActivityIndicator color="#1C1C1E" />
            ) : (
              <Send size={20} color={draft.trim() ? "#1C1C1E" : theme.textSecondary} />
            )}
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>

      {/* Own-message actions, with the read-by avatars on top */}
      <Modal
        visible={actionsMessage !== null}
        onClose={() => setActionsMessage(null)}
        title={t.chat.title}
        height="auto"
      >
        <View className="gap-3">
          {/* Read-by block only appears once somebody has actually read it. */}
          {actionsMessage?.reads && actionsMessage.reads.length > 0 && (
            <View className="gap-2">
              <Text className="text-xs font-manrope-bold uppercase" style={{ color: theme.textSecondary }}>
                {t.chat.readBy}
              </Text>
              <View className="flex-row flex-wrap gap-3">
                {actionsMessage.reads.map((read) => (
                  <View key={read.id} className="items-center gap-1 w-14">
                    <View
                      className="w-11 h-11 rounded-full justify-center items-center overflow-hidden"
                      style={{ backgroundColor: theme.surface }}
                    >
                      {read.user?.avatar ? (
                        <Image source={{ uri: read.user.avatar }} className="w-full h-full" />
                      ) : (
                        <Text className="text-xs font-manrope-bold" style={{ color: theme.text }}>
                          {(read.user?.name || read.user?.username || "?").slice(0, 2).toUpperCase()}
                        </Text>
                      )}
                    </View>
                    <Text
                      className="text-[10px] font-manrope-semibold text-center"
                      numberOfLines={1}
                      style={{ color: theme.text }}
                    >
                      {read.user?.name || read.user?.username || `#${read.userId}`}
                    </Text>
                    <Text className="text-[9px] font-manrope text-center" style={{ color: theme.textSecondary }}>
                      {formatDateTime(read.readAt)}
                    </Text>
                  </View>
                ))}
              </View>
            </View>
          )}

          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.surface }}
            onPress={() => actionsMessage && startEditing(actionsMessage)}
          >
            <Text className="font-manrope-semibold" style={{ color: theme.text }}>
              {t.common.edit}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.accent.dangerLight }}
            onPress={() => actionsMessage && handleDelete(actionsMessage)}
          >
            <Text className="font-manrope-semibold text-white">{t.common.delete}</Text>
          </TouchableOpacity>
        </View>
      </Modal>
    </View>
  );
}
