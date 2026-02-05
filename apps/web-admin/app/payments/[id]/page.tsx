/**
 * Payment Detail Page
 * View payment details, process refund
 */
import { notFound } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  fetchPayment,
  formatDate,
  formatCurrency,
  getStatusColor,
  getStatusLabel,
} from "@/lib/api";
import { RefundActions } from "./actions";
import Link from "next/link";

export const dynamic = "force-dynamic";

interface PageProps {
  params: Promise<{ id: string }>;
}

const providerLabels: Record<string, string> = {
  cash: "Наличные",
  tinkoff: "Tinkoff",
  yoomoney: "YooMoney",
  sber: "Сбербанк",
};

export default async function PaymentDetailPage({ params }: PageProps) {
  const { id } = await params;
  const payment = await fetchPayment(id);

  if (!payment) {
    notFound();
  }

  const canRefund =
    payment.status === "completed" && payment.provider !== "cash";

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <Link
            href="/payments"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            ← Назад к списку
          </Link>
          <h1 className="text-2xl font-bold mt-2">Платёж</h1>
        </div>
        <Badge className={getStatusColor(payment.status)}>
          {getStatusLabel(payment.status)}
        </Badge>
      </div>

      {/* Payment Info */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Основная информация</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <p className="text-sm text-muted-foreground">ID платежа</p>
              <p className="font-mono text-sm">{payment.id}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">ID поездки</p>
              <p className="font-mono text-sm">{payment.ride_id}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">ID пользователя</p>
              <p className="font-mono text-sm">{payment.user_id}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Сумма</p>
              <p className="text-2xl font-bold">
                {formatCurrency(payment.amount, payment.currency)}
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Детали оплаты</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <p className="text-sm text-muted-foreground">Метод оплаты</p>
              <p className="font-medium">
                {payment.method === "cash" ? "Наличные" : "Карта"}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Провайдер</p>
              <p className="font-medium">
                {providerLabels[payment.provider] || payment.provider}
              </p>
            </div>
            {payment.external_id && (
              <div>
                <p className="text-sm text-muted-foreground">
                  ID в платёжной системе
                </p>
                <p className="font-mono text-xs">{payment.external_id}</p>
              </div>
            )}
            {payment.description && (
              <div>
                <p className="text-sm text-muted-foreground">Описание</p>
                <p>{payment.description}</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Dates */}
      <Card>
        <CardHeader>
          <CardTitle>Даты</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-3">
          <div>
            <p className="text-sm text-muted-foreground">Создан</p>
            <p>{formatDate(payment.created_at)}</p>
          </div>
          {payment.paid_at && (
            <div>
              <p className="text-sm text-muted-foreground">Оплачен</p>
              <p>{formatDate(payment.paid_at)}</p>
            </div>
          )}
          {payment.refunded_at && (
            <div>
              <p className="text-sm text-muted-foreground">Возвращён</p>
              <p>{formatDate(payment.refunded_at)}</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Error Info */}
      {payment.fail_reason && (
        <Card className="border-red-200">
          <CardHeader>
            <CardTitle className="text-red-600">Ошибка</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-red-600">{payment.fail_reason}</p>
          </CardContent>
        </Card>
      )}

      {/* Refund Actions */}
      {canRefund && (
        <Card>
          <CardHeader>
            <CardTitle>Возврат средств</CardTitle>
          </CardHeader>
          <CardContent>
            <RefundActions paymentId={payment.id} amount={payment.amount} />
          </CardContent>
        </Card>
      )}

      {/* Confirm URL */}
      {payment.confirm_url && payment.status === "pending" && (
        <Card>
          <CardHeader>
            <CardTitle>Ссылка для оплаты</CardTitle>
          </CardHeader>
          <CardContent>
            <a
              href={payment.confirm_url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary hover:underline break-all"
            >
              {payment.confirm_url}
            </a>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
