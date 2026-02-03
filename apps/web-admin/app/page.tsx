/**
 * Dashboard — analytics basics (rides stats)
 */
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchRides } from "@/lib/api";

export default async function DashboardPage() {
  let rides: Awaited<ReturnType<typeof fetchRides>> = [];
  try {
    rides = await fetchRides();
  } catch {
    rides = [];
  }

  const completed = rides.filter((r) => r.status === "completed");
  const inProgress = rides.filter(
    (r) => r.status === "in_progress" || r.status === "matched"
  );
  const today = new Date().toISOString().slice(0, 10);
  const completedToday = completed.filter((r) =>
    r.updated_at.startsWith(today)
  );
  const totalRevenue = completed.reduce((sum, r) => sum + (r.price ?? 0), 0);

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold">Дашборд</h1>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Всего поездок
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{rides.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Завершено сегодня
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{completedToday.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              В процессе
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{inProgress.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Выручка (завершённые)
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{totalRevenue.toFixed(0)} ₽</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Аналитика</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Данные загружаются из Ride API. Для полной аналитики подключите
            TimescaleDB и дашборды Grafana.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
