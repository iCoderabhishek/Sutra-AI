import { AGENT_BACKEND_URL } from "./env";
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