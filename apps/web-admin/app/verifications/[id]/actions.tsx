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
import { reviewVerification } from "@/lib/api";

interface VerificationActionsProps {
  verificationId: string;
}

export function VerificationActions({
  verificationId,
}: VerificationActionsProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [showRejectDialog, setShowRejectDialog] = useState(false);
  const [rejectReason, setRejectReason] = useState("");

  async function handleApprove() {
    setLoading(true);
    const success = await reviewVerification(verificationId, true);
    setLoading(false);
    if (success) {
      router.refresh();
    } else {
      alert("Ошибка при одобрении");
    }
  }

  async function handleReject() {
    if (!rejectReason.trim()) {
      alert("Укажите причину отклонения");
      return;
    }
    setLoading(true);
    const success = await reviewVerification(
      verificationId,
      false,
      rejectReason
    );
    setLoading(false);
    setShowRejectDialog(false);
    if (success) {
      router.refresh();
    } else {
      alert("Ошибка при отклонении");
    }
  }

  return (
    <>
      <div className="flex gap-4">
        <Button
          onClick={handleApprove}
          disabled={loading}
          className="bg-green-600 hover:bg-green-700"
        >
          {loading ? "Обработка..." : "Одобрить"}
        </Button>
        <Button
          variant="destructive"
          onClick={() => setShowRejectDialog(true)}
          disabled={loading}
        >
          Отклонить
        </Button>
      </div>

      <Dialog open={showRejectDialog} onOpenChange={setShowRejectDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Отклонить заявку</DialogTitle>
            <DialogDescription>
              Укажите причину отклонения заявки на верификацию.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Input
              placeholder="Причина отклонения"
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowRejectDialog(false)}
              disabled={loading}
            >
              Отмена
            </Button>
            <Button
              variant="destructive"
              onClick={handleReject}
              disabled={loading}
            >
              {loading ? "Обработка..." : "Отклонить"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
