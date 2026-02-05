/**
 * Notification Context — manages push notifications
 */
import React, { createContext, useContext } from "react";
import { useAuth } from "./AuthContext";
import { usePushNotifications, type PushNotificationData } from "../hooks/usePushNotifications";
import * as Notifications from "expo-notifications";

type NotificationContextValue = {
  expoPushToken: string | null;
  lastNotification: Notifications.Notification | null;
};

const NotificationContext = createContext<NotificationContextValue | null>(null);

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  const { token } = useAuth();
  const { expoPushToken, notification } = usePushNotifications(token);

  return (
    <NotificationContext.Provider
      value={{
        expoPushToken,
        lastNotification: notification,
      }}
    >
      {children}
    </NotificationContext.Provider>
  );
}

export function useNotifications() {
  const ctx = useContext(NotificationContext);
  if (!ctx) throw new Error("useNotifications must be used within NotificationProvider");
  return ctx;
}
