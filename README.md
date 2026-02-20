# UEM Enterprise SaaS Platform

Plataforma UEM multi-tenant em microserviços (FastAPI + PostgreSQL + Redis + RabbitMQ) para gestão completa de endpoints em ambiente cloud.

## Como executar localmente

```bash
cp .env.example .env
python -m pip install -r backend/requirements.txt
python -m backend.scripts.seed_db
docker compose -f infra/docker-compose.yml up --build
```

## Agent Enterprise (produção)

### 1) Login automático + refresh de JWT
- `TokenManager` realiza `POST /v1/auth/login` com `UEM_AGENT_EMAIL` e `UEM_AGENT_PASSWORD`.
- Token fica em memória com lock thread-safe.
- Refresh automático antes de expirar.
- Em `401`, força novo login e repete a chamada.
- Não depende de token manual no ambiente.

### 2) Enrollment token (zero-touch)
- Primeiro boot usa `UEM_ENROLLMENT_TOKEN` e chama `POST /v1/agents/enroll`.
- Backend retorna `device_id` + `access_token` bootstrap.
- `device_id` persiste em arquivo local (`UEM_DEVICE_ID_FILE`).
- Reexecuções são idempotentes (reutilizam `device_id` persistido).

### 3) Windows Service
- Suporta comandos:
  - `uem-agent install`
  - `uem-agent uninstall`
  - `uem-agent console`
- Nome do serviço: `UEM Agent`
- Configurado para auto-start e restart em falha (`sc failure`).

### 4) mTLS opcional
- Cert paths:
  - `UEM_CLIENT_CERT_PATH`
  - `UEM_CLIENT_KEY_PATH`
  - `UEM_CA_CERT_PATH`
- Se `MTLS_DISABLED=true`, usa HTTPS sem cert cliente.

### 5) Polling de tasks + execução segura
- Polling contínuo:
  - `GET /tasks/v1/agent/commands/next?device_id=...`
- Report de status:
  - `POST /tasks/v1/agent/commands/{task_id}/status`
- Allowlist de comandos + timeout de execução + sem shell injection (`exec.Command`).
- Backoff exponencial em falhas.

### 6) Hardening
- Structured logging (`slog` JSON).
- Correlation ID por request.
- Retry com jitter.
- Graceful shutdown com `SIGTERM/SIGINT`.
- Timeouts por request/context.
- Validação de configuração no startup.

### Matriz de plataforma do agente
- **Windows 11 / Linux**: execução local permitida (allowlist + timeout).
- **Android / iOS**: polling + report + inventário; execução de shell local bloqueada (usar ações MDM).

## Endpoints principais backend

### Auth
- `POST /auth/v1/auth/login`
- `POST /auth/v1/agents/enroll`

### Device
- `POST /devices/v1/devices`
- `GET /devices/v1/devices`

### Task
- `GET /tasks/v1/agent/commands/next`
- `POST /tasks/v1/agent/commands/{task_id}/status`
