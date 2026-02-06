/**
 * Universal MapView component
 * Supports Google Maps and Yandex Maps based on server settings
 */
import React, { useEffect, useState, useRef, forwardRef, useImperativeHandle } from "react";
import { View, StyleSheet, ActivityIndicator, Text } from "react-native";
import GoogleMapView, { Marker, Region, PROVIDER_GOOGLE, MapViewProps as GoogleMapViewProps } from "react-native-maps";
import { getMapSettings, MapProvider, MapSettings } from "@/lib/api";

// Re-export types for consumers
export type { Region, MapProvider };

export type MapMarker = {
  id: string;
  coordinate: { latitude: number; longitude: number };
  title?: string;
  description?: string;
  pinColor?: string;
};

export type MapViewHandle = {
  animateToRegion: (region: Region, duration?: number) => void;
  fitToMarkers: (markerIds: string[], options?: { edgePadding?: { top: number; right: number; bottom: number; left: number }; animated?: boolean }) => void;
};

type Props = {
  initialRegion?: Region;
  region?: Region;
  markers?: MapMarker[];
  showsUserLocation?: boolean;
  onRegionChange?: (region: Region) => void;
  onRegionChangeComplete?: (region: Region) => void;
  onPress?: (event: { nativeEvent: { coordinate: { latitude: number; longitude: number } } }) => void;
  onMarkerPress?: (markerId: string) => void;
  style?: object;
  children?: React.ReactNode;
};

// Default region (Moscow)
const DEFAULT_REGION: Region = {
  latitude: 55.7558,
  longitude: 37.6173,
  latitudeDelta: 0.0922,
  longitudeDelta: 0.0421,
};

const MapView = forwardRef<MapViewHandle, Props>((props, ref) => {
  const {
    initialRegion = DEFAULT_REGION,
    region,
    markers = [],
    showsUserLocation = false,
    onRegionChange,
    onRegionChangeComplete,
    onPress,
    onMarkerPress,
    style,
    children,
  } = props;

  const [mapSettings, setMapSettings] = useState<MapSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const googleMapRef = useRef<GoogleMapView>(null);

  useEffect(() => {
    loadMapSettings();
  }, []);

  async function loadMapSettings() {
    try {
      const settings = await getMapSettings();
      setMapSettings(settings);
    } catch (err) {
      console.error("Failed to load map settings:", err);
      // Fallback to Google Maps
      setMapSettings({ provider: "google", api_key: "" });
    } finally {
      setLoading(false);
    }
  }

  // Expose methods to parent
  useImperativeHandle(ref, () => ({
    animateToRegion: (reg: Region, duration = 500) => {
      if (mapSettings?.provider === "google" && googleMapRef.current) {
        googleMapRef.current.animateToRegion(reg, duration);
      }
      // For Yandex, would need different implementation
    },
    fitToMarkers: (markerIds: string[], options) => {
      if (mapSettings?.provider === "google" && googleMapRef.current) {
        googleMapRef.current.fitToSuppliedMarkers(markerIds, options);
      }
    },
  }));

  if (loading) {
    return (
      <View style={[styles.container, styles.center, style]}>
        <ActivityIndicator size="large" color="#007AFF" />
        <Text style={styles.loadingText}>Загрузка карты...</Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={[styles.container, styles.center, style]}>
        <Text style={styles.errorText}>{error}</Text>
      </View>
    );
  }

  // Google Maps
  if (mapSettings?.provider === "google") {
    return (
      <GoogleMapView
        ref={googleMapRef}
        style={[styles.map, style]}
        provider={PROVIDER_GOOGLE}
        initialRegion={initialRegion}
        region={region}
        showsUserLocation={showsUserLocation}
        showsMyLocationButton={true}
        onRegionChange={onRegionChange}
        onRegionChangeComplete={onRegionChangeComplete}
        onPress={onPress}
      >
        {markers.map((marker) => (
          <Marker
            key={marker.id}
            identifier={marker.id}
            coordinate={marker.coordinate}
            title={marker.title}
            description={marker.description}
            pinColor={marker.pinColor}
            onPress={() => onMarkerPress?.(marker.id)}
          />
        ))}
        {children}
      </GoogleMapView>
    );
  }

  // Yandex Maps
  if (mapSettings?.provider === "yandex") {
    return (
      <YandexMapView
        style={style}
        initialRegion={initialRegion}
        region={region}
        markers={markers}
        showsUserLocation={showsUserLocation}
        onRegionChangeComplete={onRegionChangeComplete}
        onPress={onPress}
        onMarkerPress={onMarkerPress}
        apiKey={mapSettings.api_key}
      >
        {children}
      </YandexMapView>
    );
  }

  // Fallback
  return (
    <View style={[styles.container, styles.center, style]}>
      <Text style={styles.errorText}>Провайдер карт не настроен</Text>
    </View>
  );
});

