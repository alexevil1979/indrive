/**
 * Payment Methods Management Screen
 * List saved cards, set default, delete
 */
import { useState, useEffect, useCallback } from "react";
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Alert,
  TouchableOpacity,
  ActivityIndicator,
  RefreshControl,
} from "react-native";
import { Button, Card, Badge } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import {
  listPaymentMethods,
  deletePaymentMethod,
  setDefaultPaymentMethod,
  getAvailableProviders,
  PAYMENT_PROVIDERS,
  type PaymentMethod,
} from "../../lib/api";

export default function PaymentScreen() {
  const { token } = useAuth();
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [providers, setProviders] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const loadData = useCallback(async () => {
    if (!token) return;
    try {
      const [methodsData, providersData] = await Promise.all([
        listPaymentMethods(token).catch(() => []),
        getAvailableProviders(token).catch(() => ["cash"]),
      ]);
      setMethods(methodsData);
      setProviders(providersData);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [token]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const onRefresh = () => {
    setRefreshing(true);
    loadData();
  };

  const handleSetDefault = async (methodId: string) => {
    try {
      await setDefaultPaymentMethod(token!, methodId);
      setMethods((prev) =>
        prev.map((m) => ({
          ...m,
          is_default: m.id === methodId,
        }))
      );
    } catch (e) {
      Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось установить метод по умолчанию");
    }
  };

  const handleDelete = async (methodId: string) => {
    Alert.alert("Удалить карту?", "Карта будет удалена из вашего аккаунта", [
      { text: "Отмена", style: "cancel" },
      {
        text: "Удалить",
        style: "destructive",
        onPress: async () => {
          try {
            await deletePaymentMethod(token!, methodId);
            setMethods((prev) => prev.filter((m) => m.id !== methodId));
          } catch (e) {
            Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось удалить карту");
          }
        },
      },
    ]);
  };

  const getCardIcon = (brand: string) => {
    switch (brand.toLowerCase()) {
      case "visa":
        return "💳";
      case "mastercard":
        return "💳";
      case "mir":
        return "🏦";
      default:
        return "💳";
    }
  };

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color="#2563eb" />
        <Text style={styles.muted}>Загрузка...</Text>
      </View>
    );
  }

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
    >
      <Text style={styles.title}>Способы оплаты</Text>

      {/* Available providers */}
      <Card style={styles.card}>
        <Text style={styles.sectionTitle}>Доступные способы</Text>
        <View style={styles.providersRow}>
          {providers.map((p) => (
            <View key={p} style={styles.providerChip}>
              <Text style={styles.providerText}>
                {PAYMENT_PROVIDERS[p as keyof typeof PAYMENT_PROVIDERS] || p}
              </Text>
            </View>
          ))}
        </View>
      </Card>

      {/* Cash info */}
      <Card style={styles.card}>
        <View style={styles.methodRow}>
          <View style={styles.methodIcon}>
            <Text style={styles.iconText}>💵</Text>
          </View>
          <View style={styles.methodInfo}>
            <Text style={styles.methodTitle}>Наличные</Text>
            <Text style={styles.methodSubtitle}>Оплата водителю</Text>
          </View>
          <Badge label="Всегда" variant="success" />
        </View>
      </Card>

      {/* Saved cards */}
      <Text style={styles.sectionHeader}>Сохранённые карты</Text>

      {methods.length === 0 ? (
        <Card style={styles.card}>
          <Text style={styles.emptyText}>
            У вас пока нет сохранённых карт. Карта будет сохранена автоматически при первой онлайн-оплате.
          </Text>
        </Card>
      ) : (
        methods.map((method) => (
          <Card key={method.id} style={styles.card}>
            <View style={styles.methodRow}>
              <View style={styles.methodIcon}>
                <Text style={styles.iconText}>{getCardIcon(method.brand)}</Text>
              </View>
              <View style={styles.methodInfo}>
                <Text style={styles.methodTitle}>
                  {method.brand} •••• {method.last4}
                </Text>
                <Text style={styles.methodSubtitle}>
                  {String(method.exp_month).padStart(2, "0")}/{method.exp_year}
                  {" · "}
                  {PAYMENT_PROVIDERS[method.provider as keyof typeof PAYMENT_PROVIDERS] || method.provider}
                </Text>
              </View>
              {method.is_default && <Badge label="По умолчанию" variant="primary" />}
            </View>

            <View style={styles.cardActions}>
              {!method.is_default && (
                <TouchableOpacity
                  style={styles.actionBtn}
                  onPress={() => handleSetDefault(method.id)}
                >
                  <Text style={styles.actionBtnText}>Сделать основной</Text>
                </TouchableOpacity>
              )}
              <TouchableOpacity
                style={[styles.actionBtn, styles.deleteActionBtn]}
                onPress={() => handleDelete(method.id)}
              >
                <Text style={styles.deleteActionText}>Удалить</Text>
              </TouchableOpacity>
            </View>
          </Card>
        ))
      )}

      {/* Info */}
      <Card style={[styles.card, styles.infoCard]}>
        <Text style={styles.infoTitle}>ℹ️ Как добавить карту</Text>
        <Text style={styles.infoText}>
          При создании поездки выберите "Оплата картой" и отметьте "Сохранить карту для будущих платежей".
          После успешной оплаты карта появится в этом списке.
        </Text>
      </Card>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#eff6ff" },
  content: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, justifyContent: "center", alignItems: "center" },
  muted: { color: "#64748b", marginTop: 8 },
  title: { fontSize: 24, fontWeight: "bold", color: "#0f172a", marginBottom: 16 },
  card: { marginBottom: 12 },
  sectionTitle: { fontSize: 14, fontWeight: "600", color: "#64748b", marginBottom: 12 },
  sectionHeader: { fontSize: 16, fontWeight: "600", color: "#0f172a", marginTop: 16, marginBottom: 12 },
  providersRow: { flexDirection: "row", flexWrap: "wrap", gap: 8 },
  providerChip: { backgroundColor: "#dbeafe", paddingHorizontal: 12, paddingVertical: 6, borderRadius: 16 },
  providerText: { color: "#1d4ed8", fontSize: 13, fontWeight: "500" },
  methodRow: { flexDirection: "row", alignItems: "center" },
  methodIcon: { width: 48, height: 48, borderRadius: 24, backgroundColor: "#f1f5f9", justifyContent: "center", alignItems: "center", marginRight: 12 },
  iconText: { fontSize: 24 },
  methodInfo: { flex: 1 },
  methodTitle: { fontSize: 15, fontWeight: "600", color: "#0f172a" },
  methodSubtitle: { fontSize: 13, color: "#64748b", marginTop: 2 },
  cardActions: { flexDirection: "row", marginTop: 12, gap: 8 },
  actionBtn: { flex: 1, paddingVertical: 8, paddingHorizontal: 12, borderRadius: 8, backgroundColor: "#f1f5f9", alignItems: "center" },
  actionBtnText: { color: "#2563eb", fontSize: 13, fontWeight: "500" },
  deleteActionBtn: { backgroundColor: "#fef2f2" },
  deleteActionText: { color: "#dc2626", fontSize: 13, fontWeight: "500" },
  emptyText: { color: "#64748b", fontSize: 14, textAlign: "center", paddingVertical: 8 },
  infoCard: { backgroundColor: "#f0f9ff", borderColor: "#bae6fd" },
  infoTitle: { fontSize: 14, fontWeight: "600", color: "#0369a1", marginBottom: 8 },
  infoText: { fontSize: 13, color: "#0c4a6e", lineHeight: 20 },
});
