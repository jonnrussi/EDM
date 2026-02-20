from datetime import datetime, timedelta, timezone
from typing import Any

import jwt
from fastapi import HTTPException, Request

from .settings import settings


def issue_jwt(subject: str, tenant_id: str, role: str) -> str:
    now = datetime.now(timezone.utc)
    payload: dict[str, Any] = {
        "sub": subject,
        "tenant_id": tenant_id,
        "role": role,
        "iss": settings.jwt_issuer,
        "aud": settings.jwt_audience,
        "iat": now,
        "exp": now + timedelta(hours=1),
    }
    return jwt.encode(payload, settings.jwt_shared_secret, algorithm="HS256")


def verify_jwt(token: str) -> dict[str, Any]:
    try:
        return jwt.decode(
            token,
            settings.jwt_shared_secret,
            audience=settings.jwt_audience,
            issuer=settings.jwt_issuer,
            algorithms=["HS256"],
        )
    except jwt.PyJWTError as exc:
        raise HTTPException(status_code=401, detail="Invalid token") from exc


def replay_guard(request: Request) -> None:
    nonce = request.headers.get("X-Nonce")
    timestamp = request.headers.get("X-Timestamp")
    if not nonce or not timestamp:
        raise HTTPException(status_code=400, detail="Missing anti-replay headers")
