"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import {
  fetchPromos,
  createPromo,
  updatePromo,
  deletePromo,
  formatDate,
  formatCurrency,
  type Promo,
  type CreatePromoInput,
} from "@/lib/api";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";

const PROMO_TYPE_LABELS: Record<string, string> = {
  percent: "Процент",
  fixed: "Фикс. сумма",
};

export default function PromosPage() {
  const router = useRouter();
  const [promos, setPromos] = useState<Promo[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const [activeOnly, setActiveOnly] = useState(false);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [editingPromo, setEditingPromo] = useState<Promo | null>(null);
  const [formData, setFormData] = useState<Partial<CreatePromoInput>>({
    code: "",
    description: "",
    type: "percent",
    value: 10,
    min_order_value: 0,
    max_discount: 0,
    usage_limit: 0,
    per_user_limit: 0,
  });
  const [saving, setSaving] = useState(false);
  const limit = 20;

  const loadPromos = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchPromos(limit, page * limit, activeOnly);
      setPromos(data.promos ?? []);
      setTotal(data.total ?? 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load promos");
    } finally {
      setLoading(false);
    }
  }, [page, activeOnly]);

  useEffect(() => {
    loadPromos();
  }, [loadPromos]);

  const handleCreate = async () => {
    if (!formData.code || !formData.value) return;
    setSaving(true);
    try {
      await createPromo(formData as CreatePromoInput);
      setShowCreateDialog(false);
      setFormData({
        code: "",
        description: "",
        type: "percent",
        value: 10,
        min_order_value: 0,
        max_discount: 0,
        usage_limit: 0,
        per_user_limit: 0,
      });
      loadPromos();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Ошибка создания");
    } finally {
      setSaving(false);
    }
  };

  const handleEdit = (promo: Promo) => {
    setEditingPromo(promo);
    setFormData({
      description: promo.description,
      type: promo.type,
      value: promo.value,
      min_order_value: promo.min_order_value,
      max_discount: promo.max_discount,
      usage_limit: promo.usage_limit,
      per_user_limit: promo.per_user_limit,
    });
    setShowEditDialog(true);
  };

  const handleUpdate = async () => {
    if (!editingPromo) return;
    setSaving(true);
    try {
      await updatePromo(editingPromo.id, formData);
      setShowEditDialog(false);
      setEditingPromo(null);
      loadPromos();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Ошибка обновления");
    } finally {
      setSaving(false);
    }
  };

  const handleToggleActive = async (promo: Promo) => {
    try {
      await updatePromo(promo.id, { is_active: !promo.is_active });
      loadPromos();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Ошибка");
    }
  };

  const handleDelete = async (promo: Promo) => {
    if (!confirm(`Удалить промокод "${promo.code}"?`)) return;
    try {
      await deletePromo(promo.id);
      loadPromos();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Ошибка удаления");
    }
  };

  const totalPages = Math.ceil(total / limit);

  const isExpired = (promo: Promo) => {
    if (!promo.expires_at || promo.expires_at.startsWith("0001")) return false;
    return new Date(promo.expires_at) < new Date();
  };

  const isUsageLimitReached = (promo: Promo) => {
    return promo.usage_limit > 0 && promo.usage_count >= promo.usage_limit;
  };

  const getPromoStatus = (promo: Promo) => {
    if (!promo.is_active) return { label: "Неактивен", color: "bg-gray-100 text-gray-600" };
    if (isExpired(promo)) return { label: "Истёк", color: "bg-red-100 text-red-600" };
    if (isUsageLimitReached(promo)) return { label: "Исчерпан", color: "bg-orange-100 text-orange-600" };
    return { label: "Активен", color: "bg-green-100 text-green-600" };
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">Промокоды</h1>
          <p className="text-gray-500 mt-1">Всего: {total}</p>
        </div>
        <div className="flex items-center gap-4">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={activeOnly}
              onChange={(e) => {
                setActiveOnly(e.target.checked);
                setPage(0);
              }}
              className="rounded border-gray-300"
            />
            Только активные
          </label>
          <button
            onClick={() => setShowCreateDialog(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
          >
            + Создать промокод
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-50 text-red-600 rounded-lg">{error}</div>
      )}

      {loading ? (
        <div className="text-center py-12">Загрузка...</div>
      ) : promos.length === 0 ? (
        <div className="text-center py-12 text-gray-500">Промокоды не найдены</div>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Код</TableHead>
              <TableHead>Описание</TableHead>
              <TableHead>Тип</TableHead>
              <TableHead>Скидка</TableHead>
              <TableHead>Мин. заказ</TableHead>
              <TableHead>Использований</TableHead>
              <TableHead>Статус</TableHead>
              <TableHead>Действия</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {promos.map((promo) => {
              const status = getPromoStatus(promo);
              return (
                <TableRow key={promo.id}>
                  <TableCell>
                    <span className="font-mono font-bold text-lg">{promo.code}</span>
                  </TableCell>
                  <TableCell className="max-w-xs truncate">
                    {promo.description || "—"}
                  </TableCell>
                  <TableCell>{PROMO_TYPE_LABELS[promo.type]}</TableCell>
                  <TableCell className="font-medium">
                    {promo.type === "percent"
                      ? `${promo.value}%`
                      : formatCurrency(promo.value)}
                    {promo.max_discount > 0 && promo.type === "percent" && (
                      <span className="text-xs text-gray-500 block">
                        макс. {formatCurrency(promo.max_discount)}
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    {promo.min_order_value > 0
                      ? formatCurrency(promo.min_order_value)
                      : "—"}
                  </TableCell>
                  <TableCell>
                    {promo.usage_count}
                    {promo.usage_limit > 0 && (
                      <span className="text-gray-500"> / {promo.usage_limit}</span>
                    )}
                    {promo.per_user_limit > 0 && (
                      <span className="text-xs text-gray-400 block">
                        ({promo.per_user_limit} на юзера)
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    <span className={`px-2 py-1 text-xs font-medium rounded-full ${status.color}`}>
                      {status.label}
                    </span>
                    {promo.expires_at && !promo.expires_at.startsWith("0001") && (
                      <span className="text-xs text-gray-400 block mt-1">
                        до {formatDate(promo.expires_at)}
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => handleEdit(promo)}
                        className="text-blue-600 hover:text-blue-800 text-sm"
                      >
                        Редактировать
                      </button>
                      <button
                        onClick={() => handleToggleActive(promo)}
                        className={`text-sm ${
                          promo.is_active
                            ? "text-orange-600 hover:text-orange-800"
                            : "text-green-600 hover:text-green-800"
                        }`}
                      >
                        {promo.is_active ? "Деактивировать" : "Активировать"}
                      </button>
                      <button
                        onClick={() => handleDelete(promo)}
                        className="text-red-600 hover:text-red-800 text-sm"
                      >
                        Удалить
                      </button>
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-4 flex justify-center gap-2">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-3 py-1 border rounded disabled:opacity-50"
          >
            Назад
          </button>
          <span className="px-3 py-1">
            {page + 1} / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="px-3 py-1 border rounded disabled:opacity-50"
          >
            Далее
          </button>
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Создать промокод</DialogTitle>
          </DialogHeader>
          <PromoForm
            data={formData}
            onChange={setFormData}
            isNew
          />
          <div className="flex justify-end gap-2 mt-4">
            <button
              onClick={() => setShowCreateDialog(false)}
              className="px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              onClick={handleCreate}
              disabled={saving || !formData.code}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              {saving ? "Создание..." : "Создать"}
            </button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Редактировать: {editingPromo?.code}</DialogTitle>
          </DialogHeader>
          <PromoForm
            data={formData}
            onChange={setFormData}
          />
          <div className="flex justify-end gap-2 mt-4">
            <button
              onClick={() => setShowEditDialog(false)}
              className="px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              onClick={handleUpdate}
              disabled={saving}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              {saving ? "Сохранение..." : "Сохранить"}
            </button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// Promo Form Component
function PromoForm({
  data,
  onChange,
  isNew = false,
}: {
  data: Partial<CreatePromoInput>;
  onChange: (data: Partial<CreatePromoInput>) => void;
  isNew?: boolean;
}) {
  return (
    <div className="space-y-4">
      {isNew && (
        <div>
          <label className="block text-sm font-medium mb-1">Код промокода</label>
          <Input
            value={data.code ?? ""}
            onChange={(e) => onChange({ ...data, code: e.target.value.toUpperCase() })}
            placeholder="WELCOME10"
            className="uppercase"
          />
        </div>
      )}
      <div>
        <label className="block text-sm font-medium mb-1">Описание</label>
        <Input
          value={data.description ?? ""}
          onChange={(e) => onChange({ ...data, description: e.target.value })}
          placeholder="Скидка 10% на первую поездку"
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium mb-1">Тип скидки</label>
          <select
            value={data.type ?? "percent"}
            onChange={(e) => onChange({ ...data, type: e.target.value as "percent" | "fixed" })}
            className="w-full px-3 py-2 border rounded-lg"
          >
            <option value="percent">Процент (%)</option>
            <option value="fixed">Фиксированная сумма (₽)</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Значение {data.type === "percent" ? "(%)" : "(₽)"}
          </label>
          <Input
            type="number"
            value={data.value ?? 0}
            onChange={(e) => onChange({ ...data, value: Number(e.target.value) })}
            min={0}
            max={data.type === "percent" ? 100 : undefined}
          />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium mb-1">Мин. сумма заказа (₽)</label>
          <Input
            type="number"
            value={data.min_order_value ?? 0}
            onChange={(e) => onChange({ ...data, min_order_value: Number(e.target.value) })}
            min={0}
          />
          <p className="text-xs text-gray-500 mt-1">0 = без ограничений</p>
        </div>
        {data.type === "percent" && (
          <div>
            <label className="block text-sm font-medium mb-1">Макс. скидка (₽)</label>
            <Input
              type="number"
              value={data.max_discount ?? 0}
              onChange={(e) => onChange({ ...data, max_discount: Number(e.target.value) })}
              min={0}
            />
            <p className="text-xs text-gray-500 mt-1">0 = без ограничений</p>
          </div>
        )}
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium mb-1">Лимит использований</label>
          <Input
            type="number"
            value={data.usage_limit ?? 0}
            onChange={(e) => onChange({ ...data, usage_limit: Number(e.target.value) })}
            min={0}
          />
          <p className="text-xs text-gray-500 mt-1">0 = безлимитно</p>
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">На 1 пользователя</label>
          <Input
            type="number"
            value={data.per_user_limit ?? 0}
            onChange={(e) => onChange({ ...data, per_user_limit: Number(e.target.value) })}
            min={0}
          />
          <p className="text-xs text-gray-500 mt-1">0 = безлимитно</p>
        </div>
      </div>
      {isNew && (
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Начало действия</label>
            <Input
              type="datetime-local"
              value={data.starts_at ? data.starts_at.slice(0, 16) : ""}
              onChange={(e) => onChange({ ...data, starts_at: e.target.value ? new Date(e.target.value).toISOString() : undefined })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Окончание действия</label>
            <Input
              type="datetime-local"
              value={data.expires_at ? data.expires_at.slice(0, 16) : ""}
              onChange={(e) => onChange({ ...data, expires_at: e.target.value ? new Date(e.target.value).toISOString() : undefined })}
            />
          </div>
        </div>
      )}
    </div>
  );
}
