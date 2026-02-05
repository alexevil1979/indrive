"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import {
  fetchRatings,
  formatDate,
  type Rating,
  TAG_LABELS,
} from "@/lib/api";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export default function RatingsPage() {
  const router = useRouter();
  const [ratings, setRatings] = useState<Rating[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const limit = 20;

  const loadRatings = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchRatings(limit, page * limit);
      setRatings(data.ratings ?? []);
      setTotal(data.total ?? 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load ratings");
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    loadRatings();
  }, [loadRatings]);

  const renderStars = (score: number) => {
    return "★".repeat(score) + "☆".repeat(5 - score);
  };

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Отзывы и рейтинги</h1>
        <p className="text-gray-500 mt-1">Все оценки пользователей</p>
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : (
        <>
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Дата</TableHead>
                  <TableHead>Оценка</TableHead>
                  <TableHead>Роль</TableHead>
                  <TableHead>Теги</TableHead>
                  <TableHead>Комментарий</TableHead>
                  <TableHead>Поездка</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {ratings.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                      Отзывов пока нет
                    </TableCell>
                  </TableRow>
                ) : (
                  ratings.map((rating) => (
                    <TableRow key={rating.id} className="hover:bg-gray-50">
                      <TableCell className="whitespace-nowrap">
                        {formatDate(rating.created_at)}
                      </TableCell>
                      <TableCell>
                        <span className="text-yellow-500 text-lg">
                          {renderStars(rating.score)}
                        </span>
                        <span className="ml-2 text-gray-600 font-medium">
                          {rating.score}/5
                        </span>
                      </TableCell>
                      <TableCell>
                        <span
                          className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
                            rating.role === "driver"
                              ? "bg-green-100 text-green-800"
                              : "bg-blue-100 text-blue-800"
                          }`}
                        >
                          {rating.role === "driver" ? "Водитель" : "Пассажир"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {rating.tags?.map((tag) => (
                            <span
                              key={tag}
                              className="inline-flex px-2 py-0.5 text-xs bg-gray-100 text-gray-700 rounded-full"
                            >
                              {TAG_LABELS[tag] ?? tag}
                            </span>
                          )) ?? <span className="text-gray-400">—</span>}
                        </div>
                      </TableCell>
                      <TableCell className="max-w-xs">
                        {rating.comment ? (
                          <span className="text-gray-700 line-clamp-2">
                            {rating.comment}
                          </span>
                        ) : (
                          <span className="text-gray-400">—</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <button
                          onClick={() => router.push(`/rides/${rating.ride_id}`)}
                          className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                        >
                          Открыть
                        </button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="mt-4 flex items-center justify-between">
              <p className="text-sm text-gray-500">
                Показано {page * limit + 1}–{Math.min((page + 1) * limit, total)} из{" "}
                {total}
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(0, p - 1))}
                  disabled={page === 0}
                  className="px-3 py-1 text-sm border rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  ← Назад
                </button>
                <span className="px-3 py-1 text-sm text-gray-600">
                  {page + 1} / {totalPages}
                </span>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                  disabled={page >= totalPages - 1}
                  className="px-3 py-1 text-sm border rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Вперёд →
                </button>
              </div>
            </div>
          )}

          {/* Stats summary */}
          <div className="mt-6 grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="bg-white p-4 rounded-lg shadow">
              <p className="text-sm text-gray-500">Всего отзывов</p>
              <p className="text-2xl font-bold text-gray-900">{total}</p>
            </div>
            <div className="bg-white p-4 rounded-lg shadow">
              <p className="text-sm text-gray-500">Отзывы водителей</p>
              <p className="text-2xl font-bold text-green-600">
                {ratings.filter((r) => r.role === "driver").length}
              </p>
            </div>
            <div className="bg-white p-4 rounded-lg shadow">
              <p className="text-sm text-gray-500">Отзывы пассажиров</p>
              <p className="text-2xl font-bold text-blue-600">
                {ratings.filter((r) => r.role === "passenger").length}
              </p>
            </div>
            <div className="bg-white p-4 rounded-lg shadow">
              <p className="text-sm text-gray-500">Средняя оценка</p>
              <p className="text-2xl font-bold text-yellow-600">
                {ratings.length > 0
                  ? (
                      ratings.reduce((sum, r) => sum + r.score, 0) / ratings.length
                    ).toFixed(1)
                  : "—"}
              </p>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
