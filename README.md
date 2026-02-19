# UEM Enterprise SaaS Reference Architecture

Plataforma UEM multi-tenant em microserviços, desenhada para escalar em Kubernetes com segurança enterprise, observabilidade completa e modularidade.

## 1) Arquitetura textual

```text
[Endpoint Agent (Win/Linux/macOS)]
   | HTTPS + HMAC (fallback WebSocket)
   v
[API Gateway / WAF]
   | JWT/OAuth2
   +--> [Auth Service] ----> [PostgreSQL multi-tenant]
   +--> [Device Service] --> [Redis cache]
   +--> [Task Service] ----> [RabbitMQ event bus]
   +--> [Patch Service] ---> [CVE Feed Connector]
   +--> [Reporting Service] -> [Object Storage PDF/CSV]
   +--> [Notification Service]

Observability sidecar/exporters -> OpenTelemetry -> Prometheus/Grafana + ELK
```

## 2) Serviços e responsabilidades

- **Auth Service**: OAuth2/JWT, refresh token, RBAC e federation-ready.
- **Device Service**: inventário de hardware/software, status de AV e criptografia.
- **Task Service**: criação de tarefas remotas e fila assíncrona para execução por agentes.
- **Patch Service**: ingestão de CVE, aprovação e rollout gradual.
- **Reporting Service**: relatórios de compliance, vulnerabilidade, patch e saúde.
- **Notification Service**: envio de alertas por email/webhook.

## 3) Multi-tenancy

Estratégia padrão: **isolamento lógico por `tenant_id`** com filtros obrigatórios em todas as queries.
Opções futuras:
- `schema-per-tenant` para tenants premium com requisitos regulatórios.
- partição de banco por região para residência de dados.

## 4) Segurança enterprise

- TLS obrigatório; mTLS opcional para agentes gerenciados.
- JWT assinado + RBAC por permissão fina.
- Anti replay (`X-Nonce`, `X-Timestamp`) e assinatura HMAC no payload do agente.
- Rate limiting no gateway e trilha de auditoria imutável (`immutable_hash`).
- Criptografia AES para dados sensíveis (roadmap: campo-level encryption com KMS).

## 5) Fluxos principais

### 5.1 Fluxo agente-servidor
1. Agente coleta inventário local.
2. Assina payload com HMAC.
3. Envia para Device Service via HTTPS.
4. Em falha de rede/proxy, usa fallback WebSocket.

### 5.2 Fluxo de autenticação
1. Usuário autentica no Auth Service.
2. Serviço emite JWT com `sub`, `tenant_id` e `role`.
3. API Gateway encaminha token para serviços.
4. Middleware RBAC aplica permissões por endpoint.

### 5.3 Fluxo de tarefa remota
1. Técnico cria tarefa no Task Service.
2. Tarefa persistida com status `queued`.
3. Evento `task.created` publicado no RabbitMQ.
4. Worker de execução entrega tarefa ao agente do endpoint.
5. Resultado retorna para Task Service e Audit Log.

## 6) Estrutura de pastas

```text
backend/
  shared/
  services/
    auth_service/
    device_service/
    task_service/
    patch_service/
    reporting_service/
    notification_service/
  scripts/seed_db.py
agent/
frontend/
infra/
  docker-compose.yml
  k8s/uem-platform.yaml
```

## 7) CI/CD

- **CI**: lint + testes + SAST + scan de imagem + SBOM.
- **CD**: deploy progressivo (canary/blue-green) por serviço.
- **Agent updates**: canal assinado digitalmente + rollout por anel (pilot -> 10% -> 50% -> 100%).
- **Gate de segurança**: bloquear release com CVE crítico sem mitigação.

## 8) Como subir localmente

```bash
cp .env.example .env
python -m pip install -r backend/requirements.txt
python -m backend.scripts.seed_db
docker compose -f infra/docker-compose.yml up --build
```

## 9) Roadmap enterprise

- Controle remoto com gravação opcional e trilha legal.
- Sandbox de scripts por política e assinatura.
- Billing por plano (devices managed / tasks / storage).
- Logs imutáveis em storage WORM.

## 10) Client para adicionar endpoint no inventário

Foi adicionado um client no frontend (`frontend/src/api/deviceClient.ts`) que chama o endpoint `POST /devices/v1/devices` via API Gateway.

Fluxo sugerido:
1. Autenticar em `/auth/v1/auth/login` para obter JWT.
2. Informar o token no formulário "Adicionar endpoint ao inventário" no dashboard.
3. Submeter o formulário para cadastrar o endpoint no Device Service.
