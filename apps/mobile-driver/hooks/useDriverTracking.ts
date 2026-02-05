/**
 * Driver Tracking Hook — streams location to passengers via WebSocket
 */
import { useEffect, useRef, useCallback, useState } from "react";
import * as Location from "expo-location";
import { config } from "../lib/config";

type TrackingState = {
  isConnected: boolean;
  isTracking: boolean;
  lastUpdate: Date | null;
};

export function useDriverTracking(
  rideId: string | null,
  driverId: string | null,
  isActive: boolean // Should be true when ride is matched/in_progress
) {
  const [state, setState] = useState<TrackingState>({
    isConnected: false,
    isTracking: false,
    lastUpdate: null,
  });

  const wsRef = useRef<WebSocket | null>(null);
  const watchIdRef = useRef<Location.LocationSubscription | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (!rideId || !driverId || !isActive) return;

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    const wsUrl = config.notificationApiUrl
      .replace("http://", "ws://")
      .replace("https://", "wss://");
    const url = `${wsUrl}/ws/tracking?role=driver&rideId=${rideId}&driverId=${driverId}`;

    try {
      const ws = new WebSocket(url);

      ws.onopen = () => {
        setState((s) => ({ ...s, isConnected: true }));
        console.log("Driver tracking WebSocket connected");
      };

      ws.onclose = () => {
        setState((s) => ({ ...s, isConnected: false }));
        console.log("Driver tracking WebSocket disconnected");

        // Reconnect after 3 seconds if still active
        if (isActive && rideId && driverId) {
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, 3000);
        }
      };

      ws.onerror = (error) => {
        console.warn("Driver tracking WebSocket error:", error);
      };

      wsRef.current = ws;
    } catch (error) {
      console.warn("Failed to create tracking WebSocket:", error);
    }
  }, [rideId, driverId, isActive]);

  // Send location update
  const sendLocation = useCallback(
    (location: Location.LocationObject) => {
      if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
        return;
      }

      try {
        wsRef.current.send(
          JSON.stringify({
            type: "location",
            lat: location.coords.latitude,
            lng: location.coords.longitude,
            heading: location.coords.heading ?? undefined,
            speed: location.coords.speed
              ? Math.round(location.coords.speed * 3.6) // m/s to km/h
              : undefined,
          })
        );
        setState((s) => ({ ...s, lastUpdate: new Date() }));
      } catch (error) {
        console.warn("Failed to send location:", error);
      }
    },
    []
  );

  // Start location tracking
  const startTracking = useCallback(async () => {
    // Stop existing watch
    if (watchIdRef.current) {
      watchIdRef.current.remove();
      watchIdRef.current = null;
    }

    try {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== "granted") {
        console.warn("Location permission denied");
        return;
      }

      // Watch position with high accuracy
      watchIdRef.current = await Location.watchPositionAsync(
        {
          accuracy: Location.Accuracy.BestForNavigation,
          timeInterval: 3000, // Update every 3 seconds
          distanceInterval: 10, // Or every 10 meters
        },
        (location) => {
          sendLocation(location);
        }
      );

      setState((s) => ({ ...s, isTracking: true }));
      console.log("Driver location tracking started");
    } catch (error) {
      console.warn("Failed to start location tracking:", error);
    }
  }, [sendLocation]);

  // Stop tracking
  const stopTracking = useCallback(() => {
    if (watchIdRef.current) {
      watchIdRef.current.remove();
      watchIdRef.current = null;
    }
    setState((s) => ({ ...s, isTracking: false }));
    console.log("Driver location tracking stopped");
  }, []);

  // Disconnect
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    stopTracking();
    setState({ isConnected: false, isTracking: false, lastUpdate: null });
  }, [stopTracking]);

  // Effect to manage connection and tracking
  useEffect(() => {
    if (isActive && rideId && driverId) {
      connect();
      startTracking();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [isActive, rideId, driverId, connect, startTracking, disconnect]);

  return {
    isConnected: state.isConnected,
    isTracking: state.isTracking,
    lastUpdate: state.lastUpdate,
    disconnect,
  };
}
