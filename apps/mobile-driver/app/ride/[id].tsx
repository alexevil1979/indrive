/**
 * Ride detail (driver) — place bid or status, start/complete, navigation stub
 */
import { useEffect, useState } from "react";
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Alert,
  RefreshControl,
  Linking,
  TouchableOpacity,
} from "react-native";
import { useLocalSearchParams, useRouter } from "expo-router";
import { Button, Card, Badge, Input } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import { useDriverTracking } from "../../hooks/useDriverTracking";
import {
  getRide,
  placeBid,
  updateRideStatus,
  type Ride,
} from "../../lib/api";

const statusLabel: Record<string, string> = {
  requested: "Ожидает ставок",
  bidding: "Ставки",
  matched: "Пассажир принял",
  in_progress: "В пути",
  completed: "Завершена",
  cancelled: "Отменена",
};

export default function RideDetailScreen() {
  const params = useLocalSearchParams<{ id: string }>();
  const id = Array.isArray(params.id) ? params.id[0] : params.id;
  const { token, userId } = useAuth();
  const router = useRouter();
  const [ride, setRide] = useState<Ride | null>(null);
  const [price, setPrice] = useState("");
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  // Start tracking when ride is matched or in_progress
  const isActiveRide = ride?.status === "matched" || ride?.status === "in_progress";
  const { isConnected: isTrackingConnected, isTracking } = useDriverTracking(
    id ?? null,
    userId ?? null,
    isActiveRide && !!ride?.driver_id
  );

  const load = async () => {
    if (!token || !id) return;
    try {
      const r = await getRide(token, id);
      setRide(r);
    } catch {
      setRide(null);
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

  const handlePlaceBid = async () => {
    if (!token || !id) return;
    const p = parseFloat(price.replace(",", "."));
    if (isNaN(p) || p <= 0) {
      Alert.alert("Ошибка", "Введите цену в рублях");
      return;
    }
    setLoading(true);
    try {
      await placeBid(token, id, p);
      Alert.alert("Готово", "Ставка отправлена");
      load();
    } catch (e) {
      Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось сделать ставку");
    } finally {
      setLoading(false);
    }
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

  const openNavigation = () => {
    if (!ride) return;
    const { from, to } = ride;
    const url = `https://www.google.com/maps/dir/?api=1&origin=${from.lat},${from.lng}&destination=${to.lat},${to.lng}`;
    Linking.openURL(url).catch(() => {
      Alert.alert("Ошибка", "Не удалось открыть карты");
    });
  };

  if (!ride) {
    return (
      <View style={styles.center}>
        <Text style={styles.muted}>Загрузка...</Text>
      </View>
    );
  }

  const canBid = ride.status === "requested" || ride.status === "bidding";
  const isMyRide = !!ride.driver_id;
  const isMatchedOrActive = ride.status === "matched" || ride.status === "in_progress";

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

        {/* Tracking status indicator */}
        {isActiveRide && (
          <View style={styles.trackingStatus}>
            <View
              style={[
                styles.trackingDot,
                isTracking ? styles.trackingDotActive : styles.trackingDotInactive,
              ]}
            />
            <Text style={styles.trackingText}>
              {isTracking
                ? "📍 Позиция передаётся пассажиру"
                : isTrackingConnected
                  ? "Подключено, ожидание GPS..."
                  : "Подключение к трекингу..."}
            </Text>
          </View>
        )}
      </Card>

      {canBid ? (
        <Card style={styles.card}>
          <Text style={styles.sectionTitle}>Сделать ставку (₽)</Text>
          <View style={styles.bidRow}>
            <Button title="500 ₽" onPress={() => setPrice("500")} variant="outline" style={styles.bidBtn} />
            <Button title="700 ₽" onPress={() => setPrice("700")} variant="outline" style={styles.bidBtn} />
            <Button title="1000 ₽" onPress={() => setPrice("1000")} variant="outline" style={styles.bidBtn} />
          </View>
          <Input
            label="Цена (₽)"
            placeholder="500"
            value={price}
            onChangeText={setPrice}
            keyboardType="numeric"
          />
          <Button
            title="Подтвердить ставку"
            onPress={handlePlaceBid}
            variant="primary"
            disabled={loading || !price.trim()}
          />
        </Card>
      ) : null}

      {isMyRide && isMatchedOrActive ? (
        <>
          {/* Chat button */}
          <Card style={styles.card}>
            <TouchableOpacity
              style={styles.chatButton}
              onPress={() => router.push(`/chat/${id}`)}
            >
              <Text style={styles.chatButtonIcon}>💬</Text>
              <Text style={styles.chatButtonText}>Чат с пассажиром</Text>
              <Text style={styles.chatButtonArrow}>→</Text>
            </TouchableOpacity>
          </Card>

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
              <>
                <Button
                  title="Открыть навигацию (Google Maps)"
                  onPress={openNavigation}
                  variant="outline"
                  style={styles.navBtn}
                />
                <Button
                  title="Завершить поездку"
                  onPress={() => handleStatus("completed")}
                  variant="primary"
                  style={styles.actionBtn}
                />
              </>
            ) : null}
            <Button
              title="Отменить"
              onPress={() => handleStatus("cancelled")}
              variant="outline"
              style={styles.actionBtn}
            />
          </Card>
        </>
      ) : null}

      {ride.status === "completed" ? (
        <Card style={styles.card}>
          <View style={styles.completedHeader}>
            <Text style={styles.completedIcon}>✅</Text>
            <Text style={styles.completedTitle}>Поездка завершена</Text>
          </View>
          <TouchableOpacity
            style={styles.rateButton}
            onPress={() => router.push(`/rate/${id}`)}
          >
            <Text style={styles.rateButtonIcon}>⭐</Text>
            <Text style={styles.rateButtonText}>Оценить пассажира</Text>
            <Text style={styles.rateButtonArrow}>→</Text>
          </TouchableOpacity>
        </Card>
      ) : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f0fdf4" },
  content: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, justifyContent: "center", alignItems: "center", padding: 24 },
  muted: { color: "#64748b" },
  card: { marginBottom: 16 },
  row: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 8 },
  label: { fontSize: 12, color: "#64748b" },
  route: { fontSize: 14, color: "#0f172a", marginBottom: 4 },
  price: { fontSize: 16, fontWeight: "600", marginTop: 8 },
  // Tracking status
  trackingStatus: { flexDirection: "row", alignItems: "center", marginTop: 12, paddingTop: 12, borderTopWidth: 1, borderTopColor: "#e2e8f0" },
  trackingDot: { width: 10, height: 10, borderRadius: 5, marginRight: 8 },
  trackingDotActive: { backgroundColor: "#16a34a" },
  trackingDotInactive: { backgroundColor: "#f59e0b" },
  trackingText: { fontSize: 12, color: "#64748b" },
  sectionTitle: { fontSize: 14, fontWeight: "600", marginBottom: 12, color: "#0f172a" },
  bidRow: { flexDirection: "row", flexWrap: "wrap", gap: 8, marginBottom: 12 },
  bidBtn: { marginRight: 8, marginBottom: 8 },
  navBtn: { marginTop: 8 },
  actionBtn: { marginTop: 8 },
  // Chat button
  chatButton: { flexDirection: "row", alignItems: "center", paddingVertical: 12 },
  chatButtonIcon: { fontSize: 24, marginRight: 12 },
  chatButtonText: { flex: 1, fontSize: 16, fontWeight: "600", color: "#0f172a" },
  chatButtonArrow: { fontSize: 18, color: "#64748b" },
  // Completed ride
  completedHeader: { flexDirection: "row", alignItems: "center", marginBottom: 16 },
  completedIcon: { fontSize: 28, marginRight: 12 },
  completedTitle: { fontSize: 18, fontWeight: "700", color: "#16a34a" },
  rateButton: { flexDirection: "row", alignItems: "center", paddingVertical: 14, backgroundColor: "#fef3c7", borderRadius: 12, paddingHorizontal: 16 },
  rateButtonIcon: { fontSize: 24, marginRight: 12 },
  rateButtonText: { flex: 1, fontSize: 16, fontWeight: "600", color: "#92400e" },
  rateButtonArrow: { fontSize: 18, color: "#92400e" },
});
