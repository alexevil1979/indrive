/**
 * Available rides — map view with ride markers + online toggle
 */
import { useEffect, useState, useCallback } from "react";
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  RefreshControl,
  Switch,
  Dimensions,
} from "react-native";
import { useRouter } from "expo-router";
import { Card, Badge, Button } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import { useNotifications } from "../../context/NotificationContext";
import { listAvailableRides, setDriverOnline, type Ride } from "../../lib/api";
import { DriverMap } from "../../components/DriverMap";

const { height: SCREEN_HEIGHT } = Dimensions.get("window");

export default function AvailableRidesScreen() {
  const { token } = useAuth();
  const { isOnline, setIsOnline } = useNotifications();
  const router = useRouter();
  const [rides, setRides] = useState<Ride[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [viewMode, setViewMode] = useState<"map" | "list">("map");

  const load = useCallback(async () => {
    if (!token) return;
    try {
      const list = await listAvailableRides(token);
      setRides(list);
    } catch {
      setRides([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [token]);

  useEffect(() => {
    load();
    // Refresh rides every 30 seconds when online
    const interval = setInterval(() => {
      if (isOnline) load();
    }, 30000);
    return () => clearInterval(interval);
  }, [token, isOnline, load]);

  const onRefresh = () => {
    setRefreshing(true);
    load();
  };

  const handleOnlineToggle = async (value: boolean) => {
    setIsOnline(value);
    if (token) {
      await setDriverOnline(token, value);
    }
  };

  const renderItem = ({ item }: { item: Ride }) => (
    <TouchableOpacity onPress={() => router.push(`/ride/${item.id}`)} activeOpacity={0.7}>
      <Card style={styles.rideCard}>
        <View style={styles.rideHeader}>
          <Text style={styles.rideId}>#{item.id.slice(0, 8)}</Text>
          <Badge label={item.status === "requested" ? "Ожидает ставок" : "Ставки"} variant="default" />
        </View>
        <Text style={styles.rideRoute}>
          {item.from.address ?? `${item.from.lat.toFixed(2)}, ${item.from.lng.toFixed(2)}`} →{" "}
          {item.to.address ?? `${item.to.lat.toFixed(2)}, ${item.to.lng.toFixed(2)}`}
        </Text>
        <Text style={styles.hint}>Нажмите, чтобы сделать ставку</Text>
      </Card>
    </TouchableOpacity>
  );

  return (
    <View style={styles.container}>
      {/* Online toggle header */}
      <View style={styles.header}>
        <View style={styles.onlineToggle}>
          <Text style={styles.onlineLabel}>На линии</Text>
          <Switch
            value={isOnline}
            onValueChange={handleOnlineToggle}
            trackColor={{ true: "#16a34a", false: "#e2e8f0" }}
            thumbColor={isOnline ? "#fff" : "#94a3b8"}
          />
        </View>
        <View style={styles.viewToggle}>
          <TouchableOpacity
            style={[styles.viewBtn, viewMode === "map" && styles.viewBtnActive]}
            onPress={() => setViewMode("map")}
          >
            <Text style={[styles.viewBtnText, viewMode === "map" && styles.viewBtnTextActive]}>
              🗺️
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.viewBtn, viewMode === "list" && styles.viewBtnActive]}
            onPress={() => setViewMode("list")}
          >
            <Text style={[styles.viewBtnText, viewMode === "list" && styles.viewBtnTextActive]}>
              📋
            </Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Map or List view */}
      {viewMode === "map" ? (
        <View style={styles.mapContainer}>
          <DriverMap rides={rides} token={token} isOnline={isOnline} />
        </View>
      ) : (
        <View style={styles.listContainer}>
          {rides.length === 0 && !loading ? (
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyIcon}>🚗</Text>
              <Text style={styles.empty}>
                {isOnline
                  ? "Нет открытых заявок поблизости"
                  : "Включите «На линии», чтобы видеть заявки"}
              </Text>
              <Button title="Обновить" onPress={load} variant="outline" />
            </View>
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
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f0fdf4" },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    padding: 12,
    backgroundColor: "#fff",
    borderBottomWidth: 1,
    borderBottomColor: "#e2e8f0",
  },
  onlineToggle: { flexDirection: "row", alignItems: "center", gap: 8 },
  onlineLabel: { fontSize: 14, fontWeight: "600", color: "#0f172a" },
  viewToggle: { flexDirection: "row", gap: 4 },
  viewBtn: {
    width: 40,
    height: 40,
    borderRadius: 8,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: "#f1f5f9",
  },
  viewBtnActive: { backgroundColor: "#dcfce7" },
  viewBtnText: { fontSize: 18 },
  viewBtnTextActive: {},
  mapContainer: { 
    flex: 1, 
    minHeight: SCREEN_HEIGHT * 0.7,
  },
  listContainer: { flex: 1 },
  list: { padding: 16, paddingBottom: 32 },
  rideCard: { marginBottom: 12 },
  rideHeader: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 8 },
  rideId: { fontSize: 12, color: "#64748b" },
  rideRoute: { fontSize: 14, color: "#0f172a", marginBottom: 4 },
  hint: { fontSize: 12, color: "#16a34a" },
  emptyContainer: { flex: 1, justifyContent: "center", alignItems: "center", padding: 24 },
  emptyIcon: { fontSize: 48, marginBottom: 16 },
  empty: { fontSize: 15, color: "#64748b", textAlign: "center", marginBottom: 16 },
});
