import json
import asyncio
import pytest
from websockets import ConnectionClosed

pytestmark = pytest.mark.asyncio


# ============================================================
# 认证测试
# ============================================================

async def test_auth_success(raw_ws):
    """正确凭据应返回 auth_ok"""
    msg = json.dumps({"type": "auth", "username": "admin", "password": "admin123"})
    await raw_ws.send(msg)
    resp = await raw_ws.recv()
    data = json.loads(resp)
    assert data["type"] == "auth_ok"


async def test_auth_wrong_password(raw_ws):
    """错误密码应返回 auth_err"""
    msg = json.dumps({"type": "auth", "username": "admin", "password": "wrong"})
    await raw_ws.send(msg)
    resp = await raw_ws.recv()
    data = json.loads(resp)
    assert data["type"] == "auth_err"


async def test_auth_nonexistent_user(raw_ws):
    """不存在的用户名应返回 auth_err"""
    msg = json.dumps({"type": "auth", "username": "no_such_user", "password": "123456"})
    await raw_ws.send(msg)
    resp = await raw_ws.recv()
    data = json.loads(resp)
    assert data["type"] == "auth_err"


async def test_auth_non_json(raw_ws):
    """首条消息不是 JSON，服务端应关闭连接"""
    await raw_ws.send("not json")
    with pytest.raises(ConnectionClosed):
        await raw_ws.recv()


async def test_auth_missing_type(raw_ws):
    """缺少 type 字段，服务端应关闭连接"""
    msg = json.dumps({"username": "admin", "password": "admin123"})
    await raw_ws.send(msg)
    with pytest.raises(ConnectionClosed):
        await raw_ws.recv()


# ============================================================
# 消息回显测试
# ============================================================

async def test_echo_text(authenticated_ws):
    """发送文本消息，验证回显一致"""
    await authenticated_ws.send("hello world")
    resp = await authenticated_ws.recv()
    assert resp == "hello world"


async def test_echo_binary(authenticated_ws):
    """发送二进制消息，验证回显一致"""
    data = b"\x00\x01\x02\xff\xfe\xfd"
    await authenticated_ws.send(data)
    resp = await authenticated_ws.recv()
    assert resp == data


async def test_echo_multiple_messages(authenticated_ws):
    """连续发送多条消息，逐一验证回显"""
    messages = ["msg1", "msg2", "msg3", "hello", "世界"]
    for m in messages:
        await authenticated_ws.send(m)
        resp = await authenticated_ws.recv()
        assert resp == m


async def test_echo_json_message(authenticated_ws):
    """发送 JSON 消息，验证回显一致"""
    payload = json.dumps({"cmd": "ping", "seq": 1, "data": "test"})
    await authenticated_ws.send(payload)
    resp = await authenticated_ws.recv()
    assert resp == payload


async def test_echo_large_message(authenticated_ws):
    """发送大消息（64KB），验证回显一致"""
    large_msg = "A" * 65536
    await authenticated_ws.send(large_msg)
    resp = await authenticated_ws.recv()
    assert resp == large_msg


# ============================================================
# 并发测试
# ============================================================

async def test_concurrent_connections():
    """同时建立 5 个认证连接，各自发送消息验证回显"""
    from websockets.asyncio.client import connect
    import os

    ws_url = os.environ.get("LMVPN_WS_URL", "ws://localhost:8080/ws")

    async def one_session(uid):
        async with connect(ws_url, ping_interval=None) as ws:
            auth = json.dumps({"type": "auth", "username": "admin", "password": "admin123"})
            await ws.send(auth)
            resp = await ws.recv()
            assert json.loads(resp)["type"] == "auth_ok"

            msg = f"msg_from_{uid}"
            await ws.send(msg)
            resp = await ws.recv()
            assert resp == msg

    tasks = [one_session(i) for i in range(5)]
    await asyncio.gather(*tasks)


# ============================================================
# 协议测试
# ============================================================

async def test_connection_close_after_auth_failure(raw_ws):
    """认证失败后服务端应关闭连接"""
    msg = json.dumps({"type": "auth", "username": "admin", "password": "wrong"})
    await raw_ws.send(msg)
    resp = await raw_ws.recv()
    assert json.loads(resp)["type"] == "auth_err"
    with pytest.raises(ConnectionClosed):
        await raw_ws.recv()


async def test_ping_pong(authenticated_ws):
    """发送 ping，服务端应在超时前回应 pong"""
    await authenticated_ws.ping()
