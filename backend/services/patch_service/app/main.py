from uuid import uuid4

from fastapi import Depends, FastAPI, HTTPException, Query
from pydantic import BaseModel
from sqlalchemy.orm import Session

from backend.shared.db import get_db
from backend.shared.models import Patch
from backend.shared.rbac import require_permission

app = FastAPI(title="UEM Patch Service")


class PatchCreateRequest(BaseModel):
    cve_id: str
    package_name: str
    severity: str
    target_group: str = "all-devices"


class PatchApprovalRequest(BaseModel):
    mode: str = "manual"  # manual | automatic


class PatchTestRequest(BaseModel):
    pilot_group: str = "pilot-ring"
    max_devices: int = 20


class PatchDeployRequest(BaseModel):
    ring: str = "10_percent"


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/v1/patches", dependencies=[Depends(require_permission("patch:approve"))])
def create_patch(
    body: PatchCreateRequest,
    db: Session = Depends(get_db),
    user=Depends(require_permission("patch:approve")),
):
    patch = Patch(
        id=str(uuid4()),
        tenant_id=user["tenant_id"],
        cve_id=body.cve_id,
        package_name=body.package_name,
        severity=body.severity,
        target_group=body.target_group,
        status="pending_approval",
    )
    db.add(patch)
    db.commit()
    return {"patch_id": patch.id, "status": patch.status}


@app.get("/v1/patches", dependencies=[Depends(require_permission("patch:approve"))])
def list_patches(
    db: Session = Depends(get_db),
    user=Depends(require_permission("patch:approve")),
    severity: str | None = Query(default=None),
):
    query = db.query(Patch).filter(Patch.tenant_id == user["tenant_id"])
    if severity:
        query = query.filter(Patch.severity == severity)
    rows = query.all()
    return [
        {
            "id": x.id,
            "cve_id": x.cve_id,
            "package_name": x.package_name,
            "severity": x.severity,
            "target_group": x.target_group,
            "status": x.status,
        }
        for x in rows
    ]


@app.post("/v1/patches/{patch_id}/approve", dependencies=[Depends(require_permission("patch:approve"))])
def approve_patch(
    patch_id: str,
    body: PatchApprovalRequest,
    db: Session = Depends(get_db),
    user=Depends(require_permission("patch:approve")),
):
    patch = db.query(Patch).filter(Patch.id == patch_id, Patch.tenant_id == user["tenant_id"]).first()
    if not patch:
        raise HTTPException(status_code=404, detail="Patch not found")
    patch.status = "approved_auto" if body.mode == "automatic" else "approved_manual"
    db.commit()
    return {"patch_id": patch.id, "status": patch.status}


@app.post("/v1/patches/{patch_id}/test", dependencies=[Depends(require_permission("patch:deploy"))])
def test_patch(
    patch_id: str,
    body: PatchTestRequest,
    db: Session = Depends(get_db),
    user=Depends(require_permission("patch:deploy")),
):
    patch = db.query(Patch).filter(Patch.id == patch_id, Patch.tenant_id == user["tenant_id"]).first()
    if not patch:
        raise HTTPException(status_code=404, detail="Patch not found")
    patch.status = "pilot_testing"
    db.commit()
    return {
        "patch_id": patch.id,
        "status": patch.status,
        "pilot_group": body.pilot_group,
        "max_devices": body.max_devices,
    }


@app.post("/v1/patches/{patch_id}/deploy", dependencies=[Depends(require_permission("patch:deploy"))])
def deploy_patch(
    patch_id: str,
    body: PatchDeployRequest,
    db: Session = Depends(get_db),
    user=Depends(require_permission("patch:deploy")),
):
    patch = db.query(Patch).filter(Patch.id == patch_id, Patch.tenant_id == user["tenant_id"]).first()
    if not patch:
        raise HTTPException(status_code=404, detail="Patch not found")
    patch.status = f"deploying_{body.ring}"
    db.commit()
    return {"patch_id": patch.id, "status": patch.status, "ring": body.ring}


@app.get("/v1/vulnerabilities/report", dependencies=[Depends(require_permission("report:read"))])
def vulnerability_report(
    db: Session = Depends(get_db),
    user=Depends(require_permission("report:read")),
):
    rows = db.query(Patch).filter(Patch.tenant_id == user["tenant_id"]).all()
    critical = [x for x in rows if x.severity.lower() == "critical"]
    pending = [x for x in rows if "pending" in x.status]
    return {
        "total": len(rows),
        "critical": len(critical),
        "pending": len(pending),
        "security_remediation_ready": True,
    }
