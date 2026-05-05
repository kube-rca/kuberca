from __future__ import annotations

import threading

import psycopg
import pytest
from strands.types.exceptions import SessionException
from strands.types.session import Session, SessionAgent, SessionMessage, SessionType
from testcontainers.postgres import PostgresContainer

from app.clients.session_repository import PostgresSessionRepository

# ---------------------------------------------------------------------------
# Session-scoped container fixture
# ---------------------------------------------------------------------------

_PG_IMAGE = "pgvector/pgvector:pg16"


@pytest.fixture(scope="session")
def pg_dsn() -> str:  # type: ignore[return]
    """Start a PostgreSQL container once per test session and yield a psycopg DSN."""
    with PostgresContainer(image=_PG_IMAGE, driver=None) as pg:
        # get_connection_url with driver=None gives a plain postgresql:// URL.
        # psycopg.connect accepts postgresql:// URIs directly.
        url = pg.get_connection_url()
        # Ensure pgvector extension exists (used by the broader app stack, not repo itself).
        with psycopg.connect(url) as conn:
            conn.execute("CREATE EXTENSION IF NOT EXISTS vector")
            conn.commit()
        # Bootstrap the schema once; individual tests call _clean() for isolation.
        PostgresSessionRepository(dsn=url)
        yield url


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _repo(dsn: str) -> PostgresSessionRepository:
    return PostgresSessionRepository(dsn=dsn)


def _session(session_id: str = "sess-1") -> Session:
    return Session(session_id=session_id, session_type=SessionType.AGENT)


def _agent(agent_id: str = "agent-1") -> SessionAgent:
    return SessionAgent(agent_id=agent_id, state={}, conversation_manager_state={})


def _message(message_id: int = 0) -> SessionMessage:
    msg = {"role": "user", "content": [{"type": "text", "text": "hello"}]}
    return SessionMessage(message=msg, message_id=message_id)


def _clean(dsn: str) -> None:
    """Truncate all strands tables for test isolation."""
    with psycopg.connect(dsn) as conn:
        conn.execute("TRUNCATE strands_messages, strands_agents, strands_sessions CASCADE")
        conn.commit()


# ---------------------------------------------------------------------------
# create_session / read_session / delete_session
# ---------------------------------------------------------------------------


def test_create_and_read_session(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)

    session = _session("sess-create-read")
    repo.create_session(session)
    result = repo.read_session("sess-create-read")

    assert result is not None
    assert result.session_id == "sess-create-read"


