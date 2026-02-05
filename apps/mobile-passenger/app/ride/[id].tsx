/**
 * Ride detail — bids list, accept bid, payment selection
 */
import { useEffect, useState } from "react";
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Alert,
  TouchableOpacity,
  RefreshControl,
  Switch,
  Linking,
  TextInput,
} from "react-native";
import { useLocalSearchParams, useRouter } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import { Button, Card, Badge } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import { DriverTrackingMap } from "../../components/DriverTrackingMap";
import {
  getRide,
  listBids,
  acceptBid,
  updateRideStatus,
  createPayment,
  listPaymentMethods,
  getAvailableProviders,
  confirmCashPayment,
  validatePromo,
  applyPromo,
  PAYMENT_PROVIDERS,
  type Ride,
  type Bid,
  type PaymentMethod,
  type PromoResult,
} from "../../lib/api";

const statusLabel: Record<string, string> = {
  requested: "Ожидает ставок",
  bidding: "Ставки",
  matched: "Водитель найден",
  in_progress: "В пути",
  completed: "Завершена",
  cancelled: "Отменена",
};

export default function RideDetailScreen() {
  const params = useLocalSearchParams<{ id: string }>();
  const id = Array.isArray(params.id) ? params.id[0] : params.id;
  const { token, userId } = useAuth();
  const router = useRouter();
  const [ride, setRide] = useState<Ride | null>(null);
  const [bids, setBids] = useState<Bid[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  // Payment state
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethod[]>([]);
  const [providers, setProviders] = useState<string[]>([]);
  const [selectedPaymentType, setSelectedPaymentType] = useState<"cash" | "card">("cash");
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethod | null>(null);
  const [selectedProvider, setSelectedProvider] = useState<string>("tinkoff");
  const [saveCard, setSaveCard] = useState(true);
  const [paying, setPaying] = useState(false);

  // Promo state
  const [promoCode, setPromoCode] = useState("");
  const [promoResult, setPromoResult] = useState<PromoResult | null>(null);
  const [promoLoading, setPromoLoading] = useState(false);

  const load = async () => {
    if (!token || !id) return;
    try {
      const [r, b, methods, provs] = await Promise.all([
        getRide(token, id),
        listBids(token, id),
        listPaymentMethods(token).catch(() => []),
        getAvailableProviders(token).catch(() => ["cash"]),
      ]);
      setRide(r);
      setBids(b);
      setPaymentMethods(methods);
      setProviders(provs);

      // Set default method if available
      const defaultMethod = methods.find((m) => m.is_default);
      if (defaultMethod) {
        setSelectedMethod(defaultMethod);
        setSelectedPaymentType("card");
        setSelectedProvider(defaultMethod.provider);
      }
    } catch {
      setRide(null);
      setBids([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    load();
  }, [token, id]);

  const onRefresh = () => {
    setRefreshing(true);
    load();
  };

  const handleAcceptBid = (bidId: string) => {
    if (!token || !id) return;
    Alert.alert("Принять ставку?", "Подтвердите выбор водителя", [
      { text: "Отмена", style: "cancel" },
      {
        text: "Принять",
        onPress: async () => {
          try {
            const updated = await acceptBid(token, id, bidId);
            setRide(updated);
            setBids([]);
          } catch (e) {
            Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось принять ставку");
          }
        },
      },
    ]);
  };

  const handleStatus = (status: "in_progress" | "completed" | "cancelled") => {
    if (!token || !id) return;
    const labels = { in_progress: "Начать поездку", completed: "Завершить", cancelled: "Отменить" };
    Alert.alert(labels[status], "Подтвердить?", [
      { text: "Нет", style: "cancel" },
      {
        text: "Да",
        onPress: async () => {
          try {
            const updated = await updateRideStatus(token, id, status);
            setRide(updated);
          } catch (e) {
            Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось обновить статус");
          }
        },
      },
    ]);
  };

  const handleApplyPromo = async () => {
    if (!token || !id || !ride?.price || !promoCode.trim()) return;
    setPromoLoading(true);
    try {
      const result = await validatePromo(token, promoCode.trim(), ride.price);
      setPromoResult(result);
      if (!result.valid) {
        Alert.alert("Ошибка", result.error ?? "Промокод недействителен");
      }
    } catch (e) {
      Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось проверить промокод");
    } finally {
      setPromoLoading(false);
    }
  };

  const handleRemovePromo = () => {
    setPromoCode("");
    setPromoResult(null);
  };

  // Calculate final price with promo
  const finalPrice = promoResult?.valid ? promoResult.final_price : ride?.price ?? 0;

  const handlePayment = async () => {
    if (!token || !id || !ride?.price) return;
    setPaying(true);
    try {
      // Apply promo if valid
      if (promoResult?.valid && promoCode.trim()) {
        await applyPromo(token, promoCode.trim(), id, ride.price);
      }

      if (selectedPaymentType === "cash") {
        // Cash payment — just mark confirmed (driver receives cash)
        await confirmCashPayment(token, id);
        Alert.alert("Успешно", "Оплата наличными подтверждена");
        load();
      } else {
        // Card payment
        const intent = await createPayment(token, {
          ride_id: id,
          amount: finalPrice,
          method: "card",
          provider: selectedMethod ? selectedMethod.provider : selectedProvider,
          payment_method_id: selectedMethod?.id,
          save_card: !selectedMethod && saveCard,
        });

        if (intent.requires_action && intent.confirm_url) {
          // Open browser for 3DS or redirect
          const result = await WebBrowser.openBrowserAsync(intent.confirm_url, {
            dismissButtonStyle: "close",
          });
          // After redirect, refresh
          load();
        } else if (intent.status === "completed") {
          Alert.alert("Успешно", "Оплата прошла успешно");
          load();
        } else {
          Alert.alert("Ожидание", "Платёж обрабатывается, обновите страницу");
          load();
        }
      }
    } catch (e) {
      Alert.alert("Ошибка оплаты", e instanceof Error ? e.message : "Не удалось провести оплату");
    } finally {
      setPaying(false);
    }
  };

  if (loading || !ride) {
    return (
      <View style={styles.center}>
        <Text style={styles.muted}>{loading ? "Загрузка..." : "Поездка не найдена"}</Text>
      </View>
    );
  }

  const canAcceptBids =
    (ride.status === "requested" || ride.status === "bidding") && bids.length > 0;
  const isMatchedOrActive =
    ride.status === "matched" || ride.status === "in_progress";

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
    >
      <Card style={styles.card}>
        <View style={styles.row}>
          <Text style={styles.label}>Статус</Text>
          <Badge
            label={statusLabel[ride.status] ?? ride.status}
            variant={
              ride.status === "completed"
                ? "success"
                : ride.status === "cancelled"
                  ? "error"
                  : "default"
            }
          />
        </View>
        <Text style={styles.route}>
          Откуда: {ride.from.address ?? `${ride.from.lat}, ${ride.from.lng}`}
        </Text>
        <Text style={styles.route}>
          Куда: {ride.to.address ?? `${ride.to.lat}, ${ride.to.lng}`}
        </Text>
        {ride.price != null ? (
          <Text style={styles.price}>Цена: {ride.price} ₽</Text>
        ) : null}
      </Card>

      {canAcceptBids ? (
        <Card style={styles.card}>
          <Text style={styles.sectionTitle}>Ставки водителей</Text>
          {bids
            .filter((b) => b.status === "pending")
            .map((bid) => (
              <View key={bid.id} style={styles.bidRow}>
                <Text style={styles.bidPrice}>{bid.price} ₽</Text>
                <Button
                  title="Принять"
                  onPress={() => handleAcceptBid(bid.id)}
                  variant="primary"
                  style={styles.bidBtn}
                />
              </View>
            ))}
        </Card>
      ) : null}

      {isMatchedOrActive ? (
        <>
          {/* Driver tracking map */}
          {userId && (
            <Card style={styles.trackingCard}>
              <DriverTrackingMap
                rideId={id}
                passengerId={userId}
                isActive={ride.status === "matched" || ride.status === "in_progress"}
                from={ride.from}
                to={ride.to}
              />
            </Card>
          )}

          {/* Payment selection */}
          <Card style={styles.card}>
            <Text style={styles.sectionTitle}>Способ оплаты</Text>

            {/* Payment type toggle */}
            <View style={styles.paymentTypeRow}>
              <TouchableOpacity
                style={[
                  styles.paymentTypeBtn,
                  selectedPaymentType === "cash" && styles.paymentTypeBtnActive,
                ]}
                onPress={() => {
                  setSelectedPaymentType("cash");
                  setSelectedMethod(null);
                }}
              >
                <Text
                  style={[
                    styles.paymentTypeBtnText,
                    selectedPaymentType === "cash" && styles.paymentTypeBtnTextActive,
                  ]}
                >
                  💵 Наличные
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.paymentTypeBtn,
                  selectedPaymentType === "card" && styles.paymentTypeBtnActive,
                ]}
                onPress={() => setSelectedPaymentType("card")}
              >
                <Text
                  style={[
                    styles.paymentTypeBtnText,
                    selectedPaymentType === "card" && styles.paymentTypeBtnTextActive,
                  ]}
                >
                  💳 Карта
                </Text>
              </TouchableOpacity>
            </View>

            {selectedPaymentType === "card" && (
              <>
                {/* Saved cards */}
                {paymentMethods.length > 0 && (
                  <View style={styles.savedCardsSection}>
                    <Text style={styles.subsectionTitle}>Сохранённые карты</Text>
                    {paymentMethods.map((pm) => (
                      <TouchableOpacity
                        key={pm.id}
                        style={[
                          styles.savedCardRow,
                          selectedMethod?.id === pm.id && styles.savedCardRowActive,
                        ]}
                        onPress={() => {
                          setSelectedMethod(pm);
                          setSelectedProvider(pm.provider);
                        }}
                      >
                        <Text style={styles.savedCardText}>
                          {pm.brand} •••• {pm.last4}
                        </Text>
                        {pm.is_default && <Badge label="Основная" variant="primary" />}
                        {selectedMethod?.id === pm.id && <Text style={styles.checkmark}>✓</Text>}
                      </TouchableOpacity>
                    ))}
                  </View>
                )}

                {/* New card / provider selection */}
                {!selectedMethod && (
                  <View style={styles.newCardSection}>
                    <Text style={styles.subsectionTitle}>Новая карта</Text>
                    <View style={styles.providerRow}>
                      {providers
                        .filter((p) => p !== "cash")
                        .map((p) => (
                          <TouchableOpacity
                            key={p}
                            style={[
                              styles.providerBtn,
                              selectedProvider === p && styles.providerBtnActive,
                            ]}
                            onPress={() => setSelectedProvider(p)}
                          >
                            <Text
                              style={[
                                styles.providerBtnText,
                                selectedProvider === p && styles.providerBtnTextActive,
                              ]}
                            >
                              {PAYMENT_PROVIDERS[p as keyof typeof PAYMENT_PROVIDERS] || p}
                            </Text>
                          </TouchableOpacity>
                        ))}
                    </View>

                    <View style={styles.saveCardRow}>
                      <Text style={styles.saveCardText}>Сохранить карту</Text>
                      <Switch
                        value={saveCard}
                        onValueChange={setSaveCard}
                        trackColor={{ true: "#2563eb" }}
                      />
                    </View>
                  </View>
                )}

                {/* Use saved or new */}
                {paymentMethods.length > 0 && (
                  <TouchableOpacity
                    style={styles.useNewBtn}
                    onPress={() => setSelectedMethod(selectedMethod ? null : paymentMethods[0])}
                  >
                    <Text style={styles.useNewBtnText}>
                      {selectedMethod ? "Использовать новую карту" : "Выбрать из сохранённых"}
                    </Text>
                  </TouchableOpacity>
                )}
              </>
            )}
          </Card>

          {/* Promo code */}
          <Card style={styles.card}>
            <Text style={styles.sectionTitle}>🎁 Промокод</Text>
            {promoResult?.valid ? (
              <View style={styles.promoApplied}>
                <View style={styles.promoAppliedInfo}>
                  <Text style={styles.promoAppliedCode}>{promoCode.toUpperCase()}</Text>
                  <Text style={styles.promoAppliedDiscount}>
                    Скидка: {promoResult.discount.toFixed(0)} ₽
                  </Text>
                </View>
                <TouchableOpacity onPress={handleRemovePromo} style={styles.promoRemoveBtn}>
                  <Text style={styles.promoRemoveBtnText}>✕</Text>
                </TouchableOpacity>
              </View>
            ) : (
              <View style={styles.promoInputRow}>
                <TextInput
                  style={styles.promoInput}
                  placeholder="Введите промокод"
                  placeholderTextColor="#94a3b8"
                  value={promoCode}
                  onChangeText={setPromoCode}
                  autoCapitalize="characters"
                />
                <TouchableOpacity
                  style={[styles.promoApplyBtn, !promoCode.trim() && styles.promoApplyBtnDisabled]}
                  onPress={handleApplyPromo}
                  disabled={!promoCode.trim() || promoLoading}
                >
                  <Text style={styles.promoApplyBtnText}>
                    {promoLoading ? "..." : "Применить"}
                  </Text>
                </TouchableOpacity>
              </View>
            )}
            {promoResult?.valid && (
              <View style={styles.promoSummary}>
                <View style={styles.promoSummaryRow}>
                  <Text style={styles.promoSummaryLabel}>Стоимость поездки:</Text>
                  <Text style={styles.promoSummaryValue}>{ride.price} ₽</Text>
                </View>
                <View style={styles.promoSummaryRow}>
                  <Text style={styles.promoSummaryLabel}>Скидка:</Text>
                  <Text style={styles.promoSummaryDiscount}>−{promoResult.discount.toFixed(0)} ₽</Text>
                </View>
                <View style={[styles.promoSummaryRow, styles.promoSummaryTotal]}>
                  <Text style={styles.promoSummaryTotalLabel}>Итого:</Text>
                  <Text style={styles.promoSummaryTotalValue}>{promoResult.final_price.toFixed(0)} ₽</Text>
                </View>
              </View>
            )}
          </Card>

          {/* Chat button */}
          <Card style={styles.card}>
            <TouchableOpacity
              style={styles.chatButton}
              onPress={() => router.push(`/chat/${id}`)}
            >
              <Text style={styles.chatButtonIcon}>💬</Text>
              <Text style={styles.chatButtonText}>Чат с водителем</Text>
              <Text style={styles.chatButtonArrow}>→</Text>
            </TouchableOpacity>
          </Card>

          {/* Actions */}
          <Card style={styles.card}>
            <Text style={styles.sectionTitle}>Действия</Text>
            {ride.status === "matched" ? (
              <Button
                title="Начать поездку"
                onPress={() => handleStatus("in_progress")}
                variant="primary"
              />
            ) : null}
            {ride.status === "in_progress" ? (
              <>
                <Button
                  title={paying ? "Обработка..." : `Оплатить ${finalPrice.toFixed(0)} ₽`}
                  onPress={handlePayment}
                  variant="primary"
                  disabled={paying}
                />
                <Button
                  title="Завершить без оплаты"
                  onPress={() => handleStatus("completed")}
                  variant="outline"
                  style={styles.actionBtn}
                />
              </>
            ) : null}
            <Button
              title="Отменить"
              onPress={() => handleStatus("cancelled")}
              variant="outline"
              style={styles.actionBtn}
            />
          </Card>
        </>
      ) : null}

      {ride.status === "requested" || ride.status === "bidding" ? (
        <Text style={styles.hint}>Ожидайте ставки от водителей. Обновите страницу.</Text>
      ) : null}

      {ride.status === "completed" ? (
        <Card style={styles.card}>
          <View style={styles.completedHeader}>
            <Text style={styles.completedIcon}>✅</Text>
            <Text style={styles.completedTitle}>Поездка завершена</Text>
          </View>
          <TouchableOpacity
            style={styles.rateButton}
            onPress={() => router.push(`/rate/${id}`)}
          >
            <Text style={styles.rateButtonIcon}>⭐</Text>
            <Text style={styles.rateButtonText}>Оценить водителя</Text>
            <Text style={styles.rateButtonArrow}>→</Text>
          </TouchableOpacity>
        </Card>
      ) : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f8fafc" },
  content: { padding: 16, paddingBottom: 32 },
  trackingCard: { marginBottom: 16, height: 300, padding: 0, overflow: "hidden" },
  center: { flex: 1, justifyContent: "center", alignItems: "center", padding: 24 },
  muted: { color: "#64748b" },
  card: { marginBottom: 16 },
  row: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 8 },
  label: { fontSize: 12, color: "#64748b" },
  route: { fontSize: 14, color: "#0f172a", marginBottom: 4 },
  price: { fontSize: 16, fontWeight: "600", marginTop: 8 },
  sectionTitle: { fontSize: 14, fontWeight: "600", marginBottom: 12, color: "#0f172a" },
  bidRow: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", marginBottom: 12 },
  bidPrice: { fontSize: 18, fontWeight: "600" },
  bidBtn: { minWidth: 100 },
  actionBtn: { marginTop: 8 },
  hint: { fontSize: 12, color: "#64748b", textAlign: "center", marginTop: 8 },
  // Payment styles
  paymentTypeRow: { flexDirection: "row", gap: 8, marginBottom: 16 },
  paymentTypeBtn: { flex: 1, paddingVertical: 12, borderRadius: 8, backgroundColor: "#f1f5f9", alignItems: "center" },
  paymentTypeBtnActive: { backgroundColor: "#2563eb" },
  paymentTypeBtnText: { fontSize: 14, fontWeight: "500", color: "#64748b" },
  paymentTypeBtnTextActive: { color: "#fff" },
  savedCardsSection: { marginBottom: 12 },
  subsectionTitle: { fontSize: 12, fontWeight: "500", color: "#64748b", marginBottom: 8 },
  savedCardRow: { flexDirection: "row", alignItems: "center", padding: 12, borderRadius: 8, backgroundColor: "#f8fafc", marginBottom: 8, borderWidth: 1, borderColor: "#e2e8f0" },
  savedCardRowActive: { borderColor: "#2563eb", backgroundColor: "#eff6ff" },
  savedCardText: { flex: 1, fontSize: 14, fontWeight: "500" },
  checkmark: { color: "#2563eb", fontSize: 18, marginLeft: 8 },
  newCardSection: { marginBottom: 12 },
  providerRow: { flexDirection: "row", flexWrap: "wrap", gap: 8, marginBottom: 12 },
  providerBtn: { paddingHorizontal: 12, paddingVertical: 8, borderRadius: 6, backgroundColor: "#f1f5f9", borderWidth: 1, borderColor: "#e2e8f0" },
  providerBtnActive: { backgroundColor: "#dbeafe", borderColor: "#2563eb" },
  providerBtnText: { fontSize: 13, color: "#64748b" },
  providerBtnTextActive: { color: "#2563eb", fontWeight: "500" },
  saveCardRow: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", paddingVertical: 8 },
  saveCardText: { fontSize: 14, color: "#0f172a" },
  useNewBtn: { alignItems: "center", paddingVertical: 8 },
  useNewBtnText: { color: "#2563eb", fontSize: 13, fontWeight: "500" },
  // Chat button
  chatButton: { flexDirection: "row", alignItems: "center", paddingVertical: 12 },
  chatButtonIcon: { fontSize: 24, marginRight: 12 },
  chatButtonText: { flex: 1, fontSize: 16, fontWeight: "600", color: "#0f172a" },
  chatButtonArrow: { fontSize: 18, color: "#64748b" },
  // Completed ride
  completedHeader: { flexDirection: "row", alignItems: "center", marginBottom: 16 },
  completedIcon: { fontSize: 28, marginRight: 12 },
  completedTitle: { fontSize: 18, fontWeight: "700", color: "#16a34a" },
  rateButton: { flexDirection: "row", alignItems: "center", paddingVertical: 14, backgroundColor: "#fef3c7", borderRadius: 12, paddingHorizontal: 16 },
  rateButtonIcon: { fontSize: 24, marginRight: 12 },
  rateButtonText: { flex: 1, fontSize: 16, fontWeight: "600", color: "#92400e" },
  rateButtonArrow: { fontSize: 18, color: "#92400e" },
  // Promo styles
  promoInputRow: { flexDirection: "row", gap: 8 },
  promoInput: { flex: 1, height: 44, backgroundColor: "#f8fafc", borderRadius: 8, paddingHorizontal: 14, fontSize: 14, borderWidth: 1, borderColor: "#e2e8f0", textTransform: "uppercase" },
  promoApplyBtn: { paddingHorizontal: 16, height: 44, backgroundColor: "#2563eb", borderRadius: 8, justifyContent: "center", alignItems: "center" },
  promoApplyBtnDisabled: { backgroundColor: "#94a3b8" },
  promoApplyBtnText: { color: "#fff", fontWeight: "600", fontSize: 14 },
  promoApplied: { flexDirection: "row", alignItems: "center", backgroundColor: "#dcfce7", borderRadius: 8, padding: 12 },
  promoAppliedInfo: { flex: 1 },
  promoAppliedCode: { fontSize: 16, fontWeight: "700", color: "#16a34a" },
  promoAppliedDiscount: { fontSize: 13, color: "#15803d", marginTop: 2 },
  promoRemoveBtn: { width: 28, height: 28, borderRadius: 14, backgroundColor: "#fff", justifyContent: "center", alignItems: "center" },
  promoRemoveBtnText: { fontSize: 14, color: "#64748b" },
  promoSummary: { marginTop: 12, paddingTop: 12, borderTopWidth: 1, borderTopColor: "#e2e8f0" },
  promoSummaryRow: { flexDirection: "row", justifyContent: "space-between", marginBottom: 4 },
  promoSummaryLabel: { fontSize: 14, color: "#64748b" },
  promoSummaryValue: { fontSize: 14, color: "#0f172a" },
  promoSummaryDiscount: { fontSize: 14, color: "#16a34a", fontWeight: "500" },
  promoSummaryTotal: { marginTop: 8, paddingTop: 8, borderTopWidth: 1, borderTopColor: "#e2e8f0" },
  promoSummaryTotalLabel: { fontSize: 16, fontWeight: "600", color: "#0f172a" },
  promoSummaryTotalValue: { fontSize: 18, fontWeight: "700", color: "#2563eb" },
});
