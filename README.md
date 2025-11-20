
# StudyCollab Microservices

A microservices-based study collaboration platform featuring a **Go (Gin)** API Gateway, **Java (Spring Boot)** File Service, and **Next.js** Frontend, orchestrated with **gRPC**, **PostgreSQL**, and **MinIO**.

-----

## 🚀 Getting Started

### 1\. Environment Configuration (Required)

Before running anything, you must set up your environment variables.

1.  Copy the example file:

    ```bash
    cp .env.example .env
    ```


2.  **Standard `.env` Configuration:**

    ```properties
    # Database
    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=password
    POSTGRES_DB=studycollab
    POSTGRES_PORT=5432

    # MinIO (Object Storage)
    MINIO_ROOT_USER=admin
    MINIO_ROOT_PASSWORD=password123
    MINIO_BUCKET_NAME=chat-uploads

    # Secrets
    JWT_SECRET=supersecretkey

    # Frontend URLs
    NEXT_PUBLIC_API_URL=http://localhost:8080
    INTERNAL_API_URL=http://core-service:8080
    ```

-----

## ⚡ Option 1: The "Hybrid" Approach (Recommended for Dev)

**Why?** Running Next.js inside Docker on Windows/Mac can be slow due to file system syncing. This method runs the heavy backend infrastructure in Docker but keeps the Frontend on your machine for instant hot-reloading.

### Step 1: Start Backend & Infrastructure

Run the database, storage, and backend APIs in Docker:

```bash
# Starts Postgres, MinIO, Java Service, and Go Service
docker-compose up -d postgres minio file-service core-service
```

### Step 2: Start Frontend Locally

Open a new terminal window:

```bash
cd frontend
npm install
npm run dev
```

**Access the App:** [http://localhost:3000](https://www.google.com/search?q=http://localhost:3000)

> **Note:** If the Frontend cannot talk to the Backend, change `INTERNAL_API_URL` in your `.env` to `http://localhost:8080` (since the backend is exposed to your host machine).

-----

## 🐳 Option 2: Full Docker Setup

**Why?** Easiest setup. One command to run the entire universe. Best for "Demo mode" or testing the final build.

1.  Build and start all services:
    ```bash
    docker-compose up --build
    ```
2.  Wait for the logs to stabilize.

**Access the App:** [http://localhost:3000](https://www.google.com/search?q=http://localhost:3000)

-----

## 🛠 Option 3: Manual Setup (No Docker)

**Why?** If you cannot use Docker or want to debug a specific service using your IDE's debugger.

**⚠️ Prerequisites:**

  * **PostgreSQL:** Must be installed locally or running via Docker (`docker run -p 5432:5432 postgres...`).
  * **MinIO:** Must be installed locally or running via Docker (`docker run -p 9000:9000 minio...`).

### 1. Configure File Service (Java)

When running manually, this service uses its `application.yml` file.

-   Open `file-service/src/main/resources/application.yml`.
    
-   Edit the properties directly to match your local setup (e.g., change `postgres` to `localhost`).
    
    

**Run:**

Bash

```
# From root
./mvnw clean spring-boot:run -pl file-service

```


### 2. Configure Core Service (Go)

This service requires its own `.env` file when running locally.

-   Navigate to the directory: `cd core-service`
    
-   Create a new `.env` file using the example provided:
    
    Bash
    
    ```
    cp .env.example .env
    
    ```
    
-   Open `.env` and ensure `DB_HOST` and `FILE_SERVICE_ADDR` point to `localhost` instead of Docker container names.
    

**Run:**

Bash

```
go run cmd/main.go

```

### 4\. Run Frontend

```bash
cd frontend
npm run dev
```

-----

## 🔍 Service Endpoints

| Service | URL | Credentials |
| :--- | :--- | :--- |
| **Frontend** | http://localhost:3000 | - |
| **Core API (Go)** | http://localhost:8080 | - |
| **MinIO Console** | http://localhost:9001 | `admin` / `password123` |
| **PostgreSQL** | localhost:5432 | `postgres` / `password` |

-----

##  Troubleshooting

**1. "Database 'studycollab' does not exist"**
If you ran Docker before setting the environment variables, the database volume might be stale.

  * **Fix:** `docker-compose down -v` (Warning: Deletes data), then restart.

**2. "UUID-OSSP extension missing"**
The setup includes an auto-init script. If it fails, run this manually in your SQL tool:

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

