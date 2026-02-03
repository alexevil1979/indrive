import React from "react";
import { Pressable, Text, StyleSheet, type ViewStyle, type TextStyle } from "react-native";

export interface ButtonProps {
  title: string;
  onPress: () => void;
  variant?: "primary" | "secondary" | "outline";
  disabled?: boolean;
  style?: ViewStyle;
  textStyle?: TextStyle;
}

/** Shared Button — primary/secondary/outline (2026 RN defaults) */
export function Button({
  title,
  onPress,
  variant = "primary",
  disabled = false,
  style,
  textStyle,
}: ButtonProps) {
  return (
    <Pressable
      onPress={onPress}
      disabled={disabled}
      style={({ pressed }) => [
        styles.base,
        styles[variant],
        disabled && styles.disabled,
        pressed && !disabled && styles.pressed,
        style,
      ]}
    >
      <Text style={[styles.text, styles[`text_${variant}`], textStyle]}>{title}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: {
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: "center",
    justifyContent: "center",
  },
  primary: { backgroundColor: "#2563eb" },
  secondary: { backgroundColor: "#64748b" },
  outline: { backgroundColor: "transparent", borderWidth: 2, borderColor: "#2563eb" },
  disabled: { opacity: 0.5 },
  pressed: { opacity: 0.85 },
  text: { fontSize: 16, fontWeight: "600" },
  text_primary: { color: "#fff" },
  text_secondary: { color: "#fff" },
  text_outline: { color: "#2563eb" },
});
