/**
 * DriverMap — shows available ride requests on a map
 * Updates driver location in the background
 */
import { useState, useEffect, useRef, useCallback } from "react";
import {
  View,
  StyleSheet,
  Text,
  TouchableOpacity,
  ActivityIndicator,
  Platform,
  Linking,
} from "react-native";
import MapView, { Marker, Region, PROVIDER_GOOGLE, Callout } from "react-native-maps";
import * as Location from "expo-location";
import { useRouter } from "expo-router";
import { type Ride, updateDriverLocation } from "../lib/api";

type Props = {
  rides: Ride[];
  token: string | null;
  isOnline: boolean;
};

const MOSCOW_CENTER = { lat: 55.7558, lng: 37.6173 };

export function DriverMap({ rides, token, isOnline }: Props) {
  const mapRef = useRef<MapView>(null);
  const router = useRouter();
  const [currentLocation, setCurrentLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [loadingLocation, setLoadingLocation] = useState(true);
  const [region, setRegion] = useState<Region>({
    latitude: MOSCOW_CENTER.lat,
    longitude: MOSCOW_CENTER.lng,
    latitudeDelta: 0.05,
    longitudeDelta: 0.05,
  });

  // Get current location and start tracking
  useEffect(() => {
    let locationSubscription: Location.LocationSubscription | null = null;

    (async () => {
      setLoadingLocation(true);
      try {
        const { status } = await Location.requestForegroundPermissionsAsync();
        if (status !== "granted") {
          setLoadingLocation(false);
          return;
        }

        // Get initial location
        const location = await Location.getCurrentPositionAsync({
          accuracy: Location.Accuracy.Balanced,
        });
        const coords = {
          lat: location.coords.latitude,
          lng: location.coords.longitude,
        };
        setCurrentLocation(coords);
        setRegion({
          latitude: coords.lat,
          longitude: coords.lng,
          latitudeDelta: 0.05,
          longitudeDelta: 0.05,
        });

        // Update backend with location
        if (token && isOnline) {
          updateDriverLocation(token, coords.lat, coords.lng);
        }

        // Start watching location
        locationSubscription = await Location.watchPositionAsync(
          {
            accuracy: Location.Accuracy.High,
            timeInterval: 10000, // 10 seconds
            distanceInterval: 50, // 50 meters
          },
          (newLocation) => {
            const newCoords = {
              lat: newLocation.coords.latitude,
              lng: newLocation.coords.longitude,
            };
            setCurrentLocation(newCoords);
            // Update backend if online
            if (token && isOnline) {
              updateDriverLocation(token, newCoords.lat, newCoords.lng);
            }
          }
        );
      } catch (error) {
        console.warn("Location error:", error);
      } finally {
        setLoadingLocation(false);
      }
    })();

    return () => {
      locationSubscription?.remove();
    };
  }, [token, isOnline]);

  const centerOnCurrentLocation = useCallback(() => {
    if (!currentLocation || !mapRef.current) return;
    mapRef.current.animateToRegion({
      latitude: currentLocation.lat,
      longitude: currentLocation.lng,
      latitudeDelta: 0.02,
      longitudeDelta: 0.02,
    });
  }, [currentLocation]);

  const fitToRides = useCallback(() => {
    if (!mapRef.current || rides.length === 0) return;
    const coords = rides.flatMap((r) => [
      { latitude: r.from.lat, longitude: r.from.lng },
      { latitude: r.to.lat, longitude: r.to.lng },
    ]);
    if (currentLocation) {
      coords.push({ latitude: currentLocation.lat, longitude: currentLocation.lng });
    }
    mapRef.current.fitToCoordinates(coords, {
      edgePadding: { top: 80, right: 80, bottom: 80, left: 80 },
      animated: true,
    });
  }, [rides, currentLocation]);

  const openNavigation = (lat: number, lng: number) => {
    const url = Platform.select({
      ios: `maps:?daddr=${lat},${lng}`,
      android: `google.navigation:q=${lat},${lng}`,
    });
    if (url) {
      Linking.openURL(url).catch(() => {
        // Fallback to Google Maps URL
        Linking.openURL(`https://www.google.com/maps/dir/?api=1&destination=${lat},${lng}`);
      });
    }
  };

  const handleRidePress = (ride: Ride) => {
    router.push(`/ride/${ride.id}`);
  };

  return (
    <View style={styles.container}>
      {/* Online status indicator */}
      <View style={[styles.statusBanner, isOnline ? styles.statusOnline : styles.statusOffline]}>
        <View style={[styles.statusDot, isOnline ? styles.dotOnline : styles.dotOffline]} />
        <Text style={styles.statusText}>
          {isOnline ? "Вы на линии" : "Вы офлайн"}
        </Text>
      </View>

      {/* Map */}
      <MapView
        ref={mapRef}
        style={styles.map}
        provider={Platform.OS === "android" ? PROVIDER_GOOGLE : undefined}
        initialRegion={region}
        onRegionChangeComplete={setRegion}
        showsUserLocation
        showsMyLocationButton={false}
      >
        {/* Ride request markers */}
        {rides.map((ride) => (
          <Marker
            key={ride.id}
            coordinate={{ latitude: ride.from.lat, longitude: ride.from.lng }}
            onCalloutPress={() => handleRidePress(ride)}
          >
            <View style={styles.rideMarker}>
              <Text style={styles.rideMarkerText}>📍</Text>
            </View>
            <Callout tooltip onPress={() => handleRidePress(ride)}>
              <View style={styles.callout}>
                <Text style={styles.calloutTitle}>Заявка #{ride.id.slice(0, 8)}</Text>
                <Text style={styles.calloutText} numberOfLines={1}>
                  {ride.from.address || `${ride.from.lat.toFixed(4)}, ${ride.from.lng.toFixed(4)}`}
                </Text>
                <Text style={styles.calloutText} numberOfLines={1}>
                  → {ride.to.address || `${ride.to.lat.toFixed(4)}, ${ride.to.lng.toFixed(4)}`}
                </Text>
                <Text style={styles.calloutAction}>Нажмите для ставки</Text>
              </View>
            </Callout>
          </Marker>
        ))}
      </MapView>

      {/* Loading overlay */}
      {loadingLocation && (
        <View style={styles.loadingOverlay}>
          <ActivityIndicator size="large" color="#16a34a" />
          <Text style={styles.loadingText}>Определение местоположения...</Text>
        </View>
      )}

      {/* Map controls */}
      <View style={styles.mapControls}>
        <TouchableOpacity style={styles.controlButton} onPress={centerOnCurrentLocation}>
          <Text style={styles.controlIcon}>📍</Text>
        </TouchableOpacity>
        {rides.length > 0 && (
          <TouchableOpacity style={styles.controlButton} onPress={fitToRides}>
            <Text style={styles.controlIcon}>🗺️</Text>
          </TouchableOpacity>
        )}
      </View>

      {/* Ride count */}
      {rides.length > 0 && (
        <View style={styles.rideCount}>
          <Text style={styles.rideCountText}>
            Заявок поблизости: {rides.length}
          </Text>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  map: { flex: 1 },
  statusBanner: {
    position: "absolute",
    top: 12,
    left: 12,
    right: 12,
    zIndex: 10,
    flexDirection: "row",
    alignItems: "center",
    padding: 10,
    borderRadius: 20,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  statusOnline: { backgroundColor: "#dcfce7" },
  statusOffline: { backgroundColor: "#fef2f2" },
  statusDot: { width: 10, height: 10, borderRadius: 5, marginRight: 8 },
  dotOnline: { backgroundColor: "#16a34a" },
  dotOffline: { backgroundColor: "#dc2626" },
  statusText: { fontSize: 13, fontWeight: "500", color: "#0f172a" },
  loadingOverlay: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: "rgba(255,255,255,0.9)",
    justifyContent: "center",
    alignItems: "center",
  },
  loadingText: { marginTop: 8, color: "#64748b" },
  mapControls: {
    position: "absolute",
    right: 12,
    bottom: 100,
    gap: 8,
  },
  controlButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: "#fff",
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  controlIcon: { fontSize: 20 },
  rideMarker: {
    backgroundColor: "#fff",
    padding: 6,
    borderRadius: 20,
    borderWidth: 2,
    borderColor: "#16a34a",
  },
  rideMarkerText: { fontSize: 18 },
  callout: {
    backgroundColor: "#fff",
    padding: 12,
    borderRadius: 8,
    minWidth: 200,
    maxWidth: 250,
  },
  calloutTitle: { fontSize: 14, fontWeight: "600", color: "#0f172a", marginBottom: 4 },
  calloutText: { fontSize: 12, color: "#64748b", marginBottom: 2 },
  calloutAction: { fontSize: 12, color: "#16a34a", fontWeight: "500", marginTop: 8 },
  rideCount: {
    position: "absolute",
    bottom: 12,
    left: 12,
    backgroundColor: "#fff",
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 20,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  rideCountText: { fontSize: 13, fontWeight: "500", color: "#0f172a" },
});
