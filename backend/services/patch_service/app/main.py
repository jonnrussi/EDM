from fastapi import FastAPI

app = FastAPI(title="UEM Patch Service")


@app.get('/v1/patches')
def list_patches():
    return {'items': []}


@app.get('/health')
def health():
    return {'status': 'ok'}
