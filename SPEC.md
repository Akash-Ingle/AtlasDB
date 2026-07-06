# AtlasDB — Project Specification

> A production-grade distributed event streaming, search, and AI analytics platform.

**Version:** 1.0.0-draft
**Date:** 2026-07-05
**Status:** Awaiting approval before implementation

---

## Table of Contents

1. [Product Vision](#1-product-vision)
2. [Resume Positioning](#2-resume-positioning)
3. [Product Features](#3-product-features)
4. [High-Level Architecture](#4-high-level-architecture)
5. [Backend Architecture](#5-backend-architecture)
6. [Event Pipeline](#6-event-pipeline)
7. [Database Design](#7-database-design)
8. [Search Engine](#8-search-engine)
9. [Streaming Architecture](#9-streaming-architecture)
10. [AI Layer](#10-ai-layer)
11. [Observability](#11-observability)
12. [DevOps](#12-devops)
13. [Security](#13-security)
14. [Performance](#14-performance)
15. [Frontend](#15-frontend)
16. [Engineering Metrics](#16-engineering-metrics)
17. [Demo Scenarios](#17-demo-scenarios)
18. [Folder Structure](#18-folder-structure)
19. [Development Roadmap](#19-development-roadmap)
20. [Stretch Features](#20-stretch-features)

---

## 1. Product Vision

### What Is AtlasDB

AtlasDB is a **self-hosted, distributed event streaming and AI-powered analytics platform** designed for engineering teams that need to ingest, store, search, analyze, and reason about massive volumes of structured events in real time.

It combines the ingestion throughput of Kafka, the search capabilities of Elasticsearch, the columnar analytics of ClickHouse, and the intelligent investigation workflows of Datadog — unified under a single platform with an AI copilot that can autonomously investigate incidents, discover patterns, and generate insights.

AtlasDB is not a monitoring tool bolted onto an LLM. It is a **platform where AI is a first-class citizen** — the AI layer has deep access to the event pipeline, search indices, analytics engine, and alert system. It can reason over historical data, correlate anomalies across services, and execute multi-step investigations with tool calling and memory.

### Target Users

| Persona | Use Case |
|---|---|
| **Platform Engineers** | Centralize observability events from microservices into a single queryable store |
| **SREs / On-Call Engineers** | Investigate incidents with AI-assisted root cause analysis instead of grepping logs |
| **Backend Engineers** | Stream application events for real-time analytics dashboards |
| **Data Engineers** | Build event-driven pipelines with retry, DLQ, and backpressure semantics |
| **Security Engineers** | Detect anomalous patterns (brute force, data exfiltration) with AI-powered alerting |
| **Engineering Managers** | View weekly AI-generated reports on system health, incident trends, and SLA compliance |

### Real-World Use Cases

1. **Microservice Observability** — A team running 50 microservices sends structured events (request logs, error traces, deployment markers) to AtlasDB. Engineers search across services with natural language: *"Show me all 5xx errors from the payment service in the last hour."*

2. **Incident Investigation** — At 3 AM, an alert fires. Instead of manually correlating dashboards, the on-call engineer asks the AI copilot: *"What changed before the latency spike on checkout-service?"* The AI agent queries recent deployments, error rate changes, and infrastructure events, then produces a root cause summary.

3. **Fraud Detection Pipeline** — A fintech company streams transaction events. AtlasDB's anomaly detection flags unusual patterns. The AI layer explains: *"User X performed 47 transactions in 2 minutes from 3 different countries — 98.2% confidence this is anomalous based on historical behavior."*

4. **Real-Time Analytics Dashboard** — A SaaS company ingests product analytics events. AtlasDB powers a live dashboard showing signups, feature usage, and conversion funnels with sub-second query latency.

5. **Security Audit Trail** — Every API call, authentication attempt, and permission change is ingested as an immutable event. Security teams search and filter with full-text and semantic search, with AI-generated weekly security posture reports.

### Why This System Exists

Modern engineering teams face a fragmented tooling landscape:

- **Kafka** handles streaming but has no query layer.
- **Elasticsearch** handles search but is operationally expensive and has no native AI.
- **ClickHouse** handles analytics but requires ETL pipelines and has no streaming ingestion.
- **Datadog/Grafana** provide dashboards but are SaaS-only, expensive, and treat AI as an afterthought.

AtlasDB unifies these capabilities into a single, self-hosted platform where events flow through ingestion → processing → storage → search → analytics → AI — with every stage instrumented, observable, and scalable.

### Competitive Positioning

| Capability | Kafka | Elasticsearch | ClickHouse | Datadog | Grafana | **AtlasDB** |
|---|---|---|---|---|---|---|
| Event streaming | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Full-text search | ❌ | ✅ | Partial | ✅ | ❌ | ✅ |
| Semantic (vector) search | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Columnar analytics | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| AI-powered investigation | ❌ | ❌ | ❌ | Partial | ❌ | ✅ |
| Natural language queries | ❌ | ❌ | ❌ | Partial | ❌ | ✅ |
| Self-hosted | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Unified platform | ❌ | ❌ | ❌ | ✅ | Partial | ✅ |

### Scope

AtlasDB is a **backend-heavy, infrastructure-rich, AI-integrated platform**. It is not a general-purpose database. It is purpose-built for **event data** — time-series events with structured metadata, originating from applications, infrastructure, or business systems.

The scope explicitly includes:

- A high-throughput event ingestion API
- A durable, partitioned event storage layer
- A search engine supporting keyword, filtered, and semantic queries
- A real-time and batch analytics engine
- An AI layer with agentic investigation, anomaly explanation, and NL querying
- A professional dashboard frontend
- Full observability instrumentation (metrics, traces, logs)
- Production-grade DevOps (Docker, Kubernetes, CI/CD, Terraform)
- Security (auth, RBAC, encryption, audit logging)

The scope explicitly excludes:

- Being a general-purpose SQL database
- Replacing a message broker for inter-service communication (AtlasDB consumes events; it is not a message bus for application logic)
- Mobile applications
- Multi-cloud orchestration beyond Terraform patterns

---

## 2. Resume Positioning

### AI Engineer Resume

**Headline:** *Built an AI-powered analytics platform with agentic investigation, semantic search, and autonomous anomaly detection over a distributed event store.*

**Key talking points:**

- **RAG Pipeline** — Implemented retrieval-augmented generation over event data. Events are embedded and stored in a vector database (pgvector). When an engineer asks a natural language question, the system retrieves semantically relevant events, constructs context, and generates grounded answers with citation.

- **Agentic Investigation Workflow** — Built a multi-step AI agent that investigates incidents autonomously. The agent has access to tools: `query_events`, `get_metrics`, `list_deployments`, `search_similar_incidents`, `get_service_dependencies`. It plans an investigation, executes tool calls, synthesizes findings, and produces a root cause report with confidence scoring.

- **Anomaly Detection & Explanation** — Implemented statistical anomaly detection (Z-score, IQR, isolation forest) on event streams. When anomalies are detected, an LLM generates human-readable explanations by analyzing the anomalous data in context of historical baselines and correlated events.

- **Natural Language → SQL/Query Translation** — Users type natural language queries (*"Show me error rates by service for the last 24 hours"*). The system translates these into structured queries against the event store, executes them, and returns results with the generated query visible for transparency.

- **Semantic Search** — Events are embedded at ingestion time using a sentence transformer. Search supports hybrid retrieval: BM25 keyword matching combined with cosine similarity over embeddings, with reciprocal rank fusion for result merging.

- **Background AI Workers** — Designed a background job system where AI workers continuously analyze event streams: generating weekly health reports, deduplicating alerts, discovering patterns, and suggesting dashboard configurations.

- **Evaluation & Confidence** — Every AI response includes a confidence score. Implemented an evaluation framework that measures answer quality against labeled test cases for regression testing of AI behavior.

- **Memory & Context Management** — The AI copilot maintains conversation memory within investigation sessions. Long conversations are summarized to fit context windows. Cross-session memory stores resolved incidents for future reference.

### Backend / SWE Resume

**Headline:** *Designed and built a distributed event streaming platform handling millions of events with real-time processing, search, analytics, and sub-100ms API latency.*

**Key talking points:**

- **Event-Driven Architecture** — Designed a multi-stage event pipeline: HTTP ingestion → validation → async queue → stream processing → storage → indexing → analytics. Each stage is independently scalable and connected via message queues with backpressure, retry, and dead-letter semantics.

- **Distributed Systems Design** — Implemented partitioned event storage with time-based sharding. Designed the system for horizontal scaling: stateless API servers behind a load balancer, partitioned queues, and sharded storage. Handled consistency tradeoffs (eventual consistency for search index, strong consistency for event writes).

- **High-Performance Search** — Built a search subsystem supporting full-text search with inverted indices, filtered queries with bitmap indices, and time-range queries with partition pruning. Query latency under 50ms at p99 for typical workloads.

- **Concurrent Stream Processing** — Implemented a stream processor in Go that consumes events from the queue, applies transformations (enrichment, normalization, aggregation), and writes to multiple sinks (storage, search index, analytics tables) concurrently with configurable parallelism.

- **API Design** — Designed RESTful APIs with OpenAPI specs, versioning, pagination (cursor-based), filtering, rate limiting, and comprehensive error handling. WebSocket endpoints for real-time event streaming to the dashboard.

- **Caching Architecture** — Multi-layer caching: Redis for hot query results and session data, in-process LRU caches for frequently accessed metadata, cache invalidation via event-driven updates.

- **Database Design** — PostgreSQL with time-based table partitioning for event storage, covering indices for common query patterns, connection pooling via PgBouncer, read replicas for analytics queries.

- **Worker Queue System** — Built a reliable job queue for background tasks (AI analysis, report generation, alert evaluation) with priority levels, retry with exponential backoff, dead-letter queue, and worker pool management.

### Cloud / DevOps Resume

**Headline:** *Instrumented and deployed a distributed platform on Kubernetes with full observability (OpenTelemetry, Prometheus, Grafana), CI/CD, infrastructure-as-code, and autoscaling.*

**Key talking points:**

- **Kubernetes Deployment** — Designed Kubernetes manifests (Deployments, Services, ConfigMaps, Secrets, HPA, PDB) for all services. Implemented health checks (liveness, readiness, startup probes) and resource limits. Configured Horizontal Pod Autoscaler based on CPU, memory, and custom metrics (queue depth).

- **Observability Stack** — Instrumented every service with OpenTelemetry SDK for distributed tracing. Exported metrics to Prometheus. Built Grafana dashboards for system health, event pipeline throughput, search latency, AI response times, and queue depths. Structured JSON logging with correlation IDs propagated across services.

- **CI/CD Pipeline** — GitHub Actions workflows for: lint → test → build → security scan → Docker build → push to registry → deploy to staging → integration tests → promote to production. Branch protection, required checks, and automated rollback on health check failure.

- **Infrastructure as Code** — Terraform modules for provisioning cloud resources: managed PostgreSQL, Redis, object storage, Kubernetes cluster, networking, IAM. State stored remotely with locking.

- **Docker & Compose** — Multi-stage Dockerfiles optimized for layer caching and minimal image size. Docker Compose for local development with all services, databases, and observability stack runnable on a single machine.

- **Deployment Strategies** — Implemented rolling deployments (default) and blue/green deployments (for critical services) in Kubernetes. Canary deployment support via traffic splitting.

- **Autoscaling** — Horizontal Pod Autoscaler for stateless services. Custom metrics adapter for scaling based on queue depth (scale ingestion workers when queue backs up). Vertical Pod Autoscaler recommendations for right-sizing.

- **Secrets Management** — Kubernetes Secrets with external-secrets-operator pattern. Environment-specific configuration via ConfigMaps. No secrets in code or Docker images.

---

## 3. Product Features

### MVP Features

#### 1. Event Ingestion API

- **Purpose:** Accept structured events via HTTP (REST) and buffer them for processing. This is the entry point for all data in the system.
- **Technical Complexity:** Must handle high concurrency, validate schemas, assign timestamps and IDs, and enqueue events reliably without data loss. Requires careful error handling — partial batch failures, malformed payloads, backpressure when downstream is slow.
- **Why Recruiters Care:** Demonstrates ability to build high-throughput, production-grade APIs — the bread and butter of backend engineering at scale companies.

#### 2. Event Processing Pipeline

- **Purpose:** Consume events from the queue, enrich them (add metadata, normalize fields), and fan out to storage, search index, and analytics.
- **Technical Complexity:** Concurrent consumer groups, exactly-once vs at-least-once semantics, ordering guarantees within partitions, dead-letter queue for poison messages, configurable processing parallelism.
- **Why Recruiters Care:** Shows understanding of stream processing, message queue semantics, and the tradeoffs that come up in every distributed system interview.

#### 3. Event Storage

- **Purpose:** Durably store events in a queryable format with time-based partitioning.
- **Technical Complexity:** PostgreSQL with declarative table partitioning by time range. Covering indices for common access patterns (by service, by severity, by time range). Partition pruning for efficient time-range queries. Retention policies with automatic partition dropping.
- **Why Recruiters Care:** Database design, indexing strategy, and partitioning are core senior engineer competencies.

#### 4. Search

- **Purpose:** Allow users to search events by keyword, filter by fields, and constrain by time range.
- **Technical Complexity:** Full-text search using PostgreSQL `tsvector`/`tsquery` for MVP (migrating to dedicated search later). GIN indices on JSONB fields. Cursor-based pagination. Query parsing and validation.
- **Why Recruiters Care:** Search is a core product feature at companies like Elastic, Datadog, and Stripe. Understanding search internals (inverted indices, ranking, pagination) is a strong signal.

#### 5. Real-Time Dashboard

- **Purpose:** Display live event stream, basic analytics (event volume over time, top services, error rates), and search interface.
- **Technical Complexity:** WebSocket connection for live events. Efficient re-rendering with virtualized lists (handling thousands of events). Responsive layout. Time-series charts.
- **Why Recruiters Care:** Full-stack capability. Shows you can build the interface that makes backend systems usable.

#### 6. Authentication & Authorization

- **Purpose:** Secure the platform with user accounts, JWT-based auth, and API keys for programmatic access.
- **Technical Complexity:** JWT issuance and validation, refresh token rotation, API key generation with scoped permissions, password hashing (bcrypt/argon2), session management.
- **Why Recruiters Care:** Security is non-negotiable. Every production system needs auth, and implementing it correctly demonstrates attention to detail.

#### 7. Alerting

- **Purpose:** Define threshold-based alert rules (e.g., "alert when error rate exceeds 5% over 5 minutes") and notify via webhook/email.
- **Technical Complexity:** Alert rule evaluation engine that runs periodically, checks conditions against recent data, manages alert state (firing/resolved/silenced), deduplication, notification dispatch.
- **Why Recruiters Care:** Alerting is a core feature of every observability platform. Building one from scratch shows deep product and systems understanding.

#### 8. Observability

- **Purpose:** Instrument AtlasDB itself with metrics, traces, and structured logs so the platform is self-observable.
- **Technical Complexity:** OpenTelemetry SDK integration, Prometheus metrics exposition, distributed trace propagation across services, structured JSON logging with correlation IDs, Grafana dashboard provisioning.
- **Why Recruiters Care:** For DevOps/SRE roles, this is the headline feature. For backend roles, it shows production maturity.

### Advanced Features

#### 9. AI Copilot — Natural Language Queries

- **Purpose:** Users type questions in plain English. The AI translates them into structured queries, executes them, and returns results.
- **Technical Complexity:** Prompt engineering for reliable query generation. Schema-aware context injection. Query validation and sandboxing (prevent destructive queries). Streaming response via SSE. Fallback handling when the LLM produces invalid queries.
- **Why Recruiters Care:** This is the headline AI feature. It demonstrates practical LLM integration — not a toy chatbot, but a system that translates natural language into real database operations.

#### 10. Semantic Search (Vector Search)

- **Purpose:** Find events by meaning, not just keywords. "Authentication failures" should match events containing "login denied", "invalid credentials", "401 unauthorized".
- **Technical Complexity:** Embedding generation at ingestion time (sentence-transformers). Vector storage in pgvector. Hybrid retrieval combining BM25 keyword scores with cosine similarity. Reciprocal rank fusion. Index tuning (HNSW parameters).
- **Why Recruiters Care:** Vector search is one of the most in-demand AI engineering skills. Implementing it over a real data pipeline (not a toy demo) is a strong differentiator.

#### 11. Anomaly Detection

- **Purpose:** Automatically detect unusual patterns in event streams — traffic spikes, error rate changes, latency degradation.
- **Technical Complexity:** Statistical methods (Z-score over sliding windows, IQR). Configurable per metric. Integration with the alert system. Low false-positive tuning. AI-generated explanations for detected anomalies.
- **Why Recruiters Care:** Combines ML/statistics with systems engineering. Shows ability to build intelligent features, not just plumbing.

#### 12. AI Agent — Incident Investigation

- **Purpose:** An autonomous AI agent that investigates incidents by querying data, correlating events, and producing root cause analyses.
- **Technical Complexity:** Tool-calling agent architecture. Tool definitions: `query_events`, `get_metrics`, `list_recent_deployments`, `get_service_topology`, `search_similar_incidents`. Multi-step planning and execution. Conversation memory. Confidence scoring. Guardrails (max steps, timeout, cost limits).
- **Why Recruiters Care:** Agentic AI is the frontier. Building a production-grade agent with tools, planning, and guardrails demonstrates deep AI engineering capability.

#### 13. Kubernetes Deployment

- **Purpose:** Deploy AtlasDB on Kubernetes with production-grade configurations.
- **Technical Complexity:** Helm charts or Kustomize overlays. Health probes. Resource requests/limits. HPA with custom metrics. PodDisruptionBudgets. Network policies. Ingress configuration. Persistent volume claims for stateful services.
- **Why Recruiters Care:** Kubernetes is the standard for production deployments. Demonstrating real K8s manifests (not just `docker run`) is expected for platform/DevOps roles.

#### 14. CI/CD Pipeline

- **Purpose:** Automated build, test, and deploy pipeline.
- **Technical Complexity:** Multi-stage GitHub Actions workflows. Parallel test execution. Docker layer caching. Container scanning. Staged deployment (staging → production). Automated rollback on failed health checks. Branch protection rules.
- **Why Recruiters Care:** CI/CD is a core DevOps competency. A well-designed pipeline shows production engineering maturity.

#### 15. Background AI Workers

- **Purpose:** Continuously running AI workers that analyze data streams and produce insights without user interaction.
- **Technical Complexity:** Worker pool management. Job scheduling (cron-based and event-triggered). Tasks: weekly health reports, alert deduplication, pattern discovery, suggested dashboards. Rate limiting API calls to LLM providers. Cost tracking.
- **Why Recruiters Care:** Shows you can build AI features that operate at system level, not just respond to user prompts.

### Stretch Goals

#### 16. Multi-Tenancy

- **Purpose:** Support multiple teams/organizations with data isolation.
- **Technical Complexity:** Row-level security in PostgreSQL. Tenant-scoped queries. Per-tenant rate limits and quotas. Tenant-aware caching.
- **Why Recruiters Care:** Multi-tenancy is a defining challenge at SaaS companies (Datadog, Snowflake, Stripe). Discussing tenant isolation strategies shows platform maturity.

#### 17. Event Replay

- **Purpose:** Re-process historical events through updated processing logic.
- **Technical Complexity:** Idempotent processing. Checkpoint management. Progress tracking. Parallel replay across partitions.
- **Why Recruiters Care:** Event sourcing and replay are staples of distributed systems interviews.

#### 18. Custom Query Language

- **Purpose:** A domain-specific query language for event exploration (similar to Datadog's log query syntax or Splunk's SPL).
- **Technical Complexity:** Lexer, parser, AST, query planner, execution engine. Syntax: `service:payment-api status:error | stats count by endpoint | sort -count`.
- **Why Recruiters Care:** Building a query language from scratch is a strong differentiator for compiler/systems roles.

#### 19. Federated Search

- **Purpose:** Query across multiple AtlasDB instances.
- **Technical Complexity:** Query routing, result merging, distributed pagination, timeout handling.
- **Why Recruiters Care:** Distributed query execution is a core challenge at companies like Snowflake, Databricks, and CockroachDB.

#### 20. Plugin System

- **Purpose:** Allow users to write custom event processors, alert handlers, and dashboard widgets.
- **Technical Complexity:** Plugin API design, sandboxed execution, versioning, discovery.
- **Why Recruiters Care:** Platform extensibility shows senior-level API design thinking.

---

## 4. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          CLIENTS                                    │
│   SDKs (Go, Python, JS)  ·  CLI  ·  Dashboard UI  ·  REST API      │
└─────────────────────┬───────────────────────────────────────────────┘
                      │ HTTPS / WebSocket
                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       API GATEWAY                                   │
│   Rate Limiting · Auth · Request Routing · Load Balancing           │
└──────┬──────────┬──────────┬───────────┬──────────┬─────────────────┘
       │          │          │           │          │
       ▼          ▼          ▼           ▼          ▼
┌──────────┐ ┌────────┐ ┌────────┐ ┌─────────┐ ┌──────────┐
│ Ingest   │ │ Query  │ │ Search │ │ Alert   │ │ AI       │
│ Service  │ │ Service│ │ Service│ │ Service │ │ Service  │
└────┬─────┘ └───┬────┘ └───┬────┘ └────┬────┘ └────┬─────┘
     │           │          │           │           │
     ▼           │          │           │           │
┌──────────┐     │          │           │           │
│ Message  │     │          │           │           │
│ Queue    │     │          │           │           │
│ (Redis   │     │          │           │           │
│ Streams) │     │          │           │           │
└────┬─────┘     │          │           │           │
     │           │          │           │           │
     ▼           │          │           │           │
┌──────────┐     │          │           │           │
│ Stream   │     │          │           │           │
│ Processor│     │          │           │           │
└──┬───┬───┘     │          │           │           │
   │   │         │          │           │           │
   ▼   ▼         ▼          ▼           ▼           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        DATA LAYER                                   │
│  ┌────────────┐  ┌──────────┐  ┌──────────┐  ┌───────────────────┐ │
│  │ PostgreSQL │  │  Redis   │  │ pgvector │  │  Object Storage   │ │
│  │ (Events,   │  │ (Cache,  │  │ (Vector  │  │  (S3/MinIO -      │ │
│  │  Metadata, │  │  Queue,  │  │  Embeds) │  │   Backups,        │ │
│  │  Analytics)│  │  Sessions)│ │          │  │   Large Blobs)    │ │
│  └────────────┘  └──────────┘  └──────────┘  └───────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
       │                                              │
       ▼                                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     OBSERVABILITY                                   │
│  OpenTelemetry Collector · Prometheus · Grafana · Structured Logs   │
└─────────────────────────────────────────────────────────────────────┘
```

### Service Descriptions

| Service | Responsibility |
|---|---|
| **API Gateway** | Single entry point. Authenticates requests (JWT/API key), applies rate limiting, routes to backend services. In MVP, this is a reverse proxy layer within the main API; in production, it could be Kong, Traefik, or a custom Go router. |
| **Ingest Service** | Accepts event payloads, validates schema, assigns event IDs and server-side timestamps, batches events into the message queue. Optimized for throughput — minimal processing, fast acknowledgment. |
| **Message Queue** | Redis Streams acts as the durable buffer between ingestion and processing. Provides consumer groups for parallel processing, message acknowledgment, and pending message recovery. |
| **Stream Processor** | Consumes events from the queue. Enriches events (GeoIP, service metadata lookup). Normalizes fields. Writes to PostgreSQL (durable storage), updates search indices, and feeds the analytics aggregation tables. Handles failures with retry and dead-letter routing. |
| **Query Service** | Serves structured queries against the event store. Supports filtering, time ranges, aggregations. Powers the dashboard's analytics views. |
| **Search Service** | Handles keyword and semantic search. Parses search queries, executes against full-text and vector indices, ranks results, returns paginated responses. |
| **Alert Service** | Evaluates alert rules on a schedule. Checks conditions (threshold, rate-of-change, anomaly) against recent data. Manages alert lifecycle (pending → firing → resolved). Dispatches notifications. |
| **AI Service** | Hosts the AI copilot and background AI workers. Handles natural language query translation, incident investigation agent, anomaly explanation, report generation. Manages LLM API calls, prompt construction, tool execution, and response streaming. |
| **Data Layer** | PostgreSQL for durable event storage and relational data. Redis for caching, session management, and message queuing. pgvector (PostgreSQL extension) for vector embeddings. MinIO/S3 for object storage (backups, large payloads). |
| **Observability Stack** | OpenTelemetry Collector receives traces and metrics from all services. Prometheus scrapes and stores metrics. Grafana provides dashboards. All services emit structured JSON logs with trace/span IDs. |

---

## 5. Backend Architecture

### Technology Choice: Go

**Primary language:** Go (Golang)

**Rationale:**
- Native concurrency (goroutines, channels) is ideal for stream processing and high-throughput APIs
- Single binary compilation simplifies Docker images and deployment
- Excellent standard library for HTTP servers, JSON handling, and networking
- Industry standard for infrastructure software (Kubernetes, Docker, Prometheus, CockroachDB are all Go)
- Strong performance characteristics without GC pauses that affect tail latency
- Resume signal: Go is the language of choice at infrastructure companies

**Frontend:** TypeScript + React + Next.js

### Service Design Principles

Every service follows these conventions:

1. **Clean separation** — Each service is a standalone Go module with its own `main.go`, configuration, and Dockerfile.
2. **Interface-driven** — Core logic is defined by Go interfaces. Implementations can be swapped (e.g., PostgreSQL storage vs in-memory for testing).
3. **Configuration** — Environment variables with sensible defaults. Validated at startup. No config files in production.
4. **Health checks** — Every service exposes `/healthz` (liveness) and `/readyz` (readiness) endpoints.
5. **Graceful shutdown** — All services handle SIGTERM, drain in-flight requests, close database connections, and exit cleanly.
6. **Structured logging** — JSON logs with `service`, `trace_id`, `span_id`, `level`, `message`, `timestamp`.
7. **Metrics** — Prometheus metrics exposed on `/metrics`. Standard HTTP metrics (request count, latency histogram, error rate) plus service-specific metrics.
8. **Tracing** — OpenTelemetry spans for all inbound/outbound requests, database queries, and queue operations.

### Service Specifications

#### 5.1 API Service (`api-server`)

The central HTTP server that handles client-facing requests.

**Endpoints:**

```
POST   /api/v1/events              Ingest events (single or batch)
GET    /api/v1/events              Query events (with filters)
GET    /api/v1/events/:id          Get single event
GET    /api/v1/events/stream       WebSocket: live event stream

GET    /api/v1/search              Search events
POST   /api/v1/search/semantic     Semantic search

GET    /api/v1/analytics/summary   Dashboard summary
GET    /api/v1/analytics/timeseries  Time-series data
GET    /api/v1/analytics/top       Top-N queries

POST   /api/v1/alerts/rules        Create alert rule
GET    /api/v1/alerts/rules        List alert rules
PUT    /api/v1/alerts/rules/:id    Update alert rule
DELETE /api/v1/alerts/rules/:id    Delete alert rule
GET    /api/v1/alerts/history      Alert history

POST   /api/v1/ai/query            Natural language query
POST   /api/v1/ai/investigate      Start investigation
GET    /api/v1/ai/investigate/:id  Get investigation status
POST   /api/v1/ai/explain          Explain anomaly/event

POST   /api/v1/auth/register       Register user
POST   /api/v1/auth/login          Login (returns JWT)
POST   /api/v1/auth/refresh        Refresh token
POST   /api/v1/auth/api-keys       Generate API key
DELETE /api/v1/auth/api-keys/:id   Revoke API key

GET    /api/v1/system/health       System health overview
GET    /api/v1/system/metrics      Internal metrics summary
```

**Implementation details:**
- Go `net/http` with `chi` router (lightweight, idiomatic)
- Middleware chain: request ID → logging → tracing → auth → rate limit → handler
- Request validation using struct tags
- Cursor-based pagination for all list endpoints
- Consistent error response format: `{ "error": { "code": "...", "message": "...", "details": {...} } }`
- Content negotiation (JSON, optionally NDJSON for streaming)

#### 5.2 Authentication Service (`auth`)

Embedded within the API service as a middleware layer and handler group.

**Components:**
- **JWT Issuer** — Signs access tokens (15min TTL) and refresh tokens (7d TTL) using RS256
- **API Key Manager** — Generates, stores (hashed), and validates API keys with scoped permissions
- **Middleware** — Extracts and validates JWT or API key from `Authorization` header
- **Password Manager** — Argon2id hashing for user passwords
- **Session Store** — Redis-backed session cache for fast token validation

**Authorization model:**
- Roles: `admin`, `editor`, `viewer`
- API keys carry scopes: `events:write`, `events:read`, `search:read`, `alerts:manage`, `ai:query`

#### 5.3 Ingestion Service (`ingest`)

Optimized path for accepting events with minimal latency.

**Design:**
- Accepts single events or batches (up to 1000 events per request)
- Validates event schema (required fields: `source`, `type`, `timestamp`, `data`)
- Assigns server-side `event_id` (ULIDv2 — lexicographically sortable, time-ordered)
- Assigns `received_at` timestamp
- Writes to Redis Streams (the message queue) in batches
- Returns acknowledgment immediately after queue write (not after processing)
- Backpressure: returns `429 Too Many Requests` when queue depth exceeds threshold

**Event schema:**
```json
{
  "event_id": "01J5M2K...",
  "source": "payment-service",
  "type": "http_request",
  "severity": "info",
  "timestamp": "2026-07-05T12:00:00Z",
  "received_at": "2026-07-05T12:00:00.123Z",
  "data": {
    "method": "POST",
    "path": "/api/charge",
    "status": 500,
    "duration_ms": 1234,
    "user_id": "usr_abc123"
  },
  "tags": ["production", "us-east-1"],
  "metadata": {
    "host": "pod-abc-xyz",
    "version": "v2.3.1"
  }
}
```

#### 5.4 Stream Processor (`processor`)

The core event processing engine.

**Design:**
- Runs as a separate process (or set of processes) consuming from Redis Streams
- Uses consumer groups for parallel processing across multiple instances
- Processing pipeline per event:
  1. Deserialize and validate
  2. Enrich (resolve service metadata, add GeoIP if applicable)
  3. Normalize (standardize field names, coerce types)
  4. Generate embedding (async, via embedding worker)
  5. Write to PostgreSQL (durable storage)
  6. Update search index (full-text tsvector)
  7. Update analytics aggregation tables (increment counters, update histograms)
  8. Evaluate real-time alert conditions
  9. ACK message in Redis Stream
- On failure: retry up to 3 times with exponential backoff, then route to dead-letter stream
- Configurable parallelism: `PROCESSOR_WORKERS=8`

#### 5.5 Search Service (`search`)

Handles all search operations.

**Design:**
- **Keyword search:** PostgreSQL full-text search using `tsvector` column with GIN index. Supports boolean operators, phrase matching, and field-scoped search.
- **Filtered search:** JSONB `@>` containment queries with GIN index on `data` column. Supports nested field access.
- **Semantic search:** Queries pgvector using cosine distance (`<=>` operator) with HNSW index. Returns top-K nearest neighbors.
- **Hybrid search:** Combines keyword and semantic results using Reciprocal Rank Fusion (RRF). Score = Σ(1 / (k + rank_i)) across retrieval methods.
- **Time-range optimization:** All queries scoped to time partitions. Partition pruning eliminates scanning irrelevant data.

#### 5.6 Analytics Service (`analytics`)

Pre-aggregated analytics for dashboard performance.

**Design:**
- **Materialized aggregations:** Background workers maintain pre-computed aggregation tables:
  - `event_counts_1m` — event count per service per minute
  - `event_counts_1h` — rolled up hourly
  - `error_rates_1m` — error count / total count per service per minute
  - `latency_percentiles_1m` — p50, p95, p99 latency per endpoint per minute
- **Query engine:** Serves aggregation queries by reading pre-computed tables (fast) rather than scanning raw events (slow).
- **Real-time component:** For the current (incomplete) minute, queries raw events and merges with pre-computed data.
- **Retention:** Raw events: 30 days. 1-minute aggregations: 30 days. 1-hour aggregations: 1 year.

#### 5.7 Notification Service (`notifier`)

Dispatches alert notifications.

**Design:**
- Consumes alert events from the alert service
- Supports channels: webhook (HTTP POST), email (SMTP), Slack (incoming webhook)
- Template engine for notification formatting
- Retry with exponential backoff on delivery failure
- Deduplication: won't re-notify for the same alert within a configurable cooldown period

#### 5.8 AI Service (`ai`)

The intelligence layer. Detailed in [Section 10](#10-ai-layer).

#### 5.9 Worker Queue (`workers`)

Background job execution system.

**Design:**
- Job queue backed by Redis (sorted set for priority, list for FIFO)
- Job types:
  - `embed_event` — generate vector embedding for an event
  - `evaluate_alerts` — run alert rule evaluation
  - `generate_report` — AI-generated health report
  - `cleanup_partitions` — drop expired partitions
  - `aggregate_rollup` — roll up minute-level aggregations to hourly
- Worker pool with configurable concurrency per job type
- Dead-letter queue for persistently failing jobs
- Job status tracking: `pending` → `running` → `completed` / `failed`
- Prometheus metrics: queue depth, processing time, failure rate per job type

#### 5.10 Scheduler (`scheduler`)

Cron-like job scheduling.

**Design:**
- Runs periodic tasks on configurable schedules:
  - Every 1 minute: alert rule evaluation
  - Every 5 minutes: analytics rollup
  - Every 1 hour: anomaly detection scan
  - Every 24 hours: partition maintenance, AI weekly report (on Mondays)
- Uses distributed locking (Redis `SET NX EX`) to prevent duplicate execution when multiple scheduler instances run
- Emits metrics on schedule execution (last run time, duration, success/failure)

---

## 6. Event Pipeline

### Complete Event Lifecycle

```
Client SDK / API Call
        │
        ▼
┌───────────────────┐
│ 1. RECEPTION      │  API server receives HTTP POST /api/v1/events
│                   │  Assign request ID, start trace span
└───────┬───────────┘
        │
        ▼
┌───────────────────┐
│ 2. AUTHENTICATION │  Validate JWT or API key
│                   │  Check scope: events:write
│                   │  Extract tenant/user context
└───────┬───────────┘
        │
        ▼
┌───────────────────┐
│ 3. VALIDATION     │  Validate event schema (required fields, types)
│                   │  Validate field lengths and sizes
│                   │  Reject malformed events with 400
│                   │  For batches: partial success (207 Multi-Status)
└───────┬───────────┘
        │
        ▼
┌───────────────────┐
│ 4. ENRICHMENT     │  Assign event_id (ULID)
│  (Lightweight)    │  Assign received_at timestamp
│                   │  Add server metadata (ingestion host, region)
└───────┬───────────┘
        │
        ▼
┌───────────────────┐
│ 5. QUEUE WRITE    │  Write event(s) to Redis Stream (XADD)
│                   │  Stream key: events:{partition}
│                   │  Partition by: hash(source) % N
│                   │  Return 202 Accepted to client
└───────┬───────────┘
        │
        ▼
   Client receives ACK (pipeline continues async)
        │
        ▼
┌───────────────────┐
│ 6. CONSUMPTION    │  Stream Processor reads via XREADGROUP
│                   │  Consumer group ensures each event processed once
│                   │  Claim pending messages after timeout (recovery)
└───────┬───────────┘
        │
        ▼
┌───────────────────┐
│ 7. PROCESSING     │  Full enrichment:
│                   │    - Resolve service metadata from registry
│                   │    - Normalize field names to canonical schema
│                   │    - Parse structured data from string fields
│                   │    - Compute derived fields (duration buckets, etc.)
└───────┬───────────┘
        │
        ├──────────────────────────────────────────────┐
        │                                              │
        ▼                                              ▼
┌───────────────────┐                    ┌───────────────────┐
│ 8a. STORAGE       │                    │ 8b. INDEXING      │
│  INSERT into      │                    │  Update tsvector   │
│  events table     │                    │  (full-text index) │
│  (partitioned     │                    │  Enqueue embedding │
│   by time)        │                    │  generation job    │
└───────┬───────────┘                    └───────┬───────────┘
        │                                        │
        ├────────────────────────────┐           │
        │                            │           │
        ▼                            ▼           ▼
┌───────────────────┐  ┌───────────────────┐
│ 9. AGGREGATION    │  │ 10. ALERT EVAL    │
│  Increment        │  │   Check event     │
│  event_counts_1m  │  │   against active  │
│  Update error     │  │   alert rules     │
│  rate counters    │  │   (real-time      │
│  Update latency   │  │    evaluation)    │
│  percentile       │  │   Fire alert if   │
│  sketches         │  │   threshold met   │
└───────────────────┘  └───────┬───────────┘
                               │
                               ▼
                  ┌───────────────────┐
                  │ 11. NOTIFICATION  │
                  │   If alert fired: │
                  │   dispatch via    │
                  │   webhook/email/  │
                  │   Slack           │
                  └───────┬───────────┘
                          │
                          ▼
                  ┌───────────────────┐
                  │ 12. AI ANALYSIS   │
                  │   (Background)    │
                  │   Anomaly check   │
                  │   Pattern match   │
                  │   Update AI       │
                  │   knowledge base  │
                  └───────────────────┘
```

### Stage Details

**Stages 1–5 (Synchronous, client-facing):**
Total target latency: < 20ms at p99. The client receives a `202 Accepted` response as soon as the event is durably written to the Redis Stream. This decouples ingestion latency from processing latency.

**Stages 6–12 (Asynchronous, background):**
Processing latency target: < 500ms from queue write to storage + index (stages 6–8). End-to-end latency (including AI analysis): < 5s for most events, background analysis on longer schedules.

### Failure Handling

| Failure | Handling |
|---|---|
| Malformed event | Rejected at validation (stage 3) with detailed error |
| Queue full | Return 429 with `Retry-After` header. Client SDK retries with exponential backoff |
| Processor crash | Redis Stream retains unacknowledged messages. New processor instance claims pending messages after visibility timeout |
| Storage write failure | Retry 3x with exponential backoff. After 3 failures, route to dead-letter stream for manual investigation |
| Embedding generation failure | Non-blocking. Event stored without embedding. Retry job enqueued. Semantic search degrades gracefully |
| Alert evaluation failure | Logged and metric emitted. Next evaluation cycle will catch up |

---

## 7. Database Design

### Technology Selection

#### PostgreSQL (Primary Data Store)

**Why:** Mature, reliable, supports partitioning, full-text search (`tsvector`), JSONB indexing, and vector search via pgvector extension — all in one database. For a portfolio project, this eliminates operational complexity of running separate Elasticsearch, ClickHouse, and Pinecone instances while demonstrating the same engineering concepts.

**Use cases:**
- Event storage (partitioned by time)
- User accounts and authentication
- Alert rules and alert history
- Analytics aggregation tables
- System configuration

#### Redis (Cache, Queue, Sessions)

**Why:** In-memory data store with Streams (message queue), sorted sets (priority queue), pub/sub (real-time notifications), and key-value cache. Single dependency that serves multiple infrastructure roles.

**Use cases:**
- Message queue (Redis Streams) for event pipeline
- Cache for hot query results (60s TTL)
- Session store for authenticated users
- Distributed locks for scheduler
- Real-time pub/sub for WebSocket event fan-out
- Worker job queue

#### pgvector (Vector Storage)

**Why:** PostgreSQL extension that adds vector similarity search. Eliminates the need for a separate vector database while supporting HNSW and IVFFlat indices. Keeps all data in one database engine, simplifying joins between vector results and relational data.

**Use cases:**
- Event embeddings for semantic search
- Similar incident retrieval for AI investigations

#### MinIO / S3-Compatible Object Storage

**Why:** Stores large, infrequently accessed data. MinIO provides S3-compatible API for local development; production can use any S3-compatible store.

**Use cases:**
- Event partition backups
- AI-generated reports (PDF/HTML)
- Large event payloads (> 64KB)
- Database backups

### Schema Design

#### Events Table (Partitioned)

```sql
CREATE TABLE events (
    event_id        TEXT        NOT NULL,  -- ULID, sortable
    source          TEXT        NOT NULL,  -- originating service
    event_type      TEXT        NOT NULL,  -- e.g., http_request, error, deployment
    severity        TEXT        NOT NULL DEFAULT 'info',  -- debug, info, warn, error, fatal
    timestamp       TIMESTAMPTZ NOT NULL,  -- client-provided event time
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    data            JSONB       NOT NULL DEFAULT '{}',
    tags            TEXT[]      DEFAULT '{}',
    metadata        JSONB       DEFAULT '{}',

    -- Search
    search_vector   TSVECTOR,  -- auto-populated by trigger
    embedding       vector(384),  -- populated async by embedding worker

    -- Partitioning
    PRIMARY KEY (event_id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create partitions: one per day
-- Example: events_2026_07_05 FOR VALUES FROM ('2026-07-05') TO ('2026-07-06')
```

#### Indices

```sql
-- Time-range queries (partition key, pruned automatically)
-- No additional index needed; partition pruning handles this

-- Source filtering
CREATE INDEX idx_events_source ON events (source, timestamp DESC);

-- Severity filtering
CREATE INDEX idx_events_severity ON events (severity, timestamp DESC)
    WHERE severity IN ('error', 'fatal');  -- partial index, only for interesting severities

-- Full-text search
CREATE INDEX idx_events_search ON events USING GIN (search_vector);

-- JSONB data access (commonly queried fields)
CREATE INDEX idx_events_data_status ON events ((data->>'status'), timestamp DESC);
CREATE INDEX idx_events_data_path ON events ((data->>'path'), timestamp DESC);

-- Tag search
CREATE INDEX idx_events_tags ON events USING GIN (tags);

-- Vector similarity search (HNSW for approximate nearest neighbor)
CREATE INDEX idx_events_embedding ON events
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 200);
```

#### Aggregation Tables

```sql
CREATE TABLE event_counts_1m (
    bucket          TIMESTAMPTZ NOT NULL,  -- truncated to minute
    source          TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    severity        TEXT        NOT NULL,
    count           BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (bucket, source, event_type, severity)
) PARTITION BY RANGE (bucket);

CREATE TABLE latency_stats_1m (
    bucket          TIMESTAMPTZ NOT NULL,
    source          TEXT        NOT NULL,
    endpoint        TEXT        NOT NULL,
    count           BIGINT      NOT NULL,
    sum_ms          DOUBLE PRECISION NOT NULL,
    min_ms          DOUBLE PRECISION NOT NULL,
    max_ms          DOUBLE PRECISION NOT NULL,
    p50_ms          DOUBLE PRECISION,
    p95_ms          DOUBLE PRECISION,
    p99_ms          DOUBLE PRECISION,
    PRIMARY KEY (bucket, source, endpoint)
) PARTITION BY RANGE (bucket);

CREATE TABLE event_counts_1h (
    bucket          TIMESTAMPTZ NOT NULL,
    source          TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    severity        TEXT        NOT NULL,
    count           BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (bucket, source, event_type, severity)
) PARTITION BY RANGE (bucket);
```

#### Users and Auth

```sql
CREATE TABLE users (
    user_id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT        UNIQUE NOT NULL,
    password_hash   TEXT        NOT NULL,
    role            TEXT        NOT NULL DEFAULT 'viewer',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE api_keys (
    key_id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(user_id),
    key_hash        TEXT        NOT NULL,  -- bcrypt hash of the key
    key_prefix      TEXT        NOT NULL,  -- first 8 chars for identification
    name            TEXT        NOT NULL,
    scopes          TEXT[]      NOT NULL,
    last_used_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### Alert Rules and History

```sql
CREATE TABLE alert_rules (
    rule_id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT        NOT NULL,
    description     TEXT,
    condition       JSONB       NOT NULL,  -- { "metric": "error_rate", "operator": ">", "threshold": 0.05, "window": "5m", "group_by": ["source"] }
    severity        TEXT        NOT NULL DEFAULT 'warning',
    channels        JSONB       NOT NULL DEFAULT '[]',  -- [{ "type": "webhook", "url": "..." }]
    enabled         BOOLEAN     NOT NULL DEFAULT true,
    cooldown_seconds INTEGER    NOT NULL DEFAULT 300,
    created_by      UUID        REFERENCES users(user_id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE alert_events (
    alert_event_id  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id         UUID        NOT NULL REFERENCES alert_rules(rule_id),
    status          TEXT        NOT NULL,  -- firing, resolved
    fired_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ,
    value           DOUBLE PRECISION,  -- the metric value that triggered the alert
    context         JSONB,  -- snapshot of relevant data
    notified        BOOLEAN     NOT NULL DEFAULT false
);

CREATE INDEX idx_alert_events_rule ON alert_events (rule_id, fired_at DESC);
CREATE INDEX idx_alert_events_status ON alert_events (status, fired_at DESC);
```

#### AI Tables

```sql
CREATE TABLE ai_investigations (
    investigation_id UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID       REFERENCES users(user_id),
    query            TEXT       NOT NULL,
    status           TEXT       NOT NULL DEFAULT 'running',  -- running, completed, failed
    steps            JSONB      NOT NULL DEFAULT '[]',  -- array of { tool, input, output, duration_ms }
    summary          TEXT,
    confidence       DOUBLE PRECISION,
    tokens_used      INTEGER,
    started_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at     TIMESTAMPTZ
);

CREATE TABLE ai_conversations (
    message_id       UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id       UUID       NOT NULL,
    role             TEXT       NOT NULL,  -- user, assistant, system, tool
    content          TEXT       NOT NULL,
    tool_calls       JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_conversations_session ON ai_conversations (session_id, created_at);
```

### Partitioning Strategy

- **Events:** Daily partitions. Auto-created by a scheduled job 7 days in advance. Auto-dropped after retention period (configurable, default 30 days).
- **Aggregation tables:** Daily partitions for 1-minute tables, monthly partitions for 1-hour tables.
- **Partition management:** A scheduled job runs daily to create future partitions and detach/drop expired ones.

### Caching Strategy

| Data | Cache Layer | TTL | Invalidation |
|---|---|---|---|
| Query results (analytics) | Redis | 60s | Time-based expiry |
| User sessions | Redis | 15min | On logout / token refresh |
| Service metadata | In-process LRU | 5min | Periodic refresh |
| Alert rules | In-process LRU | 1min | On rule update (pub/sub) |
| Search results | Redis | 30s | Time-based expiry |
| AI conversation context | Redis | 1hr | On session end |

### Scaling Strategy

| Component | Scaling Approach |
|---|---|
| API servers | Horizontal — add replicas behind load balancer. Stateless. |
| Stream processors | Horizontal — add consumer group members. Redis Streams handles partition assignment. |
| PostgreSQL reads | Read replicas for analytics queries. Primary for writes. |
| PostgreSQL writes | Vertical scaling first. Connection pooling via PgBouncer. Batch writes. |
| Redis | Vertical for MVP. Redis Cluster for production scale. |
| Workers | Horizontal — add worker instances. Job queue distributes work. |
| Search | Read replicas for query load. Async index updates. |

---

## 8. Search Engine

### Architecture

The search engine supports four query modes, all accessible through a unified API.

#### 8.1 Keyword Search

**Implementation:** PostgreSQL full-text search.

- Events have a `search_vector` column of type `tsvector`, populated by a trigger that concatenates searchable fields: `source`, `event_type`, `severity`, `data::text`, `tags`.
- Queries are parsed into `tsquery` using `plainto_tsquery` (simple) or `to_tsquery` (advanced with boolean operators).
- Results ranked by `ts_rank_cd` (cover density ranking).
- GIN index on `search_vector` for fast lookups.

**Query examples:**
```
"connection timeout"       → phrase search
payment AND error          → boolean AND
error OR failure           → boolean OR
source:payment-api error   → field-scoped + keyword
```

**Parser:** The search service implements a lightweight query parser that translates the user-facing syntax into `tsquery` predicates + SQL `WHERE` clauses for field-scoped terms.

#### 8.2 Filtered Search

**Implementation:** SQL `WHERE` clauses over indexed columns and JSONB fields.

**Supported filters:**
```
source = "payment-service"
severity IN ("error", "fatal")
timestamp >= "2026-07-04T00:00:00Z" AND timestamp < "2026-07-05T00:00:00Z"
data->>'status' = '500'
data->>'duration_ms' > '1000'
tags @> ARRAY['production']
```

**Optimization:** Filters that include `timestamp` ranges trigger partition pruning, dramatically reducing the amount of data scanned.

#### 8.3 Semantic Search

**Implementation:** pgvector cosine similarity search.

- At query time, the search query text is embedded using the same model used at ingestion time (e.g., `all-MiniLM-L6-v2`, 384 dimensions).
- The embedding is compared against stored event embeddings using cosine distance: `ORDER BY embedding <=> $query_embedding LIMIT $k`.
- HNSW index provides approximate nearest neighbor search with sub-linear query time.

**When to use:** Semantic search is most valuable when the user's query terms don't exactly match the event text. Example: searching "authentication failures" finds events containing "login denied", "invalid credentials", "401 unauthorized".

#### 8.4 Hybrid Search (Keyword + Semantic)

**Implementation:** Reciprocal Rank Fusion (RRF).

1. Execute keyword search → get ranked list R₁
2. Execute semantic search → get ranked list R₂
3. For each document d, compute: `RRF_score(d) = 1/(k + rank₁(d)) + 1/(k + rank₂(d))` where k=60
4. Sort by RRF_score descending
5. Return top-N results

**Why RRF:** Simple, parameter-free (beyond k), and empirically effective for combining heterogeneous retrieval signals. No need to normalize scores across different ranking systems.

#### 8.5 Time-Based Query Optimization

All search queries are time-bounded (default: last 24 hours). This is critical for performance:

- Partition pruning eliminates scanning irrelevant daily partitions
- Indices are smaller per partition, leading to faster lookups
- Users almost always care about recent events

#### 8.6 Pagination

**Approach:** Cursor-based pagination using `(timestamp, event_id)` as the cursor.

```json
{
  "results": [...],
  "cursor": "eyJ0cyI6IjIwMjYtMDctMDVUMTI6MDA6MDBaIiwiZWlkIjoiMDFKNU0ySy4uLiJ9",
  "has_more": true
}
```

**Why cursor-based:** Offset-based pagination (`LIMIT/OFFSET`) degrades at high offsets because the database must scan and discard rows. Cursor-based pagination uses an indexed `WHERE` clause, providing consistent performance regardless of page depth.

#### 8.7 Indexing Strategy Summary

| Access Pattern | Index Type | Column(s) |
|---|---|---|
| Full-text keyword search | GIN | `search_vector` |
| Time-range scan | B-tree (partition key) | `timestamp` |
| Source + time | B-tree composite | `(source, timestamp DESC)` |
| Severity + time (errors only) | Partial B-tree | `(severity, timestamp DESC) WHERE severity IN ('error','fatal')` |
| JSONB field access | B-tree on expression | `((data->>'status'), timestamp DESC)` |
| Tag filtering | GIN | `tags` |
| Vector similarity | HNSW | `embedding vector_cosine_ops` |

---

## 9. Streaming Architecture

### Message Queue: Redis Streams

**Why Redis Streams (not Kafka, not RabbitMQ):**

- **Kafka** is the gold standard for event streaming but is operationally heavy (JVM, ZooKeeper/KRaft, topic management). For a portfolio project where the goal is demonstrating streaming concepts, Redis Streams provides the same semantics (append-only log, consumer groups, message acknowledgment, persistence) with dramatically simpler operations.
- **RabbitMQ** is a message broker, not a stream. It doesn't retain messages after consumption, making replay impossible.
- Redis Streams provides: persistent append-only log, consumer groups, message acknowledgment, pending entry list (PEL) for failure recovery, and `XRANGE` for replay — the same primitives that Kafka provides.

**The engineering concepts are identical:** partitioning, consumer groups, offset management, at-least-once delivery, dead-letter queues, backpressure. An interviewer asking about streaming will care about these concepts, not whether you used Kafka or Redis Streams.

### Stream Topology

```
Ingestion API
      │
      │ XADD
      ▼
┌─────────────────────────────────────────────┐
│           Redis Streams                      │
│                                              │
│  events:0  events:1  events:2  events:3     │  ← 4 partitions
│                                              │
│  dead-letter:events                          │  ← DLQ stream
│                                              │
│  notifications                               │  ← alert notifications
│                                              │
└─────┬──────────┬──────────┬──────────┬──────┘
      │          │          │          │
      ▼          ▼          ▼          ▼
  Consumer   Consumer   Consumer   Consumer
  Worker 0   Worker 1   Worker 2   Worker 3
      │          │          │          │
      ▼          ▼          ▼          ▼
   PostgreSQL + Search Index + Analytics
```

### Partitioning

Events are partitioned by `hash(source) % num_partitions`. This ensures:
- Events from the same source land on the same partition → ordering within a source is preserved
- Load is distributed across partitions → parallel processing
- Number of partitions is configurable at deployment time (default: 4 for development, 16+ for production)

### Consumer Groups

Each stream processor instance joins a consumer group (`atlas-processors`). Redis Streams automatically distributes partitions across consumers. When a consumer is added or removed, pending partitions are rebalanced.

### Message Lifecycle

```
1. XADD events:0 * field value      → Message added to stream
2. XREADGROUP GROUP atlas-processors → Consumer reads message
   CONSUMER worker-0 COUNT 10
3. Process message                    → Application logic
4. XACK events:0 atlas-processors    → Message acknowledged
   message-id
```

### At-Least-Once Delivery

Messages are not removed from the pending entry list (PEL) until explicitly acknowledged with `XACK`. If a consumer crashes:
1. The message remains in the PEL
2. After a visibility timeout (configurable, default 30s), another consumer can claim it with `XCLAIM`
3. A background goroutine in each consumer periodically scans for stale pending messages and claims them

This provides at-least-once delivery. Processing logic must be **idempotent** — writing the same event twice produces the same result (achieved via `INSERT ON CONFLICT DO NOTHING` using `event_id`).

### Dead-Letter Queue

After 3 failed processing attempts (tracked via the delivery count in PEL), the message is:
1. Written to the `dead-letter:events` stream with the original payload and error details
2. Acknowledged in the original stream (removed from PEL)
3. A metric is incremented (`atlas_dlq_messages_total`)
4. An alert can be configured on DLQ depth

DLQ messages are inspectable via an admin API and can be replayed manually.

### Backpressure

When the system is overwhelmed:

1. **Queue depth monitoring:** The ingestion service checks `XLEN` of the target stream before writing. If depth exceeds the threshold (configurable, default 100,000), it returns `429 Too Many Requests` with a `Retry-After` header.

2. **Consumer-side backpressure:** Each consumer reads in batches (`COUNT 10`). If processing is slow, the consumer naturally slows its read rate, allowing the queue to buffer.

3. **Metrics-based autoscaling:** Queue depth is exposed as a Prometheus metric. The Kubernetes HPA can scale processor pods based on this metric.

### Batch Processing

For high-volume scenarios:
- **Batch ingestion:** API accepts arrays of events (up to 1000). Written to Redis in a pipeline (single round trip).
- **Batch processing:** Consumer reads up to 100 events per `XREADGROUP` call. Writes to PostgreSQL in batch `INSERT` (multi-row insert).
- **Batch embedding:** Embedding worker processes events in batches of 32 for efficient GPU/API utilization.

### Tradeoffs

| Decision | Tradeoff |
|---|---|
| Redis Streams over Kafka | Simpler operations, lower resource usage. Gives up: native topic replication, longer retention, richer ecosystem (Kafka Connect, Schema Registry). Acceptable for portfolio scope. |
| At-least-once over exactly-once | Simpler implementation. Requires idempotent consumers. Exactly-once would require transactional outbox pattern, adding significant complexity. |
| Hash partitioning by source | Good load distribution for many sources. Skewed if one source dominates. Mitigation: monitor partition lag, re-partition if needed. |
| Visibility timeout for failure recovery | 30s timeout means a crashed consumer's messages are unavailable for 30s. Acceptable for analytics workloads. Lower timeout increases false claims on slow-but-alive consumers. |

---

## 10. AI Layer

The AI layer is a first-class subsystem, not a bolt-on. It has deep integration with the event pipeline, search engine, analytics, and alert system.

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        AI SERVICE                               │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ NL Query     │  │ Investigation│  │ Background Workers   │  │
│  │ Engine       │  │ Agent        │  │                      │  │
│  │              │  │              │  │ - Anomaly Explainer  │  │
│  │ User text    │  │ Multi-step   │  │ - Report Generator   │  │
│  │ → SQL/Query  │  │ tool-calling │  │ - Alert Deduplicator │  │
│  │ → Execute    │  │ investigation│  │ - Pattern Discoverer │  │
│  │ → Format     │  │ with memory  │  │ - Dashboard Suggester│  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │
│         │                 │                      │              │
│         ▼                 ▼                      ▼              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                   LLM Gateway                            │   │
│  │  - Provider abstraction (Claude, OpenAI, local)          │   │
│  │  - Prompt management (templates, versioning)             │   │
│  │  - Token tracking & cost estimation                      │   │
│  │  - Rate limiting & retry                                 │   │
│  │  - Response streaming (SSE)                              │   │
│  │  - Structured output parsing (JSON mode)                 │   │
│  └─────────────────────────────────────────────────────────┘   │
│         │                 │                      │              │
│         ▼                 ▼                      ▼              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                   Tool Registry                          │   │
│  │                                                          │   │
│  │  query_events      - Query event store with filters      │   │
│  │  search_events     - Full-text and semantic search       │   │
│  │  get_metrics       - Retrieve metric time-series         │   │
│  │  list_deployments  - Recent deployment events            │   │
│  │  get_service_info  - Service metadata and dependencies   │   │
│  │  get_alert_history - Recent alerts for a service         │   │
│  │  find_similar      - Vector search for similar incidents │   │
│  │  run_aggregation   - Run analytics query                 │   │
│  │  get_system_health - Current system health snapshot      │   │
│  └─────────────────────────────────────────────────────────┘   │
│         │                 │                      │              │
│         ▼                 ▼                      ▼              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Evaluation & Guardrails                      │   │
│  │  - Confidence scoring                                     │   │
│  │  - Max steps / max tokens / max cost per request          │   │
│  │  - Query sandboxing (read-only, no DDL)                   │   │
│  │  - Output validation (structured response schema)         │   │
│  │  - Hallucination mitigation (cite sources)                │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 10.1 Natural Language Query Engine

**Purpose:** Translate plain English questions into structured queries, execute them, and return formatted results.

**Flow:**

1. User submits: *"Show me login failures from yesterday"*
2. System constructs prompt with:
   - The user's question
   - Database schema (table definitions, available columns, data types)
   - Available query patterns (examples of valid queries)
   - Current time context
3. LLM generates a structured query:
   ```json
   {
     "type": "event_query",
     "filters": {
       "event_type": "authentication",
       "data.result": "failure",
       "timestamp_gte": "2026-07-04T00:00:00Z",
       "timestamp_lt": "2026-07-05T00:00:00Z"
     },
     "order_by": "timestamp DESC",
     "limit": 100
   }
   ```
4. System validates the generated query (no destructive operations, valid fields, reasonable limits)
5. System executes the query against the event store
6. Results are passed back to the LLM for natural language summarization
7. Response streamed to user via SSE with both the summary and the raw results

**Transparency:** The generated query is always shown to the user so they can verify what was executed.

**Error handling:** If the LLM generates an invalid query, the system retries once with error feedback. If still invalid, returns a helpful error: *"I couldn't translate that query. Try being more specific about the time range and fields."*

### 10.2 Incident Investigation Agent

**Purpose:** An autonomous agent that investigates incidents through multi-step reasoning with tool calling.

**Agent Loop:**

```
User: "Why is checkout-service slow?"
                │
                ▼
┌─────────────────────────────────┐
│ PLAN                            │
│ Agent decides investigation     │
│ strategy:                       │
│ 1. Check recent error rates     │
│ 2. Check recent deployments     │
│ 3. Check latency metrics        │
│ 4. Search for related events    │
│ 5. Find similar past incidents  │
└───────────┬─────────────────────┘
            │
            ▼
┌─────────────────────────────────┐
│ EXECUTE (loop)                  │
│                                 │
│ Step 1: call get_metrics(       │
│   service="checkout-service",   │
│   metric="error_rate",          │
│   window="1h")                  │
│ → Result: error rate 12%        │
│   (normal: 0.5%)                │
│                                 │
│ Step 2: call list_deployments(  │
│   service="checkout-service",   │
│   window="6h")                  │
│ → Result: v2.3.1 deployed 45m  │
│   ago                           │
│                                 │
│ Step 3: call query_events(      │
│   source="checkout-service",    │
│   severity="error",             │
│   window="1h")                  │
│ → Result: 847 errors,           │
│   "connection refused:          │
│    payment-gateway:5432"        │
│                                 │
│ Step 4: call get_service_info(  │
│   service="payment-gateway")    │
│ → Result: payment-gateway       │
│   health: degraded,             │
│   DB connection pool exhausted  │
│                                 │
│ (Agent decides it has enough    │
│  information)                   │
└───────────┬─────────────────────┘
            │
            ▼
┌─────────────────────────────────┐
│ SYNTHESIZE                      │
│                                 │
│ Root Cause: checkout-service    │
│ latency is caused by            │
│ payment-gateway's database      │
│ connection pool exhaustion.     │
│ payment-gateway's PostgreSQL    │
│ connections are maxed out,      │
│ causing "connection refused"    │
│ errors. This started after      │
│ v2.3.1 deployment 45 minutes    │
│ ago. Recommendation: rollback   │
│ v2.3.1 or increase connection   │
│ pool size on payment-gateway.   │
│                                 │
│ Confidence: 87%                 │
│ Steps taken: 4                  │
│ Tokens used: 3,200              │
└─────────────────────────────────┘
```

**Tool Definitions:**

| Tool | Description | Parameters |
|---|---|---|
| `query_events` | Query events with filters | `source`, `event_type`, `severity`, `time_range`, `limit` |
| `search_events` | Full-text or semantic search | `query`, `mode` (keyword/semantic/hybrid), `time_range` |
| `get_metrics` | Retrieve time-series metrics | `service`, `metric`, `window`, `resolution` |
| `list_deployments` | Recent deployments | `service`, `window` |
| `get_service_info` | Service metadata | `service` |
| `get_alert_history` | Recent alerts | `service`, `severity`, `window` |
| `find_similar` | Vector search for similar incidents | `description`, `top_k` |
| `run_aggregation` | Execute analytics query | `metric`, `group_by`, `window`, `function` |

**Guardrails:**
- Maximum 10 steps per investigation
- Maximum 30 seconds wall time
- Maximum 10,000 tokens per investigation
- All tool calls are logged and auditable
- Agent cannot modify data (read-only tools)

### 10.3 Anomaly Detection & Explanation

**Detection (Statistical):**
- Runs on 1-minute aggregation data
- For each metric (error rate, latency p99, event volume) per service:
  - Compute rolling mean and standard deviation over a 24-hour window
  - Flag as anomalous if current value > mean + 3σ (Z-score method)
  - Alternative: IQR method for skewed distributions
- Configurable sensitivity per metric

**Explanation (AI):**
When an anomaly is detected, the AI service:
1. Retrieves the anomalous metric data (current value, historical baseline, when it started)
2. Queries related events around the anomaly start time
3. Checks for recent deployments or infrastructure changes
4. Generates a human-readable explanation:
   > *"Error rate for payment-service spiked to 12% at 14:32 UTC (baseline: 0.5%). This coincides with deployment v2.3.1 at 14:28 UTC. The errors are predominantly 'connection refused' to the downstream database. This suggests the new deployment may have introduced a connection leak."*

### 10.4 Semantic Search (RAG Pipeline)

**Embedding Generation:**
- Model: `all-MiniLM-L6-v2` (384 dimensions) — small, fast, runs locally
- Embedding computed over: concatenation of `source`, `event_type`, `severity`, and a text summary of `data`
- Generated asynchronously by embedding worker (non-blocking to ingestion)
- Stored in `embedding` column with HNSW index

**Retrieval-Augmented Generation:**
When the AI copilot needs historical context:
1. The question/investigation context is embedded
2. Top-K similar events retrieved via vector search
3. Retrieved events are injected into the LLM prompt as context
4. LLM generates a grounded response with citations (event IDs)

**Example:** *"Have we seen this error before?"*
→ Vector search finds similar events from 2 weeks ago
→ LLM: *"Yes, a similar 'connection pool exhaustion' incident occurred on June 21 (events evt_abc, evt_def). It was resolved by increasing the pool size from 20 to 50 connections."*

### 10.5 Background AI Workers

These run on schedules, not in response to user queries.

#### Weekly Health Report
- **Schedule:** Every Monday at 9:00 AM
- **Process:**
  1. Query aggregated metrics for the past 7 days
  2. Identify trends (improving, degrading, stable) per service
  3. List notable incidents and their resolutions
  4. Compute SLA compliance (uptime, error budget)
  5. Generate a structured report with the LLM
- **Output:** Stored in database, accessible via API and dashboard

#### Alert Deduplication
- **Schedule:** Every 5 minutes
- **Process:**
  1. Fetch recent active alerts
  2. Group alerts by semantic similarity (same root cause)
  3. Merge related alerts into alert groups
  4. Reduce notification fatigue

#### Pattern Discovery
- **Schedule:** Every hour
- **Process:**
  1. Analyze recent event patterns
  2. Identify recurring error patterns, cyclical behaviors, or emerging trends
  3. Generate "insights" that surface on the dashboard
  4. Example: *"payment-service errors increase 3x every day between 14:00-15:00 UTC. This correlates with a daily batch job."*

#### Dashboard Suggestions
- **Schedule:** On new service registration or weekly
- **Process:**
  1. Analyze what events a service produces
  2. Suggest dashboard panels (error rate chart, latency histogram, event volume)
  3. Auto-generate dashboard JSON if approved by user

### 10.6 Memory & Context Management

**Within-session memory:**
- Investigation sessions maintain conversation history
- Messages stored in `ai_conversations` table
- Context window managed by truncating old messages and injecting a summary

**Cross-session memory:**
- Resolved investigations stored with summaries and root causes
- Vector-indexed for retrieval when similar incidents occur
- Agent can reference: *"Based on a similar incident on June 21, this was likely caused by..."*

**Context window strategy:**
- System prompt: ~500 tokens (role, capabilities, constraints)
- Schema context: ~300 tokens (table definitions)
- Retrieved context (RAG): ~1000 tokens (similar events/incidents)
- Conversation history: remaining window space
- Prioritize recent messages; summarize older messages

### 10.7 Evaluation Framework

**Why:** AI behavior must be testable and regression-resistant.

**Components:**
- **Test cases:** Labeled pairs of (input_question, expected_query/response)
- **Metrics:**
  - Query generation accuracy (does the generated query return correct results?)
  - Investigation completeness (did the agent find the root cause?)
  - Explanation quality (human-rated, 1-5 scale)
  - Hallucination rate (did the response contain ungrounded claims?)
  - Latency (time to first token, total response time)
- **Process:** Evaluation suite runs in CI on AI-related code changes. Results tracked over time.

### 10.8 Confidence Scoring

Every AI response includes a confidence score (0.0–1.0):

- **1.0:** Direct data retrieval, no inference needed
- **0.8–0.9:** Strong evidence from multiple correlated data points
- **0.5–0.7:** Inference based on partial evidence
- **< 0.5:** Low confidence, speculative — marked with disclaimer

The confidence score is computed by the LLM as part of its structured output (self-assessed) and validated against the number and quality of evidence sources used.

### 10.9 LLM Provider Abstraction

The AI service abstracts LLM provider details behind an interface:

```go
type LLMProvider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    CompleteStream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
    Embed(ctx context.Context, texts []string) ([][]float32, error)
}
```

**Supported providers:**
- **Claude (Anthropic)** — Primary provider for reasoning and tool calling
- **OpenAI** — Alternative provider
- **Local models (Ollama)** — For development and cost-sensitive workloads

Provider selection is configurable per use case. Embedding uses a local model by default (no API cost). Reasoning uses Claude by default.

---

## 11. Observability

### Philosophy

AtlasDB is an observability platform that is itself fully observable. This is not just good engineering — it's a recursive demonstration of the platform's value.

### 11.1 Distributed Tracing (OpenTelemetry)

**Instrumentation:**
- Every inbound HTTP request starts a trace span
- Trace context is propagated via `traceparent` header (W3C Trace Context)
- Spans created for: HTTP handlers, database queries, Redis operations, queue publish/consume, LLM API calls, external HTTP calls
- Span attributes include: `service.name`, `http.method`, `http.route`, `http.status_code`, `db.statement`, `db.duration_ms`

**Exporter:** OpenTelemetry Collector receives spans via OTLP/gRPC and exports to:
- Jaeger or Tempo (trace storage and visualization)
- Stdout in development (JSON-formatted spans)

**Trace example:**
```
[api-server] POST /api/v1/events (12ms)
  └── [api-server] validate_events (0.5ms)
  └── [api-server] redis_xadd (3ms)
      └── [redis] XADD events:2 (1ms)
```

### 11.2 Metrics (Prometheus)

Every service exposes a `/metrics` endpoint in Prometheus exposition format.

**Standard metrics (all services):**

| Metric | Type | Labels |
|---|---|---|
| `http_requests_total` | Counter | `method`, `route`, `status` |
| `http_request_duration_seconds` | Histogram | `method`, `route` |
| `http_requests_in_flight` | Gauge | — |

**Service-specific metrics:**

| Service | Metric | Type |
|---|---|---|
| Ingest | `atlas_events_ingested_total` | Counter |
| Ingest | `atlas_events_ingested_bytes_total` | Counter |
| Ingest | `atlas_ingest_batch_size` | Histogram |
| Processor | `atlas_events_processed_total` | Counter |
| Processor | `atlas_processing_duration_seconds` | Histogram |
| Processor | `atlas_processing_errors_total` | Counter |
| Queue | `atlas_queue_depth` | Gauge |
| Queue | `atlas_queue_oldest_message_age_seconds` | Gauge |
| Queue | `atlas_dlq_messages_total` | Counter |
| Search | `atlas_search_queries_total` | Counter |
| Search | `atlas_search_duration_seconds` | Histogram |
| Search | `atlas_search_results_count` | Histogram |
| AI | `atlas_ai_requests_total` | Counter |
| AI | `atlas_ai_duration_seconds` | Histogram |
| AI | `atlas_ai_tokens_used_total` | Counter |
| AI | `atlas_ai_confidence_score` | Histogram |
| DB | `atlas_db_query_duration_seconds` | Histogram |
| DB | `atlas_db_connections_active` | Gauge |
| DB | `atlas_db_connections_idle` | Gauge |
| Cache | `atlas_cache_hits_total` | Counter |
| Cache | `atlas_cache_misses_total` | Counter |
| Workers | `atlas_worker_jobs_total` | Counter |
| Workers | `atlas_worker_jobs_in_progress` | Gauge |
| Workers | `atlas_worker_job_duration_seconds` | Histogram |

### 11.3 Structured Logging

**Format:** JSON, one object per line.

```json
{
  "timestamp": "2026-07-05T12:00:00.123Z",
  "level": "info",
  "service": "processor",
  "message": "Event processed successfully",
  "trace_id": "abc123def456",
  "span_id": "789ghi",
  "event_id": "01J5M2K...",
  "source": "payment-service",
  "duration_ms": 4.2
}
```

**Levels:** `debug`, `info`, `warn`, `error`
**Correlation:** Every log line includes `trace_id` and `span_id` for correlation with traces.
**Best practices:**
- Never log sensitive data (passwords, tokens, PII)
- Log at appropriate levels (don't spam `info` for per-event processing; use `debug`)
- Include actionable context (what happened, what entity, what was the outcome)

### 11.4 Health Checks

Every service exposes:

- `GET /healthz` — **Liveness probe.** Returns 200 if the process is running. Used by Kubernetes to restart crashed containers.
- `GET /readyz` — **Readiness probe.** Returns 200 if the service can handle traffic (database connected, dependencies reachable). Used by Kubernetes to route traffic.

**Readiness checks:**
- API server: PostgreSQL ping, Redis ping
- Processor: Redis Stream reachable
- AI service: LLM provider reachable (cached check, not per-request)

### 11.5 Grafana Dashboards

Pre-configured dashboards provisioned as JSON:

1. **System Overview** — Request rate, error rate, latency (p50/p95/p99) across all services
2. **Event Pipeline** — Ingestion rate, queue depth, processing rate, processing latency, DLQ depth
3. **Search Performance** — Query rate, query latency, result counts, cache hit rate
4. **AI Performance** — Request rate, latency, tokens used, confidence distribution, cost estimate
5. **Infrastructure** — Pod CPU/memory, database connections, Redis memory, disk usage
6. **Alerts** — Active alerts, alert history, notification delivery success rate

---

## 12. DevOps

### 12.1 Docker

**Dockerfile per service:**
- Multi-stage builds: `builder` stage compiles Go binary, `runner` stage uses `gcr.io/distroless/static-debian12` (minimal, no shell, tiny attack surface)
- Binary is statically compiled (`CGO_ENABLED=0` for pure Go services; CGO required for pgvector client if using cgo bindings)
- Layer ordering optimized: dependencies cached, source code changes only rebuild later layers
- Image size target: < 30MB per service

**Example structure:**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api-server ./cmd/api-server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/api-server /bin/api-server
EXPOSE 8080
ENTRYPOINT ["/bin/api-server"]
```

### 12.2 Docker Compose (Local Development)

Single `docker-compose.yml` that runs the entire platform locally:

```yaml
services:
  api-server
  processor
  ai-service
  postgres        (with pgvector extension)
  redis
  prometheus
  grafana
  otel-collector
  minio
  dashboard       (Next.js frontend)
```

**Features:**
- Health checks on all services
- Volume mounts for database persistence
- Environment variables from `.env.development`
- Hot-reload for frontend (volume mount source code)
- Port mappings for local access (API: 8080, Dashboard: 3000, Grafana: 3001, Prometheus: 9090)

### 12.3 Kubernetes

**Manifests organized by service:**

```
k8s/
├── base/
│   ├── namespace.yaml
│   ├── api-server/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── hpa.yaml
│   │   └── pdb.yaml
│   ├── processor/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── hpa.yaml
│   ├── ai-service/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── postgres/
│   │   ├── statefulset.yaml
│   │   ├── service.yaml
│   │   └── pvc.yaml
│   ├── redis/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── monitoring/
│   │   ├── prometheus/
│   │   ├── grafana/
│   │   └── otel-collector/
│   └── ingress.yaml
└── overlays/
    ├── development/
    ├── staging/
    └── production/
```

**Key configurations:**
- **HPA** on API server: scale 2–10 pods based on CPU (target 70%) and custom metric (request latency p99 > 200ms)
- **HPA** on processor: scale 2–8 pods based on custom metric (queue depth > 1000)
- **PodDisruptionBudget:** min 1 available for API server during rolling updates
- **Resource requests/limits:** set for all containers to enable bin-packing and prevent noisy neighbors
- **Liveness/Readiness probes:** configured per service with appropriate thresholds
- **ConfigMaps:** non-secret configuration (feature flags, tuning parameters)
- **Secrets:** database credentials, API keys, JWT signing keys

### 12.4 CI/CD (GitHub Actions)

#### PR Pipeline (on pull request)

```yaml
jobs:
  lint:        golangci-lint, eslint (frontend)
  test:        go test ./... -race, jest (frontend)
  build:       go build (verify compilation)
  security:    govulncheck, trivy (container scan)
```

#### Deploy Pipeline (on merge to main)

```yaml
jobs:
  test:        full test suite
  build:       docker build + tag with git SHA
  push:        push to container registry (GHCR)
  deploy-staging:
    - kubectl apply (staging overlay)
    - wait for rollout
    - run smoke tests against staging
  deploy-production:
    - manual approval gate (for production)
    - kubectl apply (production overlay)
    - wait for rollout
    - run health checks
    - if health check fails: kubectl rollout undo
```

### 12.5 Terraform

Infrastructure as code for cloud resources:

```
terraform/
├── modules/
│   ├── networking/     VPC, subnets, security groups
│   ├── kubernetes/     Managed K8s cluster (EKS/GKE)
│   ├── database/       Managed PostgreSQL (RDS/Cloud SQL)
│   ├── redis/          Managed Redis (ElastiCache/Memorystore)
│   ├── storage/        S3 bucket for backups
│   └── iam/            Service accounts and roles
├── environments/
│   ├── staging/
│   └── production/
├── backend.tf          Remote state (S3 + DynamoDB lock)
└── variables.tf
```

### 12.6 Environment Configuration

| Environment | Database | Redis | Replicas | Features |
|---|---|---|---|---|
| `development` | Local Docker | Local Docker | 1 each | Hot reload, debug logging, no auth required |
| `staging` | Managed (small) | Managed (small) | 2 each | Full auth, staging API keys, synthetic data |
| `production` | Managed (HA) | Managed (HA) | 3+ each | Full security, real data, monitoring alerts |

### 12.7 Deployment Strategies

- **Rolling (default):** Kubernetes default. New pods created, old pods terminated one at a time. Zero downtime. Used for all services.
- **Blue/Green (critical changes):** Two full deployments maintained. Traffic switched at load balancer level. Instant rollback by switching back. Used for database schema migrations and major version bumps.

### 12.8 Secrets Management

- Development: `.env.development` file (gitignored)
- Kubernetes: `Secret` resources, populated via:
  - CI/CD pipeline (secrets stored in GitHub Secrets)
  - External Secrets Operator (for production, pulling from AWS Secrets Manager or Vault)
- Rotation: API keys and database passwords rotatable without downtime via rolling restart

---

## 13. Security

### 13.1 Authentication

**User Authentication:**
- Registration with email and password
- Password hashed with Argon2id (memory-hard, resistant to GPU attacks)
- Login returns JWT access token (15 min TTL) + refresh token (7 day TTL, stored in HttpOnly cookie)
- JWT signed with RS256 (asymmetric — services can verify without the signing key)
- Refresh token rotation: each use invalidates the old refresh token and issues a new pair

**API Key Authentication:**
- Generated via dashboard or API (requires authenticated user)
- Format: `atlas_` + 32-char random string (e.g., `atlas_sk_a1b2c3d4e5f6...`)
- Stored as bcrypt hash in database; only the prefix stored in cleartext for identification
- Scoped permissions: each key has a set of allowed scopes
- Expiration: optional, configurable per key

### 13.2 Authorization

**Role-Based Access Control (RBAC):**

| Role | Permissions |
|---|---|
| `admin` | Full access: manage users, manage alert rules, manage API keys, all read/write |
| `editor` | Create/edit alert rules, use AI features, full read access |
| `viewer` | Read-only: view events, search, view dashboards, view alerts |

**Scope-Based (API Keys):**
- `events:write` — ingest events
- `events:read` — query events
- `search:read` — search events
- `alerts:manage` — CRUD alert rules
- `ai:query` — use AI features
- `admin:all` — full access (admin API keys only)

### 13.3 Rate Limiting

- **Per-user rate limiting:** Token bucket algorithm, stored in Redis
- **Limits:**
  - Event ingestion: 10,000 events/minute per API key
  - Search queries: 100 queries/minute per user
  - AI queries: 20 queries/minute per user (LLM cost control)
  - Auth endpoints: 10 attempts/minute per IP (brute force protection)
- **Response:** `429 Too Many Requests` with `Retry-After` header and `X-RateLimit-Remaining` header

### 13.4 Audit Logging

Every mutation and sensitive operation generates an audit event:
- User login/logout
- API key creation/revocation
- Alert rule changes
- AI investigation queries
- Role/permission changes

Audit events are stored as regular events in AtlasDB (source: `atlas-audit`), making them searchable and analyzable with the same tools.

### 13.5 Encryption

- **In transit:** TLS 1.3 for all external connections. TLS for database and Redis connections in production.
- **At rest:** Database encryption via cloud provider's managed encryption. Backup encryption in S3 (SSE-S3 or SSE-KMS).
- **Sensitive fields:** API key values are bcrypt-hashed. Passwords are Argon2id-hashed. JWT signing keys are stored in Kubernetes Secrets.

### 13.6 Input Validation

- All API inputs validated against defined schemas
- SQL injection prevention via parameterized queries (never string concatenation)
- JSON payload size limits (1MB per event, 10MB per batch)
- Rate limiting on failed authentication attempts
- CORS configuration for dashboard origin only

---

## 14. Performance

### 14.1 Connection Pooling

- **PostgreSQL:** PgBouncer in transaction pooling mode. Pool size: 20 connections per API server instance. Prevents connection exhaustion under high load.
- **Redis:** Go Redis client with built-in connection pool. Pool size: 10 connections per service instance.
- **HTTP clients (outbound):** Reusable `http.Client` with connection pooling and keep-alive.

### 14.2 Caching

**Multi-layer caching strategy:**

1. **In-process LRU cache** — For immutable or slowly-changing data (service metadata, alert rules). Eliminates database round-trips for hot data. Invalidated via Redis pub/sub on change.

2. **Redis cache** — For query results with short TTL. Key structure: `cache:query:<hash_of_query_params>`. TTL: 30–60 seconds. Reduces database load for repeated queries (e.g., dashboard refreshing every 10 seconds).

3. **HTTP cache headers** — `Cache-Control` headers on analytics endpoints. Allows browser and CDN caching for static dashboard data.

### 14.3 Compression

- **HTTP responses:** gzip compression via middleware for responses > 1KB. Reduces bandwidth by ~70% for JSON payloads.
- **Event storage:** JSONB in PostgreSQL is stored in a decomposed binary format (more efficient than text JSON). TOAST compression for large fields.
- **Queue messages:** Events serialized as MessagePack (more compact than JSON) for Redis Stream entries.

### 14.4 Batch Operations

- **Batch ingestion:** Multi-row `INSERT` for event writes (up to 100 events per statement). Reduces per-event overhead by 10x.
- **Batch embedding:** Embedding model processes 32 texts per batch. Amortizes model loading overhead.
- **Batch queue reads:** Consumer reads 100 messages per `XREADGROUP` call. Reduces Redis round-trips.

### 14.5 Pagination

- **Cursor-based:** All list endpoints use cursor-based pagination. Cursor encodes `(timestamp, event_id)`. Consistent performance at any depth.
- **Page size:** Default 50, max 1000. Configurable per request.

### 14.6 Query Optimization

- **Partition pruning:** All event queries require a time range (enforced by API). Queries that span fewer partitions are faster.
- **Covering indices:** Composite indices include commonly selected columns to enable index-only scans.
- **Prepared statements:** Frequently executed queries use prepared statements, eliminating repeated query parsing.
- **Query plan caching:** PostgreSQL caches query plans for prepared statements across executions.
- **EXPLAIN analysis:** During development, all queries are analyzed with `EXPLAIN (ANALYZE, BUFFERS)` to ensure index usage.

### 14.7 Horizontal Scaling

| Component | Scaling Mechanism |
|---|---|
| API server | Add replicas (stateless). Load balancer distributes requests. |
| Processors | Add consumer group members. Redis Streams rebalances partitions. |
| Database reads | Read replicas. Route analytics queries to replicas. |
| Workers | Add worker instances. Job queue distributes work automatically. |
| Search | Read replicas for query load. Eventual consistency acceptable. |

### 14.8 Performance Testing

- **Load testing:** k6 scripts for simulating realistic workloads (event ingestion, search queries, dashboard refreshes).
- **Benchmarks:** Go benchmarks for critical hot paths (event validation, serialization, query execution).
- **Profiling:** `pprof` endpoints exposed on all Go services (protected by admin auth). CPU and memory profiling for bottleneck identification.
- **Targets:**
  - Event ingestion: > 10,000 events/second (single instance)
  - Search query: < 50ms p99
  - Analytics query: < 100ms p99
  - AI query (NL → results): < 3s p99

---

## 15. Frontend

### Technology Stack

- **Framework:** Next.js 15 (App Router, Server Components)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **Charts:** Recharts (time-series, bar, pie)
- **Data Fetching:** TanStack Query (SWR caching, polling, optimistic updates)
- **State Management:** Zustand (lightweight, minimal boilerplate)
- **Real-Time:** Native WebSocket client
- **UI Components:** shadcn/ui (Radix primitives, accessible, customizable)

### Design Principles

- **Dark mode first:** Analytics dashboards are predominantly used in dark mode. Light mode supported but dark is default.
- **Information density:** Show more data, fewer decorations. Inspired by Grafana/Datadog density.
- **Real-time by default:** Dashboard auto-refreshes. Live event stream visible. No manual refresh needed.
- **Keyboard navigable:** Power users expect keyboard shortcuts (/ for search, ? for help, k for command palette).

### Page Specifications

#### 15.1 Overview Dashboard (`/`)

The landing page. Shows system health at a glance.

**Components:**
- **Health indicator:** Green/yellow/red per service with last heartbeat time
- **Event volume chart:** Time-series line chart, last 24 hours, with 1-minute granularity
- **Error rate chart:** Per-service error rate, overlaid on single chart
- **Active alerts counter:** Badge showing firing alert count, clickable to alerts page
- **Recent events table:** Last 20 events, auto-updating via WebSocket
- **AI insights panel:** Latest pattern discoveries and recommendations from background AI workers

#### 15.2 Live Event Stream (`/events`)

Real-time event viewer.

**Components:**
- **Virtualized event list:** Displays thousands of events without DOM bloat (react-virtualized)
- **Auto-scroll:** New events appear at top, auto-pauses when user scrolls up
- **Quick filters:** Toggle by severity (info/warn/error), by source (dropdown), by time
- **Event detail panel:** Click event to see full payload, metadata, and tags in a side panel
- **Pause/resume:** Pause live stream to inspect without events scrolling away

#### 15.3 Analytics (`/analytics`)

Interactive analytics dashboard.

**Components:**
- **Time range selector:** Preset ranges (15m, 1h, 6h, 24h, 7d, 30d) and custom
- **Event volume over time:** Stacked area chart by severity or source
- **Top sources:** Bar chart of most active event sources
- **Error rate over time:** Line chart per service
- **Latency percentiles:** p50, p95, p99 line chart
- **Top endpoints:** Table of endpoints by request count and average latency

#### 15.4 Search (`/search`)

Unified search interface.

**Components:**
- **Search bar:** Full-width, prominent. Supports keyword and filter syntax.
- **Mode toggle:** Keyword / Semantic / Hybrid
- **Results list:** Paginated, with highlighted matching terms
- **Filters panel:** Sidebar with facets (source, severity, type, time range)
- **Saved searches:** Save and recall frequent queries

#### 15.5 Alerts (`/alerts`)

Alert management.

**Components:**
- **Active alerts list:** Currently firing alerts with severity badge, time, affected service
- **Alert rules list:** All configured rules with enable/disable toggle
- **Create/edit rule form:** Metric, condition, threshold, window, notification channels
- **Alert timeline:** Gantt-chart-style view of alert fire/resolve over time
- **Alert detail:** Click to see alert history, affected events, AI explanation

#### 15.6 AI Copilot (`/ai`)

Interactive AI assistant.

**Components:**
- **Chat interface:** Message-based UI (similar to ChatGPT). User types questions, AI responds with formatted results.
- **Query transparency:** When AI generates a query, show it in a collapsible block before the results.
- **Investigation mode:** "Investigate" button starts an AI investigation. Shows steps in real-time (tool calls, intermediate results, final summary).
- **Suggested queries:** Pre-populated question buttons for common queries ("What happened in the last hour?", "Any anomalies today?", "Compare error rates this week vs last week")
- **Streaming responses:** Responses stream word-by-word via SSE.

#### 15.7 System Health (`/system`)

AtlasDB's own health dashboard.

**Components:**
- **Service status grid:** Each service with health indicator, version, uptime
- **Embedded Grafana panels:** Key metrics (CPU, memory, latency) rendered via Grafana iframe or API
- **Queue depth gauge:** Current event queue depth with historical sparkline
- **Database stats:** Active connections, query rate, replication lag

#### 15.8 Settings (`/settings`)

Configuration management.

**Components:**
- **Profile:** User email, password change
- **API keys:** Generate, list, revoke API keys
- **Team management:** (admin only) Invite users, assign roles
- **Notification channels:** Configure webhook URLs, email settings
- **Data retention:** Configure partition retention period
- **AI configuration:** Select LLM provider, adjust token budgets

---

## 16. Engineering Metrics

These are the internal metrics that AtlasDB tracks about its own performance. They are exposed via Prometheus and visualized in Grafana.

### API Metrics

| Metric | Description | Alert Threshold |
|---|---|---|
| `http_request_duration_seconds` (p99) | API response latency | > 500ms for 5 minutes |
| `http_requests_total` (rate) | Request throughput | Drop > 50% from baseline |
| `http_requests_total{status=~"5.."}` (rate) | Server error rate | > 1% of total requests |

### Pipeline Metrics

| Metric | Description | Alert Threshold |
|---|---|---|
| `atlas_queue_depth` | Messages waiting in queue | > 10,000 for 5 minutes |
| `atlas_events_ingested_total` (rate) | Ingestion throughput | — |
| `atlas_events_processed_total` (rate) | Processing throughput | Lag > 1000 behind ingestion |
| `atlas_processing_duration_seconds` (p99) | Per-event processing time | > 1 second |
| `atlas_dlq_messages_total` (rate) | Dead-letter queue additions | > 0 (any DLQ entry is notable) |

### Search Metrics

| Metric | Description | Alert Threshold |
|---|---|---|
| `atlas_search_duration_seconds` (p99) | Search query latency | > 200ms |
| `atlas_search_queries_total` (rate) | Search throughput | — |

### AI Metrics

| Metric | Description | Alert Threshold |
|---|---|---|
| `atlas_ai_duration_seconds` (p99) | AI response latency | > 10 seconds |
| `atlas_ai_tokens_used_total` (rate) | Token consumption rate | > budget threshold |
| `atlas_ai_confidence_score` (avg) | Average confidence | < 0.5 |

### Cache Metrics

| Metric | Description | Alert Threshold |
|---|---|---|
| `atlas_cache_hits_total / (hits + misses)` | Cache hit rate | < 70% |

### Infrastructure Metrics

| Metric | Description | Alert Threshold |
|---|---|---|
| Container CPU utilization | Per-pod CPU | > 80% sustained |
| Container memory utilization | Per-pod memory | > 85% sustained |
| `atlas_db_connections_active` | Active DB connections | > 80% of pool size |
| `atlas_worker_jobs_in_progress` | Active workers | > 90% of worker pool |

---

## 17. Demo Scenarios

Each scenario demonstrates AtlasDB handling a realistic operational situation. These are designed for live demos or interview walkthroughs.

### Scenario 1: Traffic Spike

**Setup:** A marketing campaign drives 10x normal traffic.

**What happens in AtlasDB:**
1. Event ingestion rate spikes from 1,000/s to 10,000/s
2. Queue depth increases as processors catch up
3. HPA scales processor pods from 2 to 6 based on queue depth metric
4. Dashboard shows the traffic spike in real-time
5. Analytics correctly report elevated request volume
6. AI insight: *"Traffic spike detected at 14:00 UTC. Volume 10x above baseline. Pattern consistent with external traffic source (no corresponding error rate increase). No action required."*

### Scenario 2: Distributed Outage

**Setup:** A database goes down, causing cascading failures across services.

**What happens in AtlasDB:**
1. Error events flood in from 5 services: "connection refused", "timeout", "circuit breaker open"
2. Alert rules fire for error rate thresholds on multiple services
3. AI alert deduplication groups the alerts: *"5 related alerts detected — likely single root cause"*
4. User asks AI: *"What's causing the outage?"*
5. AI agent investigates:
   - Queries error events → all reference `orders-db.internal:5432`
   - Checks service dependencies → all affected services depend on orders-db
   - Checks infrastructure events → `orders-db pod restarted at 14:32`
   - Summary: *"Root cause: orders-db pod restart at 14:32 UTC caused connection failures for 5 downstream services. Services auto-recovered by 14:35 when the database completed startup. Confidence: 92%."*

### Scenario 3: Database Slowdown

**Setup:** A missing index causes query latency to degrade gradually.

**What happens in AtlasDB:**
1. Latency percentiles for `user-service` gradually increase over 2 hours
2. Anomaly detection flags: p99 latency is 5σ above baseline
3. Alert fires: *"user-service p99 latency > 2s (threshold: 500ms)"*
4. AI explanation: *"user-service latency has been degrading since 10:00 UTC. The degradation is gradual, not sudden, suggesting a data-dependent issue rather than a deployment. Top slow endpoints: GET /api/users/search. Recommendation: check for missing database indices on the users table."*

### Scenario 4: API Failures

**Setup:** A third-party payment API starts returning 503 errors.

**What happens in AtlasDB:**
1. Events from `payment-service` show spike in `status: 503` with external API URL
2. Alert fires for payment-service error rate
3. AI investigation identifies the pattern: *"payment-service errors are concentrated on calls to api.stripe.com. Error messages indicate 'Service Unavailable'. This is a third-party dependency failure, not an internal issue. Recommendation: enable circuit breaker for Stripe API calls and fall back to queued processing."*

### Scenario 5: Fraud Detection

**Setup:** A compromised account performs rapid-fire transactions.

**What happens in AtlasDB:**
1. Transaction events show user_123 performing 47 transactions in 2 minutes from 3 IP addresses in different countries
2. Anomaly detection flags: transaction volume for user_123 is 50σ above their historical average
3. AI analysis: *"User user_123 performed 47 transactions in 120 seconds from US, Romania, and Singapore. Historical average: 3 transactions/day from a single location. Confidence of anomalous behavior: 98.7%. Recommended action: freeze account and trigger fraud review."*

### Scenario 6: Login Attack

**Setup:** A brute-force attack on authentication endpoints.

**What happens in AtlasDB:**
1. Massive spike in `event_type: authentication, data.result: failure` events
2. Events show thousands of login attempts from a small set of IP addresses
3. Alert fires: *"Authentication failure rate > 100/minute"*
4. AI analysis: *"Detected potential brute-force attack. 4,200 failed login attempts in 10 minutes from 3 IP addresses (185.x.x.x, 91.x.x.x, 45.x.x.x). Targeted accounts: top 50 by attempt count. No successful breaches detected. Recommendation: block source IPs and enforce CAPTCHA."*

### Scenario 7: Memory Leak

**Setup:** A new deployment introduces a memory leak that slowly consumes resources.

**What happens in AtlasDB:**
1. Infrastructure events show `catalog-service` memory usage increasing linearly: 200MB → 500MB → 1.2GB over 6 hours
2. Anomaly detection flags the trend
3. AI analysis: *"catalog-service memory usage has increased linearly since deployment v3.1.0 at 08:00 UTC. Current usage: 1.2GB (limit: 2GB). At current rate, OOMKill expected in ~4 hours. This pattern is consistent with a memory leak introduced in the latest deployment. Recommendation: rollback to v3.0.9 and investigate memory allocation in v3.1.0."*

---

## 18. Folder Structure

```
atlasdb/
├── cmd/                              # Application entrypoints
│   ├── api-server/
│   │   └── main.go                   # API server entrypoint
│   ├── processor/
│   │   └── main.go                   # Stream processor entrypoint
│   ├── ai-service/
│   │   └── main.go                   # AI service entrypoint
│   ├── worker/
│   │   └── main.go                   # Background worker entrypoint
│   └── scheduler/
│       └── main.go                   # Scheduler entrypoint
│
├── internal/                         # Private application code
│   ├── api/                          # HTTP handlers, middleware, routing
│   │   ├── handlers/
│   │   │   ├── events.go
│   │   │   ├── search.go
│   │   │   ├── analytics.go
│   │   │   ├── alerts.go
│   │   │   ├── ai.go
│   │   │   ├── auth.go
│   │   │   └── system.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── ratelimit.go
│   │   │   ├── logging.go
│   │   │   ├── tracing.go
│   │   │   ├── cors.go
│   │   │   └── recovery.go
│   │   └── router.go
│   │
│   ├── auth/                         # Authentication & authorization
│   │   ├── jwt.go
│   │   ├── apikey.go
│   │   ├── password.go
│   │   └── rbac.go
│   │
│   ├── ingest/                       # Event ingestion logic
│   │   ├── ingester.go
│   │   ├── validator.go
│   │   └── batcher.go
│   │
│   ├── processor/                    # Stream processing logic
│   │   ├── consumer.go
│   │   ├── enricher.go
│   │   ├── normalizer.go
│   │   └── pipeline.go
│   │
│   ├── search/                       # Search engine
│   │   ├── engine.go
│   │   ├── keyword.go
│   │   ├── semantic.go
│   │   ├── hybrid.go
│   │   ├── parser.go                 # Query syntax parser
│   │   └── ranking.go
│   │
│   ├── analytics/                    # Analytics engine
│   │   ├── aggregator.go
│   │   ├── rollup.go
│   │   └── queries.go
│   │
│   ├── alerts/                       # Alert system
│   │   ├── engine.go
│   │   ├── evaluator.go
│   │   ├── notifier.go
│   │   └── channels/
│   │       ├── webhook.go
│   │       ├── email.go
│   │       └── slack.go
│   │
│   ├── ai/                           # AI layer
│   │   ├── service.go                # AI service orchestrator
│   │   ├── nlquery/                  # Natural language → query
│   │   │   ├── translator.go
│   │   │   └── prompts.go
│   │   ├── agent/                    # Investigation agent
│   │   │   ├── agent.go
│   │   │   ├── tools.go
│   │   │   ├── planner.go
│   │   │   └── memory.go
│   │   ├── anomaly/                  # Anomaly detection & explanation
│   │   │   ├── detector.go
│   │   │   └── explainer.go
│   │   ├── embeddings/               # Embedding generation
│   │   │   └── embedder.go
│   │   ├── workers/                  # Background AI workers
│   │   │   ├── reporter.go
│   │   │   ├── deduplicator.go
│   │   │   └── pattern_finder.go
│   │   ├── llm/                      # LLM provider abstraction
│   │   │   ├── provider.go           # Interface
│   │   │   ├── claude.go
│   │   │   ├── openai.go
│   │   │   └── ollama.go
│   │   └── eval/                     # AI evaluation framework
│   │       ├── evaluator.go
│   │       └── testcases/
│   │
│   ├── storage/                      # Data access layer
│   │   ├── postgres/
│   │   │   ├── events.go
│   │   │   ├── users.go
│   │   │   ├── alerts.go
│   │   │   ├── ai.go
│   │   │   └── migrations/
│   │   │       ├── 001_initial.sql
│   │   │       ├── 002_search_indices.sql
│   │   │       ├── 003_vector_extension.sql
│   │   │       └── 004_aggregation_tables.sql
│   │   ├── redis/
│   │   │   ├── cache.go
│   │   │   ├── queue.go
│   │   │   ├── pubsub.go
│   │   │   └── lock.go
│   │   └── s3/
│   │       └── client.go
│   │
│   ├── queue/                        # Message queue abstraction
│   │   ├── producer.go
│   │   ├── consumer.go
│   │   └── dlq.go
│   │
│   ├── worker/                       # Background job system
│   │   ├── pool.go
│   │   ├── job.go
│   │   └── scheduler.go
│   │
│   ├── telemetry/                    # Observability
│   │   ├── tracing.go
│   │   ├── metrics.go
│   │   └── logging.go
│   │
│   └── config/                       # Configuration
│       └── config.go
│
├── pkg/                              # Public/shared packages
│   ├── models/                       # Shared data models
│   │   ├── event.go
│   │   ├── user.go
│   │   ├── alert.go
│   │   └── pagination.go
│   └── client/                       # Go SDK / client library
│       └── client.go
│
├── frontend/                         # Next.js dashboard
│   ├── src/
│   │   ├── app/                      # App Router pages
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx              # Overview dashboard
│   │   │   ├── events/
│   │   │   ├── analytics/
│   │   │   ├── search/
│   │   │   ├── alerts/
│   │   │   ├── ai/
│   │   │   ├── system/
│   │   │   └── settings/
│   │   ├── components/
│   │   │   ├── ui/                   # shadcn/ui components
│   │   │   ├── charts/
│   │   │   ├── events/
│   │   │   ├── search/
│   │   │   ├── ai/
│   │   │   └── layout/
│   │   ├── lib/
│   │   │   ├── api.ts                # API client
│   │   │   ├── websocket.ts          # WebSocket client
│   │   │   └── utils.ts
│   │   ├── hooks/
│   │   │   ├── use-events.ts
│   │   │   ├── use-search.ts
│   │   │   └── use-ai.ts
│   │   └── stores/
│   │       └── app-store.ts
│   ├── public/
│   ├── package.json
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   └── next.config.ts
│
├── deploy/                           # Deployment configurations
│   ├── docker/
│   │   ├── Dockerfile.api-server
│   │   ├── Dockerfile.processor
│   │   ├── Dockerfile.ai-service
│   │   ├── Dockerfile.worker
│   │   ├── Dockerfile.frontend
│   │   └── Dockerfile.scheduler
│   ├── docker-compose.yml
│   ├── docker-compose.dev.yml
│   └── docker-compose.monitoring.yml
│
├── k8s/                              # Kubernetes manifests
│   ├── base/
│   │   ├── namespace.yaml
│   │   ├── api-server/
│   │   ├── processor/
│   │   ├── ai-service/
│   │   ├── worker/
│   │   ├── scheduler/
│   │   ├── postgres/
│   │   ├── redis/
│   │   ├── monitoring/
│   │   └── ingress.yaml
│   └── overlays/
│       ├── development/
│       ├── staging/
│       └── production/
│
├── terraform/                        # Infrastructure as code
│   ├── modules/
│   │   ├── networking/
│   │   ├── kubernetes/
│   │   ├── database/
│   │   ├── redis/
│   │   ├── storage/
│   │   └── iam/
│   ├── environments/
│   │   ├── staging/
│   │   └── production/
│   └── backend.tf
│
├── scripts/                          # Utility scripts
│   ├── setup.sh                      # Local development setup
│   ├── seed.sh                       # Seed database with sample data
│   ├── generate-events.go            # Event generator for testing
│   ├── migrate.sh                    # Run database migrations
│   └── load-test.js                  # k6 load test script
│
├── tests/                            # Integration and E2E tests
│   ├── integration/
│   │   ├── ingest_test.go
│   │   ├── search_test.go
│   │   ├── pipeline_test.go
│   │   └── ai_test.go
│   └── e2e/
│       ├── smoke_test.go
│       └── scenarios/
│
├── monitoring/                       # Observability configurations
│   ├── prometheus/
│   │   └── prometheus.yml
│   ├── grafana/
│   │   ├── provisioning/
│   │   │   ├── datasources/
│   │   │   └── dashboards/
│   │   └── dashboards/
│   │       ├── overview.json
│   │       ├── pipeline.json
│   │       ├── search.json
│   │       ├── ai.json
│   │       └── infrastructure.json
│   └── otel/
│       └── otel-collector-config.yaml
│
├── .github/
│   └── workflows/
│       ├── ci.yml                    # PR checks
│       └── deploy.yml                # Deployment pipeline
│
├── go.mod
├── go.sum
├── Makefile                          # Build and development commands
├── .env.example                      # Example environment variables
├── .gitignore
├── SPEC.md                           # This specification
└── README.md                         # Project documentation
```

---

## 19. Development Roadmap

### Phase 1: Foundation (Weeks 1–2)

**Goal:** Core event ingestion, storage, and basic querying. The system can receive events, store them, and retrieve them.

**Deliverables:**
- [ ] Go project structure with `cmd/` and `internal/` layout
- [ ] Configuration loading from environment variables
- [ ] PostgreSQL schema with partitioned events table, migrations
- [ ] Redis Streams setup (producer and consumer)
- [ ] Event ingestion API (`POST /api/v1/events`)
- [ ] Event query API (`GET /api/v1/events` with filters, pagination)
- [ ] Stream processor (consume → validate → store)
- [ ] Basic structured logging
- [ ] Docker Compose for local development (API + PostgreSQL + Redis)
- [ ] Makefile with `make dev`, `make test`, `make build`
- [ ] Unit tests for core packages

**Milestone:** Can send events via curl, see them stored in PostgreSQL, and query them back.

### Phase 2: Search & Auth (Weeks 3–4)

**Goal:** Full-text search, authentication, and the beginnings of the dashboard.

**Deliverables:**
- [ ] Full-text search with `tsvector`/`tsquery` and GIN index
- [ ] Search API (`GET /api/v1/search`)
- [ ] Query parser for search syntax (`field:value`, boolean operators)
- [ ] User registration and login
- [ ] JWT authentication with access/refresh tokens
- [ ] API key generation and validation
- [ ] Auth middleware (protect all endpoints)
- [ ] Rate limiting middleware
- [ ] Next.js frontend scaffolding
- [ ] Login page
- [ ] Overview dashboard page (event volume chart, recent events)
- [ ] Search page with results
- [ ] WebSocket endpoint for live event streaming

**Milestone:** Can log in, search events, and see a live dashboard.

### Phase 3: Analytics & Alerts (Weeks 5–6)

**Goal:** Pre-aggregated analytics, alert system, and a polished dashboard.

**Deliverables:**
- [ ] Aggregation tables (`event_counts_1m`, `latency_stats_1m`)
- [ ] Stream processor writes to aggregation tables
- [ ] Rollup worker (1-minute → 1-hour aggregations)
- [ ] Analytics API (time-series, top-N, summary)
- [ ] Analytics dashboard page (charts, time range selector)
- [ ] Alert rule CRUD API
- [ ] Alert evaluation engine (periodic condition checks)
- [ ] Alert notification dispatch (webhook)
- [ ] Alerts page (active alerts, rules management)
- [ ] Live event stream page with virtualized list
- [ ] Background worker framework with job queue

**Milestone:** Dashboard shows real-time analytics. Alerts fire and notify via webhook.

### Phase 4: AI Layer (Weeks 7–9)

**Goal:** AI copilot with natural language queries, semantic search, and investigation agent.

**Deliverables:**
- [ ] LLM provider abstraction (Claude and OpenAI implementations)
- [ ] Natural language → query translation
- [ ] AI query API (`POST /api/v1/ai/query`)
- [ ] Embedding generation worker
- [ ] pgvector setup and HNSW index
- [ ] Semantic search implementation
- [ ] Hybrid search (keyword + semantic with RRF)
- [ ] Investigation agent with tool calling
- [ ] Investigation API (`POST /api/v1/ai/investigate`)
- [ ] Anomaly detection (Z-score on aggregation data)
- [ ] Anomaly explanation with LLM
- [ ] AI copilot chat page
- [ ] Response streaming via SSE
- [ ] Confidence scoring

**Milestone:** Can ask natural language questions, get AI-powered answers. Investigation agent can autonomously diagnose issues.

### Phase 5: Observability & DevOps (Weeks 10–11)

**Goal:** Full observability stack, Docker optimization, Kubernetes manifests, CI/CD.

**Deliverables:**
- [ ] OpenTelemetry SDK integration (tracing for all services)
- [ ] Prometheus metrics on all services
- [ ] Grafana dashboards (provisioned as JSON)
- [ ] OpenTelemetry Collector configuration
- [ ] Health check endpoints (liveness, readiness)
- [ ] Multi-stage Dockerfiles for all services
- [ ] Docker Compose with full observability stack
- [ ] Kubernetes manifests (Deployments, Services, HPA, PDB)
- [ ] Kustomize overlays for dev/staging/production
- [ ] GitHub Actions CI pipeline (lint, test, build, scan)
- [ ] GitHub Actions deploy pipeline (build, push, deploy)
- [ ] System health dashboard page

**Milestone:** Full observability. Services instrumented with traces and metrics. Grafana dashboards live. CI/CD pipeline operational.

### Phase 6: Hardening & Polish (Weeks 12–13)

**Goal:** Production-grade hardening, background AI workers, performance testing, demo scenarios.

**Deliverables:**
- [ ] Background AI workers (weekly report, pattern discovery, alert deduplication)
- [ ] Dead-letter queue monitoring and admin API
- [ ] Audit logging
- [ ] Cache layer (Redis caching for query results)
- [ ] Connection pooling optimization
- [ ] Batch write optimization
- [ ] k6 load test scripts
- [ ] Event generator script (for demo data)
- [ ] Demo scenario data generators (traffic spike, outage, etc.)
- [ ] Settings page
- [ ] Terraform modules (at least one cloud — e.g., AWS with RDS, ElastiCache, EKS)
- [ ] Integration tests
- [ ] Documentation (README with architecture diagram, setup instructions, API reference)

**Milestone:** Production-ready. Can run demo scenarios end-to-end. Load tested. Documented.

---

## 20. Stretch Features

Features inspired by leading infrastructure companies.

### Datadog-Inspired

- **Service Map:** Automatically discover service dependencies from event data. Visualize as an interactive graph showing traffic flow, error rates, and latency between services. *Technical depth: graph construction from call data, real-time rendering with D3.js or similar.*

- **Log Patterns:** Automatically group similar log messages into patterns (e.g., cluster 1,000 "connection timeout to X.X.X.X:5432" messages into a single pattern). Reduces noise. *Technical depth: text clustering using locality-sensitive hashing or TF-IDF + k-means.*

### Snowflake-Inspired

- **Query History & Cost Tracking:** Track every query executed against AtlasDB with execution time, rows scanned, and (for AI queries) token cost. Expose as a queryable audit log. *Technical depth: query lifecycle instrumentation, cost attribution.*

- **Time Travel:** Query events as they existed at a previous point in time. Since events are immutable and partitioned by time, this is naturally supported. Extend to allow querying "what did the aggregation look like at 14:00 yesterday?" using point-in-time snapshots. *Technical depth: snapshot isolation, temporal queries.*

### Elastic-Inspired

- **Index Lifecycle Management:** Automatic lifecycle policies for event data: hot (recent, fast storage) → warm (older, compressed) → cold (archived to object storage) → delete. *Technical depth: tiered storage, automatic data movement, compression strategies.*

- **Custom Analyzers:** Allow users to define custom text analyzers for search (tokenizers, filters, character mappings) for domain-specific event data. *Technical depth: pluggable text analysis pipeline.*

### OpenAI / Anthropic-Inspired

- **Prompt Playground:** An interface for testing and iterating on AI prompts used by the system. Adjust system prompts, temperature, and see how they affect responses. Version control for prompts. *Technical depth: prompt versioning, A/B testing, evaluation harness.*

- **Function Calling Marketplace:** Allow users to register custom tools that the AI agent can call during investigations. Example: a user registers a `restart_service` tool so the AI can not only diagnose but remediate. *Technical depth: dynamic tool registration, sandboxed execution, approval workflows.*

### Stripe-Inspired

- **Idempotency Keys:** Support `Idempotency-Key` header on event ingestion. Guarantee that retried requests (network errors, timeouts) don't create duplicate events. *Technical depth: idempotency key storage, conflict detection, cleanup.*

- **Webhook Delivery with Retry:** For alert notifications, implement Stripe-style webhook delivery: signed payloads, delivery attempts with exponential backoff, delivery status dashboard, manual retry. *Technical depth: reliable webhook delivery, signature verification, delivery tracking.*

### Netflix-Inspired

- **Chaos Events:** Inject synthetic failure events to test alert rules and AI investigation capabilities. "What happens if payment-service starts throwing 500s?" Simulate without affecting real systems. *Technical depth: synthetic event generation, scenario modeling.*

- **Adaptive Alerting:** Alert thresholds that automatically adjust based on historical patterns. If a service naturally has higher error rates on Mondays, the threshold adapts. *Technical depth: seasonal decomposition, adaptive baselines.*

### Uber-Inspired

- **Event Sampling:** At very high volumes, intelligently sample events (keep all errors, sample 10% of info events) to manage storage costs while preserving signal. *Technical depth: priority-based sampling, reservoir sampling, maintaining statistical validity.*

### Cloudflare-Inspired

- **Edge Ingestion Simulation:** Simulate edge PoPs that buffer events locally and forward to the central cluster. Demonstrates geo-distributed ingestion resilience. *Technical depth: edge buffering, store-and-forward, consistency models.*

- **Analytics at the Edge:** Pre-aggregate event counts at the edge before forwarding to reduce central processing load. *Technical depth: distributed aggregation, merge at center.*

---

## Appendix: Interview Discussion Guide

This project is designed to support 30–40 minutes of deep technical discussion. Here are the kinds of questions an interviewer might ask, mapped to the relevant sections:

**Distributed Systems:**
- *"How would you handle 100x the current event volume?"* → Sections 9, 14
- *"What consistency guarantees does the search index have?"* → Sections 6, 8
- *"How do you handle a consumer crash mid-processing?"* → Section 9

**Database Design:**
- *"Why did you choose time-based partitioning?"* → Section 7
- *"Walk me through the indexing strategy."* → Sections 7, 8
- *"How do you handle schema evolution for events?"* → Section 7

**AI Engineering:**
- *"How does the investigation agent decide what tools to call?"* → Section 10.2
- *"How do you prevent hallucinations in the AI copilot?"* → Section 10.7
- *"Walk me through the RAG pipeline."* → Section 10.4
- *"How do you evaluate AI quality?"* → Section 10.7

**System Design:**
- *"How would you add multi-tenancy?"* → Section 3 (Stretch)
- *"What's the tradeoff between your caching and consistency?"* → Section 14
- *"How would you implement exactly-once processing?"* → Section 9

**DevOps:**
- *"Walk me through a production deployment."* → Section 12
- *"How do you handle a bad deployment?"* → Section 12.7
- *"What happens when an alert fires?"* → Sections 5.7, 17

---

*End of specification. Awaiting approval to begin implementation.*
