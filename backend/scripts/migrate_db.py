import time

from sqlalchemy import text

from backend.shared.db import Base, SessionLocal, engine
from backend.shared.models import Tenant, User


def wait_for_db(max_wait_seconds: int = 60) -> None:
    deadline = time.time() + max_wait_seconds
    while time.time() < deadline:
        try:
            with engine.connect() as conn:
                conn.execute(text("SELECT 1"))
            return
        except Exception:
            time.sleep(1)
    raise RuntimeError("Database unreachable after waiting")


def run_migrations() -> None:
    Base.metadata.create_all(bind=engine)


def run_seed_once() -> None:
    db = SessionLocal()
    try:
        tenant = db.query(Tenant).filter(Tenant.id == "tenant-default").first()
        if tenant:
            print("Seed skipped (already initialized).")
            return

        tenant = Tenant(id="tenant-default", name="Acme Corp", plan="enterprise")
        admin = User(
            id="admin-bootstrap",
            tenant_id=tenant.id,
            email="admin@acme.local",
            password_hash="$2b$12$example",
            role="org_admin",
        )
        db.add(tenant)
        db.add(admin)
        db.commit()
        print("Seed completed.")
    finally:
        db.close()


if __name__ == "__main__":
    wait_for_db()
    run_migrations()
    run_seed_once()
