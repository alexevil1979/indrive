/**
 * Tab layout — Available rides, My rides, Profile + logout
 */
import { Tabs, useRouter } from "expo-router";
import { TouchableOpacity, Text } from "react-native";
import { useAuth } from "../../context/AuthContext";

export default function TabsLayout() {
  const { logout } = useAuth();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.replace("/");
  };

  return (
    <Tabs
      screenOptions={{
        headerShown: true,
        headerTitle: "RideHail Водитель",
        headerRight: () => (
          <TouchableOpacity onPress={handleLogout} style={{ marginRight: 16 }}>
            <Text style={{ color: "#16a34a", fontSize: 14 }}>Выйти</Text>
          </TouchableOpacity>
        ),
        tabBarLabelStyle: { fontSize: 12 },
      }}
    >
      <Tabs.Screen name="index" options={{ title: "Заявки", tabBarLabel: "Заявки" }} />
      <Tabs.Screen name="rides" options={{ title: "Мои поездки", tabBarLabel: "Поездки" }} />
      <Tabs.Screen name="profile" options={{ title: "Профиль", tabBarLabel: "Профиль" }} />
    </Tabs>
  );
}
