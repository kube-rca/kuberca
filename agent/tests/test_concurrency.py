from __future__ import annotations

import asyncio
import time

import pytest

from app.core.concurrency import init_concurrency, run_in_thread_limited


def test_run_in_thread_limited_returns_result():
    init_concurrency(max_concurrent=2)
    result = asyncio.run(run_in_thread_limited(lambda: 42))
    assert result == 42


def test_semaphore_limits_concurrency():
    init_concurrency(max_concurrent=1)
    started: list[float] = []

    def slow_task(idx: int) -> int:
        started.append(time.monotonic())
        time.sleep(0.1)
        return idx

    async def _run() -> list[int]:
        return list(
            await asyncio.gather(
                run_in_thread_limited(slow_task, 0),
                run_in_thread_limited(slow_task, 1),
            )
        )

    results = asyncio.run(_run())
    assert set(results) == {0, 1}
    # With semaphore=1, second task starts after first finishes
    assert started[1] - started[0] >= 0.08


def test_exception_propagates_and_releases_semaphore():
    """blocking 함수의 예외가 전파되고 semaphore가 해제되는지 확인."""
    init_concurrency(max_concurrent=1)

    def failing_task() -> None:
        raise ValueError("boom")

    async def _run() -> None:
        with pytest.raises(ValueError, match="boom"):
            await run_in_thread_limited(failing_task)
        # Semaphore가 정상 해제되었으면 다음 호출이 즉시 성공
        result = await run_in_thread_limited(lambda: "recovered")
        assert result == "recovered"

    asyncio.run(_run())


def test_run_without_init_still_works():
    """Semaphore 미초기화 시에도 to_thread로 정상 동작."""
    from app.core import concurrency

    concurrency._semaphore = None
    result = asyncio.run(run_in_thread_limited(lambda: "ok"))
    assert result == "ok"
