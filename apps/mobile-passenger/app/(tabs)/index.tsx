/**
 * Home — interactive map + ride request (from/to, create)
 */
import { useState, useCallback } from "react";
import { View, Text, StyleSheet, ScrollView, Alert, Dimensions } from "react-native";
import { useRouter } from "expo-router";
import { Button, Card } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import { createRide, type Point } from "../../lib/api";
import { RideMap } from "../../components/RideMap";

const MOSCOW_CENTER: Point = { lat: 55.7558, lng: 37.6173, address: "Москва, центр" };
const MOSCOW_NORTH: Point = { lat: 55.76, lng: 37.62, address: "Москва, север" };

const { height: SCREEN_HEIGHT } = Dimensions.get("window");

export default function HomeScreen() {
  const { token } = useAuth();
  const router = useRouter();
  const [from, setFrom] = useState<Point>(MOSCOW_CENTER);
  const [to, setTo] = useState<Point>(MOSCOW_NORTH);
  const [loading, setLoading] = useState(false);

  const handleFromChange = useCallback((point: Point) => {
    setFrom((prev) => ({ ...prev, ...point }));
  }, []);

  const handleToChange = useCallback((point: Point) => {
    setTo((prev) => ({ ...prev, ...point }));
  }, []);

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
    <View style={styles.container}>
      {/* Map takes most of the screen */}
      <View style={styles.mapContainer}>
        <RideMap
          from={from}
          to={to}
          onFromChange={handleFromChange}
          onToChange={handleToChange}
          token={token}
        />
      </View>

      {/* Bottom card for creating ride */}
      <Card style={styles.bottomCard}>
        <Text style={styles.cardTitle}>Новая поездка</Text>
        <View style={styles.routeInfo}>
          <View style={styles.routeRow}>
            <View style={[styles.dot, { backgroundColor: "#16a34a" }]} />
            <Text style={styles.routeText} numberOfLines={1}>
              {from.address || `${from.lat.toFixed(4)}, ${from.lng.toFixed(4)}`}
            </Text>
          </View>
          <View style={styles.routeDivider} />
          <View style={styles.routeRow}>
            <View style={[styles.dot, { backgroundColor: "#dc2626" }]} />
            <Text style={styles.routeText} numberOfLines={1}>
              {to.address || `${to.lat.toFixed(4)}, ${to.lng.toFixed(4)}`}
            </Text>
          </View>
        </View>
        <Button
          title={loading ? "Создаём..." : "Найти водителя"}
          onPress={handleCreateRide}
          variant="primary"
          disabled={loading}
        />
      </Card>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f8fafc" },
  mapContainer: { 
    flex: 1, 
    minHeight: SCREEN_HEIGHT * 0.5,
    margin: 12,
    borderRadius: 16,
    overflow: "hidden",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 4,
  },
  bottomCard: { 
    margin: 12, 
    marginTop: 0,
    marginBottom: 24,
  },
  cardTitle: { 
    fontSize: 18, 
    fontWeight: "700", 
    marginBottom: 12, 
    color: "#0f172a",
    textAlign: "center",
  },
  routeInfo: { marginBottom: 16 },
  routeRow: { 
    flexDirection: "row", 
    alignItems: "center", 
    paddingVertical: 8,
  },
  dot: { 
    width: 12, 
    height: 12, 
    borderRadius: 6, 
    marginRight: 12,
  },
  routeText: { 
    flex: 1, 
    fontSize: 14, 
    color: "#334155",
  },
  routeDivider: {
    width: 2,
    height: 16,
    backgroundColor: "#e2e8f0",
    marginLeft: 5,
    marginVertical: 2,
  },
});
