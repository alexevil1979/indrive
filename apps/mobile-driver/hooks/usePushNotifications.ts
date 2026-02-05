/**
 * Push Notifications Hook for Driver App
 * Handles registration, permissions, and notification handling
 */
import { useState, useEffect, useRef, useCallback } from "react";
import { Platform, Alert, Vibration } from "react-native";
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
  passenger_id?: string;
  from_address?: string;
  to_address?: string;
  status?: string;
  [key: string]: string | undefined;
};

export function usePushNotifications(token: string | null, isOnline: boolean) {
  const router = useRouter();
  const [expoPushToken, setExpoPushToken] = useState<string | null>(null);
  const [notification, setNotification] = useState<Notifications.Notification | null>(null);
  const [newRideAlert, setNewRideAlert] = useState<PushNotificationData | null>(null);
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
          "x-user-id": token,
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

    const { status: existingStatus } = await Notifications.getPermissionsAsync();
    let finalStatus = existingStatus;

    if (existingStatus !== "granted") {
      const { status } = await Notifications.requestPermissionsAsync();
      finalStatus = status;
    }

    if (finalStatus !== "granted") {
      Alert.alert(
        "Уведомления",
        "Разрешите уведомления, чтобы получать информацию о новых заявках"
      );
      return null;
    }

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

  // Handle incoming notification
  const handleNotificationReceived = useCallback((notification: Notifications.Notification) => {
    setNotification(notification);
    const data = notification.request.content.data as PushNotificationData;
    
    // If it's a new ride and driver is online, show special alert
    if (data.type === "new_ride" && isOnline) {
      setNewRideAlert(data);
      // Vibrate to alert driver
      Vibration.vibrate([0, 500, 200, 500]);
      
      // Auto-clear after 30 seconds
      setTimeout(() => {
        setNewRideAlert((current) => 
          current?.ride_id === data.ride_id ? null : current
        );
      }, 30000);
    }
  }, [isOnline]);

  // Handle notification tap
  const handleNotificationResponse = useCallback((response: Notifications.NotificationResponse) => {
    const data = response.notification.request.content.data as PushNotificationData;
    
    if (data.type === "new_ride" && data.ride_id) {
      // Navigate to ride detail to place bid
      router.push(`/ride/${data.ride_id}`);
      setNewRideAlert(null);
    } else if (data.type === "bid_accepted" && data.ride_id) {
      // Navigate to active ride
      router.push(`/ride/${data.ride_id}`);
    } else if (data.type === "ride_cancelled" && data.ride_id) {
      // Navigate to rides list
      router.push("/(tabs)/rides");
    }
  }, [router]);

  // Dismiss new ride alert
  const dismissNewRideAlert = useCallback(() => {
    setNewRideAlert(null);
  }, []);

  // Accept new ride (navigate to bid)
  const acceptNewRide = useCallback(() => {
    if (newRideAlert?.ride_id) {
      router.push(`/ride/${newRideAlert.ride_id}`);
      setNewRideAlert(null);
    }
  }, [newRideAlert, router]);

  // Setup push notifications
  useEffect(() => {
    registerForPushNotifications().then((pushToken) => {
      if (pushToken) {
        setExpoPushToken(pushToken);
        registerTokenWithBackend(pushToken);
      }
    });

    notificationListener.current = Notifications.addNotificationReceivedListener(
      handleNotificationReceived
    );

    responseListener.current = Notifications.addNotificationResponseReceivedListener(
      handleNotificationResponse
    );

    if (Platform.OS === "android") {
      // High priority channel for new rides
      Notifications.setNotificationChannelAsync("rides", {
        name: "Новые заявки",
        importance: Notifications.AndroidImportance.HIGH,
        vibrationPattern: [0, 500, 200, 500],
        lightColor: "#16a34a",
        sound: "default",
      });
      
      // Default channel
      Notifications.setNotificationChannelAsync("default", {
        name: "RideHail Driver",
        importance: Notifications.AndroidImportance.DEFAULT,
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
  }, [
    registerForPushNotifications,
    registerTokenWithBackend,
    handleNotificationReceived,
    handleNotificationResponse,
  ]);

  // Re-register token when auth token changes
  useEffect(() => {
    if (token && expoPushToken) {
      registerTokenWithBackend(expoPushToken);
    }
  }, [token, expoPushToken, registerTokenWithBackend]);

  return {
    expoPushToken,
    notification,
    newRideAlert,
    dismissNewRideAlert,
    acceptNewRide,
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
      sound: "default",
    },
    trigger: null,
  });
}
