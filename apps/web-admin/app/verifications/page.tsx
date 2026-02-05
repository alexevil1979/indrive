/**
 * Driver Verifications Admin Page
 * Lists pending verifications, allows approve/reject
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
import {
  fetchVerifications,
  formatDate,
  getStatusColor,
  getStatusLabel,
  type DriverVerification,
} from "@/lib/api";
import Link from "next/link";

export const dynamic = "force-dynamic";

export default async function VerificationsPage() {
  let verifications: DriverVerification[] = [];
  try {
    verifications = await fetchVerifications();
  } catch {
    verifications = [];
  }

  const pending = verifications.filter((v) => v.status === "pending");
  const approved = verifications.filter((v) => v.status === "approved");
  const rejected = verifications.filter((v) => v.status === "rejected");

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Верификация водителей</h1>
        <Badge variant="outline">{verifications.length} заявок</Badge>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-yellow-600">
              Ожидают проверки
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{pending.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-green-600">
              Одобрено
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{approved.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-red-600">
              Отклонено
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{rejected.length}</p>
          </CardContent>
        </Card>
      </div>

      {/* Table */}
      <Card>
        <CardHeader>
          <CardTitle>Заявки на верификацию</CardTitle>
        </CardHeader>
        <CardContent>
          {verifications.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">
              Нет заявок на верификацию.
              <br />
              <span className="text-xs">
                Убедитесь, что ADMIN_JWT установлен и User API доступен.
              </span>
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Пользователь</TableHead>
                  <TableHead>Права</TableHead>
                  <TableHead>Авто</TableHead>
                  <TableHead>Документы</TableHead>
                  <TableHead>Статус</TableHead>
                  <TableHead>Дата</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {verifications.map((v) => (
                  <TableRow key={v.id}>
                    <TableCell className="font-mono text-xs">
                      {v.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {v.user_id.slice(0, 8)}...
                    </TableCell>
                    <TableCell>{v.license_number || "—"}</TableCell>
                    <TableCell>
                      {v.vehicle_model ? (
                        <span>
                          {v.vehicle_model} {v.vehicle_plate}
                        </span>
                      ) : (
                        "—"
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {v.documents?.length ?? 0} док.
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge className={getStatusColor(v.status)}>
                        {getStatusLabel(v.status)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(v.submitted_at)}
                    </TableCell>
                    <TableCell>
                      <Link
                        href={`/verifications/${v.id}`}
                        className="text-primary hover:underline text-sm"
                      >
                        Открыть
                      </Link>
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
