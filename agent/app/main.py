from __future__ import annotations

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.api import analysis, health
from app.core.dependencies import get_settings
from app.core.logging import configure_logging

settings = get_settings()
configure_logging(settings.log_level)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Starting kube-rca-agent on port %s", settings.port)
    yield


app = FastAPI(title="kube-rca-agent", version="1.0.0", lifespan=lifespan)
app.include_router(health.router)
app.include_router(analysis.router)
