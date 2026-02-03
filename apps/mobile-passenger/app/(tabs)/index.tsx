/**
 * Home — map stub + ride request (from/to, create)
 */
import { useState } from "react";
import { View, Text, StyleSheet, ScrollView, Alert } from "react-native";
import { useRouter } from "expo-router";
import { Button, Card, Input, MapPlaceholder } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import { createRide, type Point } from "../../lib/api";

const defaultFrom: Point = { lat: 55.7558, lng: 37.6173, address: "Москва, центр" };
const defaultTo: Point = { lat: 55.76, lng: 37.62, address: "Москва, север" };

export default function HomeScreen() {
  const { token } = useAuth();
  const router = useRouter();
  const [from, setFrom] = useState<Point>(defaultFrom);
  const [to, setTo] = useState<Point>(defaultTo);
  const [loading, setLoading] = useState(false);

  const handleCreateRide = async () => {
    if (!token) return;
    setLoading(true);
    try {
      const ride = await createRide(token, from, to);
      router.push(`/ride/${ride.id}`);
    } catch (e) {
      Alert.alert(
        "Ошибка",
        e instanceof Error ? e.message : "Не удалось создать поездку"
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <View style={styles.mapSection}>
        <Text style={styles.mapLabel}>Карта (заглушка — OpenStreetMap)</Text>
        <MapPlaceholder message="Откуда → Куда (карта в следующей версии)" />
      </View>

      <Card style={styles.card}>
        <Text style={styles.cardTitle}>Новая поездка</Text>
        <Input
          label="Откуда"
          placeholder="Широта, долгота или адрес"
          value={from.address ?? `${from.lat.toFixed(4)}, ${from.lng.toFixed(4)}`}
          onChangeText={(t) => setFrom((p) => ({ ...p, address: t }))}
        />
        <Input
          label="Куда"
          placeholder="Широта, долгота или адрес"
          value={to.address ?? `${to.lat.toFixed(4)}, ${to.lng.toFixed(4)}`}
          onChangeText={(t) => setTo((p) => ({ ...p, address: t }))}
        />
        <Button
          title={loading ? "Создаём..." : "Найти водителя (создать заявку)"}
          onPress={handleCreateRide}
          variant="primary"
          disabled={loading}
        />
      </Card>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f8fafc" },
  content: { padding: 16, paddingBottom: 32 },
  mapSection: { marginBottom: 16 },
  mapLabel: { fontSize: 12, color: "#64748b", marginBottom: 8 },
  card: { marginBottom: 16 },
  cardTitle: { fontSize: 16, fontWeight: "600", marginBottom: 12, color: "#0f172a" },
});
