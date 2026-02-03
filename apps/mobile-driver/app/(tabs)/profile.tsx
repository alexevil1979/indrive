/**
 * Driver profile — verification stub (doc upload later)
 */
import { useState } from "react";
import { View, Text, StyleSheet, ScrollView, Alert } from "react-native";
import { Button, Input, Card, Badge } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import { createDriverProfile, getProfile } from "../../lib/api";
import { useEffect } from "react";

type ProfileData = {
  profile?: { user_id: string; display_name?: string; phone?: string };
  driver_profile?: { user_id: string; verified: boolean; license_number?: string };
};

export default function ProfileScreen() {
  const { token } = useAuth();
  const [profile, setProfile] = useState<ProfileData | null>(null);
  const [licenseNumber, setLicenseNumber] = useState("");
  const [loading, setLoading] = useState(false);
  const [loadingProfile, setLoadingProfile] = useState(true);

  const loadProfile = async () => {
    if (!token) return;
    try {
      const data = (await getProfile(token)) as ProfileData;
      setProfile(data);
      if (data?.driver_profile?.license_number) {
        setLicenseNumber(data.driver_profile.license_number);
      }
    } catch {
      setProfile(null);
    } finally {
      setLoadingProfile(false);
    }
  };

  useEffect(() => {
    loadProfile();
  }, [token]);

  const handleCreateDriverProfile = async () => {
    if (!token || !licenseNumber.trim()) {
      Alert.alert("Ошибка", "Введите номер прав");
      return;
    }
    setLoading(true);
    try {
      await createDriverProfile(token, licenseNumber.trim());
      Alert.alert("Готово", "Профиль водителя создан. Верификация — заглушка (загрузка документов в следующей версии).");
      loadProfile();
    } catch (e) {
      Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось создать профиль");
    } finally {
      setLoading(false);
    }
  };

  if (loadingProfile) {
    return (
      <View style={styles.center}>
        <Text style={styles.muted}>Загрузка...</Text>
      </View>
    );
  }

  const hasDriverProfile = !!profile?.driver_profile;
  const verified = profile?.driver_profile?.verified ?? false;

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Card style={styles.card}>
        <Text style={styles.sectionTitle}>Профиль водителя</Text>
        {hasDriverProfile ? (
          <>
            <View style={styles.row}>
              <Text style={styles.label}>Статус</Text>
              <Badge
                label={verified ? "Верифицирован" : "На проверке"}
                variant={verified ? "success" : "warning"}
              />
            </View>
            {profile?.driver_profile?.license_number ? (
              <Text style={styles.text}>Права: {profile.driver_profile.license_number}</Text>
            ) : null}
            <Text style={styles.hint}>Загрузка документов (права, фото) — в следующей версии.</Text>
          </>
        ) : (
          <>
            <Input
              label="Номер водительского удостоверения"
              placeholder="Серия и номер"
              value={licenseNumber}
              onChangeText={setLicenseNumber}
            />
            <Button
              title={loading ? "Создание..." : "Создать профиль водителя"}
              onPress={handleCreateDriverProfile}
              variant="primary"
              disabled={loading}
            />
            <Text style={styles.hint}>После создания профиля станет доступна верификация (загрузка документов).</Text>
          </>
        )}
      </Card>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f0fdf4" },
  content: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, justifyContent: "center", alignItems: "center", padding: 24 },
  muted: { color: "#64748b" },
  card: { marginBottom: 16 },
  sectionTitle: { fontSize: 16, fontWeight: "600", marginBottom: 12, color: "#0f172a" },
  row: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 8 },
  label: { fontSize: 14, color: "#64748b" },
  text: { fontSize: 14, color: "#0f172a", marginBottom: 8 },
  hint: { fontSize: 12, color: "#64748b", marginTop: 8 },
});
