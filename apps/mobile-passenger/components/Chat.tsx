/**
 * Chat Component — real-time messaging with driver
 */
import { useState, useRef, useEffect } from "react";
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TextInput,
  TouchableOpacity,
  KeyboardAvoidingView,
  Platform,
  ActivityIndicator,
} from "react-native";
import { useChat, type ChatMessage } from "../hooks/useChat";

type Props = {
  rideId: string;
  userId: string;
  token: string;
  otherUserName?: string;
};

export function Chat({ rideId, userId, token, otherUserName = "Водитель" }: Props) {
  const { messages, isConnected, isLoading, sendMessage } = useChat(rideId, userId, token);
  const [inputText, setInputText] = useState("");
  const flatListRef = useRef<FlatList>(null);

  // Scroll to bottom on new messages
  useEffect(() => {
    if (messages.length > 0) {
      setTimeout(() => {
        flatListRef.current?.scrollToEnd({ animated: true });
      }, 100);
    }
  }, [messages.length]);

  const handleSend = () => {
    const text = inputText.trim();
    if (!text) return;
    
    if (sendMessage(text)) {
      setInputText("");
    }
  };

  const formatTime = (dateString: string) => {
    try {
      const date = new Date(dateString);
      return date.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
    } catch {
      return "";
    }
  };

  const renderMessage = ({ item }: { item: ChatMessage }) => (
    <View style={[styles.messageRow, item.isMe ? styles.messageRowMe : styles.messageRowOther]}>
      <View style={[styles.messageBubble, item.isMe ? styles.bubbleMe : styles.bubbleOther]}>
        {!item.isMe && <Text style={styles.senderName}>{otherUserName}</Text>}
        <Text style={[styles.messageText, item.isMe ? styles.textMe : styles.textOther]}>
          {item.text}
        </Text>
        <Text style={[styles.messageTime, item.isMe ? styles.timeMe : styles.timeOther]}>
          {formatTime(item.createdAt)}
        </Text>
      </View>
    </View>
  );

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#2563eb" />
        <Text style={styles.loadingText}>Загрузка чата...</Text>
      </View>
    );
  }

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === "ios" ? "padding" : undefined}
      keyboardVerticalOffset={100}
    >
      {/* Connection status */}
      <View style={[styles.statusBar, isConnected ? styles.statusConnected : styles.statusDisconnected]}>
        <View style={[styles.statusDot, isConnected ? styles.dotConnected : styles.dotDisconnected]} />
        <Text style={styles.statusText}>
          {isConnected ? "Чат подключён" : "Подключение..."}
        </Text>
      </View>

      {/* Messages */}
      <FlatList
        ref={flatListRef}
        data={messages}
        keyExtractor={(item) => item.id}
        renderItem={renderMessage}
        contentContainerStyle={styles.messagesList}
        showsVerticalScrollIndicator={false}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyIcon}>💬</Text>
            <Text style={styles.emptyText}>Начните диалог с водителем</Text>
          </View>
        }
      />

      {/* Input */}
      <View style={styles.inputContainer}>
        <TextInput
          style={styles.input}
          value={inputText}
          onChangeText={setInputText}
          placeholder="Сообщение..."
          placeholderTextColor="#94a3b8"
          multiline
          maxLength={500}
          editable={isConnected}
        />
        <TouchableOpacity
          style={[styles.sendButton, !inputText.trim() && styles.sendButtonDisabled]}
          onPress={handleSend}
          disabled={!inputText.trim() || !isConnected}
        >
          <Text style={styles.sendButtonText}>➤</Text>
        </TouchableOpacity>
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f8fafc",
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  loadingText: {
    marginTop: 8,
    color: "#64748b",
  },
  statusBar: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    paddingVertical: 6,
    paddingHorizontal: 12,
  },
  statusConnected: {
    backgroundColor: "#dcfce7",
  },
  statusDisconnected: {
    backgroundColor: "#fef3c7",
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 6,
  },
  dotConnected: {
    backgroundColor: "#16a34a",
  },
  dotDisconnected: {
    backgroundColor: "#f59e0b",
  },
  statusText: {
    fontSize: 12,
    color: "#0f172a",
  },
  messagesList: {
    padding: 12,
    flexGrow: 1,
  },
  messageRow: {
    marginBottom: 8,
  },
  messageRowMe: {
    alignItems: "flex-end",
  },
  messageRowOther: {
    alignItems: "flex-start",
  },
  messageBubble: {
    maxWidth: "80%",
    padding: 12,
    borderRadius: 16,
  },
  bubbleMe: {
    backgroundColor: "#2563eb",
    borderBottomRightRadius: 4,
  },
  bubbleOther: {
    backgroundColor: "#fff",
    borderBottomLeftRadius: 4,
    borderWidth: 1,
    borderColor: "#e2e8f0",
  },
  senderName: {
    fontSize: 11,
    fontWeight: "600",
    color: "#64748b",
    marginBottom: 4,
  },
  messageText: {
    fontSize: 15,
    lineHeight: 20,
  },
  textMe: {
    color: "#fff",
  },
  textOther: {
    color: "#0f172a",
  },
  messageTime: {
    fontSize: 10,
    marginTop: 4,
    alignSelf: "flex-end",
  },
  timeMe: {
    color: "rgba(255,255,255,0.7)",
  },
  timeOther: {
    color: "#94a3b8",
  },
  emptyContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 60,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 12,
  },
  emptyText: {
    fontSize: 15,
    color: "#64748b",
  },
  inputContainer: {
    flexDirection: "row",
    alignItems: "flex-end",
    padding: 12,
    backgroundColor: "#fff",
    borderTopWidth: 1,
    borderTopColor: "#e2e8f0",
  },
  input: {
    flex: 1,
    backgroundColor: "#f1f5f9",
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 10,
    fontSize: 15,
    maxHeight: 100,
    color: "#0f172a",
  },
  sendButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: "#2563eb",
    justifyContent: "center",
    alignItems: "center",
    marginLeft: 8,
  },
  sendButtonDisabled: {
    backgroundColor: "#94a3b8",
  },
  sendButtonText: {
    color: "#fff",
    fontSize: 18,
  },
});
