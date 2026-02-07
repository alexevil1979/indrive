import { NextRequest, NextResponse } from "next/server";

const INTERNAL_AUTH = process.env.INTERNAL_AUTH_URL ?? "http://localhost:9080";

export async function POST(req: NextRequest) {
  try {
    const { email, password } = await req.json();

    if (!email || !password) {
      return NextResponse.json(
        { error: "Email и пароль обязательны" },
        { status: 400 }
      );
    }

    const res = await fetch(`${INTERNAL_AUTH}/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    const data = await res.json();

    if (!res.ok) {
      return NextResponse.json(
        { error: data.error ?? "Неверные учётные данные" },
        { status: res.status }
      );
    }

    const { access_token, refresh_token, expires_in } = data;

    if (!access_token) {
      return NextResponse.json(
        { error: "Не удалось получить токен" },
        { status: 500 }
      );
    }

    const response = NextResponse.json({
      success: true,
      role: data.role ?? "",
    });

    // Set httpOnly cookies
    response.cookies.set("access_token", access_token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: expires_in ?? 86400,
    });

    response.cookies.set("refresh_token", refresh_token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: 7 * 24 * 3600, // 7 days
    });

    return response;
  } catch (error) {
    console.error("Login error:", error);
    return NextResponse.json(
      { error: "Ошибка авторизации" },
      { status: 500 }
    );
  }
}
