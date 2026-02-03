/**
 * User management — stub (Auth/User service would expose admin list)
 */
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default async function UsersPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Пользователи</h1>

      <Card>
        <CardHeader>
          <CardTitle>Управление пользователями</CardTitle>
          <p className="text-sm text-muted-foreground">
            Список пользователей (пассажиры, водители) загружается из User/Auth сервиса.
            Добавьте в Auth или User endpoint GET /admin/users (только для роли admin) и вызовите его здесь.
          </p>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground py-8 text-center">
            Заглушка: подключите API списка пользователей.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
