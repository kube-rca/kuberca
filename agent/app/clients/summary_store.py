from __future__ import annotations

import logging
from typing import Protocol

import psycopg
from psycopg.rows import dict_row


class SummaryStore(Protocol):
    def list_summaries(self, session_id: str, limit: int) -> list[str]:
        raise NotImplementedError

    def append_summary(self, session_id: str, summary: str, max_items: int) -> None:
        raise NotImplementedError


class PostgresSummaryStore(SummaryStore):
    def __init__(self, dsn: str) -> None:
        self._dsn = dsn
        self._logger = logging.getLogger(__name__)
        self._ensure_schema()

    def _connect(self) -> psycopg.Connection:
        return psycopg.connect(self._dsn, row_factory=dict_row)

    def _ensure_schema(self) -> None:
        statements = [
            """
            CREATE TABLE IF NOT EXISTS kube_rca_session_summaries (
                summary_id BIGSERIAL PRIMARY KEY,
                session_id TEXT NOT NULL,
                summary TEXT NOT NULL,
                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
            )
            """,
            """
            CREATE INDEX IF NOT EXISTS kube_rca_session_summaries_lookup_idx
            ON kube_rca_session_summaries(session_id, summary_id DESC)
            """,
        ]
        with self._connect() as conn:
            with conn.cursor() as cur:
                for statement in statements:
                    cur.execute(statement)

    def list_summaries(self, session_id: str, limit: int) -> list[str]:
        if limit <= 0:
            return []
        query = """
            SELECT summary FROM kube_rca_session_summaries
            WHERE session_id = %s
            ORDER BY summary_id DESC
            LIMIT %s
        """
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(query, (session_id, limit))
                rows = cur.fetchall()
        summaries = [row["summary"] for row in rows]
        return list(reversed(summaries))

    def append_summary(self, session_id: str, summary: str, max_items: int) -> None:
        if max_items <= 0:
            return
        try:
            with self._connect() as conn:
                with conn.cursor() as cur:
                    cur.execute(
                        """
                        INSERT INTO kube_rca_session_summaries (session_id, summary)
                        VALUES (%s, %s)
                        """,
                        (session_id, summary),
                    )
                    cur.execute(
                        """
                        DELETE FROM kube_rca_session_summaries
                        WHERE session_id = %s
                          AND summary_id NOT IN (
                            SELECT summary_id
                            FROM kube_rca_session_summaries
                            WHERE session_id = %s
                            ORDER BY summary_id DESC
                            LIMIT %s
                          )
                        """,
                        (session_id, session_id, max_items),
                    )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to append session summary: %s", exc)
