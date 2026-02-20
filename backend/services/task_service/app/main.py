from uuid import uuid4

from fastapi import Depends, FastAPI
from pydantic import BaseModel
from sqlalchemy.orm import Session

from backend.shared.bootstrap import startup_banner
from backend.shared.db import get_db
from backend.shared.events import publish_event
from backend.shared.models import Task
from backend.shared.rbac import require_permission

app = FastAPI(title="UEM Task Service")


@app.on_event("startup")
def on_startup() -> None:
    startup_banner("task-service")


class TaskRequest(BaseModel):
    device_id: str
    task_type: str
    payload: dict


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok", "service": "task-service"}


@app.post("/v1/tasks", dependencies=[Depends(require_permission("task:run"))])
def create_task(body: TaskRequest, db: Session = Depends(get_db), user=Depends(require_permission("task:run"))):
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
    publish_event("task.created", {"task_id": task.id, "tenant_id": task.tenant_id})
    return {"task_id": task.id, "status": task.status}
