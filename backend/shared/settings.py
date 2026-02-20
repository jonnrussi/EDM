from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(case_sensitive=False, extra="ignore")

    env: str = "development"
    log_level: str = "INFO"

    postgres_host: str = "postgres"
    postgres_port: int = 5432
    postgres_db: str = "uem"
    postgres_user: str = "uem_user"
    postgres_password: str = "change-me"

    redis_url: str = "redis://redis:6379/0"
    rabbitmq_url: str = "amqp://uem:uem@rabbitmq:5672/"

    jwt_issuer: str = "uem-identity"
    jwt_audience: str = "uem-api"
    jwt_shared_secret: str = Field(..., min_length=32)
    jwt_algorithm: str = "HS256"

    hmac_shared_secret: str = Field(..., min_length=32)

    @property
    def database_url(self) -> str:
        return (
            f"postgresql+psycopg://{self.postgres_user}:{self.postgres_password}"
            f"@{self.postgres_host}:{self.postgres_port}/{self.postgres_db}"
        )


settings = Settings()
