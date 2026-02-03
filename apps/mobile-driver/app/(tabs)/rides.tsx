/**
 * My rides (driver)
 */
import { useEffect, useState } from "react";
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  RefreshControl,
} from "react-native";
import { useRouter } from "expo-router";
import { Card, Badge } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import { listMyRides, type Ride } from "../../lib/api";

const statusLabel: Record<string, string> = {
  matched: "Пассажир принял",
  in_progress: "В пути",
  completed: "Завершена",
  cancelled: "Отменена",
};

export default function MyRidesScreen() {
  const { token } = useAuth();
  const router = useRouter();
  const [rides, setRides] = useState<Ride[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const load = async () => {
    if (!token) return;
    try {
      const list = await listMyRides(token);
      setRides(list);
    } catch {
      setRides([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    load();
  }, [token]);

  const onRefresh = () => {
    setRefreshing(true);
    load();
  };

  const renderItem = ({ item }: { item: Ride }) => (
    <TouchableOpacity onPress={() => router.push(`/ride/${item.id}`)} activeOpacity={0.7}>
      <Card style={styles.rideCard}>
        <View style={styles.rideHeader}>
          <Text style={styles.rideId}>#{item.id.slice(0, 8)}</Text>
          <Badge
            label={statusLabel[item.status] ?? item.status}
            variant={item.status === "completed" ? "success" : item.status === "cancelled" ? "error" : "default"}
          />
        </View>
        <Text style={styles.rideRoute}>
          {item.from.address ?? `${item.from.lat}, ${item.from.lng}`} →{" "}
          {item.to.address ?? `${item.to.lat}, ${item.to.lng}`}
        </Text>
        {item.price != null ? (
          <Text style={styles.ridePrice}>{item.price} ₽</Text>
        ) : null}
      </Card>
    </TouchableOpacity>
  );

  return (
    <View style={styles.container}>
      {rides.length === 0 && !loading ? (
        <Text style={styles.empty}>Нет поездок. Сделайте ставку в заявках.</Text>
      ) : (
        <FlatList
          data={rides}
          keyExtractor={(item) => item.id}
          renderItem={renderItem}
          contentContainerStyle={styles.list}
          refreshControl={
            <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
          }
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f0fdf4" },
  list: { padding: 16, paddingBottom: 32 },
  rideCard: { marginBottom: 12 },
  rideHeader: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 8 },
  rideId: { fontSize: 12, color: "#64748b" },
  rideRoute: { fontSize: 14, color: "#0f172a", marginBottom: 4 },
  ridePrice: { fontSize: 16, fontWeight: "600", color: "#0f172a" },
  empty: { padding: 24, textAlign: "center", color: "#64748b" },
});
