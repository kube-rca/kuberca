"""Log-injection defence helper.

User-controlled values (HTTP request fields, query parameters, headers, etc.)
must never be inserted into log records verbatim. CR/LF/NUL bytes can forge
new log lines, break structured-log parsers, or smuggle ANSI escapes into
operator terminals (CWE-117 - Log Injection / Log Forging).

Wrap any such value with :func:`sanitize_log` before passing it to a logger:

    logger.info("provider=%s", sanitize_log(request.provider))
"""

from __future__ import annotations

import re

# Strip the three control characters that are typically abused for log forging:
#   \r (CR)   - line terminator on Windows / classic Mac
#   \n (LF)   - line terminator on Unix
#   \x00 (NUL) - C-string terminator, can truncate downstream consumers
_CONTROL = re.compile(r"[\r\n\x00]")


def sanitize_log(value: object) -> str:
    """Return a log-safe string by replacing CR/LF/NUL with a single space.

    Non-string inputs are coerced via :func:`str` so the helper is safe to use
    on ints, dicts, exceptions, or any other object that may flow into a log
    statement.
    """
    return _CONTROL.sub(" ", str(value))
