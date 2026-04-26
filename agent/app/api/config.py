from __future__ import annotations

import logging
import os

from fastapi import APIRouter
from pydantic import BaseModel

from app.core.dependencies import (
    get_analysis_engine,
    get_analysis_service,
    get_chat_service,
    get_settings,
)
from app.core.log_sanitize import sanitize_log

logger = logging.getLogger(__name__)

router = APIRouter(tags=["config"])


class AIConfigUpdateRequest(BaseModel):
    provider: str  # gemini, openai, anthropic
    model_id: str


@router.post("/config/ai")
async def update_ai_config(request: AIConfigUpdateRequest):
    """Backend에서 UI 설정 변경 시 호출. lru_cache 초기화 및 환경변수 덮어쓰기."""

    logger.info(
        "Updating AI config: provider=%s, model_id=%s",
        sanitize_log(request.provider),
        sanitize_log(request.model_id),
    )

    # 1. 환경변수 덮어쓰기 (다음 Settings 로드 시 반영)
    os.environ["AI_PROVIDER"] = request.provider
    provider = request.provider.lower()
    if provider == "gemini":
        os.environ["GEMINI_MODEL_ID"] = request.model_id
    elif provider == "openai":
        os.environ["OPENAI_MODEL_ID"] = request.model_id
    elif provider == "anthropic":
        os.environ["ANTHROPIC_MODEL_ID"] = request.model_id

    # 2. @lru_cache 초기화 (의존성 체인 재생성)
    get_settings.cache_clear()
    get_analysis_engine.cache_clear()
    get_analysis_service.cache_clear()
    get_chat_service.cache_clear()

    logger.info("AI config updated and caches cleared")

    return {"status": "ok", "provider": request.provider, "model_id": request.model_id}
