from __future__ import annotations

import asyncio
import logging
from collections.abc import Callable
from typing import Any, TypeVar

from fastapi import Request

T = TypeVar("T")

_semaphore: asyncio.Semaphore | None = None

logger = logging.getLogger(__name__)


def init_concurrency(max_concurrent: int) -> None:
    """Initialize the global concurrency limiter for analysis requests."""
    global _semaphore  # noqa: PLW0603
    _semaphore = asyncio.Semaphore(max(1, max_concurrent))


async def _wait_for_disconnect(request: Request) -> None:
    """Block until the HTTP client disconnects."""
    while not await request.is_disconnected():
        await asyncio.sleep(1)


async def run_in_thread_limited(
    func: Callable[..., T], *args: Any, request: Request | None = None
) -> T:
    """Run a blocking function in a thread, limited by the global semaphore.

    If *request* is provided, monitors for client disconnection and releases
    the semaphore early so other analyses can proceed.
    """
    if _semaphore is None:
        return await asyncio.to_thread(func, *args)

    async with _semaphore:
        task = asyncio.ensure_future(asyncio.to_thread(func, *args))
        if request is None:
            return await task

        disconnect = asyncio.ensure_future(_wait_for_disconnect(request))
        done, _ = await asyncio.wait(
            {task, disconnect},
            return_when=asyncio.FIRST_COMPLETED,
        )

        if disconnect in done:
            task.cancel()
            logger.warning("client_disconnected — releasing semaphore slot")
            raise asyncio.CancelledError("client disconnected")

        disconnect.cancel()
        return task.result()
