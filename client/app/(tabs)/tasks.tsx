import { useRouter } from "expo-router";
import { Calendar, Check, CheckCircle, Plus, Repeat, X } from "lucide-react-native";
import { useCallback, useEffect, useState } from "react";
import { RefreshControl, ScrollView, Text, TouchableOpacity, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { TasksSkeleton } from "@/components/skeletons";
import { useAlert } from "@/components/ui/alert";
import Button from "@/components/ui/button";
import DatePicker from "@/components/ui/date-picker";
import Input from "@/components/ui/input";
import Modal from "@/components/ui/modal";
import { userColors } from "@/constants/colors";
import { taskApi, taskScheduleApi } from "@/lib/api";
import type { Task, TaskAssignment } from "@/lib/types";
import { useRealtimeRefresh } from "@/lib/useRealtimeRefresh";
import { useResponsiveLayout } from "@/lib/useResponsiveLayout";
import { useAuth } from "@/stores/authStore";
import { useHome } from "@/stores/homeStore";
import { interpolate, useI18n } from "@/stores/i18nStore";
import { useTheme } from "@/stores/themeStore";

type FilterType = "All" | "My" | "By Room" | "Completed";
type RecurrenceType = "daily" | "weekly" | "monthly";

export default function TasksScreen() {
  const router = useRouter();
  const insets = useSafeAreaInsets();
  const { user } = useAuth();
  const { home, rooms, isAdmin } = useHome();
  const { theme } = useTheme();
  const { t } = useI18n();
  const { alert } = useAlert();
  const { isDesktop, horizontalPadding, contentMaxWidth } = useResponsiveLayout();

  const [tasks, setTasks] = useState<Task[]>([]);
  const [assignments, setAssignments] = useState<TaskAssignment[]>([]);
  const [activeFilter, setActiveFilter] = useState<FilterType>("All");
  const [isLoading, setIsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newTaskName, setNewTaskName] = useState("");
  const [newTaskDescription, setNewTaskDescription] = useState("");
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);
  const [isDatePickerVisible, setDatePickerVisibility] = useState(false);
  const [selectedRoomId, setSelectedRoomId] = useState<number | null>(null);
  const [selectedUserIds, setSelectedUserIds] = useState<number[]>([]);
  const [reminderMinutesInput, setReminderMinutesInput] = useState("30");
  const [creating, setCreating] = useState(false);

  // Schedule modal state
  const [showScheduleModal, setShowScheduleModal] = useState(false);
  const [scheduleTaskId, setScheduleTaskId] = useState<number | null>(null);
  const [scheduleRecurrence, setScheduleRecurrence] = useState<RecurrenceType>("daily");
  const [scheduleUserIds, setScheduleUserIds] = useState<number[]>([]);
  const [creatingSchedule, setCreatingSchedule] = useState(false);
  const [selectedTaskForActions, setSelectedTaskForActions] = useState<Task | null>(null);
  const [showTaskActionsModal, setShowTaskActionsModal] = useState(false);
  const [showEditTaskModal, setShowEditTaskModal] = useState(false);
  const [editingTaskId, setEditingTaskId] = useState<number | null>(null);
  const [editTaskName, setEditTaskName] = useState("");
  const [editTaskDescription, setEditTaskDescription] = useState("");
  const [editTaskRoomId, setEditTaskRoomId] = useState<number | null>(null);
  const [editTaskDate, setEditTaskDate] = useState<Date | null>(null);
  const [editReminderMinutesInput, setEditReminderMinutesInput] = useState("30");
  const [editTaskUserIds, setEditTaskUserIds] = useState<number[]>([]);
  const [editTaskOriginalAssignments, setEditTaskOriginalAssignments] = useState<TaskAssignment[]>([]);
  const [savingEditTask, setSavingEditTask] = useState(false);
  const [isEditDatePickerVisible, setIsEditDatePickerVisible] = useState(false);

  const parseReminderMinutes = (value: string) => {
    const trimmed = value.trim();
    if (!trimmed) return 30;
    const parsed = Number.parseInt(trimmed, 10);
    if (Number.isNaN(parsed) || parsed < 0) return 30;
    return parsed;
  };

  const loadTasks = useCallback(async () => {
    if (!home || !user) {
      setIsLoading(false);
      return;
    }

    try {
      const [tasksData, assignmentsData] = await Promise.all([
        taskApi.getByHomeId(home.id),
        taskApi.getUserAssignments(home.id, user.id).catch(() => []),
      ]);

      setTasks(tasksData || []);
      setAssignments(assignmentsData || []);
    } catch (error) {
      console.error("Error loading tasks:", error);
    } finally {
      setIsLoading(false);
    }
  }, [home, user]);

  useEffect(() => {
    loadTasks();
  }, [loadTasks]);

  useRealtimeRefresh(["TASK"], loadTasks);

  const onRefresh = async () => {
    setRefreshing(true);
    await loadTasks();
    setRefreshing(false);
  };

  const showDatePicker = () => setDatePickerVisibility(true);
  const hideDatePicker = () => setDatePickerVisibility(false);

  const handleConfirmDate = (date: Date) => {
    setSelectedDate(date);
    hideDatePicker();
  };

  const handleDelete = (taskId: number) => {
    if (!home) return;

    alert(t.tasks.deleteTask, t.tasks.deleteTaskConfirm, [
      { text: t.common.cancel, style: "cancel" },
      {
        text: t.common.delete,
        style: "destructive",
        onPress: async () => {
          try {
            await taskApi.delete(home.id, taskId);
            await loadTasks();
          } catch (error) {
            console.error(error);
            alert(t.common.error, t.tasks.failedToDelete);
          }
        },
      },
    ]);
  };

  const handleToggleTask = async (task: Task) => {
    if (!home || !user) return;

    const completed = isTaskCompleted(task);

    if (completed) {
      let assignmentId = task.assignments?.find((a) => a.userId === user.id)?.id;
      if (!assignmentId) {
        assignmentId = assignments.find((a) => a.taskId === task.id && a.userId === user.id)?.id;
      }

      if (assignmentId) {
        try {
          await taskApi.markUncompleted(home.id, task.id, assignmentId);
          await loadTasks();
        } catch (error) {
          console.error("Error uncompleting task:", error);
          alert(t.common.error, t.tasks.failedToUncomplete);
        }
      }
    } else {
      alert(t.tasks.completeTask, t.tasks.completeTaskConfirm, [
        { text: t.common.cancel, style: "cancel" },
        {
          text: t.tasks.complete,
          onPress: async () => {
            try {
              await taskApi.completeTask(home.id, task.id);
              await loadTasks();
            } catch (error) {
              console.error("Error completing task:", error);
              alert(t.common.error, t.tasks.failedToComplete);
            }
          },
        },
      ]);
    }
  };

  const handleCreateTask = async () => {
    if (!home || !newTaskName.trim()) return;

    setCreating(true);
    try {
      await taskApi.create(home.id, {
        name: newTaskName.trim(),
        description: newTaskDescription.trim(),
        scheduleType: "once",
        dueDate: selectedDate ? selectedDate.toISOString() : undefined,
        reminderMinutes: parseReminderMinutes(reminderMinutesInput),
        homeId: home.id,
        roomId: selectedRoomId || undefined,
        assignUserIds: selectedUserIds.length > 0 ? selectedUserIds : undefined,
      });

      setNewTaskName("");
      setNewTaskDescription("");
      setSelectedDate(null);
      setSelectedRoomId(null);
      setSelectedUserIds([]);
      setReminderMinutesInput("30");
      setShowCreateModal(false);
      await loadTasks();
    } catch (error) {
      console.error("Error creating task:", error);
      alert(t.common.error, t.tasks.couldNotCreate);
    } finally {
      setCreating(false);
    }
  };

  // Schedule handlers
  const handleOpenScheduleModal = (taskId: number) => {
    setScheduleTaskId(taskId);
    setScheduleRecurrence("daily");
    setScheduleUserIds([]);
    setShowScheduleModal(true);
  };

  const openTaskActions = (task: Task) => {
    setSelectedTaskForActions(task);
    setShowTaskActionsModal(true);
  };

  const openEditTaskModal = (task: Task) => {
    setEditingTaskId(task.id);
    setEditTaskName(task.name || "");
    setEditTaskDescription(task.description || "");
    setEditTaskRoomId(task.roomId ?? null);
    setEditTaskDate(task.dueDate ? new Date(task.dueDate) : null);
    setEditReminderMinutesInput(String(task.reminderMinutes ?? 30));
    const activeAssignments = (task.assignments || []).filter((a) => a.status !== "completed");
    setEditTaskOriginalAssignments(activeAssignments);
    setEditTaskUserIds(activeAssignments.map((a) => a.userId));
    setShowEditTaskModal(true);
  };

  const handleSaveTaskEdit = async () => {
    if (!home || !editingTaskId || !editTaskName.trim()) return;
    setSavingEditTask(true);
    try {
      await taskApi.update(home.id, editingTaskId, {
        name: editTaskName.trim(),
        description: editTaskDescription.trim(),
        roomId: editTaskRoomId || undefined,
        dueDate: editTaskDate ? editTaskDate.toISOString() : undefined,
        reminderMinutes: parseReminderMinutes(editReminderMinutesInput),
      });

      const originalUserIds = editTaskOriginalAssignments.map((a) => a.userId);
      const userIdsToAssign = editTaskUserIds.filter((id) => !originalUserIds.includes(id));
      const assignmentsToRemove = editTaskOriginalAssignments.filter((a) => !editTaskUserIds.includes(a.userId));
      const assignDate = (editTaskDate ?? new Date()).toISOString();

      await Promise.all([
        ...userIdsToAssign.map((userId) => taskApi.assignUser(home.id, editingTaskId, userId, assignDate)),
        ...assignmentsToRemove.map((a) => taskApi.deleteAssignment(home.id, editingTaskId, a.id)),
      ]);

      setShowEditTaskModal(false);
      setEditingTaskId(null);
      await loadTasks();
    } catch (error) {
      console.error("Error editing task:", error);
      alert(t.common.error, t.tasks.failedToUpdate);
    } finally {
      setSavingEditTask(false);
    }
  };

  const handleCreateSchedule = async () => {
    if (!home || !scheduleTaskId || scheduleUserIds.length === 0) return;

    setCreatingSchedule(true);
    try {
      await taskScheduleApi.create(home.id, {
        taskId: scheduleTaskId,
        homeId: home.id,
        recurrenceType: scheduleRecurrence,
        userIds: scheduleUserIds,
      });

      setShowScheduleModal(false);
      setScheduleTaskId(null);
      setScheduleUserIds([]);
      await loadTasks();
    } catch (error) {
      console.error("Error creating schedule:", error);
      alert(t.common.error, t.tasks.schedule.failedToCreate);
    } finally {
      setCreatingSchedule(false);
    }
  };

  const handleDeleteSchedule = (task: Task) => {
    if (!home || !task.schedule) return;

    alert(t.tasks.schedule.deleteSchedule, t.tasks.schedule.deleteScheduleConfirm, [
      { text: t.common.cancel, style: "cancel" },
      {
        text: t.common.delete,
        style: "destructive",
        onPress: async () => {
          try {
            await taskScheduleApi.delete(home.id, task.schedule!.id);
            await loadTasks();
          } catch (error) {
            console.error(error);
            alert(t.common.error, t.tasks.schedule.failedToDelete);
          }
        },
      },
    ]);
  };

  const getRecurrenceLabel = (type: string) => {
    switch (type) {
      case "daily":
        return t.tasks.schedule.daily;
      case "weekly":
        return t.tasks.schedule.weekly;
      case "monthly":
        return t.tasks.schedule.monthly;
      default:
        return type;
    }
  };

  const getFilteredTasks = () => {
    const completedTasks = tasks.filter((task) => isTaskCompleted(task));
    const activeTasks = tasks.filter((task) => !isTaskCompleted(task));

    if (activeFilter === "Completed") {
      return completedTasks;
    }

    if (activeFilter === "My") {
      const myTaskIds = assignments.map((a) => a.taskId);
      return activeTasks.filter((task) => myTaskIds.includes(task.id));
    }

    return activeTasks;
  };

  // For a recurring task, assignments pile up across rotation cycles - the
  // most recent one (by assignedDate, falling back to id) is the one that
  // reflects the current cycle, not whichever happens to be first in the array.
  const getLatestAssignment = (assignmentsList: TaskAssignment[], userId?: number) => {
    const pool = userId != null ? assignmentsList.filter((a) => a.userId === userId) : assignmentsList;
    if (pool.length === 0) return undefined;
    return pool.reduce((latest, current) => {
      const latestTime = new Date(latest.assignedDate).getTime() || latest.id;
      const currentTime = new Date(current.assignedDate).getTime() || current.id;
      return currentTime >= latestTime ? current : latest;
    });
  };

  const isTaskCompleted = (task: Task) => {
    if (task.assignments && task.assignments.length > 0) {
      const userAssignment = getLatestAssignment(task.assignments, user?.id);
      if (userAssignment) return userAssignment.status === "completed";
      return getLatestAssignment(task.assignments)?.status === "completed";
    }
    const assignment = assignments.find((a) => a.taskId === task.id);
    return assignment?.status === "completed";
  };

  const getTaskAssignee = (task: Task) => {
    if (task.assignments && task.assignments.length > 0) {
      if (task.schedule) {
        // Recurring task: only the current cycle's assignee is relevant.
        const current = getLatestAssignment(task.assignments);
        return current?.user?.name || t.tasks.assigned;
      }
      const firstName = task.assignments[0].user?.name || t.tasks.assigned;
      if (task.assignments.length > 1) {
        return `${firstName} +${task.assignments.length - 1}`;
      }
      return firstName;
    }
    return t.tasks.unassigned;
  };

  const getTaskDueText = (task: Task) => {
    if (task.dueDate) {
      const date = new Date(task.dueDate);
      return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false })}`;
    }
    return t.tasks.noDueDate;
  };

  const getReminderText = (task: Task) => {
    if (!task.dueDate) return null;
    if (task.reminderMinutes === 0) return t.tasks.reminderAtDue;
    return `${task.reminderMinutes} ${t.tasks.reminderMinutesSuffix}`;
  };

  const getTaskCompletedDate = (task: Task) => {
    let assignment = task.assignments ? getLatestAssignment(task.assignments, user?.id) : undefined;
    if (!assignment) {
      assignment = assignments.find((a) => a.taskId === task.id && a.userId === user?.id);
    }
    if (!assignment && task.assignments) {
      assignment = getLatestAssignment(task.assignments.filter((a) => a.status === "completed"));
    }

    if (assignment?.completeDate) {
      const date = new Date(assignment.completeDate);
      return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false })}`;
    }
    return "";
  };

  const getMyTasksCount = () => {
    const myTaskIds = assignments.map((a) => a.taskId);
    return tasks.filter((t) => myTaskIds.includes(t.id)).length;
  };

  const renderTaskItem = (task: Task, index: number) => {
    const completed = isTaskCompleted(task);
    const colorIndex = index % userColors.length;
    const completedDate = completed ? getTaskCompletedDate(task) : "";
    const hasSchedule = !!task.schedule;
    return (
      <View
        key={task.id}
        className="rounded-24 p-5"
        style={{ backgroundColor: theme.surface, width: isDesktop ? "49%" : "100%" }}
      >
        <View className="flex-row items-center gap-4">
          <TouchableOpacity
            className="w-8 h-8 rounded-16 border-2 justify-center items-center"
            onPress={() => handleToggleTask(task)}
            activeOpacity={0.7}
            style={[
              { borderColor: theme.textSecondary },
              completed && { backgroundColor: theme.accent.pink, borderColor: theme.accent.pink },
            ]}
          >
            {completed && <Check size={16} color="#FFFFFF" strokeWidth={3} />}
          </TouchableOpacity>

          <TouchableOpacity
            className="flex-1"
            onPress={() => router.push({ pathname: "/tasks/[id]", params: { id: String(task.id) } })}
            onLongPress={() => openTaskActions(task)}
            activeOpacity={0.7}
          >
            <View className="flex-1">
              <View className="flex-row items-center gap-2 mb-2">
                <Text
                  className={`text-lg font-manrope-bold flex-1 ${completed ? "line-through opacity-50" : ""}`}
                  style={{ color: theme.text }}
                  numberOfLines={1}
                >
                  {task.name}
                </Text>
                {hasSchedule && (
                  <View
                    className="px-2 py-0.5 rounded-full flex-row items-center gap-1"
                    style={{ backgroundColor: `${theme.accent.purple}20` }}
                  >
                    <Repeat size={10} color={theme.accent.purple} />
                    <Text className="text-[10px] font-manrope-bold" style={{ color: theme.accent.purple }}>
                      {getRecurrenceLabel(task.schedule!.recurrenceType)}
                    </Text>
                  </View>
                )}
              </View>
              {task.description ? (
                <Text
                  className={`text-sm font-manrope mb-1 ${completed ? "line-through opacity-50" : ""}`}
                  style={{ color: theme.textSecondary }}
                  numberOfLines={2}
                >
                  {task.description}
                </Text>
              ) : null}
              <View className="flex-row items-center flex-wrap gap-2">
                <View className="w-2 h-2 rounded-full" style={{ backgroundColor: userColors[colorIndex] }} />
                <Text className="text-xs font-manrope-semibold" style={{ color: theme.textSecondary }}>
                  {getTaskAssignee(task)}
                </Text>
                {task.room && (
                  <>
                    <Text className="text-[10px]" style={{ color: theme.textSecondary }}>
                      •
                    </Text>
                    <Text className="text-xs font-manrope-semibold" style={{ color: theme.textSecondary }}>
                      {task.room.name}
                    </Text>
                  </>
                )}
                <Text className="text-[10px]" style={{ color: theme.textSecondary }}>
                  •
                </Text>
                <View className="flex-row items-center gap-1">
                  <Calendar size={12} color={theme.textSecondary} />
                  <Text className="text-xs font-manrope-semibold" style={{ color: theme.textSecondary }}>
                    {getTaskDueText(task)}
                  </Text>
                </View>
                {getReminderText(task) && (
                  <>
                    <Text className="text-[10px]" style={{ color: theme.textSecondary }}>
                      •
                    </Text>
                    <Text className="text-xs font-manrope-semibold" style={{ color: theme.textSecondary }}>
                      {t.tasks.reminderLabel}: {getReminderText(task)}
                    </Text>
                  </>
                )}
                {completed && completedDate && (
                  <>
                    <Text className="text-[10px]" style={{ color: theme.textSecondary }}>
                      •
                    </Text>
                    <Text className="text-xs font-manrope-semibold" style={{ color: theme.status.success }}>
                      Done: {completedDate}
                    </Text>
                  </>
                )}
              </View>
            </View>
          </TouchableOpacity>
        </View>
      </View>
    );
  };

  if (isLoading) {
    return <TasksSkeleton />;
  }

  return (
    <View className="flex-1" style={{ backgroundColor: theme.background }}>
      <ScrollView
        className="flex-1"
        contentContainerStyle={{
          paddingHorizontal: horizontalPadding,
          paddingBottom: isDesktop ? 48 : 120,
          paddingTop: insets.top + 24,
          width: "100%",
          maxWidth: contentMaxWidth,
          alignSelf: "center",
        }}
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={theme.text} />}
        showsVerticalScrollIndicator={false}
      >
        {/* Header */}
        <View className="flex-row justify-between items-start mb-6">
          <View>
            <Text className="text-4xl font-manrope-bold mb-1" style={{ color: theme.text }}>
              {t.tasks.title}
            </Text>
            <Text className="text-base font-manrope" style={{ color: theme.textSecondary }}>
              {interpolate(t.tasks.assignedToYou, { count: getMyTasksCount() })}
            </Text>
          </View>
        </View>

        {/* Filter Tabs */}
        <View className="flex-row gap-2.5 mb-6" style={{ flexWrap: isDesktop ? "wrap" : "nowrap" }}>
          {[
            { key: "All" as FilterType, label: t.tasks.filters.all },
            { key: "My" as FilterType, label: t.tasks.filters.my },
            { key: "By Room" as FilterType, label: t.tasks.filters.byRoom },
            { key: "Completed" as FilterType, label: t.tasks.filters.completed },
          ].map((filter) => (
            <TouchableOpacity
              key={filter.key}
              className={`px-5 py-3 rounded-[12px] border ${
                activeFilter === filter.key ? "bg-primary border-primary" : ""
              }`}
              style={[
                { backgroundColor: theme.surface, borderColor: theme.border },
                activeFilter === filter.key && { backgroundColor: "#1C1C1E", borderColor: "#1C1C1E" },
              ]}
              onPress={() => setActiveFilter(filter.key)}
              activeOpacity={0.8}
            >
              <Text
                className={`text-sm font-manrope-semibold ${activeFilter === filter.key ? "text-white" : ""}`}
                style={activeFilter !== filter.key ? { color: theme.textSecondary } : undefined}
              >
                {filter.label}
              </Text>
            </TouchableOpacity>
          ))}
        </View>

        {/* Tasks List */}
        {getFilteredTasks().length === 0 ? (
          <View className="items-center py-20 px-6">
            <View
              className="w-16 h-16 rounded-full justify-center items-center mb-4"
              style={{ backgroundColor: theme.surface }}
            >
              <CheckCircle size={32} color={theme.textSecondary} />
            </View>
            <Text className="text-xl font-manrope-bold mb-2 text-center" style={{ color: theme.text }}>
              {t.tasks.noTasks}
            </Text>
            <Text className="text-sm font-manrope text-center leading-5 mb-6" style={{ color: theme.textSecondary }}>
              {t.tasks.noTasksHint}
            </Text>
            <Button title={t.tasks.newTask || "Create Task"} onPress={() => setShowCreateModal(true)} />
          </View>
        ) : (
          <View
            className="gap-3"
            style={{
              flexDirection: isDesktop ? "row" : "column",
              flexWrap: isDesktop ? "wrap" : "nowrap",
              justifyContent: "space-between",
            }}
          >
            {getFilteredTasks().map((task, index) => renderTaskItem(task, index))}
          </View>
        )}
      </ScrollView>

      {/* Floating Add Task FAB */}
      <TouchableOpacity
        className="absolute bottom-[120px] right-6 w-14 h-14 rounded-[18px] justify-center items-center shadow-lg z-40"
        style={{ backgroundColor: theme.accent.purple }}
        onPress={() => setShowCreateModal(true)}
        activeOpacity={0.8}
      >
        <Plus size={28} color="#1C1C1E" strokeWidth={2.5} />
      </TouchableOpacity>

      {/* Create Task Modal */}
      <Modal visible={showCreateModal} onClose={() => setShowCreateModal(false)} title={t.tasks.newTask} height="full">
        <View className="flex-1">
          <Input
            label={t.tasks.taskName}
            placeholder={t.tasks.taskNamePlaceholder}
            value={newTaskName}
            onChangeText={setNewTaskName}
          />

          <Input
            label={t.tasks.description}
            placeholder={t.tasks.descriptionPlaceholder}
            value={newTaskDescription}
            onChangeText={setNewTaskDescription}
            multiline
            numberOfLines={3}
          />

          <View className="mb-6">
            <Text
              className="text-xs font-manrope-bold uppercase tracking-wide mb-3 ml-1"
              style={{ color: theme.textSecondary }}
            >
              {t.tasks.dueDate}
            </Text>

            <TouchableOpacity
              onPress={showDatePicker}
              className="rounded-[12px] p-4 h-14 justify-center"
              style={{ backgroundColor: theme.surface }}
            >
              <Text
                className="text-base font-manrope-semibold"
                style={{ color: selectedDate ? theme.text : theme.textSecondary }}
              >
                {selectedDate
                  ? selectedDate.toLocaleString([], {
                      year: "numeric",
                      month: "2-digit",
                      day: "2-digit",
                      hour: "2-digit",
                      minute: "2-digit",
                    })
                  : t.tasks.selectDateTime}
              </Text>
            </TouchableOpacity>

            <DatePicker
              visible={isDatePickerVisible}
              onClose={hideDatePicker}
              onConfirm={handleConfirmDate}
              value={selectedDate ?? undefined}
              mode="datetime"
              minimumDate={new Date()}
              title={t.tasks.selectDateTime}
            />
          </View>

          <Input
            label={t.tasks.reminderBefore}
            placeholder="30"
            value={reminderMinutesInput}
            onChangeText={(value) => setReminderMinutesInput(value.replace(/[^0-9]/g, ""))}
            keyboardType="number-pad"
          />

          {rooms.length > 0 && (
            <View className="mb-6">
              <Text
                className="text-xs font-manrope-bold uppercase tracking-wide mb-3 ml-1"
                style={{ color: theme.textSecondary }}
              >
                {t.tasks.room}
              </Text>
              <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                <View className="flex-row gap-2.5">
                  <TouchableOpacity
                    className="px-4.5 py-3 rounded-[12px]"
                    style={[{ backgroundColor: theme.surface }, !selectedRoomId && { backgroundColor: theme.text }]}
                    onPress={() => setSelectedRoomId(null)}
                  >
                    <Text
                      className="text-sm font-manrope-semibold"
                      style={[{ color: theme.textSecondary }, !selectedRoomId && { color: theme.background }]}
                    >
                      {t.common.none}
                    </Text>
                  </TouchableOpacity>
                  {rooms.map((room) => (
                    <TouchableOpacity
                      key={room.id}
                      className="px-4.5 py-3 rounded-[12px]"
                      style={[
                        { backgroundColor: theme.surface },
                        selectedRoomId === room.id && { backgroundColor: theme.text },
                      ]}
                      onPress={() => setSelectedRoomId(room.id)}
                    >
                      <Text
                        className="text-sm font-manrope-semibold"
                        style={[
                          { color: theme.textSecondary },
                          selectedRoomId === room.id && { color: theme.background },
                        ]}
                      >
                        {room.name}
                      </Text>
                    </TouchableOpacity>
                  ))}
                </View>
              </ScrollView>
            </View>
          )}

          {home?.memberships && home.memberships.length > 0 && (
            <View className="mb-6">
              <Text
                className="text-xs font-manrope-bold uppercase tracking-wide mb-3 ml-1"
                style={{ color: theme.textSecondary }}
              >
                {t.tasks.assignTo}
              </Text>
              <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                <View className="flex-row gap-2.5">
                  {home.memberships.map((membership) => {
                    const isSelected = selectedUserIds.includes(membership.userId);
                    return (
                      <TouchableOpacity
                        key={membership.userId}
                        className="px-4.5 py-3 rounded-[12px]"
                        style={[{ backgroundColor: theme.surface }, isSelected && { backgroundColor: theme.text }]}
                        onPress={() =>
                          setSelectedUserIds((prev) =>
                            isSelected ? prev.filter((id) => id !== membership.userId) : [...prev, membership.userId],
                          )
                        }
                      >
                        <Text
                          className="text-sm font-manrope-semibold"
                          style={[{ color: theme.textSecondary }, isSelected && { color: theme.background }]}
                        >
                          {membership.user?.name || `User ${membership.userId}`}
                        </Text>
                      </TouchableOpacity>
                    );
                  })}
                </View>
              </ScrollView>
            </View>
          )}

          <Button
            title={t.tasks.createTask}
            onPress={handleCreateTask}
            loading={creating}
            disabled={!newTaskName.trim() || creating}
            variant="yellow"
            style={{ marginTop: "auto" }}
          />
        </View>
      </Modal>

      <Modal
        visible={showTaskActionsModal}
        onClose={() => {
          setShowTaskActionsModal(false);
          setSelectedTaskForActions(null);
        }}
        title={selectedTaskForActions?.name || "Task actions"}
        height="auto"
      >
        <View className="gap-3">
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.surface }}
            onPress={() => {
              const task = selectedTaskForActions;
              setShowTaskActionsModal(false);
              setSelectedTaskForActions(null);
              if (task) {
                router.push({ pathname: "/tasks/[id]", params: { id: String(task.id) } });
              }
            }}
          >
            <Text className="font-manrope-semibold" style={{ color: theme.text }}>
              {t.common.open}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.surface }}
            onPress={() => {
              const task = selectedTaskForActions;
              setShowTaskActionsModal(false);
              setSelectedTaskForActions(null);
              if (task) {
                openEditTaskModal(task);
              }
            }}
          >
            <Text className="font-manrope-semibold" style={{ color: theme.text }}>
              {t.common.edit}
            </Text>
          </TouchableOpacity>

          {isAdmin && selectedTaskForActions && !selectedTaskForActions.schedule && (
            <TouchableOpacity
              className="h-12 rounded-xl justify-center items-center"
              style={{ backgroundColor: theme.surface }}
              onPress={() => {
                const task = selectedTaskForActions;
                setShowTaskActionsModal(false);
                setSelectedTaskForActions(null);
                if (task) {
                  handleOpenScheduleModal(task.id);
                }
              }}
            >
              <Text className="font-manrope-semibold" style={{ color: theme.text }}>
                {t.tasks.schedule.create}
              </Text>
            </TouchableOpacity>
          )}

          {isAdmin && selectedTaskForActions?.schedule && (
            <TouchableOpacity
              className="h-12 rounded-xl justify-center items-center"
              style={{ backgroundColor: theme.surface }}
              onPress={() => {
                const task = selectedTaskForActions;
                setShowTaskActionsModal(false);
                setSelectedTaskForActions(null);
                if (task) {
                  handleDeleteSchedule(task);
                }
              }}
            >
              <Text className="font-manrope-semibold" style={{ color: theme.text }}>
                {t.tasks.schedule.deleteAction}
              </Text>
            </TouchableOpacity>
          )}

          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.accent.dangerLight }}
            onPress={() => {
              const task = selectedTaskForActions;
              setShowTaskActionsModal(false);
              setSelectedTaskForActions(null);
              if (task) {
                handleDelete(task.id);
              }
            }}
          >
            <Text className="font-manrope-semibold text-white">{t.common.delete}</Text>
          </TouchableOpacity>
        </View>
      </Modal>

      <Modal
        visible={showEditTaskModal}
        onClose={() => {
          setShowEditTaskModal(false);
          setEditingTaskId(null);
        }}
        title={t.tasks.editTask}
        height="full"
      >
        <View className="flex-1">
          <Input placeholder={t.tasks.taskNamePlaceholder} value={editTaskName} onChangeText={setEditTaskName} />
          <Input
            placeholder={t.tasks.descriptionPlaceholder}
            value={editTaskDescription}
            onChangeText={setEditTaskDescription}
            multiline
            numberOfLines={3}
          />

          <TouchableOpacity
            onPress={() => setIsEditDatePickerVisible(true)}
            className="rounded-[12px] p-4 h-14 justify-center mb-4"
            style={{ backgroundColor: theme.surface }}
          >
            <Text className="text-base font-manrope-semibold" style={{ color: theme.text }}>
              {editTaskDate ? editTaskDate.toLocaleString() : t.tasks.selectDateTime}
            </Text>
          </TouchableOpacity>
          <DatePicker
            visible={isEditDatePickerVisible}
            onClose={() => setIsEditDatePickerVisible(false)}
            onConfirm={(date) => {
              setEditTaskDate(date);
              setIsEditDatePickerVisible(false);
            }}
            value={editTaskDate ?? undefined}
            mode="datetime"
            title={t.tasks.selectDateTime}
          />

          <Input
            label={t.tasks.reminderBefore}
            placeholder="30"
            value={editReminderMinutesInput}
            onChangeText={(value) => setEditReminderMinutesInput(value.replace(/[^0-9]/g, ""))}
            keyboardType="number-pad"
          />

          {rooms.length > 0 && (
            <ScrollView horizontal showsHorizontalScrollIndicator={false} className="mb-6">
              <View className="flex-row gap-2.5">
                <TouchableOpacity
                  className="px-4.5 py-3 rounded-[12px]"
                  style={[
                    { backgroundColor: theme.surface },
                    editTaskRoomId === null && { backgroundColor: theme.text },
                  ]}
                  onPress={() => setEditTaskRoomId(null)}
                >
                  <Text
                    className="text-sm font-manrope-semibold"
                    style={[{ color: theme.textSecondary }, editTaskRoomId === null && { color: theme.background }]}
                  >
                    {t.tasks.noRoom}
                  </Text>
                </TouchableOpacity>
                {rooms.map((room) => (
                  <TouchableOpacity
                    key={room.id}
                    className="px-4.5 py-3 rounded-[12px]"
                    style={[
                      { backgroundColor: theme.surface },
                      editTaskRoomId === room.id && { backgroundColor: theme.text },
                    ]}
                    onPress={() => setEditTaskRoomId(room.id)}
                  >
                    <Text
                      className="text-sm font-manrope-semibold"
                      style={[
                        { color: theme.textSecondary },
                        editTaskRoomId === room.id && { color: theme.background },
                      ]}
                    >
                      {room.name}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </ScrollView>
          )}

          {home?.memberships && home.memberships.length > 0 && (
            <View className="mb-6">
              <Text
                className="text-xs font-manrope-bold uppercase tracking-wide mb-3 ml-1"
                style={{ color: theme.textSecondary }}
              >
                {t.tasks.assignTo}
              </Text>
              <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                <View className="flex-row gap-2.5">
                  {home.memberships.map((membership) => {
                    const isSelected = editTaskUserIds.includes(membership.userId);
                    return (
                      <TouchableOpacity
                        key={membership.userId}
                        className="px-4.5 py-3 rounded-[12px]"
                        style={[{ backgroundColor: theme.surface }, isSelected && { backgroundColor: theme.text }]}
                        onPress={() =>
                          setEditTaskUserIds((prev) =>
                            isSelected ? prev.filter((id) => id !== membership.userId) : [...prev, membership.userId],
                          )
                        }
                      >
                        <Text
                          className="text-sm font-manrope-semibold"
                          style={[{ color: theme.textSecondary }, isSelected && { color: theme.background }]}
                        >
                          {membership.user?.name || `User ${membership.userId}`}
                        </Text>
                      </TouchableOpacity>
                    );
                  })}
                </View>
              </ScrollView>
            </View>
          )}

          <Button
            title={t.common.save}
            onPress={handleSaveTaskEdit}
            loading={savingEditTask}
            disabled={!editTaskName.trim() || savingEditTask}
            variant="yellow"
            style={{ marginTop: "auto" }}
          />
        </View>
      </Modal>

      {/* Schedule Modal */}
      <Modal
        visible={showScheduleModal}
        onClose={() => setShowScheduleModal(false)}
        title={t.tasks.schedule.create}
        height="full"
      >
        <View className="flex-1">
          {/* Recurrence Type */}
          <View className="mb-6">
            <Text
              className="text-xs font-manrope-bold uppercase tracking-wide mb-3 ml-1"
              style={{ color: theme.textSecondary }}
            >
              {t.tasks.schedule.recurrence}
            </Text>
            <View className="flex-row gap-2.5">
              {(["daily", "weekly", "monthly"] as RecurrenceType[]).map((type) => (
                <TouchableOpacity
                  key={type}
                  className="flex-1 py-3 rounded-[12px] items-center"
                  style={[
                    { backgroundColor: theme.surface },
                    scheduleRecurrence === type && { backgroundColor: theme.accent.purple },
                  ]}
                  onPress={() => setScheduleRecurrence(type)}
                >
                  <Text
                    className="text-sm font-manrope-bold"
                    style={[{ color: theme.textSecondary }, scheduleRecurrence === type && { color: "#1C1C1E" }]}
                  >
                    {getRecurrenceLabel(type)}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>

          {/* Rotation Users */}
          {home?.memberships && home.memberships.length > 0 && (
            <View className="mb-6">
              <Text
                className="text-xs font-manrope-bold uppercase tracking-wide mb-3 ml-1"
                style={{ color: theme.textSecondary }}
              >
                {t.tasks.schedule.rotationUsers}
              </Text>
              <Text className="text-xs font-manrope mb-3 ml-1" style={{ color: theme.textMuted }}>
                {t.tasks.schedule.rotationHint}
              </Text>
              <View className="gap-2">
                {home.memberships.map((membership, _idx) => {
                  const isSelected = scheduleUserIds.includes(membership.userId);
                  const orderIndex = scheduleUserIds.indexOf(membership.userId);
                  return (
                    <TouchableOpacity
                      key={membership.userId}
                      className="flex-row items-center px-4 py-3 rounded-[12px]"
                      style={[
                        { backgroundColor: theme.surface },
                        isSelected && {
                          backgroundColor: `${theme.accent.purple}15`,
                          borderWidth: 1,
                          borderColor: theme.accent.purple,
                        },
                      ]}
                      onPress={() =>
                        setScheduleUserIds((prev) =>
                          isSelected ? prev.filter((id) => id !== membership.userId) : [...prev, membership.userId],
                        )
                      }
                    >
                      {isSelected && (
                        <View
                          className="w-6 h-6 rounded-full justify-center items-center mr-3"
                          style={{ backgroundColor: theme.accent.purple }}
                        >
                          <Text className="text-xs font-manrope-bold" style={{ color: "#1C1C1E" }}>
                            {orderIndex + 1}
                          </Text>
                        </View>
                      )}
                      <Text
                        className="text-sm font-manrope-semibold flex-1"
                        style={{ color: isSelected ? theme.text : theme.textSecondary }}
                      >
                        {membership.user?.name || `User ${membership.userId}`}
                      </Text>
                      {isSelected && <X size={16} color={theme.textSecondary} />}
                    </TouchableOpacity>
                  );
                })}
              </View>
              {scheduleUserIds.length === 0 && (
                <Text className="text-xs font-manrope mt-2 ml-1" style={{ color: theme.status.error }}>
                  {t.tasks.schedule.selectUsers}
                </Text>
              )}
            </View>
          )}

          <Button
            title={t.tasks.schedule.create}
            onPress={handleCreateSchedule}
            loading={creatingSchedule}
            disabled={scheduleUserIds.length === 0 || creatingSchedule}
            variant="purple"
            style={{ marginTop: "auto" }}
          />
        </View>
      </Modal>
    </View>
  );
}
