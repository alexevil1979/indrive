/**
 * Ride monitoring — list rides, filter by status
 */
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { fetchRides } from "@/lib/api";

const statusLabel: Record<string, string> = {
  requested: "Ожидает ставок",
  bidding: "Ставки",
  matched: "Водитель найден",
  in_progress: "В пути",
  completed: "Завершена",
  cancelled: "Отменена",
};

export default async function RidesPage() {
  let rides: Awaited<ReturnType<typeof fetchRides>> = [];
  try {
    rides = await fetchRides();
  } catch {
    rides = [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Поездки</h1>

      <Card>
        <CardHeader>
          <CardTitle>Список поездок</CardTitle>
          <p className="text-sm text-muted-foreground">
            Данные из Ride API (до 100 записей). В продакшене добавьте пагинацию и фильтр по статусу.
          </p>
        </CardHeader>
        <CardContent>
          {rides.length === 0 ? (
            <p className="text-muted-foreground py-8 text-center">
              Нет поездок или сервис недоступен.
            </p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left py-3 px-2">ID</th>
                    <th className="text-left py-3 px-2">Статус</th>
                    <th className="text-left py-3 px-2">Откуда</th>
                    <th className="text-left py-3 px-2">Куда</th>
                    <th className="text-left py-3 px-2">Цена</th>
                    <th className="text-left py-3 px-2">Создана</th>
                  </tr>
                </thead>
                <tbody>
                  {rides.map((r) => (
                    <tr key={r.id} className="border-b border-border/50">
                      <td className="py-3 px-2 font-mono text-xs">{r.id.slice(0, 8)}</td>
                      <td className="py-3 px-2">
                        <Badge
                          variant={
                            r.status === "completed"
                              ? "success"
                              : r.status === "cancelled"
                                ? "error"
                                : "default"
                          }
                        >
                          {statusLabel[r.status] ?? r.status}
                        </Badge>
                      </td>
                      <td className="py-3 px-2">
                        {r.from.address ?? `${r.from.lat.toFixed(2)}, ${r.from.lng.toFixed(2)}`}
                      </td>
                      <td className="py-3 px-2">
                        {r.to.address ?? `${r.to.lat.toFixed(2)}, ${r.to.lng.toFixed(2)}`}
                      </td>
                      <td className="py-3 px-2">
                        {r.price != null ? `${r.price} ₽` : "—"}
                      </td>
                      <td className="py-3 px-2 text-muted-foreground">
                        {new Date(r.created_at).toLocaleString("ru")}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
