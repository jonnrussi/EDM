from uuid import uuid4

from fastapi import Depends, FastAPI
from pydantic import BaseModel
from sqlalchemy.orm import Session

from backend.shared.db import get_db
from backend.shared.models import Device
from backend.shared.rbac import require_permission

app = FastAPI(title="UEM Device Service")


class DeviceRegistration(BaseModel):
    hostname: str
    os_name: str
    os_version: str
    cpu: str
    ram_mb: int


@app.post("/v1/devices", dependencies=[Depends(require_permission("device:write"))])
def register_device(payload: DeviceRegistration, db: Session = Depends(get_db)):
    device = Device(
        id=str(uuid4()),
        tenant_id="tenant-default",
        hostname=payload.hostname,
        os_name=payload.os_name,
        os_version=payload.os_version,
        cpu=payload.cpu,
        ram_mb=payload.ram_mb,
        disk_json={"root": "120GB"},
        bios_version="unknown",
    )
    db.add(device)
    db.commit()
    return {"device_id": device.id}


@app.get("/v1/devices", dependencies=[Depends(require_permission("device:read"))])
def list_devices(db: Session = Depends(get_db)):
    rows = db.query(Device).limit(100).all()
    return [{"id": x.id, "hostname": x.hostname, "os": x.os_name} for x in rows]
