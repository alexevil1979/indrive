import React from "react";
import { View, Text, StyleSheet, type ViewStyle } from "react-native";

export interface BadgeProps {
  label: string;
  variant?: "default" | "success" | "warning" | "error";
  style?: ViewStyle;
}

/** Shared Badge — status/tag (e.g. ride status) */
export function Badge({ label, variant = "default", style }: BadgeProps) {
  return (
    <View style={[styles.badge, styles[`badge_${variant}`], style]}>
      <Text style={[styles.text, styles[`text_${variant}`]]}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    paddingVertical: 4,
    paddingHorizontal: 10,
    borderRadius: 9999,
    alignSelf: "flex-start",
  },
  badge_default: { backgroundColor: "#e2e8f0" },
  badge_success: { backgroundColor: "#dcfce7" },
  badge_warning: { backgroundColor: "#fef3c7" },
  badge_error: { backgroundColor: "#fee2e2" },
  text: { fontSize: 12, fontWeight: "600" },
  text_default: { color: "#475569" },
  text_success: { color: "#166534" },
  text_warning: { color: "#92400e" },
  text_error: { color: "#991b1b" },
});
