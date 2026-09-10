import { useRouter } from "expo-router";
import {
  ArrowLeft,
  Bath,
  BedDouble,
  Book,
  Car,
  Check,
  ChefHat,
  Coffee,
  DoorOpen,
  Dumbbell,
  Home as HomeIcon,
  Lightbulb,
  Plus,
  Sofa,
  TreePine,
  Tv,
  Utensils,
  Wifi,
  Wrench,
} from "lucide-react-native";
import { useState } from "react";
import { ScrollView, Text, TouchableOpacity, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { useAlert } from "@/components/ui/alert";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Modal from "@/components/ui/modal";
import { roomApi } from "@/lib/api";
import { useResponsiveLayout } from "@/lib/useResponsiveLayout";
import { useHome } from "@/stores/homeStore";
import { interpolate, useI18n } from "@/stores/i18nStore";
import { useTheme } from "@/stores/themeStore";

const ROOM_COLOR_OPTIONS = [
  "#FF7476",
  "#FF9F7A",
  "#FBEB9E",
  "#A8E6CF",
  "#7DD3E8",
  "#D8D4FC",
  "#F5A3D3",
  "#22C55E",
  "#F472B6",
  "#C4B5FD",
  "#94A3B8",
  "#FDE68A",
  "#6EE7B7",
];

const ROOM_ICON_OPTIONS = [
  "home",
  "utensils",
  "lightbulb",
  "coffee",
  "wrench",
  "car",
  "book",
  "bed",
  "bath",
  "sofa",
  "tv",
  "wifi",
  "gym",
  "garden",
  "chef",
] as const;

const getRoomIcon = (iconId: string | undefined, size: number, color: string) => {
  switch (iconId) {
    case "utensils":
      return <Utensils size={size} color={color} />;
    case "lightbulb":
      return <Lightbulb size={size} color={color} />;
    case "coffee":
      return <Coffee size={size} color={color} />;
    case "wrench":
      return <Wrench size={size} color={color} />;
    case "car":
      return <Car size={size} color={color} />;
    case "book":
      return <Book size={size} color={color} />;
    case "bed":
      return <BedDouble size={size} color={color} />;
    case "bath":
      return <Bath size={size} color={color} />;
    case "sofa":
      return <Sofa size={size} color={color} />;
    case "tv":
      return <Tv size={size} color={color} />;
    case "wifi":
      return <Wifi size={size} color={color} />;
    case "gym":
      return <Dumbbell size={size} color={color} />;
    case "garden":
      return <TreePine size={size} color={color} />;
    case "chef":
      return <ChefHat size={size} color={color} />;
    default:
      return <HomeIcon size={size} color={color} />;
  }
};

export default function RoomsScreen() {
  const insets = useSafeAreaInsets();
  const router = useRouter();
  const { home, rooms, isAdmin, createRoom, deleteRoom, refreshRooms } = useHome();
  const { theme } = useTheme();
  const { t } = useI18n();
  const { alert } = useAlert();
  const { isDesktop, horizontalPadding, contentMaxWidth } = useResponsiveLayout();

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [roomName, setRoomName] = useState("");
  const [selectedColor, setSelectedColor] = useState(ROOM_COLOR_OPTIONS[2]);
  const [selectedIcon, setSelectedIcon] = useState<(typeof ROOM_ICON_OPTIONS)[number]>("home");
  const [isLoading, setIsLoading] = useState(false);
  const [selectedRoom, setSelectedRoom] = useState<(typeof rooms)[number] | null>(null);
  const [showRoomActionsModal, setShowRoomActionsModal] = useState(false);
  const [editingRoomId, setEditingRoomId] = useState<number | null>(null);

  const handleSaveRoom = async () => {
    if (!roomName.trim()) return;
    if (!home) return;

    setIsLoading(true);
    try {
      if (editingRoomId) {
        await roomApi.update(home.id, editingRoomId, {
          name: roomName.trim(),
          icon: selectedIcon,
          color: selectedColor,
        });
        await refreshRooms();
      } else {
        const result = await createRoom(roomName.trim(), selectedIcon, selectedColor);
        if (!result.success) {
          alert(t.common.error, result.error || t.rooms.failedToCreate);
          return;
        }
      }

      setShowCreateModal(false);
      setEditingRoomId(null);
      setRoomName("");
      setSelectedColor(ROOM_COLOR_OPTIONS[2]);
      setSelectedIcon("home");
    } catch (error) {
      console.error("Error saving room:", error);
      alert(t.common.error, editingRoomId ? t.rooms.failedToUpdate : t.rooms.failedToCreate);
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteRoom = (roomId: number, roomName: string) => {
    alert(t.rooms.deleteRoom, interpolate(t.rooms.deleteRoomConfirm, { name: roomName }), [
      { text: t.common.cancel, style: "cancel" },
      {
        text: t.common.delete,
        style: "destructive",
        onPress: async () => {
          const result = await deleteRoom(roomId);
          if (!result.success) {
            alert(t.common.error, result.error || t.rooms.failedToDelete);
          }
        },
      },
    ]);
  };

  const openRoomActions = (room: (typeof rooms)[number]) => {
    if (!isAdmin) return;
    setSelectedRoom(room);
    setShowRoomActionsModal(true);
  };

  const openEditRoom = (room: (typeof rooms)[number]) => {
    setEditingRoomId(room.id);
    setRoomName(room.name);
    setSelectedIcon((room.icon as (typeof ROOM_ICON_OPTIONS)[number]) || "home");
    setSelectedColor(room.color || ROOM_COLOR_OPTIONS[2]);
    setShowCreateModal(true);
  };

  if (!home) {
    return (
      <View className="flex-1" style={{ paddingTop: insets.top, backgroundColor: theme.background }}>
        <View className="flex-row items-center justify-between mb-8">
          <TouchableOpacity
            onPress={() => router.back()}
            className="w-12 h-12 rounded-16 justify-center items-center"
            style={{ backgroundColor: theme.surface }}
          >
            <ArrowLeft size={24} color={theme.text} />
          </TouchableOpacity>
          <Text className="text-2xl font-manrope-bold" style={{ color: theme.text }}>
            {t.rooms.title}
          </Text>
          <View className="w-12" />
        </View>
        <View className="flex-1 justify-center items-center py-15">
          <Text className="text-base font-manrope" style={{ color: theme.textSecondary }}>
            {t.rooms.joinHomeToManage}
          </Text>
        </View>
      </View>
    );
  }

  return (
    <View className="flex-1" style={{ backgroundColor: theme.background }}>
      <ScrollView
        className="flex-1"
        contentContainerStyle={{
          paddingHorizontal: horizontalPadding,
          paddingBottom: 40,
          paddingTop: insets.top + 16,
          width: "100%",
          maxWidth: contentMaxWidth,
          alignSelf: "center",
        }}
        showsVerticalScrollIndicator={false}
      >
        {/* Header */}
        <View className="flex-row items-center justify-between mb-8">
          <TouchableOpacity
            onPress={() => router.back()}
            className="w-12 h-12 rounded-16 justify-center items-center"
            style={{ backgroundColor: theme.surface }}
          >
            <ArrowLeft size={24} color={theme.text} />
          </TouchableOpacity>
          <Text className="text-2xl font-manrope-bold" style={{ color: theme.text }}>
            {t.rooms.title}
          </Text>
          {isAdmin ? (
            <TouchableOpacity
              onPress={() => setShowCreateModal(true)}
              className="w-12 h-12 rounded-16 justify-center items-center"
              style={{ backgroundColor: theme.accent.yellow }}
            >
              <Plus size={24} color="#1C1C1E" />
            </TouchableOpacity>
          ) : (
            <View className="w-12" />
          )}
        </View>

        {/* Rooms Grid */}
        {rooms.length === 0 ? (
          <View className="items-center py-20 px-6">
            <View
              className="w-16 h-16 rounded-full justify-center items-center mb-4"
              style={{ backgroundColor: theme.surface }}
            >
              <DoorOpen size={32} color={theme.textSecondary} />
            </View>
            <Text className="text-xl font-manrope-bold mb-2 text-center" style={{ color: theme.text }}>
              {t.rooms.noRooms}
            </Text>
            <Text className="text-sm font-manrope text-center leading-5 mb-6" style={{ color: theme.textSecondary }}>
              {isAdmin ? t.rooms.noRoomsAdminHint : t.rooms.noRoomsMemberHint}
            </Text>
            {isAdmin && <Button title={t.rooms.createRoom || "Create Room"} onPress={() => setShowCreateModal(true)} />}
          </View>
        ) : (
          <View
            className="flex-row flex-wrap gap-4"
            style={{ justifyContent: isDesktop ? "flex-start" : "space-between" }}
          >
            {rooms.map((room, index) => {
              const ROOM_COLORS = [
                theme.accent.yellow,
                theme.accent.purple,
                theme.accent.pink,
                theme.surface,
                theme.border,
              ];
              const colorIndex = index % ROOM_COLORS.length;
              const backgroundColor = room.color || ROOM_COLORS[colorIndex];
              const finalTextColor =
                backgroundColor === theme.surface || backgroundColor === theme.border ? theme.text : "#1C1C1E";

              return (
                <TouchableOpacity
                  key={room.id}
                  className="rounded-28 p-6 relative"
                  style={{ backgroundColor, width: isDesktop ? "23.8%" : "47%", minHeight: 160 }}
                  onPress={() =>
                    router.push({ pathname: "/rooms/[id]", params: { id: String(room.id), name: room.name } })
                  }
                  onLongPress={() => openRoomActions(room)}
                >
                  <View className="w-14 h-14 rounded-20 justify-center items-center mb-4 bg-black/10">
                    {getRoomIcon(room.icon, 28, finalTextColor)}
                  </View>
                  <Text className="text-lg font-manrope-bold mb-1" style={{ color: finalTextColor }}>
                    {room.name}
                  </Text>
                  <Text className="text-xs font-manrope" style={{ color: finalTextColor, opacity: 0.6 }}>
                    {interpolate(t.rooms.added, { date: new Date(room.createdAt).toLocaleDateString() })}
                  </Text>
                </TouchableOpacity>
              );
            })}
          </View>
        )}
      </ScrollView>

      {/* Create Room Modal */}
      <Modal
        visible={showCreateModal}
        onClose={() => {
          setShowCreateModal(false);
          setEditingRoomId(null);
        }}
        title={editingRoomId ? t.rooms.editRoom : t.rooms.newRoom}
        height="full"
      >
        <View className="flex-1">
          <View className="items-center mb-6">
            <View
              className="w-20 h-20 rounded-3xl justify-center items-center"
              style={{ backgroundColor: selectedColor }}
            >
              {getRoomIcon(selectedIcon, 32, "#1C1C1E")}
            </View>
          </View>

          <Input placeholder={t.rooms.roomNamePlaceholder} value={roomName} onChangeText={setRoomName} />

          <View className="mb-6 gap-3">
            <View className="flex-row justify-center gap-2.5">
              {ROOM_COLOR_OPTIONS.slice(0, 7).map((color) => (
                <TouchableOpacity
                  key={color}
                  className={`w-9 h-9 rounded-full ${selectedColor === color ? "border-[3px] border-black/30" : ""}`}
                  style={{ backgroundColor: color }}
                  onPress={() => setSelectedColor(color)}
                />
              ))}
            </View>
            <View className="flex-row justify-center gap-2.5">
              {ROOM_COLOR_OPTIONS.slice(7).map((color) => (
                <TouchableOpacity
                  key={color}
                  className={`w-9 h-9 rounded-full ${selectedColor === color ? "border-[3px] border-black/30" : ""}`}
                  style={{ backgroundColor: color }}
                  onPress={() => setSelectedColor(color)}
                />
              ))}
            </View>
          </View>

          <ScrollView className="max-h-[220px]" showsVerticalScrollIndicator={false}>
            <View className="gap-3">
              {Array.from({ length: Math.ceil(ROOM_ICON_OPTIONS.length / 6) }, (_, row) => (
                <View key={row} className="flex-row justify-center gap-2.5">
                  {ROOM_ICON_OPTIONS.slice(row * 6, row * 6 + 6).map((icon) => (
                    <TouchableOpacity
                      key={icon}
                      className="w-12 h-12 rounded-full justify-center items-center"
                      style={{ backgroundColor: selectedIcon === icon ? selectedColor : theme.surface }}
                      onPress={() => setSelectedIcon(icon)}
                    >
                      {getRoomIcon(icon, 20, selectedIcon === icon ? "#1C1C1E" : theme.textSecondary)}
                    </TouchableOpacity>
                  ))}
                </View>
              ))}
            </View>
          </ScrollView>
        </View>

        <View className="flex-row gap-3 pt-4">
          <Button
            title={editingRoomId ? t.common.save : t.rooms.createRoom}
            icon={<Check size={20} color={theme.isDark ? "#1C1C1E" : "#FFFFFF"} />}
            onPress={handleSaveRoom}
            disabled={!roomName.trim() || isLoading}
            className="flex-1"
          />
        </View>
      </Modal>

      <Modal
        visible={showRoomActionsModal}
        onClose={() => {
          setShowRoomActionsModal(false);
          setSelectedRoom(null);
        }}
        title={selectedRoom?.name || "Room actions"}
        height="auto"
      >
        <View className="gap-3">
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.surface }}
            onPress={() => {
              const room = selectedRoom;
              setShowRoomActionsModal(false);
              setSelectedRoom(null);
              if (room) {
                openEditRoom(room);
              }
            }}
          >
            <Text className="font-manrope-semibold" style={{ color: theme.text }}>
              {t.common.edit}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.accent.dangerLight }}
            onPress={() => {
              const room = selectedRoom;
              setShowRoomActionsModal(false);
              setSelectedRoom(null);
              if (room) {
                handleDeleteRoom(room.id, room.name);
              }
            }}
          >
            <Text className="font-manrope-semibold text-white">{t.common.delete}</Text>
          </TouchableOpacity>
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.surface }}
            onPress={() => {
              setShowRoomActionsModal(false);
              setSelectedRoom(null);
            }}
          >
            <Text className="font-manrope-semibold" style={{ color: theme.text }}>
              {t.common.cancel}
            </Text>
          </TouchableOpacity>
        </View>
      </Modal>
    </View>
  );
}
