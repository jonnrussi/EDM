from uuid import uuid4

from fastapi import Depends, FastAPI, HTTPException, Query
from pydantic import BaseModel
from sqlalchemy.orm import Session

from backend.shared.db import get_db
from backend.shared.events import publish_event
from backend.shared.models import Task
from backend.shared.rbac import require_permission

app = FastAPI(title="UEM Task Service")


class TaskRequest(BaseModel):
    device_id: str
    task_type: str
    payload: dict


class SoftwareDeploymentRequest(BaseModel):
    device_id: str
    package_name: str
    package_version: str
    action: str = "install"
    schedule_at: str | None = None
    target_group: str | None = None


class ConfigurationTaskRequest(BaseModel):
    device_id: str
    policy_name: str
    settings: dict


class UserManagementTaskRequest(BaseModel):
    device_id: str
    action: str
    username: str
    password: str | None = None
    admin_privileges: bool = False


class RemoteControlTaskRequest(BaseModel):
    device_id: str
    unattended: bool = True
    file_transfer: bool = True
    chat_enabled: bool = True
    record_session: bool = True


class AgentCommandStatusRequest(BaseModel):
    success: bool
    output: str
    exit_code: int


def _queue_task(body: TaskRequest, db: Session, user: dict) -> dict:
    task = Task(
        id=str(uuid4()),
        tenant_id=user["tenant_id"],
        device_id=body.device_id,
        task_type=body.task_type,
        payload=body.payload,
        status="queued",
        created_by=user["sub"],
    )
    db.add(task)
    db.commit()
    publish_event("task.created", {"task_id": task.id, "tenant_id": task.tenant_id, "task_type": task.task_type})
    return {"task_id": task.id, "status": task.status, "task_type": task.task_type}


@app.post("/v1/tasks", dependencies=[Depends(require_permission("task:run"))])
def create_task(body: TaskRequest, db: Session = Depends(get_db), user=Depends(require_permission("task:run"))):
    return _queue_task(body, db, user)


@app.post("/v1/software/deploy", dependencies=[Depends(require_permission("task:run"))])
def deploy_software(
    body: SoftwareDeploymentRequest,
    db: Session = Depends(get_db),
    user=Depends(require_permission("task:run")),
):
    return _queue_task(
        TaskRequest(
            device_id=body.device_id,
            task_type=f"software_{body.action}",
            payload={
                "package_name": body.package_name,
                "package_version": body.package_version,
                "schedule_at": body.schedule_at,
                "target_group": body.target_group,
                "silent": True,
                "auto_update": True,
            },
        ),
        db,
        user,
    )


@app.post("/v1/configurations/apply", dependencies=[Depends(require_permission("task:run"))])
def apply_configuration(
    body: ConfigurationTaskRequest,
    db: Session = Depends(get_db),
    user=Depends(require_permission("task:run")),
):
    return _queue_task(
        TaskRequest(
            device_id=body.device_id,
            task_type="configuration_apply",
            payload={"policy_name": body.policy_name, "settings": body.settings},
        ),
        db,
        user,
    )


@app.post("/v1/users/manage", dependencies=[Depends(require_permission("task:run"))])
def manage_local_user(
    body: UserManagementTaskRequest,
    db: Session = Depends(get_db),
    user=Depends(require_permission("task:run")),
):
    return _queue_task(
        TaskRequest(
            device_id=body.device_id,
            task_type="local_user_management",
            payload={
                "action": body.action,
                "username": body.username,
                "password": body.password,
                "admin_privileges": body.admin_privileges,
                "ad_integration": True,
            },
        ),
        db,
        user,
    )


@app.post("/v1/remote-control/start", dependencies=[Depends(require_permission("task:run"))])
def start_remote_control(
    body: RemoteControlTaskRequest,
    db: Session = Depends(get_db),
    user=Depends(require_permission("task:run")),
):
    return _queue_task(
        TaskRequest(
            device_id=body.device_id,
            task_type="remote_control_session",
            payload={
                "unattended": body.unattended,
                "file_transfer": body.file_transfer,
                "chat_enabled": body.chat_enabled,
                "record_session": body.record_session,
                "encrypted_channel": True,
            },
        ),
        db,
        user,
    )


@app.get("/v1/agent/commands/next")
def next_agent_command(
    device_id: str = Query(...),
    db: Session = Depends(get_db),
):
    task = (
        db.query(Task)
        .filter(Task.device_id == device_id, Task.status == "queued")
        .order_by(Task.created_at.asc())
        .first()
    )
    if not task:
        return {"task_id": ""}

    task.status = "dispatched"
    db.commit()
    return {
        "task_id": task.id,
        "task_type": task.task_type,
        "command": task.payload.get("command", "echo"),
        "args": task.payload.get("args", [f"task:{task.task_type}"]),
    }


@app.post("/v1/agent/commands/{task_id}/status")
def report_agent_command_status(
    task_id: str,
    body: AgentCommandStatusRequest,
    db: Session = Depends(get_db),
):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")

    payload = dict(task.payload)
    payload["execution_result"] = {
        "success": body.success,
        "output": body.output,
        "exit_code": body.exit_code,
    }
    task.payload = payload
    task.status = "completed" if body.success else "failed"
    db.commit()
    return {"task_id": task.id, "status": task.status}
