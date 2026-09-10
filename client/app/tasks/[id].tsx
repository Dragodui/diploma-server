import { useLocalSearchParams, useRouter } from "expo-router";
import { ArrowLeft, Calendar, Check, Clock3, Repeat, Trash } from "lucide-react-native";
import { useCallback, useEffect, useMemo, useState } from "react";
import { ActivityIndicator, ScrollView, Text, TouchableOpacity, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { useAlert } from "@/components/ui/alert";
import Button from "@/components/ui/button";
import { taskApi } from "@/lib/api";
import type { Task } from "@/lib/types";
import { useRealtimeRefresh } from "@/lib/useRealtimeRefresh";
import { useResponsiveLayout } from "@/lib/useResponsiveLayout";
import { useAuth } from "@/stores/authStore";
import { useHome } from "@/stores/homeStore";
import { useI18n } from "@/stores/i18nStore";
import { useTheme } from "@/stores/themeStore";

export default function TaskDetailScreen() {
  const insets = useSafeAreaInsets();
  const router = useRouter();
  const { id } = useLocalSearchParams<{ id: string }>();
  const taskId = Number(id);
  const { theme } = useTheme();
  const { t } = useI18n();
  const { user } = useAuth();
  const { home, isAdmin } = useHome();
  const { alert } = useAlert();
  const { horizontalPadding } = useResponsiveLayout();

  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const goBack = useCallback(() => {
    if (router.canGoBack()) {
      router.back();
      return;
    }
    router.replace("/(tabs)/tasks");
  }, [router]);

  const loadTask = useCallback(async () => {
    if (!home || !Number.isFinite(taskId)) {
      setLoading(false);
      return;
    }

    try {
      const taskData = await taskApi.getById(home.id, taskId);
      setTask(taskData);
    } catch (error) {
      console.error("Failed to load task details:", error);
      setTask(null);
    } finally {
      setLoading(false);
    }
  }, [home, taskId]);

  useEffect(() => {
    loadTask();
  }, [loadTask]);

  useRealtimeRefresh(["TASK"], loadTask);

  const myAssignment = useMemo(
    () => task?.assignments?.find((assignment) => assignment.userId === user?.id),
    [task?.assignments, user?.id],
  );
  const completed = myAssignment?.status === "completed" || task?.assignments?.some((a) => a.status === "completed");

  const dueText = useMemo(() => {
    if (!task?.dueDate) return t.tasks.noDueDate;
    const date = new Date(task.dueDate);
    return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false })}`;
  }, [task?.dueDate, t.tasks.noDueDate]);

  const reminderText = useMemo(() => {
    if (!task?.dueDate) return t.tasks.noReminder;
    if ((task.reminderMinutes ?? 30) === 0) return t.tasks.reminderAtDue;
    return `${task.reminderMinutes ?? 30} ${t.tasks.reminderMinutesSuffix}`;
  }, [task?.dueDate, task?.reminderMinutes, t.tasks.noReminder, t.tasks.reminderAtDue, t.tasks.reminderMinutesSuffix]);

  const assigneesText = useMemo(() => {
    if (!task?.assignments?.length) return t.tasks.unassigned;
    const names = task.assignments.map((assignment) => assignment.user?.name).filter(Boolean) as string[];
    if (names.length === 0) return t.tasks.assigned;
    return names.join(", ");
  }, [task?.assignments, t.tasks.assigned, t.tasks.unassigned]);

  const handleToggleStatus = async () => {
    if (!home || !task || !user || submitting) return;
    setSubmitting(true);
    try {
      if (completed) {
        const assignmentId = myAssignment?.id ?? task.assignments?.find((a) => a.status === "completed")?.id;
        if (!assignmentId) return;
        await taskApi.markUncompleted(home.id, task.id, assignmentId);
      } else {
        await taskApi.completeTask(home.id, task.id);
      }
      await loadTask();
    } catch (error) {
      console.error("Failed to toggle task status:", error);
      alert(t.common.error, completed ? t.tasks.failedToUncomplete : t.tasks.failedToComplete);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = () => {
    if (!home || !task || submitting) return;

    alert(t.tasks.deleteTask, t.tasks.deleteTaskConfirm, [
      { text: t.common.cancel, style: "cancel" },
      {
        text: t.common.delete,
        style: "destructive",
        onPress: async () => {
          setSubmitting(true);
          try {
            await taskApi.delete(home.id, task.id);
            goBack();
          } catch (error) {
            console.error("Failed to delete task:", error);
            alert(t.common.error, t.tasks.failedToDelete);
          } finally {
            setSubmitting(false);
          }
        },
      },
    ]);
  };

  if (loading) {
    return (
      <View className="flex-1 items-center justify-center" style={{ backgroundColor: theme.background }}>
        <ActivityIndicator size="large" color={theme.text} />
      </View>
    );
  }

  if (!task) {
    return (
      <View className="flex-1 items-center justify-center px-6" style={{ backgroundColor: theme.background }}>
        <Text className="text-xl font-manrope-bold mb-2 text-center" style={{ color: theme.text }}>
          {t.common.error}
        </Text>
        <Text className="text-sm font-manrope text-center mb-5" style={{ color: theme.textSecondary }}>
          Task not found.
        </Text>
        <Button title={t.common.done} onPress={goBack} />
      </View>
    );
  }

  const scheduleLabel = task.schedule
    ? t.tasks.schedule[task.schedule.recurrenceType as keyof typeof t.tasks.schedule] || t.tasks.schedule.scheduled
    : null;

  const nextRunText = task.schedule?.nextRunDate
    ? t.tasks.schedule.nextRun.replace("{date}", new Date(task.schedule.nextRunDate).toLocaleString())
    : null;

  return (
    <View className="flex-1" style={{ backgroundColor: theme.background }}>
      <ScrollView
        className="flex-1"
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{
          paddingTop: insets.top + 16,
          paddingHorizontal: horizontalPadding,
          paddingBottom: 120,
          width: "100%",
          maxWidth: 960,
          alignSelf: "center",
        }}
      >
        <View className="flex-row items-center justify-between mb-4 gap-2">
          <TouchableOpacity
            onPress={goBack}
            className="w-11 h-11 rounded-2xl items-center justify-center"
            style={{ backgroundColor: theme.surface }}
            activeOpacity={0.8}
          >
            <ArrowLeft size={22} color={theme.text} />
          </TouchableOpacity>
          <Text
            className="flex-1 text-lg font-manrope-bold text-center mx-2"
            style={{ color: theme.text }}
            numberOfLines={1}
          >
            Task #{task.id}
          </Text>
          {isAdmin && (
            <TouchableOpacity
              onPress={handleDelete}
              className="w-11 h-11 rounded-2xl items-center justify-center"
              style={{ backgroundColor: theme.surface }}
              activeOpacity={0.8}
            >
              <Trash size={18} color={theme.accent.pink} />
            </TouchableOpacity>
          )}
        </View>
        <View className="rounded-3xl p-5 mb-4" style={{ backgroundColor: theme.surface }}>
          <Text className="text-2xl font-bold font-manrope" style={{ color: theme.text }}>
            {task.name || "—"}
          </Text>
        </View>
        <View className="rounded-3xl p-5 mb-4" style={{ backgroundColor: theme.surface }}>
          <Text className="text-xs font-manrope-bold uppercase mb-2" style={{ color: theme.textSecondary }}>
            {t.tasks.description}
          </Text>
          <Text className="text-base font-manrope" style={{ color: theme.text }}>
            {task.description || "—"}
          </Text>
        </View>

        <View className="rounded-3xl p-5 mb-4 gap-3" style={{ backgroundColor: theme.surface }}>
          <View className="flex-row items-center justify-between">
            <View className="flex-row items-center gap-2">
              <Calendar size={16} color={theme.textSecondary} />
              <Text className="text-sm font-manrope-semibold" style={{ color: theme.textSecondary }}>
                {t.tasks.dueDate}
              </Text>
            </View>
            <Text className="text-sm font-manrope-semibold text-right flex-1 ml-3" style={{ color: theme.text }}>
              {dueText}
            </Text>
          </View>

          <View className="flex-row items-center justify-between">
            <View className="flex-row items-center gap-2">
              <Clock3 size={16} color={theme.textSecondary} />
              <Text className="text-sm font-manrope-semibold" style={{ color: theme.textSecondary }}>
                {t.tasks.assigned}
              </Text>
            </View>
            <Text className="text-sm font-manrope-semibold text-right flex-1 ml-3" style={{ color: theme.text }}>
              {assigneesText}
            </Text>
          </View>

          <View className="flex-row items-center justify-between">
            <View className="flex-row items-center gap-2">
              <Check size={16} color={theme.textSecondary} />
              <Text className="text-sm font-manrope-semibold" style={{ color: theme.textSecondary }}>
                Status
              </Text>
            </View>
            <Text
              className="text-sm font-manrope-semibold text-right flex-1 ml-3"
              style={{ color: completed ? theme.status.success : theme.accent.pink }}
            >
              {completed ? "Completed" : "Pending"}
            </Text>
          </View>

          {task.room && (
            <View className="flex-row items-center justify-between">
              <Text className="text-sm font-manrope-semibold" style={{ color: theme.textSecondary }}>
                {t.tasks.room}
              </Text>
              <Text className="text-sm font-manrope-semibold text-right flex-1 ml-3" style={{ color: theme.text }}>
                {task.room.name}
              </Text>
            </View>
          )}

          {scheduleLabel && (
            <View className="flex-row items-center justify-between">
              <View className="flex-row items-center gap-2">
                <Repeat size={16} color={theme.accent.purple} />
                <Text className="text-sm font-manrope-semibold" style={{ color: theme.textSecondary }}>
                  {t.tasks.schedule.title}
                </Text>
              </View>
              <Text
                className="text-sm font-manrope-semibold text-right flex-1 ml-3"
                style={{ color: theme.accent.purple }}
              >
                {scheduleLabel}
              </Text>
            </View>
          )}
          {nextRunText && (
            <View className="flex-row items-center justify-between">
              <View className="flex-row items-center gap-2">
                <Clock3 size={16} color={theme.accent.purple} />
                <Text className="text-sm font-manrope-semibold" style={{ color: theme.textSecondary }}>
                  {t.tasks.schedule.title}
                </Text>
              </View>
              <Text
                className="text-sm font-manrope-semibold text-right flex-1 ml-3"
                style={{ color: theme.accent.purple }}
              >
                {nextRunText}
              </Text>
            </View>
          )}
          <View className="flex-row items-center justify-between">
            <Text className="text-sm font-manrope-semibold" style={{ color: theme.textSecondary }}>
              {t.tasks.reminderLabel}
            </Text>
            <Text className="text-sm font-manrope-semibold text-right flex-1 ml-3" style={{ color: theme.text }}>
              {reminderText}
            </Text>
          </View>
        </View>

        <Button
          title={completed ? "Undo complete" : t.tasks.complete}
          onPress={handleToggleStatus}
          loading={submitting}
          className="mb-3"
        />
      </ScrollView>
    </View>
  );
}
