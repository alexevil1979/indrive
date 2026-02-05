/**
 * Verification Detail Page
 * View documents, approve/reject verification
 */
import { notFound } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  fetchVerification,
  formatDate,
  getStatusColor,
  getStatusLabel,
} from "@/lib/api";
import { VerificationActions } from "./actions";
import Link from "next/link";

export const dynamic = "force-dynamic";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function VerificationDetailPage({ params }: PageProps) {
  const { id } = await params;
  const verification = await fetchVerification(id);

  if (!verification) {
    notFound();
  }

  const docTypeLabels: Record<string, string> = {
    license: "Водительское удостоверение",
    passport: "Паспорт",
    vehicle_reg: "СТС",
    insurance: "ОСАГО",
    photo: "Фото водителя",
    vehicle_photo: "Фото автомобиля",
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <Link
            href="/verifications"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            ← Назад к списку
          </Link>
          <h1 className="text-2xl font-bold mt-2">Заявка на верификацию</h1>
        </div>
        <Badge className={getStatusColor(verification.status)}>
          {getStatusLabel(verification.status)}
        </Badge>
      </div>

      {/* Info Card */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Информация о водителе</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <p className="text-sm text-muted-foreground">ID пользователя</p>
              <p className="font-mono text-sm">{verification.user_id}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Номер прав</p>
              <p className="font-medium">
                {verification.license_number || "Не указан"}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Дата подачи</p>
              <p>{formatDate(verification.submitted_at)}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Автомобиль</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <p className="text-sm text-muted-foreground">Модель</p>
              <p className="font-medium">
                {verification.vehicle_model || "Не указана"}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Госномер</p>
              <p className="font-medium">
                {verification.vehicle_plate || "Не указан"}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Год выпуска</p>
              <p>{verification.vehicle_year || "Не указан"}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Documents */}
      <Card>
        <CardHeader>
          <CardTitle>Загруженные документы</CardTitle>
        </CardHeader>
        <CardContent>
          {!verification.documents || verification.documents.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">
              Документы не загружены
            </p>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {verification.documents.map((doc) => (
                <div
                  key={doc.id}
                  className="border rounded-lg p-4 space-y-2"
                >
                  <div className="flex items-center justify-between">
                    <p className="font-medium text-sm">
                      {docTypeLabels[doc.doc_type] || doc.doc_type}
                    </p>
                    <Badge className={getStatusColor(doc.status)}>
                      {getStatusLabel(doc.status)}
                    </Badge>
                  </div>
                  <p className="text-xs text-muted-foreground truncate">
                    {doc.file_name}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {(doc.file_size / 1024).toFixed(1)} KB •{" "}
                    {formatDate(doc.uploaded_at)}
                  </p>
                  {doc.storage_url && (
                    <a
                      href={doc.storage_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary text-sm hover:underline"
                    >
                      Открыть документ →
                    </a>
                  )}
                  {doc.reject_reason && (
                    <p className="text-xs text-red-600 mt-2">
                      Причина отказа: {doc.reject_reason}
                    </p>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Actions */}
      {verification.status === "pending" && (
        <Card>
          <CardHeader>
            <CardTitle>Действия</CardTitle>
          </CardHeader>
          <CardContent>
            <VerificationActions verificationId={verification.id} />
          </CardContent>
        </Card>
      )}

      {/* Review Info */}
      {verification.reviewed_at && (
        <Card>
          <CardHeader>
            <CardTitle>Результат проверки</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div>
              <p className="text-sm text-muted-foreground">Дата проверки</p>
              <p>{formatDate(verification.reviewed_at)}</p>
            </div>
            {verification.reviewed_by && (
              <div>
                <p className="text-sm text-muted-foreground">Проверил</p>
                <p className="font-mono text-sm">{verification.reviewed_by}</p>
              </div>
            )}
            {verification.reject_reason && (
              <div>
                <p className="text-sm text-muted-foreground">Причина отказа</p>
                <p className="text-red-600">{verification.reject_reason}</p>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
