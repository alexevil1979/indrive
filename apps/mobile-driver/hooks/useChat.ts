/**
 * Chat Hook — WebSocket connection for ride chat
 */
import { useState, useEffect, useRef, useCallback } from "react";
import { config } from "../lib/config";

export type ChatMessage = {
  id: string;
  text: string;
  userId: string;
  isMe: boolean;
  createdAt: string;
};

type WebSocketMessage = {
  type: string;
  userId: string;
  text?: string;
  messageId?: string;
  createdAt?: string;
};

export function useChat(rideId: string | null, userId: string | null, token: string | null) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Fetch chat history
  const fetchHistory = useCallback(async () => {
    if (!rideId || !token) return;
    
    try {
      const response = await fetch(
        `${config.notificationApiUrl}/api/v1/chat/${rideId}/messages?limit=50`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );
      
      if (response.ok) {
        const data = await response.json();
        const history: ChatMessage[] = (data.messages ?? []).map(
          (msg: { id: string; text: string; user_id: string; created_at: string }) => ({
            id: msg.id,
            text: msg.text,
            userId: msg.user_id,
            isMe: msg.user_id === userId,
            createdAt: msg.created_at,
          })
        );
        setMessages(history);
      }
    } catch (error) {
      console.warn("Failed to fetch chat history:", error);
    } finally {
      setIsLoading(false);
    }
  }, [rideId, userId, token]);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (!rideId || !userId) return;

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    // Build WebSocket URL
    const wsUrl = config.notificationApiUrl
      .replace("http://", "ws://")
      .replace("https://", "wss://");
    const url = `${wsUrl}/ws/chat?rideId=${rideId}&userId=${userId}`;

    try {
      const ws = new WebSocket(url);

      ws.onopen = () => {
        setIsConnected(true);
        console.log("Chat WebSocket connected");
      };

      ws.onclose = () => {
        setIsConnected(false);
        console.log("Chat WebSocket disconnected");
        
        // Reconnect after 3 seconds
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current);
        }
        reconnectTimeoutRef.current = setTimeout(() => {
          if (rideId && userId) {
            connect();
          }
        }, 3000);
      };

      ws.onerror = (error) => {
        console.warn("Chat WebSocket error:", error);
      };

      ws.onmessage = (event) => {
        try {
          const data: WebSocketMessage = JSON.parse(event.data);
          
          if (data.type === "message" && data.text) {
            const newMessage: ChatMessage = {
              id: data.messageId ?? String(Date.now()),
              text: data.text,
              userId: data.userId,
              isMe: data.userId === userId,
              createdAt: data.createdAt ?? new Date().toISOString(),
            };
            
            setMessages((prev) => [...prev, newMessage]);
          }
        } catch (error) {
          console.warn("Failed to parse WebSocket message:", error);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.warn("Failed to create WebSocket:", error);
    }
  }, [rideId, userId]);

  // Send message
  const sendMessage = useCallback((text: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket not connected");
      return false;
    }

    try {
      wsRef.current.send(JSON.stringify({ type: "message", text }));
      return true;
    } catch (error) {
      console.warn("Failed to send message:", error);
      return false;
    }
  }, []);

  // Disconnect
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  // Initialize
  useEffect(() => {
    if (rideId && userId && token) {
      fetchHistory();
      connect();
    }

    return () => {
      disconnect();
    };
  }, [rideId, userId, token, fetchHistory, connect, disconnect]);

  return {
    messages,
    isConnected,
    isLoading,
    sendMessage,
    disconnect,
  };
}
