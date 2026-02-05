/**
 * Root layout — AuthProvider, NotificationProvider, Stack (auth vs app)
 */
import { Stack, useRouter } from "expo-router";
import { StatusBar } from "expo-status-bar";
import { useEffect } from "react";
import { AuthProvider, useAuth } from "../context/AuthContext";
import { NotificationProvider } from "../context/NotificationContext";

function RootLayoutNav() {
  const { isSignedIn, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isLoading) return;
    if (isSignedIn) {
      router.replace("/(tabs)");
    } else {
      router.replace("/");
    }
  }, [isSignedIn, isLoading, router]);

  return (
    <>
      <StatusBar style="auto" />
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="index" />
        <Stack.Screen name="register" />
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen
          name="ride/[id]"
          options={{
            headerShown: true,
            headerTitle: "Поездка",
            headerBackTitle: "Назад",
          }}
        />
        <Stack.Screen
          name="chat/[rideId]"
          options={{
            headerShown: true,
            headerTitle: "Чат с пассажиром",
            headerBackTitle: "Назад",
          }}
        />
      </Stack>
    </>
  );
}

export default function RootLayout() {
  return (
    <AuthProvider>
      <NotificationProvider>
        <RootLayoutNav />
      </NotificationProvider>
    </AuthProvider>
  );
}
