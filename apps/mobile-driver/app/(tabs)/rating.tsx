/**
 * Rating Screen — view driver rating and reviews
 */
import { useState, useEffect, useCallback } from "react";
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  RefreshControl,
  ActivityIndicator,
  FlatList,
} from "react-native";
import { Card } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import {
  getMyRating,
  getMyRatings,
  type UserRating,
  type Rating,
  TAG_LABELS,
} from "../../lib/api";

export default function RatingScreen() {
  const { token } = useAuth();
  const [rating, setRating] = useState<UserRating | null>(null);
  const [reviews, setReviews] = useState<Rating[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const loadData = useCallback(async () => {
    if (!token) return;

    try {
      const [ratingData, reviewsData] = await Promise.all([
        getMyRating(token, "driver"),
        getMyRatings(token, "driver", 50, 0),
      ]);
      setRating(ratingData);
      setReviews(reviewsData);
    } catch (e) {
      console.warn("Failed to load rating:", e);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [token]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const onRefresh = () => {
    setRefreshing(true);
    loadData();
  };

  const formatDate = (dateString: string) => {
    try {
      const date = new Date(dateString);
      return date.toLocaleDateString("ru-RU", {
        day: "numeric",
        month: "short",
        year: "numeric",
      });
    } catch {
      return "";
    }
  };

  const renderStars = (score: number) => {
    return "★".repeat(score) + "☆".repeat(5 - score);
  };

  const renderReview = ({ item }: { item: Rating }) => (
    <Card style={styles.reviewCard}>
      <View style={styles.reviewHeader}>
        <Text style={styles.reviewStars}>{renderStars(item.score)}</Text>
        <Text style={styles.reviewDate}>{formatDate(item.created_at)}</Text>
      </View>
      {item.tags && item.tags.length > 0 && (
        <View style={styles.tagsRow}>
          {item.tags.map((tag) => (
            <View key={tag} style={styles.tagBadge}>
              <Text style={styles.tagText}>{TAG_LABELS[tag] ?? tag}</Text>
            </View>
          ))}
        </View>
      )}
      {item.comment ? (
        <Text style={styles.reviewComment}>{item.comment}</Text>
      ) : null}
    </Card>
  );

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color="#16a34a" />
        <Text style={styles.loadingText}>Загрузка рейтинга...</Text>
      </View>
    );
  }

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
    >
      {/* Rating Summary */}
      <Card style={styles.summaryCard}>
        <View style={styles.summaryHeader}>
          <Text style={styles.avgScore}>
            {rating?.average_score?.toFixed(1) ?? "0.0"}
          </Text>
          <View style={styles.summaryMeta}>
            <Text style={styles.avgStars}>
              {renderStars(Math.round(rating?.average_score ?? 0))}
            </Text>
            <Text style={styles.totalRatings}>
              {rating?.total_ratings ?? 0} оценок
            </Text>
          </View>
        </View>

        {/* Score distribution */}
        {rating && rating.total_ratings > 0 && (
          <View style={styles.distribution}>
            {[5, 4, 3, 2, 1].map((score) => {
              const count =
                score === 5 ? rating.score_5_count :
                score === 4 ? rating.score_4_count :
                score === 3 ? rating.score_3_count :
                score === 2 ? rating.score_2_count :
                rating.score_1_count;
              const percentage = rating.total_ratings > 0
                ? (count / rating.total_ratings) * 100
                : 0;

              return (
                <View key={score} style={styles.distRow}>
                  <Text style={styles.distLabel}>{score} ★</Text>
                  <View style={styles.distBarBg}>
                    <View
                      style={[
                        styles.distBarFill,
                        { width: `${percentage}%` },
                      ]}
                    />
                  </View>
                  <Text style={styles.distCount}>{count}</Text>
                </View>
              );
            })}
          </View>
        )}
      </Card>

      {/* Reviews */}
      <Text style={styles.sectionTitle}>Отзывы пассажиров</Text>
      {reviews.length === 0 ? (
        <Card style={styles.emptyCard}>
          <Text style={styles.emptyIcon}>📝</Text>
          <Text style={styles.emptyText}>У вас пока нет отзывов</Text>
          <Text style={styles.emptyHint}>
            Завершите поездки, чтобы получить оценки от пассажиров
          </Text>
        </Card>
      ) : (
        reviews.map((review) => (
          <View key={review.id}>{renderReview({ item: review })}</View>
        ))
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f0fdf4",
  },
  content: {
    padding: 16,
    paddingBottom: 32,
  },
  center: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  loadingText: {
    marginTop: 12,
    color: "#64748b",
  },
  summaryCard: {
    marginBottom: 20,
    padding: 20,
  },
  summaryHeader: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 16,
  },
  avgScore: {
    fontSize: 56,
    fontWeight: "700",
    color: "#16a34a",
    marginRight: 16,
  },
  summaryMeta: {
    flex: 1,
  },
  avgStars: {
    fontSize: 24,
    color: "#fbbf24",
    marginBottom: 4,
  },
  totalRatings: {
    fontSize: 14,
    color: "#64748b",
  },
  distribution: {
    borderTopWidth: 1,
    borderTopColor: "#e2e8f0",
    paddingTop: 16,
  },
  distRow: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 8,
  },
  distLabel: {
    width: 36,
    fontSize: 13,
    color: "#64748b",
  },
  distBarBg: {
    flex: 1,
    height: 8,
    backgroundColor: "#e2e8f0",
    borderRadius: 4,
    marginHorizontal: 8,
    overflow: "hidden",
  },
  distBarFill: {
    height: "100%",
    backgroundColor: "#16a34a",
    borderRadius: 4,
  },
  distCount: {
    width: 30,
    fontSize: 13,
    color: "#64748b",
    textAlign: "right",
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: "600",
    color: "#0f172a",
    marginBottom: 12,
  },
  reviewCard: {
    marginBottom: 12,
  },
  reviewHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 8,
  },
  reviewStars: {
    fontSize: 18,
    color: "#fbbf24",
  },
  reviewDate: {
    fontSize: 12,
    color: "#94a3b8",
  },
  tagsRow: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: 6,
    marginBottom: 8,
  },
  tagBadge: {
    backgroundColor: "#dcfce7",
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  tagText: {
    fontSize: 12,
    color: "#16a34a",
    fontWeight: "500",
  },
  reviewComment: {
    fontSize: 14,
    color: "#0f172a",
    lineHeight: 20,
  },
  emptyCard: {
    alignItems: "center",
    paddingVertical: 32,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 12,
  },
  emptyText: {
    fontSize: 16,
    fontWeight: "600",
    color: "#0f172a",
    marginBottom: 4,
  },
  emptyHint: {
    fontSize: 13,
    color: "#64748b",
    textAlign: "center",
  },
});
