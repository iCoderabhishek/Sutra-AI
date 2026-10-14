import { prisma } from "@sutra/db";
import type { Request, Response, NextFunction } from "express";
import { PaginationSchema } from "../agents/schema";
import { RunNotStartedError, streamAgentRun } from "../../libs/agent-proxy";

// JobRun has no userId, so ownership is scoped through the parent Agent.
const ownedBy = (userId: string) => ({ agent: { userId } });

export const listRuns = async (req: Request, res: Response, next: NextFunction) => {
    const result = PaginationSchema.safeParse(req.query);

    if (!result.success) {
        return res.status(400).json({ message: "Invalid pagination parameters" });
    }

    const { page, limit } = result.data;
    const { agentId, status } = req.query as { agentId?: string; status?: string };

    const where = {
        ...ownedBy(req.user.userId),
        ...(agentId ? { agentId } : {}),
        ...(status ? { status: status as never } : {}),
    };

    try {
        const [runs, total] = await Promise.all([
            prisma.jobRun.findMany({
                where,
                skip: (page - 1) * limit,
                take: limit,
                orderBy: { createdAt: "desc" },
                // trace omitted: hundreds of KB per row, unused by the list view.
                select: {
                    id: true,
                    agentId: true,
                    status: true,
                    totalCost: true,
                    totalTokens: true,
                    startedAt: true,
                    finishedAt: true,
                    createdAt: true,
                    agent: { select: { name: true, template: true } },
                },
            }),
            prisma.jobRun.count({ where }),
        ]);

        return res.status(200).json({
            data: runs.map((run) => ({ ...run, totalCost: Number(run.totalCost) })),
            meta: { total, page, limit, totalPages: Math.ceil(total / limit) },
        });
    } catch (error) {
        return next(error);
    }
};

export const getRun = async (req: Request, res: Response, next: NextFunction) => {
    const { runId } = req.params as { runId: string };

    try {
        const run = await prisma.jobRun.findFirst({
            where: { id: runId, ...ownedBy(req.user.userId) },
            include: { agent: { select: { id: true, name: true, template: true } } },
        });

        if (!run) return res.status(404).json({ message: "Run not found" });

        return res.status(200).json({ ...run, totalCost: Number(run.totalCost) });
    } catch (error) {
        return next(error);
    }
};

// SSE proxy: the browser can't reach the runtime, which is loopback-bound and
// holds the shared secret. Finished runs replay the persisted trace, because the
// Redis stream has a 1h TTL.
export const streamRun = async (req: Request, res: Response, next: NextFunction) => {
    const { runId } = req.params as { runId: string };

    const run = await prisma.jobRun.findFirst({
        where: { id: runId, ...ownedBy(req.user.userId) },
        select: { id: true, status: true, trace: true },
    });

    if (!run) return res.status(404).json({ message: "Run not found" });

    res.writeHead(200, {
        "Content-Type": "text/event-stream",
        "Cache-Control": "no-cache",
        Connection: "keep-alive",
        "X-Accel-Buffering": "no",
    });

    const send = (event: unknown) => res.write(`data: ${JSON.stringify(event)}\n\n`);

    const isLive = (status: string) => status === "QUEUED" || status === "RUNNING";

    const replay = (trace: unknown) => {
        for (const event of (Array.isArray(trace) ? trace : [])) send(event);
        send({ step: "Replay complete", status: "done" });
    };

    if (!isLive(run.status)) {
        replay(run.trace);
        return res.end();
    }

    let closed = false;
    req.on("close", () => { closed = true; });

    // A just-triggered run sits in the job queue before the worker hands it to
    // the runtime, and the runtime 404s until then. Wait for it instead of failing.
    const deadline = Date.now() + STREAM_START_TIMEOUT_MS;

    while (!closed) {
        try {
            await streamAgentRun(runId, (event) => {
                if (!closed) send(event);
            });
            break;
        } catch (err) {
            if (!(err instanceof RunNotStartedError)) {
                if (!closed) {
                    send({
                        step: "Stream Error",
                        status: "error",
                        content: err instanceof Error ? err.message : String(err),
                    });
                }
                break;
            }

            // Not started yet. The worker may also have finished it without the
            // runtime (no credits, no prompt, runtime down), so check the DB.
            const latest = await prisma.jobRun.findUnique({
                where: { id: runId },
                select: { status: true, trace: true },
            });

            if (!latest || !isLive(latest.status)) {
                if (!closed) replay(latest?.trace);
                break;
            }

            if (Date.now() > deadline) {
                if (!closed) {
                    send({
                        step: "Stream Error",
                        status: "error",
                        content: "Run is still queued. Is the backend worker running?",
                    });
                }
                break;
            }

            res.write(": waiting for run to start\n\n"); // keeps proxies from timing out
            await new Promise((resolve) => setTimeout(resolve, STREAM_RETRY_MS));
        }
    }

    return res.end();
};

const STREAM_RETRY_MS = 500;
const STREAM_START_TIMEOUT_MS = 2 * 60 * 1000;
