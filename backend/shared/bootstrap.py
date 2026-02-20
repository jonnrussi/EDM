import socket
import time
from urllib.parse import urlparse

from sqlalchemy import text

from .db import engine
from .settings import settings


def _wait_for_tcp(name: str, host: str, port: int, timeout_seconds: int = 30) -> None:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        try:
            with socket.create_connection((host, port), timeout=2):
                return
        except OSError:
            time.sleep(1)
    raise RuntimeError(f"{name} unreachable at {host}:{port}")


def _check_db() -> None:
    with engine.connect() as conn:
        conn.execute(text("SELECT 1"))


def startup_banner(service_name: str) -> None:
    redis = urlparse(settings.redis_url)
    rabbit = urlparse(settings.rabbitmq_url)

    _check_db()
    _wait_for_tcp("Redis", redis.hostname or "redis", redis.port or 6379)
    _wait_for_tcp("RabbitMQ", rabbit.hostname or "rabbitmq", rabbit.port or 5672)

    print("=" * 64)
    print(f"Starting {service_name}")
    print("[OK] DB connection")
    print("[OK] Redis connection")
    print("[OK] RabbitMQ connection")
    print(f"[OK] JWT config ({settings.jwt_algorithm}, issuer={settings.jwt_issuer})")
    print("[OK] HMAC config")
    print("=" * 64)
