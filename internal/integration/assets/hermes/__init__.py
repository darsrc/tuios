"""Hermes plugin installed by dartuios to report the session a pane runs."""

# installed by dartuios
# managed by dartuios; `dartuios integration install hermes` overwrites this file and
# `dartuios integration uninstall hermes` removes it.
# DARTUIOS_INTEGRATION_ID=hermes
# DARTUIOS_INTEGRATION_VERSION=__DARTUIOS_VERSION__
#
# Reports the session id of an interactive Hermes session to the dartuios pane it
# runs in, through `dartuios agent-hook hermes`, so the pane can be resumed. The
# pane's state is left to its screen rules.

from __future__ import annotations

import json
import os
import subprocess

_dartuios = __DARTUIOS_COMMAND__
_INTERACTIVE = {"cli", "tui", "desktop", "acp"}


def _enabled() -> bool:
    return os.environ.get("DARTUIOS_ENV") == "1" or bool(os.environ.get("DARTUIOS_AGENT"))


def _report(event: str, **kwargs) -> None:
    if not _enabled() or kwargs.get("platform") not in _INTERACTIVE:
        return
    session_id = kwargs.get("session_id")
    if not isinstance(session_id, str) or not session_id:
        return
    payload = json.dumps({"hook_event_name": event, "session_id": session_id})
    try:
        options = {
            "input": payload.encode(),
            "stdout": subprocess.DEVNULL,
            "stderr": subprocess.DEVNULL,
            "timeout": 2,
            "check": False,
        }
        if os.name == "nt":
            options["creationflags"] = subprocess.CREATE_NO_WINDOW
        subprocess.run([_dartuios, "agent-hook", "hermes", "--integration", "__DARTUIOS_VERSION__"], **options)
    except Exception:
        # A report that cannot be sent must never break Hermes.
        pass


def _session_started(**kwargs) -> None:
    _report("on_session_start", **kwargs)


def _session_reset(**kwargs) -> None:
    _report("on_session_reset", **kwargs)


def register(ctx):
    ctx.register_hook("on_session_start", _session_started)
    ctx.register_hook("on_session_reset", _session_reset)
