import { Queue, Worker } from "bullmq";
import redis from "./redis";
import { prisma } from "@sutra/db";
import { hasEnoughCredits } from "./credits";

export const agentQueue = new Queue("agent-queue", { connection: redis });

export const worker = new Worker("agent-queue", async (job) => {
    try {
        const { agentId, runId } = job.data;
        console.log(`{worker} Processing agent ${agentId}`);

        const agent = await prisma.agent.findUnique({
            where: { id: agentId },
            select: { userId: true }
        });

        if (!agent || !agent.userId) {
            console.error(`{worker} Agent ${agentId} not found or has no user`);
            return;
        }

        let jobRun;
        if (runId) {
            jobRun = await prisma.jobRun.update({
                where: { id: runId },
                data: { status: "ACTIVE", startedAt: new Date() },
            });
        } else {
            jobRun = await prisma.jobRun.create({
                data: {
                    agentId: agentId,
                    status: "ACTIVE",
                    trace: [],
                    totalCost: 0,
                    totalTokens: 0,
                    startedAt: new Date(),
                    finishedAt: new Date(),
                },
            });
        }

        const canRun = await hasEnoughCredits(agent.userId, 1);

        if (!canRun.hasEnoughCredits) {
            await prisma.jobRun.update({
                where: { id: jobRun.id },
                data: {
                    status: "PAUSED",
                    trace: [{
                        type: "error",
                        message: "Insufficient credits to run this agent. Please top up your balance.",
                        timestamp: new Date().toISOString()
                    }],
                    finishedAt: new Date()
                },
            });
            console.warn(`{worker} Agent ${agentId} paused: Insufficient credits for user ${agent.userId}`);
            return; // EXIT EARLY so the job doesn't run
        }

        //  code to execute the Python backend will go here...

    } catch (error) {
        console.log(`{worker} Error processing agent ${job?.data?.agentId}`, error);
    }
}, { connection: redis });