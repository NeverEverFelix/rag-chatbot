# rag-chatbot
# 🧠 rag-chatbot

## 1. Project Overview

**RAG Chatbot** is a full-stack, production-grade Retrieval-Augmented Generation (RAG) system that combines semantic search with OpenAI’s GPT models to deliver accurate, context-aware answers from a curated knowledge base.

🔗 Live: [https://www.felixmoronge.com](https://www.felixmoronge.com)

Built with:
- Microservices (Go + Python)
- PostgreSQL with pgvector
- Containerized services via Docker
- Kubernetes on AWS EKS, provisioned with Terraform and Helm
- NGINX Ingress + TLS via Let’s Encrypt

This project serves both as a portfolio-grade deployment and a reproducible open-source reference.

---

## 2. Features

- 🔍 **Semantic Search with pgvector**: Retrieves relevant chunks using cosine similarity.
- 🧠 **Retrieval-Augmented Generation**: Enhances GPT answers with context.
- ⚙️ **Dual Microservice Architecture**:  
  - `embed-api` (Python): Generates embeddings  
  - `go-api` (Go): Handles queries and streams answers
- 📦 **Containerized & Scalable**: Deployed via Docker, Kubernetes, Helm, and Terraform.
- 📡 **Streaming Chat Responses**: Uses `text/event-stream` for low-latency UX.
- 🔐 **Secrets Management**: OpenAI keys + service URLs via Kubernetes Secrets.
- 🌐 **Ingress Routing + TLS**: Route-based service access via NGINX + Let’s Encrypt.
- 📄 **Extendable Knowledge Base**: Embed your own `.txt` data using `prepare_chunks_and_embed.py`.

---

## 3. Architecture

```
                    ┌──────────────────────────────┐
                    │     User Web Interface       │
                    │ (https://www.felixmoronge.com)│
                    └──────────────┬───────────────┘
                                   │
                                   ▼
                       ┌────────────────────┐
                       │     NGINX Ingress   │
                       │ + TLS (Let's Encrypt)│
                       └──────┬───────┬──────┘
                              │       │
     ┌────────────────────┐   │       │   ┌─────────────────────┐
     │    /api/ask        │◄──┘       └──►│      /embed         │
     │    (go-api)        │                │    (embed-api)     │
     └──────┬─────────────┘                └────────┬────────────┘
            │                                         │
            ▼                                         ▼
  ┌────────────────────┐                   ┌────────────────────────┐
  │ Query Top-K Chunks │                   │ Generate Embeddings     │
  │ via pgvector       │◄─────────────────▶│ via sentence-transformers│
  └────────┬───────────┘                   └─────────┬──────────────┘
           │                                         │
           ▼                                         ▼
     ┌────────────────────────────┐       ┌────────────────────────┐
     │   PostgreSQL + pgvector    │       │ Embedded Knowledge Base│
     │    (ragdb on EKS)          │       │   (custom .txt chunks) │
     └────────────────────────────┘       └────────────────────────┘
```

---

## 4. Tech Stack

### 🧠 AI / NLP
- OpenAI GPT-3.5 Turbo  
- `sentence-transformers` (`all-MiniLM-L6-v2`)

### 🖥️ Microservices
- Go (Gin) – `go-api`
- Python (FastAPI) – `embed-api`

### 🗃️ Database
- PostgreSQL 15  
- `pgvector` extension

### 🚢 Infrastructure
- Docker  
- Kubernetes (EKS)  
- Helm (Postgres deployment)  
- Terraform (cluster + ECR + IAM)

### 🌐 Networking & TLS
- NGINX Ingress Controller  
- Cert-Manager (Let’s Encrypt)

### 🔐 Secrets
- Kubernetes Secrets  
- IAM inline policies for ECR

---

## 5. Infrastructure Overview (Terraform + EKS)


This project is deployed to Amazon Elastic Kubernetes Service (EKS) and fully provisioned using Terraform for reproducibility and scalability. It follows cloud-native best practices, enabling a production-grade RAG chatbot backend with minimal ops overhead.

| Component         | Tool                | Description |
|------------------|---------------------|-------------|
| EKS Cluster       | Terraform            | Control plane, node groups, VPC, IAM |
| Container Registry| AWS ECR              | Stores Go/Python/Postgres images |
| Database          | Helm + Terraform     | PostgreSQL 15 + pgvector with EBS |
| Ingress           | NGINX + Cert-Manager | Routing + TLS |
| Secrets           | Kubernetes Secrets   | Manages API keys and internal configs |

### Helm Values (`pgvector-values.yaml`)
```yaml
auth:
  username: NeverEverFelix
  password: "*****"
  database: ragdb

primary:
  persistence:
    enabled: true
    storageClass: ebs-sc
    size: 8Gi
```

### Secrets Example
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rag-secrets
type: Opaque
stringData:
  OPENAI_API_KEY: <your-key>
  EMBEDDING_SERVICE_URL: http://embed-api:5001/embed
```

---

## 6. Microservices Overview

### 🟦 `go-api`
- **Lang**: Go (Gin)  
- **Port**: `8080`  
- **Endpoint**: `/api/ask`

**Responsibilities:**
- Receives questions
- Sends to `embed-api`
- Retrieves top-k chunks
- Streams OpenAI response

**Env Vars:**
- `POSTGRES_DSN`
- `EMBEDDING_SERVICE_URL`
- `OPENAI_API_KEY`

---

### 🟨 `embed-api`
- **Lang**: Python (FastAPI)  
- **Port**: `5001`  
- **Endpoint**: `/embed`

**Responsibilities:**
- Returns embedding for any input text
- Populates database via `prepare_chunks_and_embed.py`

**Env Vars:**
- `POSTGRES_DSN`

---

## 7. Local Development

### ✅ Requirements
- Docker  
- Go 1.19+  
- Python 3.10+  
- PostgreSQL 15 + `pgvector`  
- OpenAI API Key

---

### 🐳 Docker (Recommended)

**Embed API**
```bash
cd rag-embed-py
docker build -t embed-api .
docker run -p 5001:5001 embed-api
```

**Go API**
```bash
cd rag-api-go
docker build -t go-api .
docker run -p 8080:8080 \
  -e OPENAI_API_KEY=<your-key> \
  -e EMBEDDING_SERVICE_URL=http://localhost:5001/embed \
  -e POSTGRES_DSN=postgres://postgres:password@localhost:5432/ragdb?sslmode=disable \
  go-api
```

---

### 🧪 Native Mode

**Embed API**
```bash
cd rag-embed-py
python3 -m venv venv && source venv/bin/activate
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 5001
```

**Go API**
```bash
cd rag-api-go
go run main.go
```

**Ingest Text Chunks**
```bash
python prepare_chunks_and_embed.py path/to/your/text.txt
```

---

## 8. Deployment Guide

### 📦 Step 1: Provision Infra
```bash
terraform init
terraform apply
```

Includes:
- EKS Cluster
- ECR Repos
- IAM Roles & Policies

### 🐘 Step 2: Deploy pgvector via Helm
```bash
helm install pgvector oci://registry-1.docker.io/bitnami/postgresql \
  -f pgvector-values.yaml
```
Enable the extension:
```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

### 🐳 Step 3: Push Docker Images to ECR
```bash
# Embed
docker build -t felixmoronge-rag-embed-py .
docker tag felixmoronge-rag-embed-py:latest <ECR_URL>
docker push <ECR_URL>

# Go
docker build -t felixmoronge-rag-api-go .
docker tag felixmoronge-rag-api-go:latest <ECR_URL>
docker push <ECR_URL>
```

### ☸️ Step 4: Apply K8s Manifests
```bash
kubectl apply -f postgres.yaml
kubectl apply -f rag-chatbot-k8s.yaml
```

### 🌐 Step 5: Ingress
```bash
kubectl apply -f portfolio-ingress.yaml
```

---

## 9. API Reference

### 🔹 `POST /api/ask`

**Request:**
```json
{
  "question": "How does this system retrieve relevant documents?"
}
```

**Response:**  
Streamed via `text/event-stream`
```text
This system uses vector similarity search powered by pgvector...
```

---

### 🔹 `POST /embed`

**Request:**
```json
{
  "text": "What is RAG?"
}
```

**Response:**
```json
{
  "embedding": [0.02, -0.18, ..., 0.07]
}
```

---

### ⚠️ Error Codes
- `400`: Malformed request
- `500`: Upstream failure or OpenAI error
- `502`: Bad Gateway / Timeout


## 10. Architecture Diagrams

### 🧭 High-Level System Architecture

```mermaid
graph TD
    A[User Interface<br/>felixmoronge.com] --> B[NGINX Ingress<br/>+ TLS via Cert-Manager]
    B --> C[go-api<br/>/api/ask]
    B --> D[embed-api<br/>/embed]

    C -->|Calls embed-api| D
    C -->|Queries| E[(PostgreSQL + pgvector)]

    D -->|Returns Embedding| C
    E -->|Returns Top-K Chunks| C
```

---

### 📡 Request Flow (RAG Pipeline)

```mermaid
sequenceDiagram
    participant User
    participant Ingress
    participant go-api
    participant embed-api
    participant PostgreSQL
    participant OpenAI

    User->>Ingress: POST /api/ask
    Ingress->>go-api: Route to go-api
    go-api->>embed-api: POST /embed (question)
    embed-api-->>go-api: Embedding vector
    go-api->>PostgreSQL: SELECT top-K chunks (ANN search)
    PostgreSQL-->>go-api: Top-K context chunks
    go-api->>OpenAI: Stream GPT-3.5 with context
    OpenAI-->>go-api: Token stream
    go-api-->>User: Streamed response (SSE)
```

---

### 🏗️ Infrastructure Provisioning (Terraform + Helm + Kubernetes)

```mermaid
graph LR
    subgraph Terraform
      A[EKS Cluster]
      B[ECR Repos]
      C[IAM Roles + Policies]
      D[Helm Provider]
    end

    subgraph Helm
      E[PostgreSQL with pgvector]
    end

    subgraph Kubernetes
      F[go-api Deployment]
      G[embed-api Deployment]
      H[Postgres Deployment]
      I[Secrets & ConfigMaps]
      J[NGINX Ingress]
    end

    A --> F
    A --> G
    A --> H
    D --> E
    C --> B
    I --> F
    I --> G
    J --> F
    J --> G
```

---

> ✅ GitHub automatically renders these diagrams when viewed on the site. For local previewing, use [Mermaid Live Editor](https://mermaid.live) or VSCode with the *Markdown Preview Mermaid Support* extension.
