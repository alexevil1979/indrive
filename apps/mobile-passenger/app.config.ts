import { ExpoConfig, ConfigContext } from "expo/config";

const GOOGLE_MAPS_API_KEY = process.env.GOOGLE_MAPS_API_KEY ?? "";

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: "RideHail Passenger",
  slug: "ridehail-passenger",
  version: "0.1.0",
  orientation: "portrait",
  icon: "./assets/icon.png",
  scheme: "ridehail",
  userInterfaceStyle: "automatic",
  splash: {
    image: "./assets/splash.png",
    resizeMode: "contain",
    backgroundColor: "#2563eb",
  },
  assetBundlePatterns: ["**/*"],
  ios: {
    supportsTablet: true,
    bundleIdentifier: "com.ridehail.passenger",
    config: {
      googleMapsApiKey: GOOGLE_MAPS_API_KEY,
    },
  },
  android: {
    adaptiveIcon: {
      foregroundImage: "./assets/adaptive-icon.png",
      backgroundColor: "#2563eb",
    },
    package: "com.ridehail.passenger",
    config: {
      googleMaps: {
        apiKey: GOOGLE_MAPS_API_KEY,
      },
    },
  },
  plugins: ["expo-router"],
});
