import json

import pika

from .settings import settings


def publish_event(routing_key: str, payload: dict) -> None:
    params = pika.URLParameters(settings.rabbitmq_url)
    with pika.BlockingConnection(params) as connection:
        channel = connection.channel()
        channel.exchange_declare(exchange="uem.events", exchange_type="topic", durable=True)
        channel.basic_publish(
            exchange="uem.events",
            routing_key=routing_key,
            body=json.dumps(payload).encode(),
            properties=pika.BasicProperties(content_type="application/json", delivery_mode=2),
        )
