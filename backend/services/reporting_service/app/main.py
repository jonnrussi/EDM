import csv
from io import StringIO
from uuid import uuid4

from fastapi import Depends, FastAPI
from sqlalchemy.orm import Session

from backend.shared.db import get_db
from backend.shared.models import Device, Patch, Report
from backend.shared.rbac import require_permission

app = FastAPI(title="UEM Reporting Service")


@app.get('/health')
def health():
    return {'status': 'ok'}


@app.get('/v1/reports/compliance', dependencies=[Depends(require_permission('report:read'))])
def compliance_report(db: Session = Depends(get_db), user=Depends(require_permission('report:read'))):
    devices = db.query(Device).filter(Device.tenant_id == user['tenant_id']).all()
    compliant = [x for x in devices if x.encryption_status == 'enabled' and x.antivirus_status == 'enabled']
    return {
        'total_devices': len(devices),
        'compliant_devices': len(compliant),
        'compliance_percent': round((len(compliant) / len(devices)) * 100, 2) if devices else 0,
    }


@app.get('/v1/reports/vulnerabilities', dependencies=[Depends(require_permission('report:read'))])
def vulnerability_report(db: Session = Depends(get_db), user=Depends(require_permission('report:read'))):
    patches = db.query(Patch).filter(Patch.tenant_id == user['tenant_id']).all()
    return {
        'total_patches': len(patches),
        'critical_patches': len([x for x in patches if x.severity.lower() == 'critical']),
        'pending_approval': len([x for x in patches if x.status == 'pending_approval']),
    }


@app.get('/v1/reports/patch-status', dependencies=[Depends(require_permission('report:read'))])
def patch_status_report(db: Session = Depends(get_db), user=Depends(require_permission('report:read'))):
    patches = db.query(Patch).filter(Patch.tenant_id == user['tenant_id']).all()
    grouped: dict[str, int] = {}
    for patch in patches:
        grouped[patch.status] = grouped.get(patch.status, 0) + 1
    return grouped


@app.get('/v1/reports/device-health', dependencies=[Depends(require_permission('report:read'))])
def device_health_report(db: Session = Depends(get_db), user=Depends(require_permission('report:read'))):
    devices = db.query(Device).filter(Device.tenant_id == user['tenant_id']).all()
    return [
        {
            'id': d.id,
            'hostname': d.hostname,
            'antivirus_status': d.antivirus_status,
            'encryption_status': d.encryption_status,
            'usb_control_status': d.usb_control_status,
            'browser_control_status': d.browser_control_status,
        }
        for d in devices
    ]


@app.get('/v1/reports/export.csv', dependencies=[Depends(require_permission('report:read'))])
def export_csv(db: Session = Depends(get_db), user=Depends(require_permission('report:read'))):
    devices = db.query(Device).filter(Device.tenant_id == user['tenant_id']).all()
    buffer = StringIO()
    writer = csv.writer(buffer)
    writer.writerow(['id', 'hostname', 'os', 'ram_mb', 'antivirus_status', 'encryption_status'])
    for d in devices:
        writer.writerow([d.id, d.hostname, d.os_name, d.ram_mb, d.antivirus_status, d.encryption_status])

    report = Report(
        id=str(uuid4()),
        tenant_id=user['tenant_id'],
        report_type='device_health',
        storage_path='inline://csv',
        format='csv',
    )
    db.add(report)
    db.commit()
    return {'report_id': report.id, 'csv': buffer.getvalue()}
