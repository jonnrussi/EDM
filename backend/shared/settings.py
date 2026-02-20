from pydantic import model_validator
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    env: str = "development"
    postgres_host: str = "postgres"
    postgres_port: int = 5432
    postgres_db: str = "uem"
    postgres_user: str = "uem_user"
    postgres_password: str = "change-me"
    database_url: str | None = None
    redis_url: str = "redis://redis:6379/0"
    rabbitmq_url: str = "amqp://uem:uem@rabbitmq:5672/"
    jwt_issuer: str = "uem-identity"
    jwt_audience: str = "uem-api"
    jwt_shared_secret: str = ""
    hmac_shared_secret: str = ""

    @property
    def resolved_database_url(self) -> str:
        if self.database_url:
            return self.database_url
        return (
            f"postgresql+psycopg://{self.postgres_user}:{self.postgres_password}"
            f"@{self.postgres_host}:{self.postgres_port}/{self.postgres_db}"
        )

    @model_validator(mode="after")
    def validate_required_secrets(self):
        if not self.jwt_shared_secret.strip():
            raise ValueError("JWT_SHARED_SECRET must be set")
        if not self.hmac_shared_secret.strip():
            raise ValueError("HMAC_SHARED_SECRET must be set")
        if not self.resolved_database_url.strip():
            raise ValueError("DATABASE_URL must be set")
        return self


settings = Settings()
