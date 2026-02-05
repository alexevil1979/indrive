/**
 * RideMap — interactive map for selecting pickup/destination
 * Shows nearby drivers as markers
 */
import { useState, useEffect, useRef, useCallback } from "react";
import {
  View,
  StyleSheet,
  Text,
  TouchableOpacity,
  ActivityIndicator,
  Platform,
} from "react-native";
import MapView, { Marker, Region, PROVIDER_GOOGLE } from "react-native-maps";
import * as Location from "expo-location";
import { type Point, type NearbyDriver, getNearbyDrivers } from "../lib/api";

type SelectionMode = "from" | "to" | null;

type Props = {
  from: Point;
  to: Point;
  onFromChange: (point: Point) => void;
  onToChange: (point: Point) => void;
  token: string | null;
};

const MOSCOW_CENTER = { lat: 55.7558, lng: 37.6173 };

export function RideMap({ from, to, onFromChange, onToChange, token }: Props) {
  const mapRef = useRef<MapView>(null);
  const [selectionMode, setSelectionMode] = useState<SelectionMode>(null);
  const [nearbyDrivers, setNearbyDrivers] = useState<NearbyDriver[]>([]);
  const [loadingLocation, setLoadingLocation] = useState(false);
  const [region, setRegion] = useState<Region>({
    latitude: from.lat || MOSCOW_CENTER.lat,
    longitude: from.lng || MOSCOW_CENTER.lng,
    latitudeDelta: 0.02,
    longitudeDelta: 0.02,
  });

  // Get current location on mount
  useEffect(() => {
    (async () => {
      setLoadingLocation(true);
      try {
        const { status } = await Location.requestForegroundPermissionsAsync();
        if (status === "granted") {
          const location = await Location.getCurrentPositionAsync({
            accuracy: Location.Accuracy.Balanced,
          });
          const newFrom: Point = {
            lat: location.coords.latitude,
            lng: location.coords.longitude,
            address: "Моё местоположение",
          };
          onFromChange(newFrom);
          setRegion({
            latitude: location.coords.latitude,
            longitude: location.coords.longitude,
            latitudeDelta: 0.02,
            longitudeDelta: 0.02,
          });
        }
      } catch {
        // Use default location
      } finally {
        setLoadingLocation(false);
      }
    })();
  }, []);

  // Fetch nearby drivers when from changes
  useEffect(() => {
    if (!token || !from.lat || !from.lng) return;
    const fetchDrivers = async () => {
      const drivers = await getNearbyDrivers(token, from.lat, from.lng, 5000);
      setNearbyDrivers(drivers);
    };
    fetchDrivers();
    // Refresh every 30 seconds
    const interval = setInterval(fetchDrivers, 30000);
    return () => clearInterval(interval);
  }, [token, from.lat, from.lng]);

  const handleMapPress = useCallback(
    (e: { nativeEvent: { coordinate: { latitude: number; longitude: number } } }) => {
      if (!selectionMode) return;
      const { latitude, longitude } = e.nativeEvent.coordinate;
      const point: Point = { lat: latitude, lng: longitude };

      if (selectionMode === "from") {
        onFromChange(point);
      } else {
        onToChange(point);
      }
      setSelectionMode(null);
    },
    [selectionMode, onFromChange, onToChange]
  );

  const handleMarkerDragEnd = useCallback(
    (type: "from" | "to", e: { nativeEvent: { coordinate: { latitude: number; longitude: number } } }) => {
      const { latitude, longitude } = e.nativeEvent.coordinate;
      const point: Point = { lat: latitude, lng: longitude };
      if (type === "from") {
        onFromChange(point);
      } else {
        onToChange(point);
      }
    },
    [onFromChange, onToChange]
  );

  const fitToMarkers = useCallback(() => {
    if (!mapRef.current) return;
    const coords = [
      { latitude: from.lat, longitude: from.lng },
      { latitude: to.lat, longitude: to.lng },
    ];
    mapRef.current.fitToCoordinates(coords, {
      edgePadding: { top: 80, right: 80, bottom: 80, left: 80 },
      animated: true,
    });
  }, [from, to]);

  return (
    <View style={styles.container}>
      {/* Selection mode indicator */}
      {selectionMode && (
        <View style={styles.selectionBanner}>
          <Text style={styles.selectionText}>
            Нажмите на карту, чтобы выбрать {selectionMode === "from" ? "точку отправления" : "точку назначения"}
          </Text>
          <TouchableOpacity onPress={() => setSelectionMode(null)}>
            <Text style={styles.cancelText}>Отмена</Text>
          </TouchableOpacity>
        </View>
      )}

      {/* Map */}
      <MapView
        ref={mapRef}
        style={styles.map}
        provider={Platform.OS === "android" ? PROVIDER_GOOGLE : undefined}
        initialRegion={region}
        onRegionChangeComplete={setRegion}
        onPress={handleMapPress}
        showsUserLocation
        showsMyLocationButton
      >
        {/* From marker */}
        <Marker
          coordinate={{ latitude: from.lat, longitude: from.lng }}
          title="Откуда"
          description={from.address}
          pinColor="#16a34a"
          draggable
          onDragEnd={(e) => handleMarkerDragEnd("from", e)}
        />

        {/* To marker */}
        <Marker
          coordinate={{ latitude: to.lat, longitude: to.lng }}
          title="Куда"
          description={to.address}
          pinColor="#dc2626"
          draggable
          onDragEnd={(e) => handleMarkerDragEnd("to", e)}
        />

        {/* Nearby drivers */}
        {nearbyDrivers.map((driver) => (
          <Marker
            key={driver.driver_id}
            coordinate={{ latitude: driver.lat, longitude: driver.lng }}
            title="Водитель"
            description={`${Math.round(driver.distance)}м`}
          >
            <View style={styles.driverMarker}>
              <Text style={styles.driverMarkerText}>🚗</Text>
            </View>
          </Marker>
        ))}
      </MapView>

      {/* Loading overlay */}
      {loadingLocation && (
        <View style={styles.loadingOverlay}>
          <ActivityIndicator size="large" color="#2563eb" />
          <Text style={styles.loadingText}>Определение местоположения...</Text>
        </View>
      )}

      {/* Bottom controls */}
      <View style={styles.controls}>
        <View style={styles.controlRow}>
          <TouchableOpacity
            style={[styles.controlBtn, selectionMode === "from" && styles.controlBtnActive]}
            onPress={() => setSelectionMode(selectionMode === "from" ? null : "from")}
          >
            <View style={[styles.dot, { backgroundColor: "#16a34a" }]} />
            <Text style={styles.controlBtnText} numberOfLines={1}>
              {from.address || `${from.lat.toFixed(4)}, ${from.lng.toFixed(4)}`}
            </Text>
          </TouchableOpacity>
        </View>
        <View style={styles.controlRow}>
          <TouchableOpacity
            style={[styles.controlBtn, selectionMode === "to" && styles.controlBtnActive]}
            onPress={() => setSelectionMode(selectionMode === "to" ? null : "to")}
          >
            <View style={[styles.dot, { backgroundColor: "#dc2626" }]} />
            <Text style={styles.controlBtnText} numberOfLines={1}>
              {to.address || `${to.lat.toFixed(4)}, ${to.lng.toFixed(4)}`}
            </Text>
          </TouchableOpacity>
        </View>
        <TouchableOpacity style={styles.fitBtn} onPress={fitToMarkers}>
          <Text style={styles.fitBtnText}>Показать маршрут</Text>
        </TouchableOpacity>

        {nearbyDrivers.length > 0 && (
          <Text style={styles.driversInfo}>
            🚗 Водителей поблизости: {nearbyDrivers.length}
          </Text>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, minHeight: 350 },
  map: { flex: 1, borderRadius: 12 },
  selectionBanner: {
    position: "absolute",
    top: 0,
    left: 0,
    right: 0,
    zIndex: 10,
    backgroundColor: "#2563eb",
    padding: 12,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    borderTopLeftRadius: 12,
    borderTopRightRadius: 12,
  },
  selectionText: { color: "#fff", fontSize: 13, flex: 1 },
  cancelText: { color: "#fff", fontWeight: "600", marginLeft: 12 },
  loadingOverlay: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: "rgba(255,255,255,0.9)",
    justifyContent: "center",
    alignItems: "center",
    borderRadius: 12,
  },
  loadingText: { marginTop: 8, color: "#64748b" },
  controls: {
    position: "absolute",
    bottom: 12,
    left: 12,
    right: 12,
    backgroundColor: "#fff",
    borderRadius: 12,
    padding: 12,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 4,
  },
  controlRow: { marginBottom: 8 },
  controlBtn: {
    flexDirection: "row",
    alignItems: "center",
    padding: 10,
    borderRadius: 8,
    backgroundColor: "#f8fafc",
    borderWidth: 1,
    borderColor: "#e2e8f0",
  },
  controlBtnActive: { borderColor: "#2563eb", backgroundColor: "#eff6ff" },
  controlBtnText: { flex: 1, fontSize: 13, color: "#0f172a", marginLeft: 8 },
  dot: { width: 12, height: 12, borderRadius: 6 },
  fitBtn: {
    backgroundColor: "#2563eb",
    padding: 12,
    borderRadius: 8,
    alignItems: "center",
  },
  fitBtnText: { color: "#fff", fontWeight: "600", fontSize: 14 },
  driversInfo: { marginTop: 8, fontSize: 12, color: "#64748b", textAlign: "center" },
  driverMarker: {
    backgroundColor: "#fff",
    padding: 4,
    borderRadius: 20,
    borderWidth: 2,
    borderColor: "#2563eb",
  },
  driverMarkerText: { fontSize: 16 },
});
