import logging
from collections.abc import Sequence

from sample_app.domain.shared.events import DomainEvent


class LoggingEventPublisher:
    def __init__(self, logger: logging.Logger) -> None:
        self._logger = logger

    def publish(self, events: Sequence[DomainEvent]) -> None:
        for event in events:
            self._logger.info("domain event name=%s event=%s", event.name, event)
