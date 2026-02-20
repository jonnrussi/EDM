import hashlib
import hmac
from datetime import datetime, timedelta, timezone
from typing import Any

import jwt
from fastapi import HTTPException, Request

from .settings import settings


class Unauthorized(HTTPException):
    def __init__(self, detail: str):
        super().__init__(status_code=401, detail=detail)


class Forbidden(HTTPException):
    def __init__(self, detail: str):
        super().__init__(status_code=403, detail=detail)


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
    return jwt.encode(payload, settings.jwt_shared_secret, algorithm=settings.jwt_algorithm)


def verify_jwt(token: str) -> dict[str, Any]:
    try:
        return jwt.decode(
            token,
            settings.jwt_shared_secret,
            audience=settings.jwt_audience,
            issuer=settings.jwt_issuer,
            algorithms=[settings.jwt_algorithm],
        )
    except jwt.PyJWTError as exc:
        detail = f"Invalid token: {exc}" if settings.env == "development" else "Invalid token"
        raise Unauthorized(detail) from exc


def _hmac_error(reason: str) -> Forbidden:
    detail = f"HMAC validation failed: {reason}" if settings.env == "development" else "Invalid request signature"
    return Forbidden(detail)


async def verify_hmac_request(request: Request) -> None:
    nonce = request.headers.get("X-Nonce")
    timestamp = request.headers.get("X-Timestamp")
    signature = request.headers.get("X-Signature")

    if not nonce or not timestamp or not signature:
        raise _hmac_error("missing X-Nonce, X-Timestamp, or X-Signature")

    try:
        ts = int(timestamp)
    except ValueError as exc:
        raise _hmac_error("X-Timestamp is not an integer") from exc

    now = int(datetime.now(timezone.utc).timestamp())
    if abs(now - ts) > 300:
        raise _hmac_error("timestamp outside 5 minute window")

    body = await request.body()
    payload = body + nonce.encode() + timestamp.encode()
    expected = hmac.new(settings.hmac_shared_secret.encode(), payload, hashlib.sha256).hexdigest()

    if not hmac.compare_digest(expected, signature):
        raise _hmac_error("signature mismatch")
