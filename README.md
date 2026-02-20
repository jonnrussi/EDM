# UEM Enterprise SaaS - Deployment Friendly Blueprint

## Unified Configuration Strategy

- **JWT padrão único HS256** (`JWT_SHARED_SECRET`) compartilhado por todos os serviços.
- **Sem dependência de RSA files** (`JWT_PRIVATE_KEY_PATH` / `JWT_PUBLIC_KEY_PATH` removidos).
- **`.env` centralizado** consumido por todos os containers em `docker-compose`.
- **Fail-fast de configuração**:
  - backend falha ao iniciar se `JWT_SHARED_SECRET` ou `HMAC_SHARED_SECRET` estiverem ausentes/curtos;
  - agent falha ao iniciar sem `UEM_HMAC_SECRET`, `UEM_AGENT_EMAIL`, `UEM_AGENT_PASSWORD`.
- **Startup banner** valida DB/Redis/RabbitMQ + JWT/HMAC.

## Single-command Deployment

```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up --build
```

O compose executa automaticamente:
1. Health checks de infra (Postgres/Redis/RabbitMQ).
2. Serviço `migrate` com wait-for-db + `create_all` + seed idempotente.
3. Sobe `auth-service`, `device-service`, `task-service` só após migração.
4. Gateway sobe somente após health dos serviços.

## Improved `.env` contract

Use `.env.example` como template e garanta:
- mesmo `JWT_SHARED_SECRET` para todos;
- mesmo `HMAC_SHARED_SECRET` no backend e `UEM_HMAC_SECRET` no agent;
- endpoints de auth/device configurados para o agent.

## FastAPI JWT Config (HS256)

Exemplo em produção no código:

```python
return jwt.encode(payload, settings.jwt_shared_secret, algorithm=settings.jwt_algorithm)
```

Validação:

```python
jwt.decode(
    token,
    settings.jwt_shared_secret,
    audience=settings.jwt_audience,
    issuer=settings.jwt_issuer,
    algorithms=[settings.jwt_algorithm],
)
```

## Unified HMAC Contract

Contrato único (Go e Python):

```text
signature = HEX(HMAC_SHA256(secret, body + nonce + timestamp))
```

Headers obrigatórios:
- `X-Nonce`
- `X-Timestamp` (epoch seconds)
- `X-Signature`

No backend (`ENV=development`) o 403 retorna razão detalhada (ex: timestamp fora de janela, assinatura inválida).

## Agent Login Flow

1. Agent chama `POST /v1/auth/login` com email/senha.
2. Token JWT fica em cache em memória até próximo do `exp`.
3. Agent envia inventário com `Authorization: Bearer` + HMAC.
4. Em `401`, força relogin e tenta novamente uma vez.
5. Logs explícitos para erro JWT/HMAC/schema/transporte.

## Developer Experience

```bash
make up     # docker compose up --build
make down   # stop + remove volumes
make seed   # roda migrate/seed manualmente se necessário
make logs   # follow logs
```

## Deployment Safety Guarantees

- Evita segredo JWT divergente: **uma variável compartilhada**.
- Evita ausência de segredo HMAC: validação obrigatória no startup.
- Evita serviços sem DB: `migrate` espera DB e serviços dependem do sucesso.
- Evita agent inválido: valida configuração antes de executar.
