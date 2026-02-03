/**
 * Ride detail — bids list, accept bid (bidding UI)
 */
import { useEffect, useState } from "react";
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Alert,
  TouchableOpacity,
  RefreshControl,
} from "react-native";
import { useLocalSearchParams, useRouter } from "expo-router";
import { Button, Card, Badge } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import {
  getRide,
  listBids,
  acceptBid,
  updateRideStatus,
  type Ride,
  type Bid,
} from "../../lib/api";

const statusLabel: Record<string, string> = {
  requested: "Ожидает ставок",
  bidding: "Ставки",
  matched: "Водитель найден",
  in_progress: "В пути",
  completed: "Завершена",
  cancelled: "Отменена",
};

export default function RideDetailScreen() {
  const params = useLocalSearchParams<{ id: string }>();
  const id = Array.isArray(params.id) ? params.id[0] : params.id;
  const { token } = useAuth();
  const router = useRouter();
  const [ride, setRide] = useState<Ride | null>(null);
  const [bids, setBids] = useState<Bid[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const load = async () => {
    if (!token || !id) return;
    try {
      const [r, b] = await Promise.all([getRide(token, id), listBids(token, id)]);
      setRide(r);
      setBids(b);
    } catch {
      setRide(null);
      setBids([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    load();
  }, [token, id]);

  const onRefresh = () => {
    setRefreshing(true);
    load();
  };

  const handleAcceptBid = (bidId: string) => {
    if (!token || !id) return;
    Alert.alert("Принять ставку?", "Подтвердите выбор водителя", [
      { text: "Отмена", style: "cancel" },
      {
        text: "Принять",
        onPress: async () => {
          try {
            const updated = await acceptBid(token, id, bidId);
            setRide(updated);
            setBids([]);
          } catch (e) {
            Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось принять ставку");
          }
        },
      },
    ]);
  };

  const handleStatus = (status: "in_progress" | "completed" | "cancelled") => {
    if (!token || !id) return;
    const labels = { in_progress: "Начать поездку", completed: "Завершить", cancelled: "Отменить" };
    Alert.alert(labels[status], "Подтвердить?", [
      { text: "Нет", style: "cancel" },
      {
        text: "Да",
        onPress: async () => {
          try {
            const updated = await updateRideStatus(token, id, status);
            setRide(updated);
          } catch (e) {
            Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось обновить статус");
          }
        },
      },
    ]);
  };

  if (loading || !ride) {
    return (
      <View style={styles.center}>
        <Text style={styles.muted}>{loading ? "Загрузка..." : "Поездка не найдена"}</Text>
      </View>
    );
  }

  const canAcceptBids =
    (ride.status === "requested" || ride.status === "bidding") && bids.length > 0;
  const isMatchedOrActive =
    ride.status === "matched" || ride.status === "in_progress";

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
    >
      <Card style={styles.card}>
        <View style={styles.row}>
          <Text style={styles.label}>Статус</Text>
          <Badge
            label={statusLabel[ride.status] ?? ride.status}
            variant={
              ride.status === "completed"
                ? "success"
                : ride.status === "cancelled"
                  ? "error"
                  : "default"
            }
          />
        </View>
        <Text style={styles.route}>
          Откуда: {ride.from.address ?? `${ride.from.lat}, ${ride.from.lng}`}
        </Text>
        <Text style={styles.route}>
          Куда: {ride.to.address ?? `${ride.to.lat}, ${ride.to.lng}`}
        </Text>
        {ride.price != null ? (
          <Text style={styles.price}>Цена: {ride.price} ₽</Text>
        ) : null}
      </Card>

      {canAcceptBids ? (
        <Card style={styles.card}>
          <Text style={styles.sectionTitle}>Ставки водителей</Text>
          {bids
            .filter((b) => b.status === "pending")
            .map((bid) => (
              <View key={bid.id} style={styles.bidRow}>
                <Text style={styles.bidPrice}>{bid.price} ₽</Text>
                <Button
                  title="Принять"
                  onPress={() => handleAcceptBid(bid.id)}
                  variant="primary"
                  style={styles.bidBtn}
                />
              </View>
            ))}
        </Card>
      ) : null}

      {isMatchedOrActive ? (
        <Card style={styles.card}>
          <Text style={styles.sectionTitle}>Действия</Text>
          {ride.status === "matched" ? (
            <Button
              title="Начать поездку"
              onPress={() => handleStatus("in_progress")}
              variant="primary"
            />
          ) : null}
          {ride.status === "in_progress" ? (
            <Button
              title="Завершить поездку"
              onPress={() => handleStatus("completed")}
              variant="primary"
              style={styles.actionBtn}
            />
          ) : null}
          <Button
            title="Отменить"
            onPress={() => handleStatus("cancelled")}
            variant="outline"
            style={styles.actionBtn}
          />
        </Card>
      ) : null}

      {ride.status === "requested" || ride.status === "bidding" ? (
        <Text style={styles.hint}>Ожидайте ставки от водителей. Обновите страницу.</Text>
      ) : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f8fafc" },
  content: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, justifyContent: "center", alignItems: "center", padding: 24 },
  muted: { color: "#64748b" },
  card: { marginBottom: 16 },
  row: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 8 },
  label: { fontSize: 12, color: "#64748b" },
  route: { fontSize: 14, color: "#0f172a", marginBottom: 4 },
  price: { fontSize: 16, fontWeight: "600", marginTop: 8 },
  sectionTitle: { fontSize: 14, fontWeight: "600", marginBottom: 12, color: "#0f172a" },
  bidRow: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", marginBottom: 12 },
  bidPrice: { fontSize: 18, fontWeight: "600" },
  bidBtn: { minWidth: 100 },
  actionBtn: { marginTop: 8 },
  hint: { fontSize: 12, color: "#64748b", textAlign: "center", marginTop: 8 },
});
