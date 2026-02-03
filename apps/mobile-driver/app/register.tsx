/**
 * Registration — driver
 */
import { useState } from "react";
import { View, Text, StyleSheet, KeyboardAvoidingView, Platform, Alert } from "react-native";
import { useRouter } from "expo-router";
import { Button, Input, Card } from "@ridehail/ui";
import { useAuth } from "../context/AuthContext";

export default function RegisterScreen() {
  const { register } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleRegister = async () => {
    setError("");
    if (!email.trim() || !password) {
      setError("Введите email и пароль");
      return;
    }
    if (password.length < 8) {
      setError("Пароль не менее 8 символов");
      return;
    }
    setLoading(true);
    try {
      await register(email.trim(), password);
      router.replace("/(tabs)");
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Ошибка регистрации";
      setError(msg);
      Alert.alert("Ошибка", msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === "ios" ? "padding" : "height"}
    >
      <View style={styles.header}>
        <Text style={styles.title}>Регистрация водителя</Text>
        <Text style={styles.subtitle}>Водитель</Text>
      </View>

      <Card style={styles.card}>
        <Input
          label="Email"
          placeholder="email@example.com"
          value={email}
          onChangeText={setEmail}
          keyboardType="email-address"
          autoCapitalize="none"
        />
        <Input
          label="Пароль (мин. 8 символов)"
          placeholder="••••••••"
          value={password}
          onChangeText={setPassword}
          secureTextEntry
        />
        {error ? <Text style={styles.errorText}>{error}</Text> : null}
        <Button
          title={loading ? "Регистрация..." : "Зарегистрироваться"}
          onPress={handleRegister}
          variant="primary"
          disabled={loading}
        />
        <Button
          title="Назад к входу"
          onPress={() => router.back()}
          variant="outline"
          style={styles.backBtn}
        />
      </Card>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f0fdf4",
    padding: 24,
    paddingTop: 60,
  },
  header: { marginBottom: 24 },
  title: { fontSize: 24, fontWeight: "700", color: "#0f172a" },
  subtitle: { fontSize: 14, color: "#16a34a", marginTop: 4 },
  card: { marginBottom: 24 },
  errorText: { color: "#dc2626", fontSize: 12, marginBottom: 8 },
  backBtn: { marginTop: 12 },
});
