/**
 * Driver Verification Screen
 * Start verification, upload documents, check status
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
import * as ImagePicker from "expo-image-picker";
import { Button, Input, Card, Badge } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import {
  startVerification,
  getVerificationStatus,
  uploadDocument,
  listDocuments,
  deleteDocument,
  DOC_TYPES,
  VERIFICATION_STATUSES,
  type DriverVerification,
  type DriverDocument,
  type StartVerificationInput,
} from "../../lib/api";

export default function VerificationScreen() {
  const { token } = useAuth();
  const [verification, setVerification] = useState<DriverVerification | null>(null);
  const [documents, setDocuments] = useState<DriverDocument[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [uploading, setUploading] = useState<string | null>(null);

  // Form state
  const [licenseNumber, setLicenseNumber] = useState("");
  const [vehicleModel, setVehicleModel] = useState("");
  const [vehiclePlate, setVehiclePlate] = useState("");
  const [vehicleYear, setVehicleYear] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const loadData = useCallback(async () => {
    if (!token) return;
    try {
      const [verif, docs] = await Promise.all([
        getVerificationStatus(token),
        listDocuments(token).catch(() => []),
      ]);
      setVerification(verif);
      setDocuments(docs);
    } catch {
      setVerification(null);
      setDocuments([]);
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

  const handleStartVerification = async () => {
    if (!token || !licenseNumber.trim()) {
      Alert.alert("Ошибка", "Введите номер водительского удостоверения");
      return;
    }

    const input: StartVerificationInput = {
      license_number: licenseNumber.trim(),
    };
    if (vehicleModel.trim()) input.vehicle_model = vehicleModel.trim();
    if (vehiclePlate.trim()) input.vehicle_plate = vehiclePlate.trim();
    if (vehicleYear.trim()) {
      const year = parseInt(vehicleYear, 10);
      if (!isNaN(year)) input.vehicle_year = year;
    }

    setSubmitting(true);
    try {
      const result = await startVerification(token, input);
      setVerification(result);
      Alert.alert("Готово", "Заявка на верификацию создана. Загрузите необходимые документы.");
    } catch (e) {
      Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось создать заявку");
    } finally {
      setSubmitting(false);
    }
  };

  const pickAndUpload = async (docType: string) => {
    const permission = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!permission.granted) {
      Alert.alert("Ошибка", "Нужен доступ к галерее");
      return;
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ImagePicker.MediaTypeOptions.Images,
      quality: 0.8,
      allowsEditing: false,
    });

    if (result.canceled || !result.assets?.[0]) return;

    const asset = result.assets[0];
    const fileName = asset.fileName || `${docType}_${Date.now()}.jpg`;
    const mimeType = asset.mimeType || "image/jpeg";

    setUploading(docType);
    try {
      const doc = await uploadDocument(token!, docType, {
        uri: asset.uri,
        name: fileName,
        type: mimeType,
      });
      setDocuments((prev) => [...prev.filter((d) => d.doc_type !== docType), doc]);
      Alert.alert("Готово", "Документ загружен");
    } catch (e) {
      Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось загрузить");
    } finally {
      setUploading(null);
    }
  };

  const handleDelete = async (docId: string) => {
    Alert.alert("Удалить документ?", "Это действие нельзя отменить", [
      { text: "Отмена", style: "cancel" },
      {
        text: "Удалить",
        style: "destructive",
        onPress: async () => {
          try {
            await deleteDocument(token!, docId);
            setDocuments((prev) => prev.filter((d) => d.id !== docId));
          } catch (e) {
            Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось удалить");
          }
        },
      },
    ]);
  };

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color="#16a34a" />
        <Text style={styles.muted}>Загрузка...</Text>
      </View>
    );
  }

  // No verification started
  if (!verification) {
    return (
      <ScrollView style={styles.container} contentContainerStyle={styles.content}>
        <Text style={styles.title}>Верификация водителя</Text>
        <Text style={styles.subtitle}>
          Заполните данные и загрузите документы для подтверждения статуса водителя.
        </Text>

        <Card style={styles.card}>
          <Input
            label="Номер водительского удостоверения *"
            placeholder="Серия и номер"
            value={licenseNumber}
            onChangeText={setLicenseNumber}
          />
          <Input
            label="Модель автомобиля"
            placeholder="Toyota Camry"
            value={vehicleModel}
            onChangeText={setVehicleModel}
          />
          <Input
            label="Госномер"
            placeholder="А123БВ777"
            value={vehiclePlate}
            onChangeText={setVehiclePlate}
            autoCapitalize="characters"
          />
          <Input
            label="Год выпуска"
            placeholder="2020"
            value={vehicleYear}
            onChangeText={setVehicleYear}
            keyboardType="numeric"
          />
          <Button
            title={submitting ? "Отправка..." : "Подать заявку"}
            onPress={handleStartVerification}
            variant="primary"
            disabled={submitting}
          />
        </Card>
      </ScrollView>
    );
  }

  // Verification exists
  const statusLabel = VERIFICATION_STATUSES[verification.status as keyof typeof VERIFICATION_STATUSES] || verification.status;
  const statusVariant = verification.status === "approved" ? "success" : verification.status === "rejected" ? "danger" : "warning";

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
    >
      <View style={styles.header}>
        <Text style={styles.title}>Верификация</Text>
        <Badge label={statusLabel} variant={statusVariant} />
      </View>

      {verification.reject_reason && (
        <Card style={[styles.card, styles.rejectCard]}>
          <Text style={styles.rejectTitle}>Причина отклонения</Text>
          <Text style={styles.rejectText}>{verification.reject_reason}</Text>
        </Card>
      )}

      <Card style={styles.card}>
        <Text style={styles.sectionTitle}>Данные заявки</Text>
        <View style={styles.row}>
          <Text style={styles.label}>Права:</Text>
          <Text style={styles.value}>{verification.license_number || "—"}</Text>
        </View>
        {verification.vehicle_model && (
          <View style={styles.row}>
            <Text style={styles.label}>Авто:</Text>
            <Text style={styles.value}>
              {verification.vehicle_model} {verification.vehicle_plate}
            </Text>
          </View>
        )}
        {verification.vehicle_year && (
          <View style={styles.row}>
            <Text style={styles.label}>Год:</Text>
            <Text style={styles.value}>{verification.vehicle_year}</Text>
          </View>
        )}
      </Card>

      {/* Documents */}
      <Card style={styles.card}>
        <Text style={styles.sectionTitle}>Документы</Text>
        {verification.status === "pending" && (
          <Text style={styles.hint}>Загрузите документы для проверки</Text>
        )}

        {Object.entries(DOC_TYPES).map(([type, label]) => {
          const doc = documents.find((d) => d.doc_type === type);
          const isUploading = uploading === type;

          return (
            <View key={type} style={styles.docRow}>
              <View style={styles.docInfo}>
                <Text style={styles.docLabel}>{label}</Text>
                {doc ? (
                  <View style={styles.docStatus}>
                    <Badge
                      label={doc.status === "approved" ? "✓" : doc.status === "rejected" ? "✗" : "…"}
                      variant={doc.status === "approved" ? "success" : doc.status === "rejected" ? "danger" : "default"}
                    />
                    <Text style={styles.docName} numberOfLines={1}>
                      {doc.file_name}
                    </Text>
                  </View>
                ) : (
                  <Text style={styles.docMissing}>Не загружен</Text>
                )}
              </View>
              <View style={styles.docActions}>
                {isUploading ? (
                  <ActivityIndicator size="small" color="#16a34a" />
                ) : verification.status === "pending" ? (
                  <>
                    <TouchableOpacity
                      style={styles.uploadBtn}
                      onPress={() => pickAndUpload(type)}
                    >
                      <Text style={styles.uploadBtnText}>{doc ? "⟳" : "+"}</Text>
                    </TouchableOpacity>
                    {doc && (
                      <TouchableOpacity
                        style={styles.deleteBtn}
                        onPress={() => handleDelete(doc.id)}
                      >
                        <Text style={styles.deleteBtnText}>×</Text>
                      </TouchableOpacity>
                    )}
                  </>
                ) : null}
              </View>
            </View>
          );
        })}
      </Card>

      {verification.status === "approved" && (
        <Card style={[styles.card, styles.successCard]}>
          <Text style={styles.successText}>
            🎉 Вы верифицированы! Теперь вы можете принимать заказы.
          </Text>
        </Card>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f0fdf4" },
  content: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, justifyContent: "center", alignItems: "center" },
  muted: { color: "#64748b", marginTop: 8 },
  title: { fontSize: 24, fontWeight: "bold", color: "#0f172a", marginBottom: 8 },
  subtitle: { fontSize: 14, color: "#64748b", marginBottom: 16 },
  header: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 16 },
  card: { marginBottom: 16 },
  sectionTitle: { fontSize: 16, fontWeight: "600", color: "#0f172a", marginBottom: 12 },
  row: { flexDirection: "row", marginBottom: 8 },
  label: { fontSize: 14, color: "#64748b", width: 80 },
  value: { fontSize: 14, color: "#0f172a", flex: 1 },
  hint: { fontSize: 12, color: "#64748b", marginBottom: 12 },
  rejectCard: { backgroundColor: "#fef2f2", borderColor: "#fecaca" },
  rejectTitle: { fontSize: 14, fontWeight: "600", color: "#dc2626", marginBottom: 4 },
  rejectText: { fontSize: 14, color: "#b91c1c" },
  successCard: { backgroundColor: "#dcfce7", borderColor: "#86efac" },
  successText: { fontSize: 14, color: "#166534", textAlign: "center" },
  docRow: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", paddingVertical: 12, borderBottomWidth: 1, borderBottomColor: "#e2e8f0" },
  docInfo: { flex: 1 },
  docLabel: { fontSize: 14, fontWeight: "500", color: "#0f172a", marginBottom: 2 },
  docStatus: { flexDirection: "row", alignItems: "center", gap: 8 },
  docName: { fontSize: 12, color: "#64748b", maxWidth: 150 },
  docMissing: { fontSize: 12, color: "#94a3b8" },
  docActions: { flexDirection: "row", gap: 8 },
  uploadBtn: { width: 32, height: 32, borderRadius: 16, backgroundColor: "#16a34a", justifyContent: "center", alignItems: "center" },
  uploadBtnText: { color: "#fff", fontSize: 18, fontWeight: "bold" },
  deleteBtn: { width: 32, height: 32, borderRadius: 16, backgroundColor: "#dc2626", justifyContent: "center", alignItems: "center" },
  deleteBtnText: { color: "#fff", fontSize: 20, fontWeight: "bold" },
});
