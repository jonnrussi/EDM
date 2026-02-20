from backend.scripts.migrate_db import run_seed_once, wait_for_db


if __name__ == "__main__":
    wait_for_db()
    run_seed_once()
