/**
 * Payments Admin Page
 * Lists all payments, allows refund
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
  fetchPayments,
  formatDate,
  formatCurrency,
  getStatusColor,
  getStatusLabel,
  type Payment,
} from "@/lib/api";
import Link from "next/link";

export const dynamic = "force-dynamic";

const providerLabels: Record<string, string> = {
  cash: "Наличные",
  tinkoff: "Tinkoff",
  yoomoney: "YooMoney",
  sber: "Сбербанк",
};

export default async function PaymentsPage() {
  let payments: Payment[] = [];
  try {
    payments = await fetchPayments();
  } catch {
    payments = [];
  }

  // Stats
  const completed = payments.filter((p) => p.status === "completed");
  const pending = payments.filter((p) => p.status === "pending");
  const failed = payments.filter((p) => p.status === "failed");
  const refunded = payments.filter((p) => p.status === "refunded");

  const totalAmount = completed.reduce((sum, p) => sum + p.amount, 0);
  const refundedAmount = refunded.reduce((sum, p) => sum + p.amount, 0);

  // Group by provider
  const byProvider: Record<string, number> = {};
  completed.forEach((p) => {
    byProvider[p.provider] = (byProvider[p.provider] || 0) + p.amount;
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Платежи</h1>
        <Badge variant="outline">{payments.length} платежей</Badge>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-green-600">
              Оплачено
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatCurrency(totalAmount)}</p>
            <p className="text-xs text-muted-foreground">
              {completed.length} платежей
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-yellow-600">
              Ожидают оплаты
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{pending.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-red-600">
              Ошибки
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{failed.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-purple-600">
              Возвраты
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{formatCurrency(refundedAmount)}</p>
            <p className="text-xs text-muted-foreground">
              {refunded.length} возвратов
            </p>
          </CardContent>
        </Card>
      </div>

      {/* By Provider */}
      <div className="grid gap-4 md:grid-cols-4">
        {Object.entries(byProvider).map(([provider, amount]) => (
          <Card key={provider}>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">
                {providerLabels[provider] || provider}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xl font-bold">{formatCurrency(amount)}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Table */}
      <Card>
        <CardHeader>
          <CardTitle>История платежей</CardTitle>
        </CardHeader>
        <CardContent>
          {payments.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">
              Нет платежей.
              <br />
              <span className="text-xs">
                Убедитесь, что ADMIN_JWT установлен и Payment API доступен.
              </span>
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Поездка</TableHead>
                  <TableHead>Сумма</TableHead>
                  <TableHead>Метод</TableHead>
                  <TableHead>Провайдер</TableHead>
                  <TableHead>Статус</TableHead>
                  <TableHead>Дата</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {payments.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell className="font-mono text-xs">
                      {p.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {p.ride_id.slice(0, 8)}...
                    </TableCell>
                    <TableCell className="font-medium">
                      {formatCurrency(p.amount, p.currency)}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {p.method === "cash" ? "Нал" : "Карта"}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {providerLabels[p.provider] || p.provider}
                    </TableCell>
                    <TableCell>
                      <Badge className={getStatusColor(p.status)}>
                        {getStatusLabel(p.status)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(p.created_at)}
                    </TableCell>
                    <TableCell>
                      <Link
                        href={`/payments/${p.id}`}
                        className="text-primary hover:underline text-sm"
                      >
                        Детали
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
