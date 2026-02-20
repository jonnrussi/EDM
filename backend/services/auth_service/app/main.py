from uuid import uuid4

from fastapi import Depends, FastAPI, HTTPException
from pydantic import BaseModel, EmailStr
from sqlalchemy.orm import Session

from backend.shared.db import get_db
from backend.shared.models import Device, User
from backend.shared.security import issue_jwt

app = FastAPI(title="UEM Auth Service")


class LoginRequest(BaseModel):
    email: EmailStr
    password: str


class EnrollRequest(BaseModel):
    enrollment_token: str
    hostname: str
    platform: str


@app.post("/v1/auth/login")
def login(data: LoginRequest, db: Session = Depends(get_db)):
    user = db.query(User).filter(User.email == data.email).first()
    if not user:
        user = User(
            id=str(uuid4()),
            tenant_id="tenant-default",
            email=data.email,
            password_hash="stub",
            role="org_admin",
        )
        db.add(user)
        db.commit()
    token = issue_jwt(subject=user.id, tenant_id=user.tenant_id, role=user.role)
    return {"access_token": token, "token_type": "bearer"}


@app.post("/v1/agents/enroll")
def enroll_agent(body: EnrollRequest, db: Session = Depends(get_db)):
    if body.enrollment_token != "enroll-dev-token":
        raise HTTPException(status_code=401, detail="Invalid enrollment token")

    existing = db.query(Device).filter(Device.hostname == body.hostname).first()
    if existing:
        device_id = existing.id
    else:
        device = Device(
            id=str(uuid4()),
            tenant_id="tenant-default",
            hostname=body.hostname,
            os_name=body.platform,
            os_version="unknown",
            cpu="unknown",
            ram_mb=0,
            serial_number="unknown",
            disk_json={},
            bios_version="unknown",
            installed_software_json=[],
            software_usage_json={},
            prohibited_software_alerts=[],
        )
        db.add(device)
        db.commit()
        device_id = device.id

    jwt_token = issue_jwt(subject=device_id, tenant_id="tenant-default", role="technician")
    return {"device_id": device_id, "access_token": jwt_token}
