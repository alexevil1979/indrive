/**
 * Tab layout — Home (map + request), Rides (my rides) + logout
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
        headerTitle: "RideHail",
        headerRight: () => (
          <TouchableOpacity onPress={handleLogout} style={{ marginRight: 16 }}>
            <Text style={{ color: "#2563eb", fontSize: 14 }}>Выйти</Text>
          </TouchableOpacity>
        ),
        tabBarLabelStyle: { fontSize: 12 },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{ title: "Главная", tabBarLabel: "Главная" }}
      />
      <Tabs.Screen
        name="rides"
        options={{ title: "Мои поездки", tabBarLabel: "Поездки" }}
      />
    </Tabs>
  );
}
