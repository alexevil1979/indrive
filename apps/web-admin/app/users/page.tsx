/**
 * Users Admin Page
 * Lists all users with roles and verification status
 */
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { fetchUsers, formatDate, type User } from "@/lib/api";

export const dynamic = "force-dynamic";

const roleLabels: Record<string, string> = {
  passenger: "Пассажир",
  driver: "Водитель",
  admin: "Администратор",
};

const roleBadgeColors: Record<string, string> = {
  passenger: "bg-blue-100 text-blue-800",
  driver: "bg-green-100 text-green-800",
  admin: "bg-purple-100 text-purple-800",
};

export default async function UsersPage() {
  let users: User[] = [];
  try {
    users = await fetchUsers();
  } catch {
    users = [];
  }

  // Stats
  const passengers = users.filter((u) => u.role === "passenger");
  const drivers = users.filter((u) => u.role === "driver");
  const admins = users.filter((u) => u.role === "admin");
  const verified = users.filter((u) => u.verified);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Пользователи</h1>
        <Badge variant="outline">{users.length} пользователей</Badge>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-blue-600">
              Пассажиры
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{passengers.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-green-600">
              Водители
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{drivers.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-purple-600">
              Администраторы
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{admins.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-emerald-600">
              Верифицированы
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{verified.length}</p>
          </CardContent>
        </Card>
      </div>

      {/* Table */}
      <Card>
        <CardHeader>
          <CardTitle>Список пользователей</CardTitle>
        </CardHeader>
        <CardContent>
          {users.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">
              Нет пользователей.
              <br />
              <span className="text-xs">
                Убедитесь, что ADMIN_JWT установлен и Auth API имеет endpoint
                /api/v1/admin/users.
              </span>
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Телефон</TableHead>
                  <TableHead>Имя</TableHead>
                  <TableHead>Роль</TableHead>
                  <TableHead>Верификация</TableHead>
                  <TableHead>Регистрация</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell className="font-mono text-xs">
                      {user.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>{user.phone || "—"}</TableCell>
                    <TableCell>{user.name || "—"}</TableCell>
                    <TableCell>
                      <Badge
                        className={
                          roleBadgeColors[user.role] ||
                          "bg-gray-100 text-gray-800"
                        }
                      >
                        {roleLabels[user.role] || user.role}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {user.verified ? (
                        <Badge className="bg-green-100 text-green-800">
                          Да
                        </Badge>
                      ) : (
                        <Badge className="bg-gray-100 text-gray-800">Нет</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(user.created_at)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
