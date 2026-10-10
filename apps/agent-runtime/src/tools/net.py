"""
Safe outbound HTTP for tools that take a URL from the model.

The model picks URLs, and scraped pages can steer it, so a URL tool must not be
usable to reach the runtime's own network (localhost, Redis, cloud metadata at
169.254.169.254, LAN hosts). Every request and every redirect hop is checked.

Known limit: the host is resolved, checked, then resolved again by httpx, so a
DNS record that flips between the two lookups (rebinding) is not covered.
"""

import asyncio
import ipaddress
import socket
from urllib.parse import urljoin, urlparse

import httpx

MAX_BYTES = 2 * 1024 * 1024
TIMEOUT_SECONDS = 10
MAX_REDIRECTS = 3
USER_AGENT = "SutraAgent/0.1 (+https://github.com/iCoderabhishek/Sutra-AI)"


class UnsafeURL(ValueError):
    """The URL is malformed or points at a non-public address."""


def _is_public(ip: str) -> bool:
    addr = ipaddress.ip_address(ip)
    return not (
        addr.is_private or addr.is_loopback or addr.is_link_local or addr.is_reserved
        or addr.is_multicast or addr.is_unspecified
    )


async def check_public_url(url: str) -> str:
    """Return the URL if it is http(s) and every address it resolves to is public."""
    parsed = urlparse(url.strip())
    if parsed.scheme not in ("http", "https") or not parsed.hostname:
        raise UnsafeURL("only http and https URLs are allowed")

    try:
        infos = await asyncio.to_thread(socket.getaddrinfo, parsed.hostname, parsed.port or 443)
    except socket.gaierror:
        raise UnsafeURL(f"could not resolve host '{parsed.hostname}'")

    for info in infos:
        if not _is_public(info[4][0]):
            raise UnsafeURL(f"'{parsed.hostname}' points to a private or local address")
    return url.strip()


async def fetch_text(url: str) -> tuple[str, str]:
    """
    GET a public URL, following up to MAX_REDIRECTS redirects (each re-checked).
    Returns (final_url, body). The body is capped at MAX_BYTES.
    """
    url = await check_public_url(url)
    async with httpx.AsyncClient(
        timeout=TIMEOUT_SECONDS,
        follow_redirects=False,
        headers={"User-Agent": USER_AGENT},
    ) as client:
        for _ in range(MAX_REDIRECTS + 1):
            async with client.stream("GET", url) as res:
                if res.is_redirect:
                    url = await check_public_url(urljoin(url, res.headers.get("location", "")))
                    continue
                res.raise_for_status()

                chunks, size = [], 0
                async for chunk in res.aiter_bytes():
                    size += len(chunk)
                    if size > MAX_BYTES:
                        break
                    chunks.append(chunk)
                body = b"".join(chunks)
                return url, body.decode(res.encoding or "utf-8", errors="replace")

    raise UnsafeURL("too many redirects")
