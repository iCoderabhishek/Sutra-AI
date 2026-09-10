import type { Agent } from "@sutra/db";
import { agentQueue } from "../../libs/queue";

export const scheduleAgent = async (agent: Agent) => {
    if (agent.schedule) {
        // Since schedule is a Json object, we cast it to extract the cron pattern
        const scheduleData = agent.schedule as { pattern?: string, cron?: string };
        const cronPattern = scheduleData.pattern || scheduleData.cron;

        if (cronPattern) {
            await agentQueue.add(
                "agent-run",
                { agentId: agent.id },
                {
                    repeat: { pattern: cronPattern },
                    jobId: `agent-${agent.id}` // Assigns a predictable ID to prevent duplicate schedules
                }
            );
            console.log(`Scheduled agent ${agent.id} with pattern ${cronPattern}`);
        } else {
            console.warn(`Agent ${agent.id} has schedule data but no 'pattern' or 'cron' string was found.`);
        }
    }
}