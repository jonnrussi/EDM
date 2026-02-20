# UEM Enterprise SaaS Platform

Plataforma UEM multi-tenant em microserviços (FastAPI + PostgreSQL + Redis + RabbitMQ) para gestão completa de endpoints em ambiente cloud.

## Como executar localmente

```bash
cp .env.example .env
python -m pip install -r backend/requirements.txt
python -m backend.scripts.seed_db
docker compose -f infra/docker-compose.yml up --build
```

## Endpoints principais

### Autenticação
- `POST /auth/v1/auth/login`

### Inventário + Segurança + MDM (Device Service)
- `POST /devices/v1/devices` (cadastro endpoint)
- `GET /devices/v1/devices` (inventário detalhado hardware/software)
- `PATCH /devices/v1/devices/{device_id}` (compliance/políticas: AV, criptografia, BitLocker, USB, browser)
- `DELETE /devices/v1/devices/{device_id}`
- `POST /devices/v1/mobile-devices` (enroll Android/iOS)
- `POST /devices/v1/mobile-devices/{id}/lock`
- `POST /devices/v1/mobile-devices/{id}/wipe`

### Patches (Patch Service)
- `POST /patch/v1/patches`
- `GET /patch/v1/patches`
- `POST /patch/v1/patches/{patch_id}/approve` (manual/automática)
- `POST /patch/v1/patches/{patch_id}/test` (pilot ring)
- `POST /patch/v1/patches/{patch_id}/deploy` (rollout por anel/grupo)
- `GET /patch/v1/vulnerabilities/report`

### Distribuição de software + Configuração + Usuários + Remoto + Automação (Task Service)
- `POST /tasks/v1/software/deploy` (instalação/desinstalação/update)
- `POST /tasks/v1/configurations/apply` (firewall, browser, drives, impressoras, scripts)
- `POST /tasks/v1/users/manage` (contas locais, reset senha, privilégio admin, AD-ready)
- `POST /tasks/v1/remote-control/start` (acesso remoto, chat, file transfer, gravação)
- `POST /tasks/v1/tasks` (automações customizadas / workflows)

### Relatórios (Reporting Service)
- `GET /reporting/v1/reports/compliance`
- `GET /reporting/v1/reports/vulnerabilities`
- `GET /reporting/v1/reports/patch-status`
- `GET /reporting/v1/reports/device-health`
- `GET /reporting/v1/reports/export.csv`

## Cobertura das funções pedidas

### 1) Gerenciamento de patches
- ✅ Windows e terceiros via `patches` + `software/deploy`
- ✅ Aprovação manual/auto
- ✅ Teste antes da implantação
- ✅ Relatórios de vulnerabilidade
- ✅ Correção de falhas de segurança

### 2) Distribuição de software
- ✅ Instalação remota
- ✅ Desinstalação silenciosa
- ✅ Atualização automática
- ✅ Repositório central (baseado em pacote/versionamento)
- ✅ Deploy por horário/grupo

### 3) Gerenciamento de configuração
- ✅ Políticas de segurança
- ✅ Firewall
- ✅ Impressoras
- ✅ Drives de rede
- ✅ Políticas de navegador
- ✅ Scripts personalizados

### 4) Gerenciamento de usuários
- ✅ Criação/gerenciamento de contas locais
- ✅ Reset de senha remoto
- ✅ Privilégios administrativos
- ✅ Integração AD (ready em payload)

### 5) Controle remoto
- ✅ Sessão remota
- ✅ Transferência de arquivos
- ✅ Chat com usuário
- ✅ Sessões não assistidas
- ✅ Gravação de sessão

### 6) Inventário hardware/software
- ✅ CPU, RAM, disco, BIOS, serial
- ✅ Software instalado
- ✅ Alertas de software proibido
- ✅ Monitoramento de uso

### 7) Segurança e compliance
- ✅ Controle USB
- ✅ Criptografia de disco
- ✅ Controle de navegador
- ✅ BitLocker status
- ✅ Conformidade corporativa

### 8) MDM
- ✅ Android/iOS
- ✅ Políticas móveis
- ✅ Bloqueio remoto
- ✅ Wipe remoto
- ✅ Controle de apps móveis (status)

### 9) Automação
- ✅ Agendamento de tarefas
- ✅ Workflows automáticos
- ✅ Auto-remediação via tasks
- ✅ Deploy por grupos dinâmicos

### 10) Cloud
- ✅ Gestão fora da rede corporativa via API Gateway
- ✅ Home office/remoto via internet
- ✅ Sem servidor local obrigatório

## Frontend

O dashboard possui formulário para cadastro de endpoint, listagem de inventário e remoção de endpoints via client `frontend/src/api/deviceClient.ts`.


## Agent corporativo (service + mTLS + polling + execução + status)

O agente agora suporta o fluxo exigido:
- Rodar como serviço (loop contínuo no `main`)
- Autenticação com certificado do dispositivo (mTLS)
- Polling contínuo de comandos (`GET /tasks/v1/agent/commands/next?device_id=...`)
- Receber comando, executar localmente com allowlist de segurança
- Reportar status/saída (`POST /tasks/v1/agent/commands/{task_id}/status`)

Variáveis do agente:
- `UEM_DEVICE_ID`
- `UEM_DEVICE_URL`
- `UEM_TASK_URL`
- `UEM_CLIENT_CERT_PATH`
- `UEM_CLIENT_KEY_PATH`
- `UEM_CA_CERT_PATH`
- `UEM_ALLOWED_COMMANDS`
- `UEM_POLL_INTERVAL_SEC`
- `UEM_HMAC_SECRET`
