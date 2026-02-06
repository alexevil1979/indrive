"use client";

import { useEffect, useState } from "react";
import {
  fetchSettings,
  updateSettings,
  AppSettings,
  MapProvider,
} from "@/lib/api";

export default function SettingsPage() {
  const [settings, setSettings] = useState<AppSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  // Form state
  const [mapProvider, setMapProvider] = useState<MapProvider>("google");
  const [googleApiKey, setGoogleApiKey] = useState("");
  const [yandexApiKey, setYandexApiKey] = useState("");
  const [language, setLanguage] = useState("ru");
  const [currency, setCurrency] = useState("RUB");

  useEffect(() => {
    loadSettings();
  }, []);

  async function loadSettings() {
    try {
      setLoading(true);
      setError(null);
      const data = await fetchSettings();
      setSettings(data);
      setMapProvider(data.map_provider);
      setGoogleApiKey(data.google_maps_api_key ?? "");
      setYandexApiKey(data.yandex_maps_api_key ?? "");
      setLanguage(data.default_language);
      setCurrency(data.default_currency);
    } catch (err) {
      setError("Не удалось загрузить настройки");
      console.error(err);
    } finally {
      setLoading(false);
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      setSaving(true);
      setError(null);
      setSuccess(false);

      await updateSettings({
        map_provider: mapProvider,
        google_maps_api_key: googleApiKey || undefined,
        yandex_maps_api_key: yandexApiKey || undefined,
        default_language: language,
        default_currency: currency,
      });

      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
      await loadSettings();
    } catch (err) {
      setError("Не удалось сохранить настройки");
      console.error(err);
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-6">Настройки</h1>
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="h-10 bg-gray-200 rounded w-1/2 mb-4"></div>
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="h-10 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Настройки</h1>

      {error && (
        <div className="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      {success && (
        <div className="mb-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
          Настройки сохранены!
        </div>
      )}

      <form onSubmit={handleSubmit} className="max-w-2xl">
        {/* Map Provider Selection */}
        <section className="mb-8 bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            🗺️ Провайдер карт
          </h2>
          <p className="text-sm text-gray-600 mb-4">
            Выберите провайдера карт для мобильных приложений. После изменения
            пользователям нужно перезапустить приложение.
          </p>

          <div className="space-y-3">
            <label className="flex items-center p-4 border rounded-lg cursor-pointer hover:bg-gray-50 transition">
              <input
                type="radio"
                name="mapProvider"
                value="google"
                checked={mapProvider === "google"}
                onChange={(e) => setMapProvider(e.target.value as MapProvider)}
                className="w-4 h-4 text-blue-600"
              />
              <div className="ml-3">
                <span className="font-medium">Google Maps</span>
                <p className="text-sm text-gray-500">
                  Глобальное покрытие, детальные карты, требуется API ключ
                </p>
              </div>
            </label>

            <label className="flex items-center p-4 border rounded-lg cursor-pointer hover:bg-gray-50 transition">
              <input
                type="radio"
                name="mapProvider"
                value="yandex"
                checked={mapProvider === "yandex"}
                onChange={(e) => setMapProvider(e.target.value as MapProvider)}
                className="w-4 h-4 text-blue-600"
              />
              <div className="ml-3">
                <span className="font-medium">Яндекс Карты</span>
                <p className="text-sm text-gray-500">
                  Лучшее покрытие России и СНГ, требуется API ключ
                </p>
              </div>
            </label>
          </div>
        </section>

        {/* API Keys */}
        <section className="mb-8 bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            🔑 API ключи
          </h2>
          <p className="text-sm text-gray-600 mb-4">
            Введите API ключи для картографических сервисов. Ключ активного
            провайдера будет передаваться мобильным приложениям.
          </p>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Google Maps API Key
                {mapProvider === "google" && (
                  <span className="ml-2 text-xs text-green-600">(активен)</span>
                )}
              </label>
              <input
                type="password"
                value={googleApiKey}
                onChange={(e) => setGoogleApiKey(e.target.value)}
                placeholder="AIza..."
                className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500">
                Получить ключ:{" "}
                <a
                  href="https://console.cloud.google.com/apis/credentials"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:underline"
                >
                  Google Cloud Console
                </a>
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Яндекс Карты API Key
                {mapProvider === "yandex" && (
                  <span className="ml-2 text-xs text-green-600">(активен)</span>
                )}
              </label>
              <input
                type="password"
                value={yandexApiKey}
                onChange={(e) => setYandexApiKey(e.target.value)}
                placeholder="..."
                className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500">
                Получить ключ:{" "}
                <a
                  href="https://developer.tech.yandex.ru/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:underline"
                >
                  Yandex Developer Console
                </a>
              </p>
            </div>
          </div>
        </section>

        {/* Localization */}
        <section className="mb-8 bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            🌍 Локализация
          </h2>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Язык по умолчанию
              </label>
              <select
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
                className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="ru">Русский</option>
                <option value="en">English</option>
                <option value="kk">Қазақша</option>
                <option value="uz">O&apos;zbek</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Валюта по умолчанию
              </label>
              <select
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
                className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="RUB">RUB (₽)</option>
                <option value="KZT">KZT (₸)</option>
                <option value="UZS">UZS</option>
                <option value="USD">USD ($)</option>
              </select>
            </div>
          </div>
        </section>

        {/* Last update info */}
        {settings?.updated_at && (
          <p className="text-sm text-gray-500 mb-4">
            Последнее обновление:{" "}
            {new Date(settings.updated_at).toLocaleString("ru-RU")}
          </p>
        )}

        {/* Submit button */}
        <button
          type="submit"
          disabled={saving}
          className="px-6 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {saving ? "Сохранение..." : "Сохранить настройки"}
        </button>
      </form>
    </div>
  );
}