def test_read_session_missing_returns_none(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    assert repo.read_session("does-not-exist") is None


def test_create_session_duplicate_raises(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    session = _session("sess-dup")
    repo.create_session(session)

    with pytest.raises(SessionException, match="already exists"):
        repo.create_session(session)


def test_delete_session_removes_row(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    session = _session("sess-delete")
    repo.create_session(session)
    repo.delete_session("sess-delete")
    assert repo.read_session("sess-delete") is None


def test_delete_session_cascades_to_agents_and_messages(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)

    sess = _session("sess-cascade")
    repo.create_session(sess)
    repo.create_agent("sess-cascade", _agent("a1"))
    repo.create_message("sess-cascade", "a1", _message(0))

    repo.delete_session("sess-cascade")

    assert repo.read_session("sess-cascade") is None
    assert repo.read_agent("sess-cascade", "a1") is None


# ---------------------------------------------------------------------------
# create_agent / read_agent / update_agent
# ---------------------------------------------------------------------------


def test_create_and_read_agent(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)

    repo.create_session(_session("sess-agent"))
    repo.create_agent("sess-agent", _agent("a1"))
    result = repo.read_agent("sess-agent", "a1")

    assert result is not None
    assert result.agent_id == "a1"


def test_read_agent_missing_returns_none(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    repo.create_session(_session("sess-agent-missing"))
    assert repo.read_agent("sess-agent-missing", "nonexistent") is None


def test_create_agent_duplicate_raises(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    repo.create_session(_session("sess-agent-dup"))
    repo.create_agent("sess-agent-dup", _agent("a1"))

    with pytest.raises(SessionException, match="already exists"):
        repo.create_agent("sess-agent-dup", _agent("a1"))


def test_update_agent_stores_new_state(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)

    repo.create_session(_session("sess-update-agent"))
    repo.create_agent("sess-update-agent", _agent("a1"))

    updated = SessionAgent(
        agent_id="a1",
        state={"turn": 1},
        conversation_manager_state={"__name__": "SlidingWindowConversationManager"},
    )
    repo.update_agent("sess-update-agent", updated)
    result = repo.read_agent("sess-update-agent", "a1")

    assert result is not None
    assert result.state == {"turn": 1}


def test_update_agent_nonexistent_raises(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    repo.create_session(_session("sess-update-agent-missing"))

    with pytest.raises(SessionException, match="does not exist"):
        repo.update_agent("sess-update-agent-missing", _agent("ghost"))


# ---------------------------------------------------------------------------
# create_message / read_message / update_message / list_messages
# ---------------------------------------------------------------------------


def _setup_session_and_agent(repo: PostgresSessionRepository, session_id: str) -> None:
    repo.create_session(Session(session_id=session_id, session_type=SessionType.AGENT))
    repo.create_agent(session_id, _agent("a1"))


def test_create_and_read_message(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    _setup_session_and_agent(repo, "sess-msg")

    repo.create_message("sess-msg", "a1", _message(0))
    result = repo.read_message("sess-msg", "a1", 0)

    assert result is not None
    assert result.message_id == 0


def test_read_message_missing_returns_none(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    _setup_session_and_agent(repo, "sess-msg-missing")
    assert repo.read_message("sess-msg-missing", "a1", 99) is None


def test_create_message_idempotent_on_conflict(pg_dsn: str) -> None:
    """ON CONFLICT DO NOTHING — second insert does not raise."""
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    _setup_session_and_agent(repo, "sess-msg-idempotent")

    msg = _message(0)
    repo.create_message("sess-msg-idempotent", "a1", msg)
    repo.create_message("sess-msg-idempotent", "a1", msg)  # must not raise

    result = repo.read_message("sess-msg-idempotent", "a1", 0)
    assert result is not None


def test_update_message_stores_new_content(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    _setup_session_and_agent(repo, "sess-msg-update")

    repo.create_message("sess-msg-update", "a1", _message(0))

    updated_msg = SessionMessage(
        message={"role": "assistant", "content": [{"type": "text", "text": "reply"}]},
        message_id=0,
    )
    repo.update_message("sess-msg-update", "a1", updated_msg)

    result = repo.read_message("sess-msg-update", "a1", 0)
    assert result is not None
    assert result.message["role"] == "assistant"


def test_update_message_nonexistent_raises(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    _setup_session_and_agent(repo, "sess-msg-update-missing")

    with pytest.raises(SessionException, match="does not exist"):
        repo.update_message("sess-msg-update-missing", "a1", _message(99))


def test_list_messages_ordered(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    _setup_session_and_agent(repo, "sess-list-msgs")

    for i in range(5):
        msg = SessionMessage(
            message={"role": "user", "content": [{"type": "text", "text": f"msg {i}"}]},
            message_id=i,
        )
        repo.create_message("sess-list-msgs", "a1", msg)

    results = repo.list_messages("sess-list-msgs", "a1")
    assert [m.message_id for m in results] == list(range(5))


def test_list_messages_with_limit_and_offset(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    _setup_session_and_agent(repo, "sess-list-paged")

    for i in range(6):
        msg = SessionMessage(
            message={"role": "user", "content": [{"type": "text", "text": f"msg {i}"}]},
            message_id=i,
        )
        repo.create_message("sess-list-paged", "a1", msg)

    page = repo.list_messages("sess-list-paged", "a1", limit=3, offset=2)
    assert len(page) == 3
    assert page[0].message_id == 2


# ---------------------------------------------------------------------------
# read_conversation_manager_name
# ---------------------------------------------------------------------------


def test_read_conversation_manager_name_present(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)

    repo.create_session(_session("sess-cm"))
    agent = SessionAgent(
        agent_id="a1",
        state={},
        conversation_manager_state={"__name__": "SlidingWindowConversationManager"},
    )
    repo.create_agent("sess-cm", agent)

    name = repo.read_conversation_manager_name("sess-cm")
    assert name == "SlidingWindowConversationManager"


def test_read_conversation_manager_name_no_session_returns_none(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)
    assert repo.read_conversation_manager_name("sess-no-cm") is None


# ---------------------------------------------------------------------------
# session_lock — advisory lock round-trip
# ---------------------------------------------------------------------------


def test_session_lock_acquires_and_releases(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)

    with repo.session_lock("sess-lock"):
        # While inside the context, we should be holding the advisory lock.
        # Verify by trying a second connection that selects pg_advisory_lock with trylock.
        with psycopg.connect(pg_dsn) as conn:
            row = conn.execute(
                "SELECT pg_try_advisory_lock(hashtext('strands:sess-lock')::bigint)"
            ).fetchone()
            # The lock is already held by the first connection; trylock should return false.
            assert row is not None
            lock_acquired = row[0]
            assert lock_acquired is False


def test_session_lock_released_after_context(pg_dsn: str) -> None:
    _clean(pg_dsn)
    repo = _repo(pg_dsn)

    with repo.session_lock("sess-lock-release"):
        pass

    # After exiting the context, the lock should be released.
    with psycopg.connect(pg_dsn) as conn:
        row = conn.execute(
            "SELECT pg_try_advisory_lock(hashtext('strands:sess-lock-release')::bigint)"
        ).fetchone()
        assert row is not None
        assert row[0] is True
        # Release the lock we just acquired in this connection.
        conn.execute("SELECT pg_advisory_unlock(hashtext('strands:sess-lock-release')::bigint)")


# ---------------------------------------------------------------------------
# Concurrency — parallel creates do not corrupt state
# ---------------------------------------------------------------------------


def test_concurrent_session_creates_are_isolated(pg_dsn: str) -> None:
    """Multiple threads writing distinct sessions must not interfere."""
    _clean(pg_dsn)
    errors: list[Exception] = []

    def _write(idx: int) -> None:
        try:
            repo = _repo(pg_dsn)
            sid = f"sess-concurrent-{idx}"
            repo.create_session(Session(session_id=sid, session_type=SessionType.AGENT))
            repo.create_agent(sid, _agent("a1"))
            repo.create_message(sid, "a1", _message(0))
        except Exception as exc:  # noqa: BLE001
            errors.append(exc)

    threads = [threading.Thread(target=_write, args=(i,)) for i in range(8)]
    for t in threads:
        t.start()
    for t in threads:
        t.join()

    assert errors == [], f"concurrent writes produced errors: {errors}"

    # Verify all sessions are readable
    repo = _repo(pg_dsn)
    for i in range(8):
        assert repo.read_session(f"sess-concurrent-{i}") is not None
