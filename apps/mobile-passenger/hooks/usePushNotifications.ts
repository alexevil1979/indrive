/**
 * Push Notifications Hook
 * Handles registration, permissions, and notification handling
 */
import { useState, useEffect, useRef, useCallback } from "react";
import { Platform, Alert } from "react-native";
import * as Device from "expo-device";
import * as Notifications from "expo-notifications";
import Constants from "expo-constants";
import { useRouter } from "expo-router";
import { config } from "../lib/config";

// Configure how notifications are displayed when app is in foreground
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
  }),
});

export type PushNotificationData = {
  type?: string;
  ride_id?: string;
  bid_id?: string;
  status?: string;
  [key: string]: string | undefined;
};

export function usePushNotifications(token: string | null) {
  const router = useRouter();
  const [expoPushToken, setExpoPushToken] = useState<string | null>(null);
  const [notification, setNotification] = useState<Notifications.Notification | null>(null);
  const notificationListener = useRef<Notifications.EventSubscription>();
  const responseListener = useRef<Notifications.EventSubscription>();

  // Register device token with backend
  const registerTokenWithBackend = useCallback(async (pushToken: string) => {
    if (!token) return;
    
    try {
      const response = await fetch(`${config.notificationApiUrl}/api/v1/device-tokens`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-user-id": token, // Using JWT as user identifier
        },
        body: JSON.stringify({
          token: pushToken,
          platform: Platform.OS,
        }),
      });
      
      if (!response.ok) {
        console.warn("Failed to register push token with backend");
      }
    } catch (error) {
      console.warn("Error registering push token:", error);
    }
  }, [token]);

  // Register for push notifications
  const registerForPushNotifications = useCallback(async () => {
    if (!Device.isDevice) {
      console.log("Push notifications require a physical device");
      return null;
    }

    // Check existing permissions
    const { status: existingStatus } = await Notifications.getPermissionsAsync();
    let finalStatus = existingStatus;

    // Request permissions if not already granted
    if (existingStatus !== "granted") {
      const { status } = await Notifications.requestPermissionsAsync();
      finalStatus = status;
    }

    if (finalStatus !== "granted") {
      Alert.alert(
        "Уведомления",
        "Разрешите уведомления, чтобы получать информацию о ставках и статусе поездки"
      );
      return null;
    }

    // Get push token
    try {
      const projectId = Constants.expoConfig?.extra?.eas?.projectId;
      const tokenData = await Notifications.getExpoPushTokenAsync({
        projectId,
      });
      return tokenData.data;
    } catch (error) {
      console.warn("Error getting push token:", error);
      return null;
    }
  }, []);

  // Handle notification tap (when user taps on notification)
  const handleNotificationResponse = useCallback((response: Notifications.NotificationResponse) => {
    const data = response.notification.request.content.data as PushNotificationData;
    
    if (data.type === "new_bid" && data.ride_id) {
      // Navigate to ride detail to see the bid
      router.push(`/ride/${data.ride_id}`);
    } else if (data.type === "ride_status" && data.ride_id) {
      // Navigate to ride detail
      router.push(`/ride/${data.ride_id}`);
    } else if (data.type === "driver_arrived" && data.ride_id) {
      // Navigate to ride detail
      router.push(`/ride/${data.ride_id}`);
    }
  }, [router]);

  // Setup push notifications
  useEffect(() => {
    // Register and get token
    registerForPushNotifications().then((pushToken) => {
      if (pushToken) {
        setExpoPushToken(pushToken);
        registerTokenWithBackend(pushToken);
      }
    });

    // Listen for incoming notifications (app is open)
    notificationListener.current = Notifications.addNotificationReceivedListener((notification) => {
      setNotification(notification);
    });

    // Listen for notification taps
    responseListener.current = Notifications.addNotificationResponseReceivedListener(
      handleNotificationResponse
    );

    // Setup Android notification channel
    if (Platform.OS === "android") {
      Notifications.setNotificationChannelAsync("default", {
        name: "RideHail",
        importance: Notifications.AndroidImportance.MAX,
        vibrationPattern: [0, 250, 250, 250],
        lightColor: "#2563eb",
      });
    }

    return () => {
      if (notificationListener.current) {
        Notifications.removeNotificationSubscription(notificationListener.current);
      }
      if (responseListener.current) {
        Notifications.removeNotificationSubscription(responseListener.current);
      }
    };
  }, [registerForPushNotifications, registerTokenWithBackend, handleNotificationResponse]);

  // Re-register token when auth token changes
  useEffect(() => {
    if (token && expoPushToken) {
      registerTokenWithBackend(expoPushToken);
    }
  }, [token, expoPushToken, registerTokenWithBackend]);

  return {
    expoPushToken,
    notification,
  };
}

// Helper to schedule local notification (for testing)
export async function scheduleLocalNotification(
  title: string,
  body: string,
  data?: PushNotificationData
) {
  await Notifications.scheduleNotificationAsync({
    content: {
      title,
      body,
      data: data ?? {},
    },
    trigger: null, // Immediate
  });
}
