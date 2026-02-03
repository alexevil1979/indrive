/**
 * Auth context — driver app (role=driver)
 */
import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { login as apiLogin, register as apiRegister, type TokenResponse } from "../lib/api";

const TOKEN_KEY = "@ridehail_driver_access_token";

type AuthState = {
  token: string | null;
  userId: string | null;
  isLoading: boolean;
  isSignedIn: boolean;
};

type AuthContextValue = AuthState & {
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  setToken: (token: string | null) => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>({
    token: null,
    userId: null,
    isLoading: true,
    isSignedIn: false,
  });

  const loadStoredToken = useCallback(async () => {
    try {
      const token = await AsyncStorage.getItem(TOKEN_KEY);
      setState((s) => ({
        ...s,
        token,
        isSignedIn: !!token,
        isLoading: false,
      }));
    } catch {
      setState((s) => ({ ...s, isLoading: false }));
    }
  }, []);

  useEffect(() => {
    loadStoredToken();
  }, [loadStoredToken]);

  const setToken = useCallback((token: string | null) => {
    setState((s) => ({
      ...s,
      token,
      isSignedIn: !!token,
    }));
    if (token) AsyncStorage.setItem(TOKEN_KEY, token);
    else AsyncStorage.removeItem(TOKEN_KEY);
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const data: TokenResponse = await apiLogin(email, password);
      setToken(data.access_token);
    },
    [setToken]
  );

  const register = useCallback(
    async (email: string, password: string) => {
      const data: TokenResponse = await apiRegister(email, password, "driver");
      setToken(data.access_token);
    },
    [setToken]
  );

  const logout = useCallback(async () => {
    setToken(null);
  }, [setToken]);

  const value: AuthContextValue = {
    ...state,
    login,
    register,
    logout,
    setToken,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
