import { Queue, Worker } from "bullmq";
import redis from "./redis";

export const agentQueue = new Queue("agent-queue", { connection: redis });

export const worker = new Worker("agent-queue", async (job) => {
    try {
        const { agentId } = job.data;
        console.log(`{worker} Processing agent ${agentId}`);
    } catch (error) {
        console.log(`{worker} Error processing agent ${job.data.agentId}`, error);
    }
}, { connection: redis });