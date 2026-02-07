import { ExpoConfig, ConfigContext } from "expo/config";

const GOOGLE_MAPS_API_KEY = process.env.GOOGLE_MAPS_API_KEY ?? "";

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: "RideHail Driver",
  slug: "ridehail-driver",
  version: "0.1.0",
  orientation: "portrait",
  scheme: "ridehail-driver",
  userInterfaceStyle: "automatic",
  newArchEnabled: false,
  assetBundlePatterns: ["**/*"],
  ios: {
    supportsTablet: true,
    bundleIdentifier: "com.ridehail.driver",
    config: {
      googleMapsApiKey: GOOGLE_MAPS_API_KEY,
    },
    infoPlist: {
      NSLocationWhenInUseUsageDescription:
        "Приложению нужен доступ к геолокации для отображения вашей позиции на карте",
      NSLocationAlwaysAndWhenInUseUsageDescription:
        "Приложению нужен доступ к геолокации для передачи позиции пассажиру во время поездки",
      NSCameraUsageDescription:
        "Приложению нужен доступ к камере для загрузки документов верификации",
      NSPhotoLibraryUsageDescription:
        "Приложению нужен доступ к фото для загрузки документов верификации",
    },
  },
  android: {
    package: "com.ridehail.driver",
    config: {
      googleMaps: {
        apiKey: GOOGLE_MAPS_API_KEY,
      },
    },
    permissions: [
      "ACCESS_COARSE_LOCATION",
      "ACCESS_FINE_LOCATION",
      "ACCESS_BACKGROUND_LOCATION",
      "CAMERA",
      "READ_EXTERNAL_STORAGE",
      "WRITE_EXTERNAL_STORAGE",
      "VIBRATE",
      "RECEIVE_BOOT_COMPLETED",
    ],
  },
  plugins: [
    "expo-router",
    [
      "expo-location",
      {
        locationAlwaysAndWhenInUsePermission:
          "Приложению нужен доступ к геолокации для передачи позиции пассажиру",
        isAndroidBackgroundLocationEnabled: true,
      },
    ],
    [
      "expo-image-picker",
      {
        photosPermission:
          "Приложению нужен доступ к фото для загрузки документов",
        cameraPermission:
          "Приложению нужен доступ к камере для загрузки документов",
      },
    ],
    "expo-notifications",
  ],
});
