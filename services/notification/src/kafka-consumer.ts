/**
 * Kafka consumer stub — subscribe to ride.matched, ride.status.changed → send push
 * Run separately or integrate in index.ts when KAFKA_BROKERS set.
 * Example: on ride.matched → sendPushToUser(driverId, "Новая поездка", "Пассажир принял вашу ставку")
 */
export async function startKafkaConsumer(): Promise<void> {
  const brokers = process.env.KAFKA_BROKERS;
  if (!brokers) {
    console.log("KAFKA_BROKERS not set — notification consumer disabled");
    return;
  }
  console.log("Kafka consumer stub — would connect to", brokers);
  // TODO: use kafkajs or node-rdkafka to consume ride.matched, ride.status.changed
  // and call sendPushToUser for driver/passenger
}