// Yandex Maps implementation (WebView-based for Expo compatibility)
type YandexMapProps = Props & { apiKey: string };

function YandexMapView(props: YandexMapProps) {
  const {
    initialRegion = DEFAULT_REGION,
    markers = [],
    showsUserLocation,
    onRegionChangeComplete,
    onPress,
    onMarkerPress,
    style,
    apiKey,
  } = props;

  // For Expo Go compatibility, we use a WebView-based Yandex Maps
  // In production build, you could use react-native-yamap
  const [webViewLoaded, setWebViewLoaded] = useState(false);

  // Generate Yandex Maps HTML
  const html = `
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="utf-8">
      <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
      <script src="https://api-maps.yandex.ru/2.1/?apikey=${apiKey}&lang=ru_RU" type="text/javascript"></script>
      <style>
        * { margin: 0; padding: 0; }
        html, body, #map { width: 100%; height: 100%; }
      </style>
    </head>
    <body>
      <div id="map"></div>
      <script type="text/javascript">
        ymaps.ready(function() {
          var map = new ymaps.Map('map', {
            center: [${initialRegion.latitude}, ${initialRegion.longitude}],
            zoom: 14,
            controls: ['zoomControl', 'geolocationControl']
          });

          // Add markers
          ${markers.map(m => `
            var marker_${m.id.replace(/-/g, '_')} = new ymaps.Placemark(
              [${m.coordinate.latitude}, ${m.coordinate.longitude}],
              { balloonContent: '${m.title || ""}' },
              { preset: 'islands#${m.pinColor || "blue"}DotIcon' }
            );
            marker_${m.id.replace(/-/g, '_')}.events.add('click', function() {
              window.ReactNativeWebView.postMessage(JSON.stringify({ type: 'markerPress', id: '${m.id}' }));
            });
            map.geoObjects.add(marker_${m.id.replace(/-/g, '_')});
          `).join('\n')}

          // Map click handler
          map.events.add('click', function(e) {
            var coords = e.get('coords');
            window.ReactNativeWebView.postMessage(JSON.stringify({
              type: 'mapPress',
              coordinate: { latitude: coords[0], longitude: coords[1] }
            }));
          });

          // Region change handler
          map.events.add('boundschange', function() {
            var center = map.getCenter();
            var zoom = map.getZoom();
            window.ReactNativeWebView.postMessage(JSON.stringify({
              type: 'regionChange',
              region: {
                latitude: center[0],
                longitude: center[1],
                latitudeDelta: 180 / Math.pow(2, zoom),
                longitudeDelta: 360 / Math.pow(2, zoom)
              }
            }));
          });

          ${showsUserLocation ? `
          // User location
          ymaps.geolocation.get({ provider: 'browser' }).then(function(result) {
            map.geoObjects.add(result.geoObjects);
          });
          ` : ''}

          window.ReactNativeWebView.postMessage(JSON.stringify({ type: 'mapReady' }));
        });
      </script>
    </body>
    </html>
  `;

  // Dynamic import WebView to avoid SSR issues
  const WebView = require('react-native-webview').WebView;

  const handleMessage = (event: { nativeEvent: { data: string } }) => {
    try {
      const data = JSON.parse(event.nativeEvent.data);
      switch (data.type) {
        case 'mapReady':
          setWebViewLoaded(true);
          break;
        case 'mapPress':
          onPress?.({ nativeEvent: { coordinate: data.coordinate } });
          break;
        case 'markerPress':
          onMarkerPress?.(data.id);
          break;
        case 'regionChange':
          onRegionChangeComplete?.(data.region);
          break;
      }
    } catch (err) {
      console.error('WebView message parse error:', err);
    }
  };

  return (
    <View style={[styles.map, style]}>
      {!webViewLoaded && (
        <View style={[styles.container, styles.center, StyleSheet.absoluteFill]}>
          <ActivityIndicator size="large" color="#FC3F1D" />
          <Text style={styles.loadingText}>Загрузка Яндекс Карт...</Text>
        </View>
      )}
      <WebView
        source={{ html }}
        style={styles.map}
        onMessage={handleMessage}
        javaScriptEnabled={true}
        domStorageEnabled={true}
        startInLoadingState={true}
        scalesPageToFit={true}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  center: {
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: "#f5f5f5",
  },
  map: {
    flex: 1,
    width: "100%",
    height: "100%",
  },
  loadingText: {
    marginTop: 10,
    color: "#666",
    fontSize: 14,
  },
  errorText: {
    color: "#ff3b30",
    fontSize: 14,
    textAlign: "center",
  },
});

MapView.displayName = "MapView";

export default MapView;
export { Marker } from "react-native-maps";
