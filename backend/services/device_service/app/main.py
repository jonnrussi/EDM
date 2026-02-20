from datetime import datetime
from uuid import uuid4

from fastapi import Depends, FastAPI, HTTPException, Query
from pydantic import BaseModel
from sqlalchemy.orm import Session

from backend.shared.db import get_db
from backend.shared.models import Device, MobileDevice
from backend.shared.rbac import require_permission

app = FastAPI(title="UEM Device Service")


class DeviceRegistration(BaseModel):
    hostname: str
    os_name: str
    os_version: str
    cpu: str
    ram_mb: int
    serial_number: str = "unknown"
    disk_json: dict = {"root": "120GB"}
    bios_version: str = "unknown"
    installed_software_json: list[str] = []
    software_usage_json: dict = {}


class DevicePolicyUpdate(BaseModel):
    antivirus_status: str | None = None
    encryption_status: str | None = None
    bitlocker_status: str | None = None
    usb_control_status: str | None = None
    browser_control_status: str | None = None


class MobileDeviceRegistration(BaseModel):
    platform: str
    owner_email: str


@app.post("/v1/devices", dependencies=[Depends(require_permission("device:write"))])
def register_device(
    payload: DeviceRegistration,
    db: Session = Depends(get_db),
    user=Depends(require_permission("device:write")),
):
    prohibited = [x for x in payload.installed_software_json if x.lower() in {"utorrent", "anydesk"}]
    device = Device(
        id=str(uuid4()),
        tenant_id=user["tenant_id"],
        hostname=payload.hostname,
        os_name=payload.os_name,
        os_version=payload.os_version,
        cpu=payload.cpu,
        ram_mb=payload.ram_mb,
        serial_number=payload.serial_number,
        disk_json=payload.disk_json,
        bios_version=payload.bios_version,
        installed_software_json=payload.installed_software_json,
        software_usage_json=payload.software_usage_json,
        prohibited_software_alerts=prohibited,
        last_seen=datetime.utcnow(),
    )
    db.add(device)
    db.commit()
    return {"device_id": device.id, "prohibited_software_alerts": prohibited}


@app.get("/v1/devices", dependencies=[Depends(require_permission("device:read"))])
def list_devices(
    db: Session = Depends(get_db),
    user=Depends(require_permission("device:read")),
    os_name: str | None = Query(default=None),
    hostname: str | None = Query(default=None),
    limit: int = Query(default=100, ge=1, le=500),
):
    query = db.query(Device).filter(Device.tenant_id == user["tenant_id"])
    if os_name:
        query = query.filter(Device.os_name.ilike(f"%{os_name}%"))
    if hostname:
        query = query.filter(Device.hostname.ilike(f"%{hostname}%"))

    rows = query.order_by(Device.last_seen.desc()).limit(limit).all()
    return [
        {
            "id": x.id,
            "hostname": x.hostname,
            "os": x.os_name,
            "os_version": x.os_version,
            "cpu": x.cpu,
            "ram_mb": x.ram_mb,
            "serial_number": x.serial_number,
            "disk_json": x.disk_json,
            "bios_version": x.bios_version,
            "installed_software": x.installed_software_json,
            "software_usage": x.software_usage_json,
            "prohibited_software_alerts": x.prohibited_software_alerts,
            "antivirus_status": x.antivirus_status,
            "encryption_status": x.encryption_status,
            "bitlocker_status": x.bitlocker_status,
            "usb_control_status": x.usb_control_status,
            "browser_control_status": x.browser_control_status,
            "last_seen": x.last_seen.isoformat(),
        }
        for x in rows
    ]


@app.patch("/v1/devices/{device_id}", dependencies=[Depends(require_permission("device:write"))])
def update_device_policy(
    device_id: str,
    body: DevicePolicyUpdate,
    db: Session = Depends(get_db),
    user=Depends(require_permission("device:write")),
):
    device = (
        db.query(Device)
        .filter(Device.id == device_id, Device.tenant_id == user["tenant_id"])
        .first()
    )
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")

    if body.antivirus_status is not None:
        device.antivirus_status = body.antivirus_status
    if body.encryption_status is not None:
        device.encryption_status = body.encryption_status
    if body.bitlocker_status is not None:
        device.bitlocker_status = body.bitlocker_status
    if body.usb_control_status is not None:
        device.usb_control_status = body.usb_control_status
    if body.browser_control_status is not None:
        device.browser_control_status = body.browser_control_status

    db.commit()
    return {"status": "updated", "device_id": device.id}


@app.delete("/v1/devices/{device_id}", dependencies=[Depends(require_permission("device:write"))])
def delete_device(
    device_id: str,
    db: Session = Depends(get_db),
    user=Depends(require_permission("device:write")),
):
    device = (
        db.query(Device)
        .filter(Device.id == device_id, Device.tenant_id == user["tenant_id"])
        .first()
    )
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")

    db.delete(device)
    db.commit()
    return {"status": "deleted", "device_id": device_id}


@app.post("/v1/mobile-devices", dependencies=[Depends(require_permission("device:write"))])
def enroll_mobile_device(
    payload: MobileDeviceRegistration,
    db: Session = Depends(get_db),
    user=Depends(require_permission("device:write")),
):
    mobile = MobileDevice(
        id=str(uuid4()),
        tenant_id=user["tenant_id"],
        platform=payload.platform,
        owner_email=payload.owner_email,
        compliance_status="enrolled",
    )
    db.add(mobile)
    db.commit()
    return {"mobile_device_id": mobile.id}


@app.post("/v1/mobile-devices/{mobile_device_id}/lock", dependencies=[Depends(require_permission("device:write"))])
def lock_mobile_device(
    mobile_device_id: str,
    db: Session = Depends(get_db),
    user=Depends(require_permission("device:write")),
):
    mobile = (
        db.query(MobileDevice)
        .filter(MobileDevice.id == mobile_device_id, MobileDevice.tenant_id == user["tenant_id"])
        .first()
    )
    if not mobile:
        raise HTTPException(status_code=404, detail="Mobile device not found")
    return {"status": "locked", "mobile_device_id": mobile_device_id}


@app.post("/v1/mobile-devices/{mobile_device_id}/wipe", dependencies=[Depends(require_permission("device:write"))])
def wipe_mobile_device(
    mobile_device_id: str,
    db: Session = Depends(get_db),
    user=Depends(require_permission("device:write")),
):
    mobile = (
        db.query(MobileDevice)
        .filter(MobileDevice.id == mobile_device_id, MobileDevice.tenant_id == user["tenant_id"])
        .first()
    )
    if not mobile:
        raise HTTPException(status_code=404, detail="Mobile device not found")
    return {"status": "wipe_requested", "mobile_device_id": mobile_device_id}
