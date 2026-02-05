"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { refundPayment, formatCurrency } from "@/lib/api";

interface RefundActionsProps {
  paymentId: string;
  amount: number;
}

export function RefundActions({ paymentId, amount }: RefundActionsProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [showDialog, setShowDialog] = useState(false);
  const [refundAmount, setRefundAmount] = useState(amount.toString());
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  async function handleRefund() {
    setError("");
    const parsedAmount = parseFloat(refundAmount);
    
    if (isNaN(parsedAmount) || parsedAmount <= 0) {
      setError("Укажите корректную сумму");
      return;
    }
    
    if (parsedAmount > amount) {
      setError("Сумма возврата не может превышать сумму платежа");
      return;
    }

    setLoading(true);
    const result = await refundPayment(paymentId, parsedAmount, reason);
    setLoading(false);

    if (result.success) {
      setShowDialog(false);
      router.refresh();
    } else {
      setError(result.error ?? "Ошибка при возврате");
    }
  }

  return (
    <>
      <div className="space-y-4">
        <p className="text-sm text-muted-foreground">
          Вы можете оформить полный или частичный возврат средств.
        </p>
        <Button onClick={() => setShowDialog(true)} variant="destructive">
          Оформить возврат
        </Button>
      </div>

      <Dialog open={showDialog} onOpenChange={setShowDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Возврат средств</DialogTitle>
            <DialogDescription>
              Максимальная сумма возврата: {formatCurrency(amount)}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <label className="text-sm font-medium">Сумма возврата</label>
              <Input
                type="number"
                value={refundAmount}
                onChange={(e) => setRefundAmount(e.target.value)}
                placeholder="Сумма"
                min={0}
                max={amount}
                step={0.01}
              />
            </div>
            <div>
              <label className="text-sm font-medium">Причина (опционально)</label>
              <Input
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="Причина возврата"
              />
            </div>
            {error && <p className="text-sm text-red-600">{error}</p>}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDialog(false)}
              disabled={loading}
            >
              Отмена
            </Button>
            <Button
              variant="destructive"
              onClick={handleRefund}
              disabled={loading}
            >
              {loading ? "Обработка..." : "Вернуть средства"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
