import { AGENT_BACKEND_URL, AGENT_SHARED_SECRET } from "./env";
import type { AgentRun } from "./types";

export const triggerAgentRun = async (payload: {
    goal: string;
    tools: string[];
}): Promise<AgentRun> => {

    try {
        // fetch the agent url
        const res = await fetch(`${AGENT_BACKEND_URL}/agents/run`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Agent-Secret': AGENT_SHARED_SECRET,
            },
            body: JSON.stringify(payload)
        });

        if (!res.ok) {
            return { error: "Failed to trigger agent" };
        }

        const data = await res.json();
        return data as AgentRun;
    } catch (err) {
        console.error("Agent proxy error:", err);
        return { error: "Failed to trigger agent" };
    }
}