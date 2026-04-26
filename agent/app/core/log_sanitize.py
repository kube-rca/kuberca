"""Log-injection defence helper.

User-controlled values (HTTP request fields, query parameters, headers, etc.)
must never be inserted into log records verbatim. CR/LF/NUL bytes can forge
new log lines, break structured-log parsers, or smuggle ANSI escapes into
operator terminals (CWE-117 - Log Injection / Log Forging).

Wrap any such value with :func:`sanitize_log` before passing it to a logger:

    logger.info("provider=%s", sanitize_log(request.provider))
"""

from __future__ import annotations


def sanitize_log(value: object) -> str:
    """Return a log-safe string by replacing CR/LF/NUL with a single space.

    Non-string inputs are coerced via :func:`str` so the helper is safe to use
    on ints, dicts, exceptions, or any other object that may flow into a log
    statement.

    Implementation uses chained ``str.replace`` calls rather than ``re.sub``
    because CodeQL's Python data-flow library recognises the
    ``str.replace`` chain pattern as a built-in barrier for the
    ``py/log-injection`` rule, while ``re.sub`` is not auto-detected as a
    sanitiser. The behaviour (CR/LF -> space, NUL stripped) is identical.
    """
    return str(value).replace("\r", " ").replace("\n", " ").replace("\x00", " ")
