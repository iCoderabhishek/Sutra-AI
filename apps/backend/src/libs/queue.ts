import { Queue, Worker } from "bullmq";
import { Redis } from "ioredis";
import { REDIS_URL } from "./env";
import { prisma } from "@sutra/db";
import { hasEnoughCredits, deductCredits, usdToCredits } from "./credits";
import { streamAgentRun, triggerAgentRun, type TraceEvent } from "./agent-proxy";

const connection = () => new Redis(REDIS_URL, { maxRetriesPerRequest: null });

// Agent.prompt and Agent.instruction are Prisma Json columns, and the Zod
// schema accepts objects, so they arrive as either a bare string or a wrapper
const asText = (value: unknown): string | undefined => {
    if (value === null || value === undefined) return undefined;
    if (typeof value === "string") return value.trim() || undefined;

    if (typeof value === "object") {
        const bag = value as Record<string, unknown>;
        for (const key of ["goal", "prompt", "text", "content", "instruction", "value"]) {
            const found = bag[key];
            if (typeof found === "string" && found.trim()) return found.trim();
        }
        return JSON.stringify(value);
    }

    return String(value);
};

export const agentQueue = new Queue("agent-queue", { connection: connection() });

export const worker = new Worker("agent-queue", async (job) => {
    const { agentId, runId } = job.data;
    console.log(`{worker} Processing agent ${agentId}`);

    // user is included so the run can be delivered to the agent's owner.
    const agent = await prisma.agent.findUnique({
        where: { id: agentId },
        include: { user: true },
    });

    if (!agent || !agent.userId) {
        console.error(`{worker} Agent ${agentId} not found or has no user`);
        return;
    }

    const jobRun = runId
        ? await prisma.jobRun.update({
            where: { id: runId },
            data: { status: "RUNNING", startedAt: new Date() },
        })
        : await prisma.jobRun.create({
            data: { agentId, status: "RUNNING", startedAt: new Date() },
        });

    const canRun = await hasEnoughCredits(agent.userId, 1);

    if (!canRun.hasEnoughCredits) {
        await prisma.jobRun.update({
            where: { id: jobRun.id },
            data: {
                status: "INSUFFICIENT_CREDITS",
                trace: [{
                    type: "error",
                    message: "Insufficient credits to run this agent. Please top up your balance.",
                    timestamp: new Date().toISOString()
                }],
                finishedAt: new Date()
            },
        });
        console.warn(`{worker} Agent ${agentId} paused: Insufficient credits for user ${agent.userId}`);
        return;
    }

    const goal = asText(agent.prompt);

    if (!goal) {
        await prisma.jobRun.update({
            where: { id: jobRun.id },
            data: {
                status: "FAILED",
                trace: [{
                    type: "error",
                    message: "Agent has no prompt, nothing to run.",
                    timestamp: new Date().toISOString(),
                }],
                finishedAt: new Date(),
            },
        });
        console.error(`{worker} Agent ${agentId} has no usable prompt`);
        return;
    }

    // Delivery is opt-in: only agents configured with send_email get an
    // address, so enabling the tool is what turns on emailing.
    const wantsEmail = agent.tools.includes("send_email");

    const triggered = await triggerAgentRun({
        goal,
        tools: agent.tools,
        run_id: jobRun.id,
        template: agent.template ?? undefined,
        instruction: asText(agent.instruction),
        email: wantsEmail ? agent.user?.email : undefined,
    });

    if (triggered.error) {
        await prisma.jobRun.update({
            where: { id: jobRun.id },
            data: {
                status: "FAILED",
                trace: [{ type: "error", message: triggered.error, timestamp: new Date().toISOString() }],
                finishedAt: new Date(),
            },
        });
        throw new Error(triggered.error);
    }

    const trace: TraceEvent[] = [];
    let totalTokens = 0;
    let costUsd = 0;
    let failed = false;

    try {
        await streamAgentRun(jobRun.id, (event) => {
            trace.push(event);
            if (event.cost) {
                totalTokens = event.cost.total_tokens;
                costUsd = event.cost.cost_usd;
            }
            if (event.status === "error") failed = true;
        });
    } catch (err) {
        failed = true;
        trace.push({
            step: "Stream Error",
            status: "error",
            content: err instanceof Error ? err.message : String(err),
        });
    }

    await prisma.jobRun.update({
        where: { id: jobRun.id },
        data: {
            status: failed ? "FAILED" : "SUCCEEDED",
            trace: trace as any,
            totalCost: costUsd,
            totalTokens,
            finishedAt: new Date(),
        },
    });

    const credits = Math.max(usdToCredits(costUsd), 1);
    const deducted = await deductCredits(agent.userId, credits);

    if (!deducted) {
        console.warn(`{worker} Could not deduct ${credits} credits from user ${agent.userId}`);
    }

    console.log(`{worker} Run ${jobRun.id} ${failed ? "failed" : "succeeded"} — ${totalTokens} tokens, $${costUsd}`);
}, { connection: connection() });
