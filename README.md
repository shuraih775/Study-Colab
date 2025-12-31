
# StudyCollab Microservices

StudyCollab is a **microservices-based study collaboration platform** built with:

* **Core Service:** Go (Gin) – API Gateway + Chat (WebSocket)
* **File Service:** Java (Spring Boot) – File metadata + presigned URL generation
* **Frontend:** Next.js
* **Storage:** PostgreSQL + MinIO
* **Inter-service communication:** gRPC
* **Orchestration:** Docker Compose
* **Developer workflow:** Makefile (standard entrypoint)

---

## Getting Started 




## 1️⃣ Environment Configuration (Required)

Create your local environment file:

```bash
cp .env.example .env
```

### Required `.env` variables

```properties
# --- Postgres Configuration ---
POSTGRES_USER=postgres
POSTGRES_PASSWORD=123456
POSTGRES_DB=studycollab
POSTGRES_PORT=5432

# --- MinIO Configuration ---
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=password123
MINIO_BUCKET_NAME=studycolab-private-assets-dev

# Internal (Docker network) endpoint
MINIO_ENDPOINT=http://minio:9000

# External (Browser-accessible) endpoint
MINIO_PUBLIC_ENDPOINT=http://localhost:9000

MINIO_CORS_ALLOW_ORIGIN=http://localhost:3000
MINIO_REGION=ap-south-1

# --- Application Secrets ---
JWT_SECRET=secret-internal-key

# --- Frontend URLs ---
NEXT_PUBLIC_API_URL=http://localhost:8080
INTERNAL_API_URL=http://core-service:8080
```

### Why there are *two* MinIO endpoints

* `MINIO_ENDPOINT`
  → Used **inside Docker** (services talk to MinIO via container DNS)

* `MINIO_PUBLIC_ENDPOINT`
  → Used for **presigned URLs** returned to the browser

This separation is **intentional and required**.

---

## ⚡ Development Workflow (Recommended)

### Start backend services

```bash
make up
```

This starts:

* PostgreSQL
* MinIO
* File Service (Java)
* Core Service (Go)

All services run in Docker and share the same network.

---

### Start frontend locally (hot reload)

In a new terminal:

```bash
cd frontend
npm install
npm run dev
```

**App URL:**
👉 [http://localhost:3000](http://localhost:3000)

This hybrid setup is **intentional**:

* Backend is stable and containerized
* Frontend reloads instantly

---

## 🐳 Full Docker Setup (All-in-One)

If you want everything inside Docker (slower, but simple):

```bash
make up-all
```

Then open:

👉 [http://localhost:3000](http://localhost:3000)

---

## 🧰 Useful Makefile Commands

```bash
make up            # Start core backend services
make up-all        # Start everything (including frontend)
make down          # Stop services
make restart       # Restart backend services
make logs          # Tail logs
make ps            # Show container status
make clean         # Stop + remove containers
make clean-volumes # Stop + wipe volumes (DATA LOSS)
```

> ⚠️ `clean-volumes` deletes **Postgres + MinIO data**. Use carefully.

---

## 🔍 Service Endpoints

| Service       | URL                                            | Notes               |
| ------------- | ---------------------------------------------- | ------------------- |
| Frontend      | [http://localhost:3000](http://localhost:3000) | Next.js             |
| Core API      | [http://localhost:8080](http://localhost:8080) | REST + WebSocket    |
| MinIO API     | [http://localhost:9000](http://localhost:9000) | Object storage      |
| MinIO Console | [http://localhost:9001](http://localhost:9001) | admin / password123 |
| PostgreSQL    | localhost:5432                                 | postgres / 123456   |

---

##  File Upload & Chat Architecture 

* Frontend uploads files **directly to MinIO** using presigned URLs
* Backend **never** proxies file uploads
* Chat messages store **file_id**, not URLs
* Download URLs are generated **on-demand** and **expire**
* Presigned URLs are **never stored in DB**

This design is:

* secure
* scalable
* production-correct

---

## 🛠 Troubleshooting

### 1. Database does not exist

If Docker was started before `.env` was set correctly:

```bash
make clean-volumes
make up
```

⚠️ This deletes local data.

---

### 2. UUID extension missing (mostly will not occur)

Run once in Postgres:

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

