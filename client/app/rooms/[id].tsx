import { useLocalSearchParams, useRouter } from "expo-router";
import { ArrowLeft, Calendar, Fan, Lightbulb, Plus, Power, Trash2, Tv } from "lucide-react-native";
import { useCallback, useEffect, useState } from "react";
import { ActivityIndicator, FlatList, ScrollView, Switch, Text, TouchableOpacity, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { useAlert } from "@/components/ui/alert";
import Button from "@/components/ui/button";
import DatePicker from "@/components/ui/date-picker";
import Input from "@/components/ui/input";
import Modal from "@/components/ui/modal";
import { smarthomeApi, taskApi } from "@/lib/api";
import type { HAState, SmartDevice, Task } from "@/lib/types";
import { useResponsiveLayout } from "@/lib/useResponsiveLayout";
import { useHome } from "@/stores/homeStore";
import { interpolate, useI18n } from "@/stores/i18nStore";
import { useTheme } from "@/stores/themeStore";

export default function RoomDetailScreen() {
  const insets = useSafeAreaInsets();
  const router = useRouter();
  const { id, name } = useLocalSearchParams<{ id: string; name: string }>();
  const { home, isAdmin } = useHome();
  const { theme } = useTheme();
  const { t } = useI18n();
  const { alert } = useAlert();
  const { isDesktop, horizontalPadding, contentMaxWidth } = useResponsiveLayout();

  const [devices, setDevices] = useState<SmartDevice[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [deviceStates, setDeviceStates] = useState<Record<string, HAState>>({});
  const [loadingTasks, setLoadingTasks] = useState(true);
  const [loading, setLoading] = useState(true);

  // Add Device Modal State
  const [showAddModal, setShowAddModal] = useState(false);
  const [discoveredDevices, setDiscoveredDevices] = useState<HAState[]>([]);
  const [discovering, setDiscovering] = useState(false);
  const [selectedEntity, setSelectedEntity] = useState<string | null>(null);
  const [deviceName, setDeviceName] = useState("");
  const [addingDevice, setAddingDevice] = useState(false);
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [newTaskName, setNewTaskName] = useState("");
  const [newTaskDescription, setNewTaskDescription] = useState("");
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);
  const [isDatePickerVisible, setDatePickerVisibility] = useState(false);
  const [reminderMinutesInput, setReminderMinutesInput] = useState("30");
  const [creatingTask, setCreatingTask] = useState(false);

  const roomId = parseInt(id, 10);

  const fetchDashboardData = useCallback(async () => {
    if (!home) return;
    try {
      const [devicesResult, states, tasksResult] = await Promise.all([
        smarthomeApi.getDevicesByRoom(home.id, roomId).catch(() => []),
        smarthomeApi.getAllStates(home.id).catch(() => []),
        taskApi.getByHomeId(home.id).catch(() => []),
      ]);
      setDevices(devicesResult);
      const stateMap: Record<string, HAState> = {};
      states.forEach((s) => (stateMap[s.entityId] = s));
      setDeviceStates(stateMap);
      setTasks((tasksResult || []).filter((task) => task.roomId === roomId));
    } catch (error) {
      console.error("Failed to fetch room dashboard data:", error);
    } finally {
      setLoading(false);
      setLoadingTasks(false);
    }
  }, [home, roomId]);

  useEffect(() => {
    fetchDashboardData();
  }, [fetchDashboardData]);

  const handleDiscover = async () => {
    if (!home) return;
    setDiscovering(true);
    try {
      const results = await smarthomeApi.discover(home.id);
      setDiscoveredDevices(results);
    } catch (_error) {
      alert(t.common.error, t.smartHome.failedToDiscoverDevices);
    } finally {
      setDiscovering(false);
    }
  };

  const handleAddDevice = async () => {
    if (!home || !selectedEntity || !deviceName.trim()) return;

    setAddingDevice(true);
    try {
      await smarthomeApi.addDevice(home.id, {
        entityId: selectedEntity,
        name: deviceName.trim(),
        roomId: roomId,
      });
      setShowAddModal(false);
      setDeviceName("");
      setSelectedEntity(null);
      fetchDashboardData();
    } catch (err: any) {
      alert(t.common.error, err.response?.data?.error || t.smartHome.failedToAddDevice);
    } finally {
      setAddingDevice(false);
    }
  };

  const handleDeleteDevice = (deviceId: number, deviceName: string) => {
    if (!home) return;
    alert(t.smartHome.deleteDeviceTitle, interpolate(t.smartHome.deleteDeviceConfirm, { name: deviceName }), [
      { text: t.common.cancel, style: "cancel" },
      {
        text: t.common.delete,
        style: "destructive",
        onPress: async () => {
          try {
            await smarthomeApi.deleteDevice(home.id, deviceId);
            fetchDashboardData();
          } catch (_error) {
            alert(t.common.error, t.smartHome.failedToDeleteDevice);
          }
        },
      },
    ]);
  };

  const handleControl = async (device: SmartDevice, state: boolean) => {
    if (!home) return;
    const service = state ? "turn_on" : "turn_off";

    // Optimistic update
    const newState = { ...deviceStates[device.entityId], state: state ? "on" : "off" };
    setDeviceStates((prev) => ({ ...prev, [device.entityId]: newState }));

    try {
      await smarthomeApi.controlDevice(home.id, device.id, service);
    } catch (error) {
      console.error("Control failed", error);
      fetchDashboardData();
    }
  };

  const showDatePicker = () => setDatePickerVisibility(true);
  const hideDatePicker = () => setDatePickerVisibility(false);

  const handleConfirmDate = (date: Date) => {
    setSelectedDate(date);
    hideDatePicker();
  };

  const parseReminderMinutes = (value: string) => {
    const trimmed = value.trim();
    if (!trimmed) return 30;
    const parsed = Number.parseInt(trimmed, 10);
    if (Number.isNaN(parsed) || parsed < 0) return 30;
    return parsed;
  };

  const handleCreateTask = async () => {
    if (!home || !newTaskName.trim()) return;

    setCreatingTask(true);
    try {
      await taskApi.create(home.id, {
        name: newTaskName.trim(),
        description: newTaskDescription.trim(),
        scheduleType: "once",
        dueDate: selectedDate ? selectedDate.toISOString() : undefined,
        reminderMinutes: parseReminderMinutes(reminderMinutesInput),
        homeId: home.id,
        roomId: roomId,
      });
      setShowTaskModal(false);
      setNewTaskName("");
      setNewTaskDescription("");
      setSelectedDate(null);
      setReminderMinutesInput("30");
      await fetchDashboardData();
    } catch (_error) {
      alert(t.common.error, t.tasks.couldNotCreate);
    } finally {
      setCreatingTask(false);
    }
  };

  const isTaskCompleted = (task: Task) => {
    if (task.assignments && task.assignments.length > 0) {
      return task.assignments.some((assignment) => assignment.status === "completed");
    }
    return false;
  };

  const formatTaskDueText = (task: Task) => {
    if (!task.dueDate) return t.tasks.noDueDate;
    const dueDate = new Date(task.dueDate);
    return `${dueDate.toLocaleDateString()} ${dueDate.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false })}`;
  };

  const taskStats = tasks.reduce(
    (acc, task) => {
      if (isTaskCompleted(task)) {
        acc.completed += 1;
      } else {
        acc.pending += 1;
      }
      return acc;
    },
    { pending: 0, completed: 0 },
  );

  const getIcon = (type: string, size = 24, color = "black") => {
    switch (type) {
      case "light":
        return <Lightbulb size={size} color={color} />;
      case "switch":
        return <Power size={size} color={color} />;
      case "fan":
        return <Fan size={size} color={color} />;
      case "media_player":
        return <Tv size={size} color={color} />;
      default:
        return <Power size={size} color={color} />;
    }
  };

  const renderDevice = ({ item }: { item: SmartDevice }) => {
    const state = deviceStates[item.entityId];
    const isOn = state?.state === "on";
    const isOffline = state?.state === "unavailable" || state?.state === "unknown";

    return (
      <View className="flex-row items-center p-4 rounded-20 mb-3" style={{ backgroundColor: theme.surface }}>
        <View className="w-10 h-10 justify-center items-center bg-black/5 rounded-12 mr-4">
          {getIcon(item.type, 24, isOn ? theme.accent.yellow : theme.text)}
        </View>
        <View className="flex-1">
          <Text className="text-base font-manrope-bold" style={{ color: theme.text }}>
            {item.name}
          </Text>
          <Text className="text-xs font-manrope" style={{ color: theme.textSecondary }}>
            {state?.state || t.smartHome.unknownState}
          </Text>
        </View>

        <Switch
          value={isOn}
          onValueChange={(val) => handleControl(item, val)}
          disabled={isOffline}
          trackColor={{ false: theme.border, true: theme.accent.cyan }}
        />

        {isAdmin && (
          <TouchableOpacity className="ml-3 p-2" onPress={() => handleDeleteDevice(item.id, item.name)}>
            <Trash2 size={16} color={theme.accent.danger} />
          </TouchableOpacity>
        )}
      </View>
    );
  };

  return (
    <View className="flex-1" style={{ backgroundColor: theme.background }}>
      {/* Header */}
      <View
        className="flex-row items-center justify-between mb-6"
        style={{
          paddingTop: insets.top + 16,
          paddingHorizontal: horizontalPadding,
          width: "100%",
          maxWidth: contentMaxWidth,
          alignSelf: "center",
        }}
      >
        <TouchableOpacity
          onPress={() => router.back()}
          className="w-12 h-12 rounded-16 justify-center items-center"
          style={{ backgroundColor: theme.surface }}
        >
          <ArrowLeft size={24} color={theme.text} />
        </TouchableOpacity>
        <Text className="text-2xl font-manrope-bold" style={{ color: theme.text }}>
          {name}
        </Text>
        {isAdmin ? (
          <TouchableOpacity
            onPress={() => {
              setShowAddModal(true);
              handleDiscover();
            }}
            className="w-12 h-12 rounded-16 justify-center items-center"
            style={{ backgroundColor: theme.accent.yellow }}
          >
            <Plus size={24} color="#1C1C1E" />
          </TouchableOpacity>
        ) : (
          <View className="w-12" />
        )}
      </View>

      {/* Content */}
      <ScrollView
        className="flex-1"
        contentContainerStyle={{
          paddingHorizontal: horizontalPadding,
          paddingBottom: 40,
          width: "100%",
          maxWidth: contentMaxWidth,
          alignSelf: "center",
        }}
        showsVerticalScrollIndicator={false}
      >
        {loading || loadingTasks ? (
          <ActivityIndicator size="large" color={theme.accent.cyan} />
        ) : (
          <>
            <View className="flex-row gap-3 mb-5">
              <View className="flex-1 rounded-20 p-4" style={{ backgroundColor: theme.surface }}>
                <Text className="text-xs font-manrope-bold uppercase" style={{ color: theme.textSecondary }}>
                  {t.tasks.title}
                </Text>
                <Text className="text-2xl font-manrope-bold mt-1" style={{ color: theme.text }}>
                  {tasks.length}
                </Text>
              </View>
              <View className="flex-1 rounded-20 p-4" style={{ backgroundColor: theme.surface }}>
                <Text className="text-xs font-manrope-bold uppercase" style={{ color: theme.textSecondary }}>
                  {t.common.pending}
                </Text>
                <Text className="text-2xl font-manrope-bold mt-1" style={{ color: theme.accent.pink }}>
                  {taskStats.pending}
                </Text>
              </View>
              <View className="flex-1 rounded-20 p-4" style={{ backgroundColor: theme.surface }}>
                <Text className="text-xs font-manrope-bold uppercase" style={{ color: theme.textSecondary }}>
                  {t.smartHome.allDevices}
                </Text>
                <Text className="text-2xl font-manrope-bold mt-1" style={{ color: theme.accent.cyan }}>
                  {devices.length}
                </Text>
              </View>
            </View>

            <View style={{ flexDirection: isDesktop ? "row" : "column", gap: 24, alignItems: "flex-start" }}>
              <View className="mb-6" style={{ flex: isDesktop ? 1 : undefined, width: "100%" }}>
                <View className="flex-row items-center justify-between mb-3">
                  <Text className="text-lg font-manrope-bold" style={{ color: theme.text }}>
                    {t.rooms.roomTasks}
                  </Text>
                  <View className="flex-row items-center gap-2">
                    <TouchableOpacity
                      className="px-3 py-2"
                      style={{ backgroundColor: theme.surface, borderRadius: "12px" }}
                      onPress={() => router.push("/(tabs)/tasks")}
                    >
                      <Text className="text-xs font-manrope-semibold" style={{ color: theme.textSecondary }}>
                        {t.tasks.title}
                      </Text>
                    </TouchableOpacity>
                    <TouchableOpacity
                      className="px-3 py-2"
                      style={{ backgroundColor: theme.accent.yellow, borderRadius: "12px" }}
                      onPress={() => setShowTaskModal(true)}
                    >
                      <Text className="text-xs font-manrope-bold" style={{ color: "#1C1C1E" }}>
                        {t.tasks.newTask}
                      </Text>
                    </TouchableOpacity>
                  </View>
                </View>

                {tasks.length === 0 ? (
                  <View className="rounded-20 p-4" style={{ backgroundColor: theme.surface }}>
                    <Text style={{ color: theme.textSecondary }}>{t.tasks.noTasks}</Text>
                  </View>
                ) : (
                  <View className="gap-3">
                    {tasks.map((task) => {
                      const completed = isTaskCompleted(task);
                      return (
                        <TouchableOpacity
                          key={task.id}
                          className="rounded-20 p-4"
                          style={{ backgroundColor: theme.surface }}
                          onPress={() => router.push({ pathname: "/tasks/[id]", params: { id: String(task.id) } })}
                          activeOpacity={0.8}
                        >
                          <View className="flex-row items-start justify-between gap-3">
                            <View className="flex-1">
                              <Text
                                className={`text-base font-manrope-bold ${completed ? "line-through opacity-60" : ""}`}
                                style={{ color: theme.text }}
                              >
                                {task.name}
                              </Text>
                              {task.description ? (
                                <Text
                                  className={`text-sm mt-1 ${completed ? "line-through opacity-50" : ""}`}
                                  style={{ color: theme.textSecondary }}
                                  numberOfLines={2}
                                >
                                  {task.description}
                                </Text>
                              ) : null}
                              <View className="flex-row items-center gap-1 mt-2">
                                <Calendar size={12} color={theme.textSecondary} />
                                <Text className="text-xs font-manrope-semibold" style={{ color: theme.textSecondary }}>
                                  {formatTaskDueText(task)}
                                </Text>
                              </View>
                            </View>
                            {completed ? (
                              <View
                                className="px-2 py-1 rounded-full"
                                style={{ backgroundColor: `${theme.status.success}20` }}
                              >
                                <Text className="text-xs font-manrope-bold" style={{ color: theme.status.success }}>
                                  {t.common.done}
                                </Text>
                              </View>
                            ) : (
                              <View
                                className="px-2 py-1 rounded-full"
                                style={{ backgroundColor: theme.accent.pinkLight }}
                              >
                                <Text className="text-xs font-manrope-bold" style={{ color: theme.accent.pink }}>
                                  {t.tasks.schedule.active}
                                </Text>
                              </View>
                            )}
                          </View>
                        </TouchableOpacity>
                      );
                    })}
                  </View>
                )}
              </View>

              <View className="mb-4" style={{ flex: isDesktop ? 1 : undefined, width: "100%" }}>
                <Text className="text-lg font-manrope-bold mb-3" style={{ color: theme.text }}>
                  {t.smartHome.allDevices}
                </Text>
                {devices.length === 0 ? (
                  <View className="rounded-20 p-4" style={{ backgroundColor: theme.surface }}>
                    <Text style={{ color: theme.textSecondary }}>{t.smartHome.noDevicesInRoom}</Text>
                  </View>
                ) : (
                  devices.map((device) => <View key={device.id}>{renderDevice({ item: device })}</View>)
                )}
              </View>
            </View>
          </>
        )}
      </ScrollView>

      {/* Create Task Modal */}
      <Modal visible={showTaskModal} onClose={() => setShowTaskModal(false)} title={t.tasks.newTask} height="full">
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
          <Button
            title={t.tasks.createTask}
            onPress={handleCreateTask}
            loading={creatingTask}
            disabled={!newTaskName.trim() || creatingTask}
            variant="yellow"
            style={{ marginTop: "auto" }}
          />
        </View>
      </Modal>

      {/* Add Device Modal */}
      <Modal visible={showAddModal} onClose={() => setShowAddModal(false)} title={t.smartHome.addDevice} height="full">
        <View className="flex-1">
          {selectedEntity ? (
            <View>
              <TouchableOpacity onPress={() => setSelectedEntity(null)} className="mb-4">
                <Text style={{ color: theme.accent.cyan }}>{t.smartHome.backToList}</Text>
              </TouchableOpacity>
              <Input
                label={t.smartHome.deviceNameLabel}
                value={deviceName}
                onChangeText={setDeviceName}
                placeholder={t.smartHome.roomDeviceNamePlaceholder}
              />
              <Text className="mb-2" style={{ color: theme.textSecondary }}>
                Entity: {selectedEntity}
              </Text>
              <Button
                title={t.common.add || "Add"}
                onPress={handleAddDevice}
                loading={addingDevice}
                variant="primary"
              />
            </View>
          ) : (
            <>
              <Text className="mb-4" style={{ color: theme.textSecondary }}>
                Select a device to add to {name}
              </Text>
              {discovering ? (
                <ActivityIndicator color={theme.accent.cyan} />
              ) : (
                <FlatList
                  data={discoveredDevices}
                  keyExtractor={(item) => item.entityId}
                  renderItem={({ item }) => (
                    <TouchableOpacity
                      className="py-3 border-b"
                      style={{ borderBottomColor: theme.border }}
                      onPress={() => {
                        setSelectedEntity(item.entityId);
                        setDeviceName(item.attributes.friendlyName || item.entityId);
                      }}
                    >
                      <Text className="font-bold" style={{ color: theme.text }}>
                        {item.attributes.friendlyName || item.entityId}
                      </Text>
                      <Text className="text-xs" style={{ color: theme.textSecondary }}>
                        {item.entityId}
                      </Text>
                    </TouchableOpacity>
                  )}
                />
              )}
            </>
          )}
        </View>
      </Modal>
    </View>
  );
}
