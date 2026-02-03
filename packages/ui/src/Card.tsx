import React from "react";
import { View, StyleSheet, type ViewStyle } from "react-native";

export interface CardProps {
  children: React.ReactNode;
  style?: ViewStyle;
}

/** Shared Card — container with shadow/border (RN-safe) */
export function Card({ children, style }: CardProps) {
  return <View style={[styles.card, style]}>{children}</View>;
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: "#fff",
    borderRadius: 12,
    padding: 16,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
    elevation: 3,
  },
});
