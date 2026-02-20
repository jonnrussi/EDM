from fastapi import FastAPI

app = FastAPI(title="UEM Notification Service")


@app.get('/v1/notifications')
def list_notifications():
    return {'items': []}


@app.get('/health')
def health():
    return {'status': 'ok'}
