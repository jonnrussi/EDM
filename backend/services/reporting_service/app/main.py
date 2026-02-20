from fastapi import FastAPI

app = FastAPI(title="UEM Reporting Service")


@app.get('/v1/reports')
def list_reports():
    return {'items': []}


@app.get('/health')
def health():
    return {'status': 'ok'}
