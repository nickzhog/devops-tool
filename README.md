# Distributed Telemetry Collector

![Go Version](https://img.shields.io/badge/Go-1.19-00ADD8?style=flat&logo=go)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![License](https://img.shields.io/badge/license-MIT-blue)

A high-performance, distributed telemetry and alerting system designed for reliable metric aggregation across microservice environments. 

This project implements a robust Agent-Server architecture capable of ingesting runtime metrics, hardware stats, and custom application telemetry via multiprotocol transports (HTTP/REST and gRPC). It focuses on data integrity, secure transmission, and fault tolerance.

## 🏗 Architecture

The system is split into two primary decoupled components:
1. **Agent:** A lightweight daemon deployed on target nodes that gathers system metrics and pushes them to the central server using batching and retry mechanisms.
2. **Server:** A high-throughput ingestion node that validates, processes, and persists incoming telemetry into a relational database (PostgreSQL) or in-memory cache (Redis).

## ✨ Key Features

* **Multiprotocol Support:** Seamlessly accepts data via gRPC or standard HTTP/REST endpoints.
* **End-to-End Security:**
  * **Payload Encryption:** Asymmetric RSA encryption ensures metric data cannot be intercepted in transit.
  * **Data Integrity:** HMAC-SHA256 signatures validate the authenticity of incoming payloads.
  * **Network Security:** Built-in CIDR-based IP filtering middleware drops unauthorized traffic at the edge.
* **High-Throughput Processing:** Utilizes worker pools and batched database inserts to minimize I/O overhead and handle traffic spikes.
* **Resilience & Reliability:** * Graceful shutdown implementations prevent data loss during deployments.
  * Configurable exponential backoff for database connections and agent retries.
* **Data Persistence:** Persistent storage layer backed by PostgreSQL (`pgx`), high-speed caching via Redis, and JSON file backup options.

## 🚀 Quick Start

### Prerequisites
* Go 1.19
* Docker & Docker Compose
* Make

### Running Locally

The easiest way to spin up the entire infrastructure (Server, Agent, and Redis) is via Docker Compose:

```bash
# Clone the repository
git clone https://github.com/nickzhog/devops-tool.git
cd devops-tool

# Spin up the environment
docker-compose up -d --build
```

### Manual Build & Development

```bash
# Build the server
go build -o bin/server cmd/server/main.go

# Build the agent
go build -o bin/agent cmd/agent/main.go

# Useful Make commands
make test   # Run unit tests with coverage
make lint   # Run static analysis
make protoc # Regenerate gRPC code from protobuf definitions
```

## ⚙️ Configuration

The system is highly configurable via environment variables and CLI flags. Environment variables take precedence over flags.

### Server Configuration
| Environment Variable | Flag | Default | Description |
|---|---|---|---|
| `ADDRESS` | `-a` | `:8080` | Bind address for the HTTP server |
| `ADDRESS_GRPC` | `-g` | `:3200` | Bind address for the gRPC server |
| `DATABASE_DSN` | `-d` | `""` | PostgreSQL connection string |
| `REDIS_ADDR` | `-redis_addr` | `""` | Redis address (takes priority over database DSN) |
| `REDIS_PASSWORD` | `-redis_psw` | `""` | Redis password (optional) |
| `REDIS_DB` | `-redis_db` | `0` | Redis database index |
| `STORE_FILE` | `-f` | `/tmp/devops-metrics-db.json` | Path for JSON metrics backup file |
| `STORE_INTERVAL` | `-i` | `1s` | Interval for periodically saving metrics to file |
| `RESTORE` | `-r` | `true` | Restore metrics from file on server startup |
| `TRUSTED_SUBNET` | `-t` | `""` | CIDR notation for allowed IP ranges |
| `KEY` | `-k` | `""` | Secret key for HMAC signature validation |
| `CRYPTO_KEY` | `-crypto-key`| `""` | Path to the RSA private key for payload decryption |

### Agent Configuration
| Environment Variable | Flag | Default | Description |
|---|---|---|---|
| `ADDRESS` | `-a` | `http://127.0.0.1:8080` | Target HTTP server address |
| `ADDRESS_GRPC` | `-g` | `""` | Target gRPC server address (used over HTTP if set) |
| `POLL_INTERVAL` | `-p` | `2s` | Frequency of gathering metrics |
| `REPORT_INTERVAL` | `-r` | `10s` | Frequency of pushing metrics to the server |
| `KEY` | `-k` | `""` | Secret key for generating HMAC signatures |
| `CRYPTO_KEY` | `-crypto-key`| `""` | Path to the RSA public key for payload encryption |

## 🛠 Tech Stack

* **Language:** Go (Golang) 1.19
* **RPC Framework:** gRPC / Protocol Buffers
* **Database & Cache:** PostgreSQL (`jackc/pgx/v4`), Redis (`redis/go-redis/v9`)
* **Infrastructure:** Docker, Docker Compose
* **Routing:** `go-chi/chi`
* **Configuration:** `caarlos0/env`
