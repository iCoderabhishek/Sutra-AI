import asyncio
import secrets
import uuid
from contextlib import asynccontextmanager

from fastapi import Depends, FastAPI, Header, HTTPException
from fastapi.responses import StreamingResponse

from models.requests import RunRequest
from models.events import TraceEvent
from agent.loop import run_agent
from libs.env import settings
from libs.redis_client import redis_client

# One Redis Stream per run: pub/sub would drop events published before the SSE
# client subscribes, a stream replays from 0.
END = "__END__"
TTL = 60 * 60


def key(run_id: str) -> str:
    return f"sutra:run:{run_id}:events"


async def publish(run_id: str, payload: str) -> None:
    await redis_client.xadd(key(run_id), {"data": payload})
    await redis_client.expire(key(run_id), TTL)


# Caps concurrent agent loops. A leaked shared secret would otherwise let one
run_slots = asyncio.Semaphore(settings.MAX_CONCURRENT_RUNS)


async def require_secret(x_agent_secret: str = Header(default="")) -> None:
    # compare_digest, not ==, so a wrong guess can't be narrowed down by timing.
    if not secrets.compare_digest(x_agent_secret, settings.AGENT_SHARED_SECRET):
        raise HTTPException(status_code=401, detail="Unauthorized")


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Refuse to boot unauthenticated rather than silently exposing /run.
    if len(settings.AGENT_SHARED_SECRET) < 32:
        raise RuntimeError(
            "AGENT_SHARED_SECRET must be set to at least 32 characters. "
            "Generate one with: python -c \"import secrets; print(secrets.token_urlsafe(32))\""
        )
    await redis_client.ping()
    yield
    await redis_client.aclose()


app = FastAPI(title="Sutra Agent Runtime", version="0.1.0", lifespan=lifespan)


@app.get("/")
async def health():
    return {"status": "ok", "service": "agent-runtime"}


@app.post("/agents/run", dependencies=[Depends(require_secret)])
async def start_run(body: RunRequest):
    run_id = body.run_id or str(uuid.uuid4())

    async def emit(event: dict):
        await publish(run_id, TraceEvent(**event).model_dump_json())

    # Creates the stream so /stream can tell an unknown run from a silent one.
    await emit({"step": "Run queued", "status": "running"})

    async def run_and_signal():
        try:
            # Queued rather than rejected, so a scheduled run still completes.
            if run_slots.locked():
                await emit({
                    "step": "Waiting for a free run slot",
                    "status": "running",
                })

            async with run_slots:
                await run_agent(
                    goal=body.goal,
                    stream_callback=emit,
                    tools=body.tools,
                    system_prompt=body.system_prompt,
                    template=body.template,
                    instruction=body.instruction,
                    email=body.email,
                )
        except Exception as e:
            await emit({"step": "Runtime Error", "status": "error", "content": str(e)})
        finally:
            await publish(run_id, END)

    asyncio.create_task(run_and_signal())

    return {"run_id": run_id}


@app.get("/agents/run/{run_id}/stream", dependencies=[Depends(require_secret)])
async def stream_run(run_id: str):
    if not await redis_client.exists(key(run_id)):
        raise HTTPException(status_code=404, detail=f"No run found with id '{run_id}'")

    async def event_generator():
        last_id = "0-0"
        while True:
            response = await redis_client.xread({key(run_id): last_id}, count=100, block=15_000)
            if not response:
                yield ": keep-alive\n\n"
                continue
            for _stream, entries in response:
                for entry_id, fields in entries:
                    last_id = entry_id
                    payload = fields["data"]
                    if payload == END:
                        return
                    yield f"data: {payload}\n\n"

    return StreamingResponse(
        event_generator(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="127.0.0.1", port=8000, reload=True)
