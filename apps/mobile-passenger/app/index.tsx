/**
 * Login screen — RideHail v0.1 Passenger
 */
import { useState } from "react";
import { View, Text, StyleSheet, KeyboardAvoidingView, Platform, Alert } from "react-native";
import { useRouter } from "expo-router";
import { Button, Input, Card } from "@ridehail/ui";
import { useAuth } from "../context/AuthContext";

export default function LoginScreen() {
  const { login } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleLogin = async () => {
    setError("");
    if (!email.trim() || !password) {
      setError("Введите email и пароль");
      return;
    }
    setLoading(true);
    try {
      await login(email.trim(), password);
      router.replace("/(tabs)");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка входа");
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
        <Text style={styles.title}>RideHail v0.1</Text>
        <Text style={styles.subtitle}>Пассажир</Text>
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
          label="Пароль"
          placeholder="••••••••"
          value={password}
          onChangeText={setPassword}
          secureTextEntry
        />
        {error ? <Text style={styles.errorText}>{error}</Text> : null}
        <Button
          title={loading ? "Вход..." : "Войти"}
          onPress={handleLogin}
          variant="primary"
          disabled={loading}
        />
        <Button
          title="Регистрация"
          onPress={() => router.push("/register")}
          variant="outline"
          style={styles.registerBtn}
        />
      </Card>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f8fafc",
    padding: 24,
    paddingTop: 60,
  },
  header: { marginBottom: 24 },
  title: { fontSize: 28, fontWeight: "700", color: "#0f172a" },
  subtitle: { fontSize: 16, color: "#64748b", marginTop: 4 },
  card: { marginBottom: 24 },
  errorText: { color: "#dc2626", fontSize: 12, marginBottom: 8 },
  registerBtn: { marginTop: 12 },
});
