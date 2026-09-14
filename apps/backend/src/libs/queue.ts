import { Queue, Worker } from "bullmq";
import redis from "./redis";
import { prisma } from "@sutra/db";

export const agentQueue = new Queue("agent-queue", { connection: redis });

export const worker = new Worker("agent-queue", async (job) => {
    try {
        const { agentId, runId } = job.data;
        console.log(`{worker} Processing agent ${agentId}`);

        // Reuse existing JobRun (manual trigger) or create a new one (cron-scheduled)
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

        // TODO: code to execute the Python backend will go here...

    } catch (error) {
        console.log(`{worker} Error processing agent ${job?.data?.agentId}`, error);
    }
}, { connection: redis });