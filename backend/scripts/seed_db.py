from uuid import uuid4

from backend.shared.db import Base, SessionLocal, engine
from backend.shared.models import Tenant, User


def run_seed() -> None:
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    tenant = Tenant(id="tenant-default", name="Acme Corp", plan="enterprise")
    admin = User(
        id=str(uuid4()),
        tenant_id=tenant.id,
        email="admin@acme.local",
        password_hash="$2b$12$example",
        role="org_admin",
    )
    db.merge(tenant)
    db.merge(admin)
    db.commit()
    db.close()
    print("Seed concluído.")


if __name__ == "__main__":
    run_seed()
