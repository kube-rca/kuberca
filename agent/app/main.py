from __future__ import annotations

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.api import analysis, chat, config, health
from app.core.concurrency import init_concurrency
from app.core.dependencies import get_settings
from app.core.logging import configure_logging

settings = get_settings()
configure_logging(settings.log_level)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_concurrency(settings.max_concurrent_analyses)

    # Eagerly initialize analysis engine and session schema
    # before any requests are served — avoids race condition
    # on concurrent first requests (CREATE TABLE IF NOT EXISTS).
    from app.core.dependencies import get_analysis_engine

    engine = get_analysis_engine()
    if engine:
        logger.info("Analysis engine pre-initialized (session schema ready)")
    else:
        logger.warning("Analysis engine not initialized (no AI provider configured)")

    logger.info(
        "Starting kube-rca-agent on port %s (max_concurrent_analyses=%d)",
        settings.port,
        settings.max_concurrent_analyses,
    )
    yield


app = FastAPI(title="kube-rca-agent", version="1.0.0", lifespan=lifespan)
app.include_router(health.router)
app.include_router(analysis.router)
app.include_router(chat.router)
app.include_router(config.router)
