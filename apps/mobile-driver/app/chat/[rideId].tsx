/**
 * Chat Screen — full screen chat with passenger for a ride
 */
import { View, StyleSheet } from "react-native";
import { useLocalSearchParams } from "expo-router";
import { useAuth } from "../../context/AuthContext";
import { Chat } from "../../components/Chat";

export default function ChatScreen() {
  const params = useLocalSearchParams<{ rideId: string }>();
  const rideId = Array.isArray(params.rideId) ? params.rideId[0] : params.rideId;
  const { token, userId } = useAuth();

  if (!rideId || !token || !userId) {
    return null;
  }

  return (
    <View style={styles.container}>
      <Chat
        rideId={rideId}
        userId={userId}
        token={token}
        otherUserName="Пассажир"
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
});
