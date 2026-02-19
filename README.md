# 🚀 High-Performance Distributed Rate Limiter

A high-throughput, distributed rate-limiting middleware built with **Go** and **Redis**. Designed to handle massive concurrent traffic spikes with microsecond-level latency and strict atomicity. Fully containerized for cloud-native deployment.

## ✨ Core Features

* **Distributed State Management:** Uses Redis to maintain rate limit states across multiple service instances.
* **Strict Atomicity:** Implements the Token Bucket algorithm entirely within **Redis Lua scripts** to guarantee zero race conditions and prevent over-allocation under high concurrency.
* **Cloud-Native Ready:** Fully Dockerized with multi-stage builds and Docker Compose for one-click deployment.
* **High Throughput & Low Latency:** Capable of handling ~13,600+ req/sec on a single node with an average response time of <1ms.

## 🛠️ Tech Stack

* **Language:** Go (Golang)
* **Storage/Cache:** Redis
* **Deployment:** Docker, Docker Compose
* **Testing:** `hey` (Load testing)

## 🐳 Quick Start

You can run the entire stack (API Server + Redis) with a single command using Docker Compose:

```bash
git clone [https://github.com/DEM-YU/distributed-rate-limiter.git](https://github.com/DEM-YU/distributed-rate-limiter.git)
cd distributed-rate-limiter
docker-compose up -d --build
