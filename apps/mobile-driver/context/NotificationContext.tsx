/**
 * Notification Context for Driver App
 * Manages push notifications and new ride alerts
 */
import React, { createContext, useContext, useState, useCallback } from "react";
import { View, Text, StyleSheet, TouchableOpacity, Modal } from "react-native";
import { useAuth } from "./AuthContext";
import {
  usePushNotifications,
  type PushNotificationData,
} from "../hooks/usePushNotifications";
import * as Notifications from "expo-notifications";

type NotificationContextValue = {
  expoPushToken: string | null;
  lastNotification: Notifications.Notification | null;
  isOnline: boolean;
  setIsOnline: (online: boolean) => void;
};

const NotificationContext = createContext<NotificationContextValue | null>(null);

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  const { token } = useAuth();
  const [isOnline, setIsOnline] = useState(false);
  
  const {
    expoPushToken,
    notification,
    newRideAlert,
    dismissNewRideAlert,
    acceptNewRide,
  } = usePushNotifications(token, isOnline);

  return (
    <NotificationContext.Provider
      value={{
        expoPushToken,
        lastNotification: notification,
        isOnline,
        setIsOnline,
      }}
    >
      {children}
      
      {/* New Ride Alert Modal */}
      <Modal
        visible={!!newRideAlert}
        transparent
        animationType="slide"
        onRequestClose={dismissNewRideAlert}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.alertCard}>
            <Text style={styles.alertIcon}>🚗</Text>
            <Text style={styles.alertTitle}>Новая заявка!</Text>
            
            {newRideAlert?.from_address && (
              <Text style={styles.alertAddress}>
                📍 {newRideAlert.from_address}
              </Text>
            )}
            {newRideAlert?.to_address && (
              <Text style={styles.alertAddress}>
                🏁 {newRideAlert.to_address}
              </Text>
            )}
            
            <View style={styles.alertButtons}>
              <TouchableOpacity
                style={[styles.alertBtn, styles.dismissBtn]}
                onPress={dismissNewRideAlert}
              >
                <Text style={styles.dismissBtnText}>Пропустить</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.alertBtn, styles.acceptBtn]}
                onPress={acceptNewRide}
              >
                <Text style={styles.acceptBtnText}>Сделать ставку</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </NotificationContext.Provider>
  );
}

export function useNotifications() {
  const ctx = useContext(NotificationContext);
  if (!ctx) throw new Error("useNotifications must be used within NotificationProvider");
  return ctx;
}

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: "rgba(0,0,0,0.5)",
    justifyContent: "center",
    alignItems: "center",
    padding: 20,
  },
  alertCard: {
    backgroundColor: "#fff",
    borderRadius: 20,
    padding: 24,
    width: "100%",
    maxWidth: 340,
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 12,
    elevation: 8,
  },
  alertIcon: {
    fontSize: 48,
    marginBottom: 12,
  },
  alertTitle: {
    fontSize: 22,
    fontWeight: "700",
    color: "#0f172a",
    marginBottom: 16,
  },
  alertAddress: {
    fontSize: 15,
    color: "#334155",
    marginBottom: 8,
    textAlign: "center",
  },
  alertButtons: {
    flexDirection: "row",
    gap: 12,
    marginTop: 20,
  },
  alertBtn: {
    flex: 1,
    paddingVertical: 14,
    paddingHorizontal: 20,
    borderRadius: 12,
    alignItems: "center",
  },
  dismissBtn: {
    backgroundColor: "#f1f5f9",
  },
  dismissBtnText: {
    color: "#64748b",
    fontWeight: "600",
    fontSize: 15,
  },
  acceptBtn: {
    backgroundColor: "#16a34a",
  },
  acceptBtnText: {
    color: "#fff",
    fontWeight: "600",
    fontSize: 15,
  },
});
