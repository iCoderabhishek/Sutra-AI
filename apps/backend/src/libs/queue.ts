import { Queue, Worker } from "bullmq";
import redis from "./redis";

export const agentQueue = new Queue("agent-queue", { connection: redis });

export const worker = new Worker("agent-queue", async (job) => {
    const { agentId } = job.data;
    console.log(`{worker} Processing agent ${agentId}`);

}, { connection: redis });