// Design system colors extracted from PDF designs

// Dark Theme (Primary - from PDF designs)
export const darkTheme = {
  // Background colors
  background: "#1C1C1E",
  surface: "#2C2C2E",
  surfaceLight: "#3A3A3C",

  // Text colors
  text: "#FFFFFF",
  textSecondary: "#8E8E93",
  textMuted: "rgba(255, 255, 255, 0.5)",

  // Border colors
  border: "#3A3A3C",
  borderLight: "rgba(255, 255, 255, 0.1)",

  // Input colors
  inputBackground: "#2C2C2E",
  inputText: "#FFFFFF",
  inputPlaceholder: "#8E8E93",
  inputBorder: "transparent",
  inputBorderFocused: "#D4C4A0", // Golden border on focus (register screen)

  // Navigation
  navBackground: "#2C2C2E",
  navIconActive: "#FFFFFF",
  navIconInactive: "#8E8E93",
  tabIconActiveBackground: "#FFFFFF",
};

// Light Theme (White theme - inverted)
export const lightTheme = {
  // Background colors
  background: "#FFFFFF",
  surface: "#F5F5F7",
  surfaceLight: "#E8E8ED",

  // Text colors
  text: "#1C1C1E",
  textSecondary: "#8E8E93",
  textMuted: "rgba(0, 0, 0, 0.5)",

  // Border colors
  border: "#E8E8ED",
  borderLight: "rgba(0, 0, 0, 0.1)",

  // Input colors
  inputBackground: "#F5F5F7",
  inputText: "#1C1C1E",
  inputPlaceholder: "#8E8E93",
  inputBorder: "#E8E8ED",
  inputBorderFocused: "#D4C4A0",

  // Navigation
  navBackground: "#FFFFFF",
  navIconActive: "#1C1C1E",
  navIconInactive: "#8E8E93",
  tabIconActiveBackground: "#D8D4FC",
};

// Accent colors (same for both themes - from PDF)
export const accentColors = {
  // Primary accent - Lavender/Purple (used for buttons, polls, active states)
  purple: "#D8D4FC",
  purpleDark: "#B8B4DC",

  // Secondary accent - Yellow/Cream (used for tasks, cards)
  yellow: "#FBEB9E",
  yellowDark: "#D4C4A0",

  // Tertiary accent - Coral/Pink (used for expenses, alerts)
  pink: "#FF7476",
  pinkLight: "rgba(255, 116, 118, 0.2)",
  pinkBadge: "rgba(255, 116, 118, 0.3)",

  // Mint/Green (shopping cards)
  mint: "#A8E6CF",
  mintDark: "#88C6AF",

  // Green (success states, confirmations)
  green: "#22C55E",
  greenLight: "rgba(34, 197, 94, 0.2)",

  // Cyan/Teal (search buttons)
  cyan: "#7DD3E8",

  // Danger/Logout
  danger: "#8B3A3A",
  dangerLight: "rgba(199, 58, 58, 0.9)",
};

// Status colors
export const statusColors = {
  success: "#22C55E",
  error: "#EF4444",
  warning: "#FB923C",
  info: "#60A5FA",
};

// User indicator colors (for task assignees)
export const userColors = [
  "#60A5FA", // Blue - Mike
  "#FBEB9E", // Yellow - Sarah
  "#FF7476", // Pink/Red - Jack
  "#22C55E", // Green - Dariusz
  "#D8D4FC", // Purple
];

// Shopping list category colors
export const categoryColors = {
  dinner: "#D8D4FC", // Lavender
  snacks: "#FBEB9E", // Yellow/Cream
  birthday: "#FF7476", // Coral/Pink
  other: "#A8E6CF", // Mint/Green
};

// Default export for backward compatibility
const Colors = {
  // Base colors
  black: "#1C1C1E",
  white: "#FFFFFF",

  // Gray scale
  gray50: "#F5F5F7",
  gray100: "#E8E8ED",
  gray200: "#D1D1D6",
  gray400: "#8E8E93",
  gray500: "#6E6E73",
  gray600: "#48484A",
  gray700: "#3A3A3C",
  gray800: "#2C2C2E",

  // Primary dark theme colors (for backward compatibility)
  primaryDark: "#1C1C1E",
  secondaryDark: "#2C2C2E",
  background: "#1C1C1E",
  surface: "#2C2C2E",
  text: "#FFFFFF",
  textSecondary: "#8E8E93",
  border: "#3A3A3C",

  // Accent colors
  accentYellow: "#FBEB9E",
  accentPurple: "#D8D4FC",
  accentPink: "#FF7476",
  accentMint: "#A8E6CF",
  accentCyan: "#7DD3E8",

  // Status colors
  green500: "#22C55E",
  red500: "#EF4444",
  blue400: "#60A5FA",
  green400: "#4ade80",
  orange400: "#FB923C",
};

export default Colors;
