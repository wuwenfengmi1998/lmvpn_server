import os
import json
import pytest
import logging
from websockets.asyncio.client import connect

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

WS_URL = os.environ.get("LMVPN_WS_URL", "ws://localhost:8080/ws")
AUTH_USERNAME = os.environ.get("LMVPN_AUTH_USERNAME", "admin")
AUTH_PASSWORD = os.environ.get("LMVPN_AUTH_PASSWORD", "admin123")


@pytest.fixture
async def raw_ws():
    """仅建立 WebSocket 连接，不认证"""
    logger.info("connecting to %s", WS_URL)
    async with connect(WS_URL, ping_interval=None) as ws:
        yield ws


@pytest.fixture
async def authenticated_ws():
    """建立 WebSocket 连接并完成认证"""
    logger.info("connecting to %s", WS_URL)
    async with connect(WS_URL, ping_interval=None) as ws:
        auth_msg = json.dumps({"type": "auth", "username": AUTH_USERNAME, "password": AUTH_PASSWORD})
        await ws.send(auth_msg)
        resp = await ws.recv()
        resp_data = json.loads(resp)
        assert resp_data.get("type") == "auth_ok", f"认证失败: {resp_data}"
        logger.info("authenticated as %s", AUTH_USERNAME)
        yield ws
