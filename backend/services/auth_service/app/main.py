from uuid import uuid4

from fastapi import Depends, FastAPI
from pydantic import BaseModel, EmailStr
from sqlalchemy.orm import Session

from backend.shared.bootstrap import startup_banner
from backend.shared.db import get_db
from backend.shared.models import User
from backend.shared.security import issue_jwt

app = FastAPI(title="UEM Auth Service")


@app.on_event("startup")
def on_startup() -> None:
    startup_banner("auth-service")


class LoginRequest(BaseModel):
    email: EmailStr
    password: str


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok", "service": "auth-service"}


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
