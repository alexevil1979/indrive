/**
 * Driver Tracking Hook — subscribes to driver location via WebSocket
 */
import { useEffect, useRef, useCallback, useState } from "react";
import { config } from "../lib/config";

export type DriverLocation = {
  lat: number;
  lng: number;
  heading?: number; // direction in degrees
  speed?: number;   // km/h
  timestamp: number;
};

type TrackingState = {
  isConnected: boolean;
  isDriverOnline: boolean;
  driverLocation: DriverLocation | null;
};

type WebSocketMessage = {
  type: string;
  rideId?: string;
  driverId?: string;
  location?: DriverLocation;
};

export function useDriverTracking(
  rideId: string | null,
  passengerId: string | null,
  isActive: boolean // Should be true when ride is matched/in_progress
) {
  const [state, setState] = useState<TrackingState>({
    isConnected: false,
    isDriverOnline: false,
    driverLocation: null,
  });

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (!rideId || !passengerId || !isActive) return;

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    const wsUrl = config.notificationApiUrl
      .replace("http://", "ws://")
      .replace("https://", "wss://");
    const url = `${wsUrl}/ws/tracking?role=passenger&rideId=${rideId}&passengerId=${passengerId}`;

    try {
      const ws = new WebSocket(url);

      ws.onopen = () => {
        setState((s) => ({ ...s, isConnected: true }));
        console.log("Passenger tracking WebSocket connected");
      };

      ws.onclose = () => {
        setState((s) => ({ ...s, isConnected: false }));
        console.log("Passenger tracking WebSocket disconnected");

        // Reconnect after 3 seconds if still active
        if (isActive && rideId && passengerId) {
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, 3000);
        }
      };

      ws.onerror = (error) => {
        console.warn("Passenger tracking WebSocket error:", error);
      };

      ws.onmessage = (event) => {
        try {
          const data: WebSocketMessage = JSON.parse(event.data);

          switch (data.type) {
            case "driver_location":
              if (data.location) {
                setState((s) => ({
                  ...s,
                  isDriverOnline: true,
                  driverLocation: data.location!,
                }));
              }
              break;

            case "driver_online":
              setState((s) => ({ ...s, isDriverOnline: true }));
              break;

            case "driver_offline":
              setState((s) => ({
                ...s,
                isDriverOnline: false,
                driverLocation: null,
              }));
              break;

            default:
              console.log("Unknown tracking message type:", data.type);
          }
        } catch (error) {
          console.warn("Failed to parse tracking message:", error);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.warn("Failed to create tracking WebSocket:", error);
    }
  }, [rideId, passengerId, isActive]);

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
    setState({
      isConnected: false,
      isDriverOnline: false,
      driverLocation: null,
    });
  }, []);

  // Effect to manage connection
  useEffect(() => {
    if (isActive && rideId && passengerId) {
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [isActive, rideId, passengerId, connect, disconnect]);

  return {
    isConnected: state.isConnected,
    isDriverOnline: state.isDriverOnline,
    driverLocation: state.driverLocation,
    disconnect,
  };
}
