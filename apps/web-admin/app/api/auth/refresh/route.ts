import { NextRequest, NextResponse } from "next/server";

const INTERNAL_AUTH = process.env.INTERNAL_AUTH_URL ?? "http://localhost:9080";

export async function POST(req: NextRequest) {
  try {
    const refreshToken = req.cookies.get("refresh_token")?.value;

    if (!refreshToken) {
      return NextResponse.json(
        { error: "Refresh token отсутствует" },
        { status: 401 }
      );
    }

    const res = await fetch(`${INTERNAL_AUTH}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    const data = await res.json();

    if (!res.ok) {
      // Clear invalid cookies
      const response = NextResponse.json(
        { error: data.error ?? "Сессия истекла" },
        { status: 401 }
      );
      response.cookies.delete("access_token");
      response.cookies.delete("refresh_token");
      return response;
    }

    const { access_token, refresh_token, expires_in } = data;

    const response = NextResponse.json({ success: true });

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
      maxAge: 7 * 24 * 3600,
    });

    return response;
  } catch (error) {
    console.error("Refresh error:", error);
    return NextResponse.json(
      { error: "Ошибка обновления токена" },
      { status: 500 }
    );
  }
}
