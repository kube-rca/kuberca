from __future__ import annotations

import asyncio
from collections.abc import Callable
from typing import Any, TypeVar

T = TypeVar("T")

_semaphore: asyncio.Semaphore | None = None


def init_concurrency(max_concurrent: int) -> None:
    """Initialize the global concurrency limiter for analysis requests."""
    global _semaphore  # noqa: PLW0603
    _semaphore = asyncio.Semaphore(max(1, max_concurrent))


async def run_in_thread_limited(func: Callable[..., T], *args: Any) -> T:
    """Run a blocking function in a thread, limited by the global semaphore."""
    if _semaphore is None:
        return await asyncio.to_thread(func, *args)
    async with _semaphore:
        return await asyncio.to_thread(func, *args)
