from __future__ import annotations

import logging
from collections.abc import Iterator
from contextlib import contextmanager
from typing import Any

import psycopg
from psycopg.errors import UniqueViolation
from psycopg.rows import dict_row
from psycopg.types.json import Jsonb
from strands.session import SessionRepository
from strands.types.exceptions import SessionException
from strands.types.session import Session, SessionAgent, SessionMessage


class PostgresSessionRepository(SessionRepository):
    def __init__(self, dsn: str) -> None:
        self._dsn = dsn
        self._logger = logging.getLogger(__name__)
        self._ensure_schema()

    def _connect(self) -> psycopg.Connection:
        return psycopg.connect(self._dsn, row_factory=dict_row)

    def _ensure_schema(self) -> None:
        statements = [
            """
            CREATE TABLE IF NOT EXISTS strands_sessions (
                session_id TEXT PRIMARY KEY,
                data JSONB NOT NULL
            )
            """,
            """
            CREATE TABLE IF NOT EXISTS strands_agents (
                session_id TEXT NOT NULL,
                agent_id TEXT NOT NULL,
                data JSONB NOT NULL,
                PRIMARY KEY (session_id, agent_id),
                FOREIGN KEY (session_id) REFERENCES strands_sessions(session_id) ON DELETE CASCADE
            )
            """,
            """
            CREATE TABLE IF NOT EXISTS strands_messages (
                session_id TEXT NOT NULL,
                agent_id TEXT NOT NULL,
                message_id INTEGER NOT NULL,
                data JSONB NOT NULL,
                PRIMARY KEY (session_id, agent_id, message_id),
                FOREIGN KEY (session_id, agent_id)
                    REFERENCES strands_agents(session_id, agent_id)
                    ON DELETE CASCADE
            )
            """,
            """
            CREATE INDEX IF NOT EXISTS strands_messages_lookup_idx
            ON strands_messages(session_id, agent_id, message_id)
            """,
        ]

        with self._connect() as conn:
            with conn.cursor() as cur:
                for statement in statements:
                    cur.execute(statement)

    @contextmanager
    def session_lock(self, session_id: str) -> Iterator[None]:
        lock_key = f"strands:{session_id}"
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    "SELECT pg_advisory_lock(hashtext(%s)::bigint)",
                    (lock_key,),
                )
            try:
                yield
            finally:
                with conn.cursor() as cur:
                    cur.execute(
                        "SELECT pg_advisory_unlock(hashtext(%s)::bigint)",
                        (lock_key,),
                    )

    def create_session(self, session: Session, **kwargs: Any) -> Session:
        try:
            with self._connect() as conn:
                with conn.cursor() as cur:
                    cur.execute(
                        "INSERT INTO strands_sessions (session_id, data) VALUES (%s, %s)",
                        (session.session_id, Jsonb(session.to_dict())),
                    )
        except UniqueViolation as exc:
            raise SessionException(f"Session {session.session_id} already exists") from exc

        return session

    def read_session(self, session_id: str, **kwargs: Any) -> Session | None:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    "SELECT data FROM strands_sessions WHERE session_id = %s",
                    (session_id,),
                )
                row = cur.fetchone()
        if not row:
            return None
        return Session.from_dict(row["data"])

    def create_agent(self, session_id: str, session_agent: SessionAgent, **kwargs: Any) -> None:
        try:
            with self._connect() as conn:
                with conn.cursor() as cur:
                    cur.execute(
                        """
                        INSERT INTO strands_agents (session_id, agent_id, data)
                        VALUES (%s, %s, %s)
                        """,
                        (session_id, session_agent.agent_id, Jsonb(session_agent.to_dict())),
                    )
        except UniqueViolation as exc:
            raise SessionException(
                f"Agent {session_agent.agent_id} in session {session_id} already exists"
            ) from exc

    def read_agent(self, session_id: str, agent_id: str, **kwargs: Any) -> SessionAgent | None:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT data FROM strands_agents
                    WHERE session_id = %s AND agent_id = %s
                    """,
                    (session_id, agent_id),
                )
                row = cur.fetchone()
        if not row:
            return None
        return SessionAgent.from_dict(row["data"])

    def update_agent(self, session_id: str, session_agent: SessionAgent, **kwargs: Any) -> None:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    UPDATE strands_agents
                    SET data = %s
                    WHERE session_id = %s AND agent_id = %s
                    """,
                    (Jsonb(session_agent.to_dict()), session_id, session_agent.agent_id),
                )
                if cur.rowcount == 0:
                    raise SessionException(
                        f"Agent {session_agent.agent_id} in session {session_id} does not exist"
                    )

    def create_message(
        self, session_id: str, agent_id: str, session_message: SessionMessage, **kwargs: Any
    ) -> None:
        try:
            with self._connect() as conn:
                with conn.cursor() as cur:
                    cur.execute(
                        """
                        INSERT INTO strands_messages (session_id, agent_id, message_id, data)
                        VALUES (%s, %s, %s, %s)
                        """,
                        (
                            session_id,
                            agent_id,
                            session_message.message_id,
                            Jsonb(session_message.to_dict()),
                        ),
                    )
        except UniqueViolation as exc:
            raise SessionException(
                f"Message {session_message.message_id} already exists for agent {agent_id}"
            ) from exc

    def read_message(
        self, session_id: str, agent_id: str, message_id: int, **kwargs: Any
    ) -> SessionMessage | None:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT data FROM strands_messages
                    WHERE session_id = %s AND agent_id = %s AND message_id = %s
                    """,
                    (session_id, agent_id, message_id),
                )
                row = cur.fetchone()
        if not row:
            return None
        return SessionMessage.from_dict(row["data"])

    def update_message(
        self, session_id: str, agent_id: str, session_message: SessionMessage, **kwargs: Any
    ) -> None:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    UPDATE strands_messages
                    SET data = %s
                    WHERE session_id = %s AND agent_id = %s AND message_id = %s
                    """,
                    (
                        Jsonb(session_message.to_dict()),
                        session_id,
                        agent_id,
                        session_message.message_id,
                    ),
                )
                if cur.rowcount == 0:
                    raise SessionException(f"Message {session_message.message_id} does not exist")

    def list_messages(
        self,
        session_id: str,
        agent_id: str,
        limit: int | None = None,
        offset: int = 0,
        **kwargs: Any,
    ) -> list[SessionMessage]:
        query = """
            SELECT data FROM strands_messages
            WHERE session_id = %s AND agent_id = %s
            ORDER BY message_id ASC
        """
        params: list[Any] = [session_id, agent_id]
        if limit is not None:
            query += " LIMIT %s"
            params.append(limit)
        if offset:
            query += " OFFSET %s"
            params.append(offset)

        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(query, params)
                rows = cur.fetchall()
        return [SessionMessage.from_dict(row["data"]) for row in rows]
