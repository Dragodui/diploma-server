import {
  Apple,
  ArrowLeft,
  Baby,
  Beef,
  Book,
  Cake,
  Candy,
  Car,
  Carrot,
  Check,
  ChevronRight,
  Coffee,
  Cookie,
  Dog,
  Fish,
  Gift,
  Home,
  Lightbulb,
  Milk,
  Pill,
  Plus,
  Scissors,
  Shirt,
  ShoppingCart,
  Sparkles,
  Trash2,
  Utensils,
  Wine,
  Wrench,
} from "lucide-react-native";
import { useCallback, useEffect, useRef, useState } from "react";
import { RefreshControl, ScrollView, Text, TouchableOpacity, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { ShoppingSkeleton } from "@/components/skeletons";
import { useAlert } from "@/components/ui/alert";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Modal from "@/components/ui/modal";
import { shoppingApi } from "@/lib/api";
import type { ShoppingCategory, ShoppingItem } from "@/lib/types";
import { useRealtimeRefresh } from "@/lib/useRealtimeRefresh";
import { useResponsiveLayout } from "@/lib/useResponsiveLayout";
import { useHome } from "@/stores/homeStore";
import { useI18n } from "@/stores/i18nStore";
import { useTheme } from "@/stores/themeStore";

// Category colors matching PDF
const CATEGORY_COLORS = ["#D8D4FC", "#FBEB9E", "#FF7476", "#A8E6CF", "#7DD3E8", "#F5A3D3"];

// Available icons for categories
const ICON_OPTIONS = [
  { id: "utensils", label: "Food" },
  { id: "shopping-cart", label: "Cart" },
  { id: "coffee", label: "Coffee" },
  { id: "wine", label: "Drinks" },
  { id: "milk", label: "Dairy" },
  { id: "beef", label: "Meat" },
  { id: "fish", label: "Fish" },
  { id: "carrot", label: "Veggies" },
  { id: "apple", label: "Fruits" },
  { id: "candy", label: "Sweets" },
  { id: "cake", label: "Bakery" },
  { id: "cookie", label: "Snacks" },
  { id: "pill", label: "Medicine" },
  { id: "baby", label: "Baby" },
  { id: "dog", label: "Pets" },
  { id: "shirt", label: "Clothes" },
  { id: "sparkles", label: "Cleaning" },
  { id: "scissors", label: "Beauty" },
  { id: "home", label: "Home" },
  { id: "lightbulb", label: "Electronics" },
  { id: "wrench", label: "Tools" },
  { id: "car", label: "Auto" },
  { id: "book", label: "Books" },
  { id: "gift", label: "Gifts" },
];

const getIconComponent = (iconId: string, size: number = 24, color: string = "#1C1C1E") => {
  switch (iconId) {
    case "utensils":
      return <Utensils size={size} color={color} />;
    case "shopping-cart":
      return <ShoppingCart size={size} color={color} />;
    case "coffee":
      return <Coffee size={size} color={color} />;
    case "wine":
      return <Wine size={size} color={color} />;
    case "milk":
      return <Milk size={size} color={color} />;
    case "beef":
      return <Beef size={size} color={color} />;
    case "fish":
      return <Fish size={size} color={color} />;
    case "carrot":
      return <Carrot size={size} color={color} />;
    case "apple":
      return <Apple size={size} color={color} />;
    case "candy":
      return <Candy size={size} color={color} />;
    case "cake":
      return <Cake size={size} color={color} />;
    case "cookie":
      return <Cookie size={size} color={color} />;
    case "pill":
      return <Pill size={size} color={color} />;
    case "baby":
      return <Baby size={size} color={color} />;
    case "dog":
      return <Dog size={size} color={color} />;
    case "shirt":
      return <Shirt size={size} color={color} />;
    case "sparkles":
      return <Sparkles size={size} color={color} />;
    case "scissors":
      return <Scissors size={size} color={color} />;
    case "home":
      return <Home size={size} color={color} />;
    case "lightbulb":
      return <Lightbulb size={size} color={color} />;
    case "wrench":
      return <Wrench size={size} color={color} />;
    case "car":
      return <Car size={size} color={color} />;
    case "book":
      return <Book size={size} color={color} />;
    case "gift":
      return <Gift size={size} color={color} />;
    default:
      return <Utensils size={size} color={color} />;
  }
};

// Color options for creating new lists
const COLOR_OPTIONS = [
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

const getItemTimestamp = (value?: string) => {
  if (!value) return 0;
  const parsed = new Date(value).getTime();
  return Number.isNaN(parsed) ? 0 : parsed;
};

const sortShoppingItems = (shoppingItems: ShoppingItem[]) =>
  [...shoppingItems].sort((left, right) => {
    if (left.isBought !== right.isBought) {
      return left.isBought ? 1 : -1;
    }

    if (!left.isBought) {
      return getItemTimestamp(left.createdAt) - getItemTimestamp(right.createdAt);
    }

    const boughtDiff = getItemTimestamp(right.boughtDate) - getItemTimestamp(left.boughtDate);
    if (boughtDiff !== 0) {
      return boughtDiff;
    }

    return getItemTimestamp(right.createdAt) - getItemTimestamp(left.createdAt);
  });

export default function ShoppingScreen() {
  const insets = useSafeAreaInsets();
  const { home } = useHome();
  const { theme } = useTheme();
  const { t } = useI18n();
  const { alert } = useAlert();
  const { isDesktop, horizontalPadding, contentMaxWidth } = useResponsiveLayout();

  const [categories, setCategories] = useState<ShoppingCategory[]>([]);
  const [items, setItems] = useState<Record<number, ShoppingItem[]>>({});
  const [activeCategory, setActiveCategory] = useState<ShoppingCategory | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  // Create category modal
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState("");
  const [selectedColor, setSelectedColor] = useState(COLOR_OPTIONS[0]);
  const [selectedIcon, setSelectedIcon] = useState(ICON_OPTIONS[0].id);
  const [creatingCategory, setCreatingCategory] = useState(false);
  const [editingCategoryId, setEditingCategoryId] = useState<number | null>(null);
  const [showCategoryActionsModal, setShowCategoryActionsModal] = useState(false);
  const [selectedCategoryForActions, setSelectedCategoryForActions] = useState<ShoppingCategory | null>(null);

  // Delete category
  const [deletingCategory, setDeletingCategory] = useState(false);

  // Create item modal
  const [showItemModal, setShowItemModal] = useState(false);
  const [newItemName, setNewItemName] = useState("");
  const [pendingItemNames, setPendingItemNames] = useState<string[]>([]);
  const [creatingItem, setCreatingItem] = useState(false);
  const [selectedItemForActions, setSelectedItemForActions] = useState<ShoppingItem | null>(null);
  const [showItemActionsModal, setShowItemActionsModal] = useState(false);
  const [showEditItemModal, setShowEditItemModal] = useState(false);
  const [editItemName, setEditItemName] = useState("");
  const [savingItemEdit, setSavingItemEdit] = useState(false);
  const suppressRealtimeRefreshUntilRef = useRef(0);

  const loadShoppingData = useCallback(async () => {
    if (!home) {
      setIsLoading(false);
      return;
    }

    try {
      const categoriesData = await shoppingApi.getCategories(home.id);
      setCategories(categoriesData || []);

      if (categoriesData && categoriesData.length > 0) {
        const results = await Promise.all(
          categoriesData.map((category) => shoppingApi.getCategoryItems(home.id, category.id).catch(() => [])),
        );
        const itemsData: Record<number, ShoppingItem[]> = {};
        categoriesData.forEach((category, i) => {
          itemsData[category.id] = sortShoppingItems(results[i] || []);
        });
        setItems(itemsData);
      }
    } catch (error) {
      console.error("Error loading shopping data:", error);
    } finally {
      setIsLoading(false);
    }
  }, [home]);

  useEffect(() => {
    loadShoppingData();
  }, [loadShoppingData]);

  const handleRealtimeRefresh = useCallback(() => {
    if (Date.now() < suppressRealtimeRefreshUntilRef.current) {
      return;
    }
    loadShoppingData();
  }, [loadShoppingData]);

  useRealtimeRefresh(["SHOPPING_CATEGORY", "SHOPPING_ITEM"], handleRealtimeRefresh);

  const onRefresh = async () => {
    setRefreshing(true);
    await loadShoppingData();
    setRefreshing(false);
  };

  const handleCreateCategory = async () => {
    if (!home || !newCategoryName.trim()) return;

    setCreatingCategory(true);
    try {
      if (editingCategoryId) {
        await shoppingApi.editCategory(home.id, editingCategoryId, {
          name: newCategoryName.trim(),
          icon: selectedIcon,
          color: selectedColor,
        });
      } else {
        await shoppingApi.createCategory(home.id, {
          name: newCategoryName.trim(),
          icon: selectedIcon,
          color: selectedColor,
        });
      }

      setNewCategoryName("");
      setSelectedIcon(ICON_OPTIONS[0].id);
      setSelectedColor(COLOR_OPTIONS[0]);
      setShowCategoryModal(false);
      setEditingCategoryId(null);
      await loadShoppingData();
    } catch (error) {
      console.error("Error creating category:", error);
    } finally {
      setCreatingCategory(false);
    }
  };

  const openCategoryActions = (category: ShoppingCategory) => {
    setSelectedCategoryForActions(category);
    setShowCategoryActionsModal(true);
  };

  const openEditCategory = (category: ShoppingCategory) => {
    setEditingCategoryId(category.id);
    setNewCategoryName(category.name || "");
    setSelectedIcon(category.icon || ICON_OPTIONS[0].id);
    setSelectedColor(category.color || COLOR_OPTIONS[0]);
    setShowCategoryModal(true);
  };

  const openItemModal = () => {
    setNewItemName("");
    setPendingItemNames([]);
    setShowItemModal(true);
  };

  const closeItemModal = () => {
    setShowItemModal(false);
    setNewItemName("");
    setPendingItemNames([]);
  };

  const handleAddPendingItem = () => {
    const name = newItemName.trim();
    if (!name) return;
    setPendingItemNames((prev) => [...prev, name]);
    setNewItemName("");
  };

  const handleCreateItem = async () => {
    if (!home || !activeCategory) return;

    const itemsToCreate = [...pendingItemNames];
    const currentName = newItemName.trim();
    if (currentName) {
      itemsToCreate.push(currentName);
    }
    if (itemsToCreate.length === 0) return;

    setCreatingItem(true);
    try {
      await shoppingApi.createItems(home.id, {
        categoryId: activeCategory.id,
        items: itemsToCreate.map((itemName) => ({ name: itemName })),
      });

      setNewItemName("");
      setPendingItemNames([]);
      setShowItemModal(false);
      await loadShoppingData();
    } catch (error) {
      console.error("Error creating item:", error);
    } finally {
      setCreatingItem(false);
    }
  };

  const toggleItemBought = async (itemId: number) => {
    if (!home) return;

    const categoryEntry = Object.entries(items).find(([, categoryItems]) =>
      categoryItems.some((item) => item.id === itemId),
    );
    if (!categoryEntry) return;

    const [categoryKey, categoryItems] = categoryEntry;
    const categoryId = Number(categoryKey);
    const currentItem = categoryItems.find((item) => item.id === itemId);
    if (!currentItem) return;

    const previousCategoryItems = categoryItems;
    const nowIso = new Date().toISOString();
    suppressRealtimeRefreshUntilRef.current = Date.now() + 1200;

    setItems((prev) => ({
      ...prev,
      [categoryId]: sortShoppingItems(
        (prev[categoryId] || []).map((item) =>
          item.id === itemId
            ? {
                ...item,
                isBought: !item.isBought,
                boughtDate: item.isBought ? undefined : nowIso,
              }
            : item,
        ),
      ),
    }));

    try {
      await shoppingApi.markBought(home.id, itemId);
    } catch (error) {
      console.error("Error toggling item:", error);
      suppressRealtimeRefreshUntilRef.current = 0;
      setItems((prev) => ({
        ...prev,
        [categoryId]: previousCategoryItems,
      }));
    }
  };

  const handleDeleteItem = async (itemId: number) => {
    if (!home) return;

    try {
      await shoppingApi.deleteItem(home.id, itemId);
      await loadShoppingData();
    } catch (error) {
      console.error("Error deleting item:", error);
    }
  };

  const openItemActions = (item: ShoppingItem) => {
    setSelectedItemForActions(item);
    setShowItemActionsModal(true);
  };

  const handleEditItem = async () => {
    if (!home || !selectedItemForActions || !editItemName.trim()) return;
    setSavingItemEdit(true);
    try {
      await shoppingApi.editItem(home.id, selectedItemForActions.id, { name: editItemName.trim() });
      setShowEditItemModal(false);
      setSelectedItemForActions(null);
      await loadShoppingData();
    } catch (error) {
      console.error("Error editing item:", error);
    } finally {
      setSavingItemEdit(false);
    }
  };

  const handleDeleteCategory = async (category: ShoppingCategory) => {
    if (!home) return;
    setDeletingCategory(true);
    try {
      await shoppingApi.deleteCategory(home.id, category.id);
      if (activeCategory?.id === category.id) {
        setActiveCategory(null);
      }
      await loadShoppingData();
    } catch (error) {
      console.error("Error deleting category:", error);
    } finally {
      setDeletingCategory(false);
    }
  };

  const openDeleteCategory = (category: ShoppingCategory) => {
    if (deletingCategory) return;
    alert(t.common.delete, `${t.common.delete} "${category.name}"?`, [
      { text: t.common.cancel, style: "cancel" },
      {
        text: t.common.delete,
        style: "destructive",
        onPress: () => {
          handleDeleteCategory(category);
        },
      },
    ]);
  };

  const getCategoryIcon = (category: ShoppingCategory) => {
    return getIconComponent(category.icon || "utensils", 24, "#1C1C1E");
  };

  const getActiveItems = () => {
    if (!activeCategory) return [];
    return items[activeCategory.id] || [];
  };

  if (isLoading) {
    return <ShoppingSkeleton />;
  }

  // List detail view
  if (activeCategory) {
    const categoryItems = getActiveItems();
    const pendingItems = categoryItems.filter((item) => !item.isBought);
    const boughtItems = categoryItems.filter((item) => item.isBought);
    const categoryColor = activeCategory.color || CATEGORY_COLORS[0];

    return (
      <View className="flex-1" style={{ backgroundColor: theme.background }}>
        <ScrollView
          className="flex-1"
          contentContainerStyle={{
            paddingHorizontal: horizontalPadding,
            paddingBottom: isDesktop ? 48 : 120,
            paddingTop: insets.top + 16,
            width: "100%",
            maxWidth: isDesktop ? 960 : contentMaxWidth,
            alignSelf: "center",
          }}
          showsVerticalScrollIndicator={false}
        >
          {/* Detail Header */}
          <View className="flex-row items-center mb-8 gap-3">
            <TouchableOpacity
              className="w-12 h-12 rounded-2xl justify-center items-center"
              style={{ backgroundColor: theme.surface }}
              onPress={() => setActiveCategory(null)}
            >
              <ArrowLeft size={22} color={theme.text} />
            </TouchableOpacity>
            <Text className="flex-1 text-2xl font-manrope-bold" style={{ color: theme.text }}>
              {activeCategory.name}
            </Text>
            {/* <View
              className="w-14 h-14 rounded-[18px] justify-center items-center"
              style={{ backgroundColor: categoryColor }}
            >
              {getCategoryIcon(activeCategory)}
            </View> */}
          </View>

          {/* Items List */}
          {pendingItems.length === 0 && boughtItems.length === 0 ? (
            <View className="items-center py-20 px-6">
              <View
                className="w-16 h-16 rounded-full justify-center items-center mb-4"
                style={{ backgroundColor: theme.surface }}
              >
                <ShoppingCart size={32} color={theme.textSecondary} />
              </View>
              <Text className="text-xl font-manrope-bold mb-2 text-center" style={{ color: theme.text }}>
                {t.shopping.noItems || "No Items"}
              </Text>
              <Text className="text-sm font-manrope text-center leading-5 mb-6" style={{ color: theme.textSecondary }}>
                {t.shopping.noItemsHint || "Your shopping list is empty."}
              </Text>
              <Button title={t.shopping.addItem || "Add Item"} onPress={openItemModal} />
            </View>
          ) : (
            <View className="gap-4">
              {pendingItems.map((item) => (
                <TouchableOpacity
                  key={item.id}
                  className="flex-row items-center gap-4 py-2"
                  onPress={() => toggleItemBought(item.id)}
                  onLongPress={() => openItemActions(item)}
                  activeOpacity={0.95}
                >
                  <View
                    className="w-8 h-8 rounded-full border-2 justify-center items-center"
                    style={[{ borderColor: theme.textSecondary }]}
                  />
                  <View className="flex-1">
                    <Text className="text-lg font-manrope-semibold" style={{ color: theme.text }}>
                      {item.name}
                    </Text>
                    {item.user?.name && (
                      <Text className="text-xs font-manrope mt-0.5" style={{ color: theme.textSecondary }}>
                        {t.shopping.addedByUser}: {item.user.name}
                      </Text>
                    )}
                  </View>
                </TouchableOpacity>
              ))}

              {pendingItems.length > 0 && boughtItems.length > 0 && (
                <View className="mt-3 pt-4" style={{ borderTopWidth: 1, borderTopColor: theme.border }} />
              )}

              {boughtItems.map((item) => (
                <TouchableOpacity
                  key={item.id}
                  className="flex-row items-center gap-4 py-2"
                  onPress={() => toggleItemBought(item.id)}
                  onLongPress={() => openItemActions(item)}
                  activeOpacity={0.95}
                >
                  <View
                    className="w-8 h-8 rounded-full border-2 justify-center items-center"
                    style={[
                      { borderColor: theme.textSecondary },
                      {
                        backgroundColor: theme.accent.purple,
                        borderColor: theme.accent.purple,
                      },
                    ]}
                  >
                    <Check size={16} color="#1C1C1E" strokeWidth={3} />
                  </View>
                  <View className="flex-1">
                    <Text
                      className="text-lg font-manrope-semibold line-through opacity-50"
                      style={{ color: theme.text }}
                    >
                      {item.name}
                    </Text>
                    {item.user?.name && (
                      <Text className="text-xs font-manrope mt-0.5 opacity-50" style={{ color: theme.textSecondary }}>
                        {t.shopping.addedByUser}: {item.user.name}
                      </Text>
                    )}
                  </View>
                  <TouchableOpacity className="p-2" onPress={() => handleDeleteItem(item.id)}>
                    <Trash2 size={18} color={theme.textSecondary} />
                  </TouchableOpacity>
                </TouchableOpacity>
              ))}
            </View>
          )}
        </ScrollView>

        {/* Add Item FAB */}
        <TouchableOpacity
          className="absolute bottom-[120px] right-6 w-14 h-14 rounded-[18px] justify-center items-center shadow-lg"
          style={{ backgroundColor: theme.accent.cyan }}
          onPress={openItemModal}
          activeOpacity={0.8}
        >
          <Plus size={28} color="#1C1C1E" strokeWidth={2.5} />
        </TouchableOpacity>

        {/* Add Item Modal */}
        <Modal visible={showItemModal} onClose={closeItemModal} title={t.shopping.addItem} height="full">
          <View className="flex-1">
            <Input
              label={t.shopping.itemName}
              placeholder={t.shopping.itemNamePlaceholder}
              value={newItemName}
              onChangeText={setNewItemName}
            />

            <Button
              title={t.common.add}
              onPress={handleAddPendingItem}
              disabled={!newItemName.trim() || creatingItem}
              variant="yellow"
              style={{ marginBottom: 16 }}
            />

            {pendingItemNames.length > 0 && (
              <View className="mb-4">
                <Text className="text-xs font-manrope-bold uppercase mb-2" style={{ color: theme.textSecondary }}>
                  {t.common.items} ({pendingItemNames.length})
                </Text>
                <View className="gap-2">
                  {pendingItemNames.map((itemName, index) => (
                    <View
                      key={`${itemName}-${index}`}
                      className="flex-row items-center justify-between px-3 py-2.5 rounded-xl"
                      style={{ backgroundColor: theme.surface }}
                    >
                      <Text className="text-sm font-manrope-semibold flex-1 mr-3" style={{ color: theme.text }}>
                        {itemName}
                      </Text>
                      <TouchableOpacity
                        onPress={() =>
                          setPendingItemNames((prev) => prev.filter((_, itemIndex) => itemIndex !== index))
                        }
                        className="p-1"
                      >
                        <Trash2 size={16} color={theme.textSecondary} />
                      </TouchableOpacity>
                    </View>
                  ))}
                </View>
              </View>
            )}

            <Button
              title={t.common.done}
              onPress={handleCreateItem}
              loading={creatingItem}
              disabled={(pendingItemNames.length === 0 && !newItemName.trim()) || creatingItem}
              variant="purple"
              style={{ marginTop: "auto" }}
            />
          </View>
        </Modal>
      </View>
    );
  }

  // Main shopping lists view - matches PDF exactly
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
              {t.shopping.title}
            </Text>
            <Text className="text-base font-manrope" style={{ color: theme.textSecondary }}>
              {t.shopping.myLists}
            </Text>
          </View>
        </View>

        {/* Category Grid - matches PDF layout */}
        {categories.length === 0 ? (
          <View className="items-center py-20 px-6">
            <View
              className="w-16 h-16 rounded-full justify-center items-center mb-4"
              style={{ backgroundColor: theme.surface }}
            >
              <ShoppingCart size={32} color={theme.textSecondary} />
            </View>
            <Text className="text-xl font-manrope-bold mb-2 text-center" style={{ color: theme.text }}>
              {t.shopping.noLists || "No Shopping Lists"}
            </Text>
            <Text className="text-sm font-manrope text-center leading-5 mb-6" style={{ color: theme.textSecondary }}>
              {t.shopping.noListsHint || "Create your first shopping list category to start adding items."}
            </Text>
            <Button
              title={t.shopping.newList || "New List"}
              onPress={() => setShowCategoryModal(true)}
              variant="purple"
            />
          </View>
        ) : (
          <View
            className="flex-row flex-wrap gap-3"
            style={{ justifyContent: isDesktop ? "flex-start" : "space-between" }}
          >
            {categories.map((category) => {
              const categoryColor = category.color || CATEGORY_COLORS[0];
              const itemCount = items[category.id]?.length || 0;

              return (
                <TouchableOpacity
                  key={category.id}
                  className="w-[47%] rounded-3xl p-[18px] justify-between"
                  style={{
                    backgroundColor: categoryColor,
                    aspectRatio: 0.9,
                    width: isDesktop ? "23.8%" : "47%",
                  }}
                  onPress={() => setActiveCategory(category)}
                  onLongPress={() => openCategoryActions(category)}
                  activeOpacity={0.9}
                >
                  <View className="w-10 h-10 rounded-xl bg-black/5 justify-center items-center">
                    {getCategoryIcon(category)}
                  </View>
                  <View className="flex-1 justify-end">
                    <Text className="text-xl font-manrope-bold text-[#1C1C1E] mb-1">{category.name}</Text>
                    <Text className="text-[13px] font-manrope-medium text-black/50">
                      {itemCount} {t.common.items}
                    </Text>
                  </View>
                  <View className="absolute bottom-[18px] right-[18px]">
                    <View className="w-8 h-8 rounded-full bg-black/10 justify-center items-center">
                      <ChevronRight size={16} color="rgba(0,0,0,1)" />
                    </View>
                  </View>
                </TouchableOpacity>
              );
            })}
          </View>
        )}
      </ScrollView>

      {/* Floating Add List FAB */}
      <TouchableOpacity
        className="absolute bottom-[120px] right-6 w-14 h-14 rounded-[18px] justify-center items-center shadow-lg z-40"
        style={{ backgroundColor: theme.accent.cyan }}
        onPress={() => setShowCategoryModal(true)}
        activeOpacity={0.8}
      >
        <Plus size={28} color="#1C1C1E" strokeWidth={2.5} />
      </TouchableOpacity>

      {/* Create Category Modal - matches PDF design */}
      <Modal
        visible={showCategoryModal}
        onClose={() => {
          setShowCategoryModal(false);
          setEditingCategoryId(null);
        }}
        title={editingCategoryId ? "Edit list" : t.shopping.newList}
        height="full"
      >
        <View className="flex-1">
          {/* Icon Preview */}
          <View className="items-center mb-6">
            <View
              className="w-20 h-20 rounded-3xl justify-center items-center"
              style={{ backgroundColor: selectedColor }}
            >
              {getIconComponent(selectedIcon, 32, "#1C1C1E")}
            </View>
          </View>

          {/* Title Input */}
          <Input placeholder={t.shopping.title_input} value={newCategoryName} onChangeText={setNewCategoryName} />

          {/* Color Picker */}
          <View className="mb-6 gap-3">
            <View className="flex-row justify-center gap-2.5">
              {COLOR_OPTIONS.slice(0, 7).map((color) => (
                <TouchableOpacity
                  key={color}
                  className={`w-9 h-9 rounded-full ${selectedColor === color ? "border-[3px] border-black/30" : ""}`}
                  style={{ backgroundColor: color }}
                  onPress={() => setSelectedColor(color)}
                />
              ))}
            </View>
            <View className="flex-row justify-center gap-2.5">
              {COLOR_OPTIONS.slice(7).map((color) => (
                <TouchableOpacity
                  key={color}
                  className={`w-9 h-9 rounded-full ${selectedColor === color ? "border-[3px] border-black/30" : ""}`}
                  style={{ backgroundColor: color }}
                  onPress={() => setSelectedColor(color)}
                />
              ))}
            </View>
          </View>

          {/* Icon Picker */}
          <ScrollView className="max-h-[220px]" showsVerticalScrollIndicator={false}>
            <View className="gap-3">
              {[0, 1, 2, 3].map((row) => (
                <View key={row} className="flex-row justify-center gap-2.5">
                  {ICON_OPTIONS.slice(row * 6, row * 6 + 6).map((icon) => (
                    <TouchableOpacity
                      key={icon.id}
                      className="w-12 h-12 rounded-full justify-center items-center"
                      style={{
                        backgroundColor: selectedIcon === icon.id ? selectedColor : theme.surface,
                      }}
                      onPress={() => setSelectedIcon(icon.id)}
                    >
                      {getIconComponent(icon.id, 20, selectedIcon === icon.id ? "#1C1C1E" : theme.textSecondary)}
                    </TouchableOpacity>
                  ))}
                </View>
              ))}
            </View>
          </ScrollView>
        </View>

        <View className="flex-row gap-3 pt-4">
          <Button
            title={editingCategoryId ? t.common.save : t.shopping.newList}
            icon={<Check size={20} color={theme.isDark ? "#1C1C1E" : "#FFFFFF"} />}
            onPress={handleCreateCategory}
            disabled={!newCategoryName.trim() || creatingCategory}
            className="flex-1"
          />
        </View>
      </Modal>

      <Modal
        visible={showCategoryActionsModal}
        onClose={() => {
          setShowCategoryActionsModal(false);
          setSelectedCategoryForActions(null);
        }}
        title={selectedCategoryForActions?.name || t.shopping.title}
        height="auto"
      >
        <View className="gap-3">
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.surface }}
            onPress={() => {
              const category = selectedCategoryForActions;
              setShowCategoryActionsModal(false);
              setSelectedCategoryForActions(null);
              if (category) openEditCategory(category);
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
              const category = selectedCategoryForActions;
              setShowCategoryActionsModal(false);
              setSelectedCategoryForActions(null);
              if (category) openDeleteCategory(category);
            }}
          >
            <Text className="font-manrope-semibold text-white">{t.common.delete}</Text>
          </TouchableOpacity>
        </View>
      </Modal>

      <Modal
        visible={showItemActionsModal}
        onClose={() => {
          setShowItemActionsModal(false);
          setSelectedItemForActions(null);
        }}
        title={selectedItemForActions?.name || t.shopping.addItem}
        height="auto"
      >
        <View className="gap-3">
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: theme.surface }}
            onPress={() => {
              const item = selectedItemForActions;
              setShowItemActionsModal(false);
              if (item) {
                setEditItemName(item.name);
                setShowEditItemModal(true);
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
              const item = selectedItemForActions;
              setShowItemActionsModal(false);
              setSelectedItemForActions(null);
              if (item) handleDeleteItem(item.id);
            }}
          >
            <Text className="font-manrope-semibold text-white">{t.common.delete}</Text>
          </TouchableOpacity>
        </View>
      </Modal>

      <Modal
        visible={showEditItemModal}
        onClose={() => {
          setShowEditItemModal(false);
          setSelectedItemForActions(null);
        }}
        title={t.shopping.editItem}
        height="auto"
      >
        <View className="gap-4">
          <Input value={editItemName} onChangeText={setEditItemName} placeholder={t.shopping.itemNamePlaceholder} />
          <TouchableOpacity
            className="h-12 rounded-xl justify-center items-center"
            style={{ backgroundColor: editItemName.trim() ? theme.text : theme.textSecondary }}
            onPress={handleEditItem}
            disabled={!editItemName.trim() || savingItemEdit}
          >
            <Text className="font-manrope-semibold" style={{ color: theme.background }}>
              {t.common.save}
            </Text>
          </TouchableOpacity>
        </View>
      </Modal>
    </View>
  );
}
