/**
 * Driver Tracking Map — displays real-time driver location on a map
 */
import { useRef, useEffect, useCallback } from "react";
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Platform,
  ActivityIndicator,
} from "react-native";
import MapView, { Marker, Polyline, Region, PROVIDER_GOOGLE } from "react-native-maps";
import { useDriverTracking, type DriverLocation } from "../hooks/useDriverTracking";

type Point = {
  lat: number;
  lng: number;
  address?: string;
};

type Props = {
  rideId: string;
  passengerId: string;
  isActive: boolean;
  from: Point;
  to: Point;
};

export function DriverTrackingMap({ rideId, passengerId, isActive, from, to }: Props) {
  const mapRef = useRef<MapView>(null);
  const { isConnected, isDriverOnline, driverLocation } = useDriverTracking(
    rideId,
    passengerId,
    isActive
  );

  // Fit map to show all markers
  const fitToMarkers = useCallback(() => {
    if (!mapRef.current) return;

    const coordinates = [
      { latitude: from.lat, longitude: from.lng },
      { latitude: to.lat, longitude: to.lng },
    ];

    if (driverLocation) {
      coordinates.push({
        latitude: driverLocation.lat,
        longitude: driverLocation.lng,
      });
    }

    mapRef.current.fitToCoordinates(coordinates, {
      edgePadding: { top: 80, right: 40, bottom: 80, left: 40 },
      animated: true,
    });
  }, [from, to, driverLocation]);

  // Center on driver
  const centerOnDriver = useCallback(() => {
    if (!mapRef.current || !driverLocation) return;

    mapRef.current.animateToRegion({
      latitude: driverLocation.lat,
      longitude: driverLocation.lng,
      latitudeDelta: 0.01,
      longitudeDelta: 0.01,
    });
  }, [driverLocation]);

  // Initial fit
  useEffect(() => {
    setTimeout(fitToMarkers, 500);
  }, [fitToMarkers]);

  // Calculate rotation for driver marker
  const getDriverRotation = (heading?: number): number => {
    return heading ?? 0;
  };

  return (
    <View style={styles.container}>
      {/* Connection status */}
      <View
        style={[
          styles.statusBanner,
          isConnected ? styles.statusConnected : styles.statusDisconnected,
        ]}
      >
        <View
          style={[
            styles.statusDot,
            isConnected ? styles.dotConnected : styles.dotDisconnected,
          ]}
        />
        <Text style={styles.statusText}>
          {!isConnected
            ? "Подключение..."
            : isDriverOnline
              ? "Водитель на связи"
              : "Ожидание водителя..."}
        </Text>
        {driverLocation && driverLocation.speed && (
          <Text style={styles.speedText}>{driverLocation.speed} км/ч</Text>
        )}
      </View>

      {/* Map */}
      <MapView
        ref={mapRef}
        style={styles.map}
        provider={Platform.OS === "android" ? PROVIDER_GOOGLE : undefined}
        initialRegion={{
          latitude: from.lat,
          longitude: from.lng,
          latitudeDelta: 0.02,
          longitudeDelta: 0.02,
        }}
        showsUserLocation
        showsMyLocationButton={false}
      >
        {/* Pickup marker */}
        <Marker
          coordinate={{ latitude: from.lat, longitude: from.lng }}
          title="Откуда"
          description={from.address}
          pinColor="#16a34a"
        />

        {/* Destination marker */}
        <Marker
          coordinate={{ latitude: to.lat, longitude: to.lng }}
          title="Куда"
          description={to.address}
          pinColor="#dc2626"
        />

        {/* Driver marker */}
        {driverLocation && (
          <Marker
            coordinate={{
              latitude: driverLocation.lat,
              longitude: driverLocation.lng,
            }}
            title="Водитель"
            rotation={getDriverRotation(driverLocation.heading)}
            anchor={{ x: 0.5, y: 0.5 }}
          >
            <View style={styles.driverMarker}>
              <Text style={styles.driverMarkerText}>🚗</Text>
            </View>
          </Marker>
        )}

        {/* Route line from driver to pickup (if driver is en route) */}
        {driverLocation && (
          <Polyline
            coordinates={[
              { latitude: driverLocation.lat, longitude: driverLocation.lng },
              { latitude: from.lat, longitude: from.lng },
            ]}
            strokeColor="#2563eb"
            strokeWidth={3}
            lineDashPattern={[10, 5]}
          />
        )}

        {/* Route line from pickup to destination */}
        <Polyline
          coordinates={[
            { latitude: from.lat, longitude: from.lng },
            { latitude: to.lat, longitude: to.lng },
          ]}
          strokeColor="#16a34a"
          strokeWidth={3}
        />
      </MapView>

      {/* Loading overlay */}
      {!isConnected && (
        <View style={styles.loadingOverlay}>
          <ActivityIndicator size="large" color="#2563eb" />
          <Text style={styles.loadingText}>Подключение к трекингу...</Text>
        </View>
      )}

      {/* Map controls */}
      <View style={styles.controls}>
        <TouchableOpacity style={styles.controlButton} onPress={fitToMarkers}>
          <Text style={styles.controlIcon}>📍</Text>
        </TouchableOpacity>
        {driverLocation && (
          <TouchableOpacity style={styles.controlButton} onPress={centerOnDriver}>
            <Text style={styles.controlIcon}>🚗</Text>
          </TouchableOpacity>
        )}
      </View>

      {/* Driver info card */}
      {isDriverOnline && driverLocation && (
        <View style={styles.driverCard}>
          <View style={styles.driverCardRow}>
            <Text style={styles.driverCardIcon}>🚗</Text>
            <View style={styles.driverCardInfo}>
              <Text style={styles.driverCardTitle}>Водитель в пути</Text>
              <Text style={styles.driverCardSubtitle}>
                Последнее обновление:{" "}
                {new Date(driverLocation.timestamp).toLocaleTimeString("ru-RU", {
                  hour: "2-digit",
                  minute: "2-digit",
                  second: "2-digit",
                })}
              </Text>
            </View>
          </View>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    position: "relative",
  },
  statusBanner: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    paddingVertical: 8,
    paddingHorizontal: 16,
    position: "absolute",
    top: 0,
    left: 0,
    right: 0,
    zIndex: 10,
  },
  statusConnected: {
    backgroundColor: "#dcfce7",
  },
  statusDisconnected: {
    backgroundColor: "#fef3c7",
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 8,
  },
  dotConnected: {
    backgroundColor: "#16a34a",
  },
  dotDisconnected: {
    backgroundColor: "#f59e0b",
  },
  statusText: {
    fontSize: 13,
    fontWeight: "500",
    color: "#0f172a",
  },
  speedText: {
    fontSize: 12,
    color: "#64748b",
    marginLeft: 8,
  },
  map: {
    flex: 1,
  },
  loadingOverlay: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: "rgba(255,255,255,0.8)",
    justifyContent: "center",
    alignItems: "center",
    zIndex: 5,
  },
  loadingText: {
    marginTop: 12,
    fontSize: 14,
    color: "#64748b",
  },
  driverMarker: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: "#2563eb",
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 4,
    elevation: 5,
  },
  driverMarkerText: {
    fontSize: 20,
  },
  controls: {
    position: "absolute",
    right: 16,
    bottom: 100,
    gap: 8,
  },
  controlButton: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: "#fff",
    justifyContent: "center",
    alignItems: "center",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.15,
    shadowRadius: 4,
    elevation: 3,
  },
  controlIcon: {
    fontSize: 20,
  },
  driverCard: {
    position: "absolute",
    bottom: 16,
    left: 16,
    right: 16,
    backgroundColor: "#fff",
    borderRadius: 12,
    padding: 16,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.15,
    shadowRadius: 8,
    elevation: 4,
  },
  driverCardRow: {
    flexDirection: "row",
    alignItems: "center",
  },
  driverCardIcon: {
    fontSize: 32,
    marginRight: 12,
  },
  driverCardInfo: {
    flex: 1,
  },
  driverCardTitle: {
    fontSize: 16,
    fontWeight: "600",
    color: "#0f172a",
  },
  driverCardSubtitle: {
    fontSize: 12,
    color: "#64748b",
    marginTop: 2,
  },
});
