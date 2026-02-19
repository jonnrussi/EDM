from fastapi import Depends, HTTPException
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

from .security import verify_jwt

bearer_scheme = HTTPBearer()

ROLE_PERMISSIONS = {
    "super_admin": {"*"},
    "org_admin": {"device:read", "device:write", "task:run", "report:read", "patch:approve"},
    "technician": {"device:read", "task:run", "patch:deploy"},
    "auditor": {"device:read", "report:read", "audit:read"},
    "viewer": {"device:read", "report:read"},
}


def require_permission(permission: str):
    def checker(credentials: HTTPAuthorizationCredentials = Depends(bearer_scheme)):
        payload = verify_jwt(credentials.credentials)
        role = payload.get("role", "viewer")
        permissions = ROLE_PERMISSIONS.get(role, set())
        if "*" not in permissions and permission not in permissions:
            raise HTTPException(status_code=403, detail="Permission denied")
        return payload

    return checker
