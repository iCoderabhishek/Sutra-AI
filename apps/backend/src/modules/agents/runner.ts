import type { Agent, JobRun } from "@sutra/db";
import { agentQueue } from "../../libs/queue";
import { scheduleAgent } from "./scheduler";

export const runAgent = async (agent: Agent, run: JobRun) => {

    if (agent.schedule) {
        const scheduleData = agent.schedule as { pattern?: string, cron?: string };
        const cronPattern = scheduleData.pattern || scheduleData.cron;

        if (cronPattern) {
            const jobScheduler = await agentQueue.getJobScheduler(`agent-${agent.id}`);

            if (!jobScheduler) {
                await scheduleAgent(agent);
            }
        } else {
            console.warn(`Agent ${agent.id} has schedule data but no 'pattern' or 'cron' string was found.`);
        }
    }

    // Enqueue the immediate job (works for both manual triggers and scheduled runs)
    await agentQueue.add("agent-run", { agentId: agent.id, runId: run.id });
}