import React from "react";
import { View, Text, StyleSheet } from "react-native";

export interface MapPlaceholderProps {
  message?: string;
}

/** Map placeholder — stub until Leaflet/OpenStreetMap or RN maps integrated */
export function MapPlaceholder({ message = "Map (OpenStreetMap stub)" }: MapPlaceholderProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.text}>{message}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    height: 200,
    backgroundColor: "#e2e8f0",
    borderRadius: 12,
    alignItems: "center",
    justifyContent: "center",
  },
  text: { fontSize: 14, color: "#64748b" },
});
