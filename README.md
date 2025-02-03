# GoodBlast Tournament API

## Overview
GoodBlast is a **casual Match-3 game** with a **Tournament System** that allows players to compete daily and track their rankings.

This project is built with **Golang 1.23**, **PostgreSQL**, **Redis**, and **Kafka**, ensuring high performance and scalability.

## Features
- **User Management**: Registration, login, and progress tracking.
- **Tournament System**: Daily tournaments, automatic creation at midnight.
- **Leaderboard**: Global and country-based rankings using Redis.
- **Dynamic Configuration**: Updates via GitHub-based config.
- **Asynchronous Processing**: Kafka event-driven architecture.
- **Caching**: Redis for leaderboard optimization.

## Tech Stack
- **Backend**: Golang (Gin Framework)
- **Database**: PostgreSQL
- **Caching**: Redis (ZSet for leaderboards)
- **Messaging**: Kafka (Confluent Cloud)
- **Configuration Management**: GitHub-based dynamic config
- **Authentication**: JWT (Paseto Token)

---

## Getting Started

### 1. Clone the Repository
```sh
$ git clone https://github.com/erkurtharun/goodblast.git
$ cd goodblast
```

### 2. Install Dependencies
Ensure you have **Golang 1.23**, **Docker**, and **Docker Compose** installed.

### 3. Environment Variables
Rename `.env.example` to `app.env` and update the required configuration values:

#### **app.env**
```ini
# Application Configuration
AppName=goodblast
Env=development
Port=8080

# Database Configuration
GoodBlastDBUrl=<your_db_url>
GoodBlastDBName=<your_db_name>
PostgresUsername=<your_db_username>
PostgresPassword=<your_db_password>

# Authentication
TokenSecretKey=<your_secret_key>

# Feature Toggle
ToggleConfigURL=<github_config_url>
GithubToken=<your_github_token>

# Kafka Configuration
KafkaBootstrapServers=<your_kafka_servers>
KafkaSecurityProtocol=SASL_SSL
KafkaSaslMechanism=PLAIN
KafkaSaslUsername=<your_kafka_username>
KafkaSaslPassword=<your_kafka_password>
KafkaClientId=<your_client_id>
KafkaSessionTimeout=45000
KafkaConsumerGroupId=go-group-1
KafkaAutoOffsetReset=earliest

# Redis Configuration
RedisHost=<your_redis_host>
RedisPort=<your_redis_port>
RedisPassword=<your_redis_password>
RedisDB=0
```

#### **Dynamic Configurations (GitHub Managed)**
```json
{
  "tokenTTL": 24,
  "coinPerLevel": 100,
  "tournamentCutoffHour": 23,
  "minimumTournamentEntryLevel": 10,
  "tournamentEntranceCoins": 500,
  "reward1": 5000,
  "reward2": 3000,
  "reward3": 2000,
  "reward4to10": 1000,
  "tournamentEntryTopic": "tournament_entrance",
  "userProgressUpdateTopic": "user_progress_update",
  "leaderboardUpdateTopic": "leaderboard_update"
}
```

### 4. Run the Application
#### With Docker Compose
```sh
$ docker-compose up --build
```

#### Without Docker
```sh
$ go run main.go
```

### 5. API Documentation (Swagger UI)
Once the server is running, visit:
```
http://localhost:8080/swagger/index.html
```

---

## Tournament Scheduling
- A **new tournament** is automatically created **every night at 00:00 UTC**.
- This process runs as a **scheduled job** that ensures uninterrupted tournament creation.

---

## Performance Optimizations

✅ **Asynchronous Processing** via Kafka for user progress & tournament entry.  
✅ **Redis Caching** for leaderboard queries (reduces load on Redis).  
✅ **PostgreSQL Transaction Optimization** for user & tournament writes.  
✅ **Event-Driven Updates** for user progress, leaderboard updates, and tournament entries.  
✅ **In-Memory Caching** for frequently accessed API responses to improve performance and reduce database/Redis load.

---

## Feature Toggle Project
The application uses a **dynamic configuration system** managed via a GitHub repository. You can find the **Toggles Project** here:  
🔗 [GitHub Toggles Repository](https://github.com/erkurtharun/Toggles)

---

## License
This project is licensed under the **MIT License**.

---

## Author
👤 **Harun ERKURT**  
📧 Email: erkurtharun@gmail.com
🔗 [GitHub](https://github.com/erkurtharun) | [LinkedIn](https://linkedin.com/in/erkurtharun)

