import type { Agent } from "@sutra/db";
import { prisma } from "@sutra/db";
import { agentQueue } from "../../libs/queue";

export const schedulerId = (agentId: string) => `agent-${agentId}`;

export const cronOf = (schedule: unknown): string | undefined => {
    if (!schedule || typeof schedule !== "object") return undefined;
    const data = schedule as { pattern?: string; cron?: string };
    return data.pattern || data.cron || undefined;
};

export const scheduleAgent = async (agent: Agent) => {
    const cronPattern = cronOf(agent.schedule);

    if (!cronPattern) {
        if (agent.schedule) {
            console.warn(`Agent ${agent.id} has schedule data but no 'pattern' or 'cron' string.`);
        }
        return;
    }

    await agentQueue.upsertJobScheduler(
        schedulerId(agent.id),
        { pattern: cronPattern },
        { name: "agent-run", data: { agentId: agent.id } },
    );
    console.log(`Scheduled agent ${agent.id} with pattern ${cronPattern}`);
};

export const unscheduleAgent = async (agentId: string) => {
    try {
        await agentQueue.removeJobScheduler(schedulerId(agentId));
        console.log(`Unscheduled agent ${agentId}`);
    } catch (error) {
        console.warn(`Could not remove scheduler for agent ${agentId}:`, error);
    }
};

export const syncAgentSchedule = async (agent: Agent) => {
    if (agent.status === "ACTIVE" && cronOf(agent.schedule)) {
        await scheduleAgent(agent);
    } else {
        await unscheduleAgent(agent.id);
    }
};

// BullMQ stores repeatable jobs in Redis only, so a flushed Redis would drop
// every cron silently. Re-register from Postgres on boot.
export const reconcileSchedules = async () => {
    try {
        const agents = await prisma.agent.findMany({ where: { status: "ACTIVE" } });

        let restored = 0;
        for (const agent of agents) {
            if (!cronOf(agent.schedule)) continue;
            await scheduleAgent(agent);
            restored += 1;
        }

        console.log(`{scheduler} Reconciled ${restored} agent schedule(s) on boot`);
    } catch (error) {
        console.error("{scheduler} Failed to reconcile schedules:", error);
    }
};
