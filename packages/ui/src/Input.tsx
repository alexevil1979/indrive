import React from "react";
import { TextInput, View, Text, StyleSheet, type TextInputProps } from "react-native";

export interface InputProps extends Omit<TextInputProps, "style"> {
  label?: string;
  error?: string;
  containerStyle?: object;
  inputStyle?: object;
}

/** Shared Input — label + error (accessible, secure text for password) */
export function Input({
  label,
  error,
  containerStyle,
  inputStyle,
  ...rest
}: InputProps) {
  return (
    <View style={[styles.container, containerStyle]}>
      {label ? <Text style={styles.label}>{label}</Text> : null}
      <TextInput
        style={[styles.input, error && styles.inputError, inputStyle]}
        placeholderTextColor="#94a3b8"
        {...rest}
      />
      {error ? <Text style={styles.error}>{error}</Text> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { marginBottom: 16 },
  label: { fontSize: 14, fontWeight: "500", marginBottom: 4, color: "#334155" },
  input: {
    borderWidth: 1,
    borderColor: "#e2e8f0",
    borderRadius: 8,
    paddingVertical: 12,
    paddingHorizontal: 16,
    fontSize: 16,
    color: "#0f172a",
  },
  inputError: { borderColor: "#dc2626" },
  error: { fontSize: 12, color: "#dc2626", marginTop: 4 },
});
