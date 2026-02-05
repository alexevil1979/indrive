/**
 * Rate Screen — rate driver after ride completion
 */
import { useState, useEffect } from "react";
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  TextInput,
  Alert,
  ScrollView,
  ActivityIndicator,
} from "react-native";
import { useLocalSearchParams, useRouter } from "expo-router";
import { Button, Card } from "@ridehail/ui";
import { useAuth } from "../../context/AuthContext";
import {
  submitRating,
  getRatingTags,
  getRide,
  type RatingTag,
  type Ride,
} from "../../lib/api";

const STARS = [1, 2, 3, 4, 5];

export default function RateScreen() {
  const params = useLocalSearchParams<{ rideId: string }>();
  const rideId = Array.isArray(params.rideId) ? params.rideId[0] : params.rideId;
  const { token } = useAuth();
  const router = useRouter();

  const [ride, setRide] = useState<Ride | null>(null);
  const [score, setScore] = useState(0);
  const [comment, setComment] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [availableTags, setAvailableTags] = useState<RatingTag[]>([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!token || !rideId) return;

    const loadData = async () => {
      setLoading(true);
      try {
        const [rideData, tagsData] = await Promise.all([
          getRide(token, rideId),
          getRatingTags(token, "driver"),
        ]);
        setRide(rideData);
        setAvailableTags(tagsData);
      } catch (e) {
        console.warn("Failed to load data:", e);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [token, rideId]);

  const toggleTag = (tagId: string) => {
    setSelectedTags((prev) =>
      prev.includes(tagId) ? prev.filter((t) => t !== tagId) : [...prev, tagId]
    );
  };

  const handleSubmit = async () => {
    if (score === 0) {
      Alert.alert("Ошибка", "Пожалуйста, выберите оценку");
      return;
    }

    if (!token || !rideId) return;

    setSubmitting(true);
    try {
      await submitRating(
        token,
        rideId,
        score,
        comment.trim() || undefined,
        selectedTags.length > 0 ? selectedTags : undefined
      );
      Alert.alert("Спасибо!", "Ваша оценка отправлена", [
        { text: "OK", onPress: () => router.back() },
      ]);
    } catch (e) {
      Alert.alert("Ошибка", e instanceof Error ? e.message : "Не удалось отправить оценку");
    } finally {
      setSubmitting(false);
    }
  };

  const handleSkip = () => {
    Alert.alert("Пропустить оценку?", "Вы можете оценить поездку позже.", [
      { text: "Отмена", style: "cancel" },
      { text: "Пропустить", onPress: () => router.back() },
    ]);
  };

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color="#2563eb" />
        <Text style={styles.loadingText}>Загрузка...</Text>
      </View>
    );
  }

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.title}>Как прошла поездка?</Text>
        {ride?.driver_id && (
          <Text style={styles.subtitle}>Оцените вашего водителя</Text>
        )}
      </View>

      {/* Stars */}
      <Card style={styles.card}>
        <View style={styles.starsContainer}>
          {STARS.map((star) => (
            <TouchableOpacity
              key={star}
              onPress={() => setScore(star)}
              style={styles.starButton}
            >
              <Text style={[styles.star, score >= star && styles.starActive]}>
                {score >= star ? "★" : "☆"}
              </Text>
            </TouchableOpacity>
          ))}
        </View>
        <Text style={styles.scoreLabel}>
          {score === 0 && "Нажмите на звезду"}
          {score === 1 && "Очень плохо"}
          {score === 2 && "Плохо"}
          {score === 3 && "Нормально"}
          {score === 4 && "Хорошо"}
          {score === 5 && "Отлично!"}
        </Text>
      </Card>

      {/* Tags */}
      {score > 0 && availableTags.length > 0 && (
        <Card style={styles.card}>
          <Text style={styles.sectionTitle}>
            {score >= 4 ? "Что понравилось?" : "Что было не так?"}
          </Text>
          <View style={styles.tagsContainer}>
            {availableTags.map((tag) => (
              <TouchableOpacity
                key={tag.id}
                style={[
                  styles.tagButton,
                  selectedTags.includes(tag.id) && styles.tagButtonActive,
                ]}
                onPress={() => toggleTag(tag.id)}
              >
                <Text
                  style={[
                    styles.tagText,
                    selectedTags.includes(tag.id) && styles.tagTextActive,
                  ]}
                >
                  {tag.label}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </Card>
      )}

      {/* Comment */}
      {score > 0 && (
        <Card style={styles.card}>
          <Text style={styles.sectionTitle}>Комментарий (необязательно)</Text>
          <TextInput
            style={styles.commentInput}
            placeholder="Расскажите подробнее..."
            placeholderTextColor="#94a3b8"
            value={comment}
            onChangeText={setComment}
            multiline
            maxLength={500}
          />
        </Card>
      )}

      {/* Actions */}
      <View style={styles.actions}>
        <Button
          title={submitting ? "Отправка..." : "Отправить оценку"}
          onPress={handleSubmit}
          variant="primary"
          disabled={score === 0 || submitting}
        />
        <TouchableOpacity style={styles.skipButton} onPress={handleSkip}>
          <Text style={styles.skipText}>Пропустить</Text>
        </TouchableOpacity>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f8fafc",
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
  header: {
    alignItems: "center",
    marginBottom: 24,
    paddingTop: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: "700",
    color: "#0f172a",
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 15,
    color: "#64748b",
  },
  card: {
    marginBottom: 16,
  },
  starsContainer: {
    flexDirection: "row",
    justifyContent: "center",
    marginBottom: 12,
  },
  starButton: {
    padding: 8,
  },
  star: {
    fontSize: 44,
    color: "#cbd5e1",
  },
  starActive: {
    color: "#fbbf24",
  },
  scoreLabel: {
    textAlign: "center",
    fontSize: 16,
    fontWeight: "600",
    color: "#0f172a",
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: "600",
    color: "#0f172a",
    marginBottom: 12,
  },
  tagsContainer: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: 8,
  },
  tagButton: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: "#f1f5f9",
    borderWidth: 1,
    borderColor: "#e2e8f0",
  },
  tagButtonActive: {
    backgroundColor: "#dbeafe",
    borderColor: "#2563eb",
  },
  tagText: {
    fontSize: 13,
    color: "#64748b",
  },
  tagTextActive: {
    color: "#2563eb",
    fontWeight: "500",
  },
  commentInput: {
    backgroundColor: "#f1f5f9",
    borderRadius: 12,
    padding: 14,
    fontSize: 15,
    minHeight: 100,
    textAlignVertical: "top",
    color: "#0f172a",
  },
  actions: {
    marginTop: 8,
  },
  skipButton: {
    alignItems: "center",
    paddingVertical: 16,
  },
  skipText: {
    color: "#64748b",
    fontSize: 14,
  },
});
