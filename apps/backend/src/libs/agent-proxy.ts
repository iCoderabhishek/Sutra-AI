import { AGENT_BACKEND_URL, AGENT_SHARED_SECRET } from "./env";
import type { AgentRun } from "./types";

const authHeaders = {
    'Content-Type': 'application/json',
    'X-Agent-Secret': AGENT_SHARED_SECRET,
};

export type RunPayload = {
    goal: string;
    run_id: string;
    tools?: string[];
    system_prompt?: string;
    template?: string;
    instruction?: string;
    email?: string;
};

export const triggerAgentRun = async (payload: RunPayload): Promise<AgentRun> => {

    try {
        const res = await fetch(`${AGENT_BACKEND_URL}/agents/run`, {
            method: 'POST',
            headers: authHeaders,
            body: JSON.stringify(payload)
        });

        if (!res.ok) {
            return { error: `Agent runtime returned ${res.status}` };
        }

        const data = await res.json();
        return data as AgentRun;
    } catch (err) {
        console.error("Agent proxy error:", err);
        return { error: "Failed to trigger agent" };
    }
}

export type TraceEvent = {
    step: string;
    status: "running" | "done" | "error";
    iteration?: number;
    content?: string;
    args?: Record<string, unknown>;
    result_preview?: string;
    cost?: {
        input_tokens: number;
        output_tokens: number;
        total_tokens: number;
        cost_usd: number;
        model: string;
    };
};

export const streamAgentRun = async (
    runId: string,
    onEvent: (event: TraceEvent) => void
): Promise<void> => {
    const res = await fetch(`${AGENT_BACKEND_URL}/agents/run/${runId}/stream`, {
        headers: authHeaders,
    });

    if (!res.ok || !res.body) {
        throw new Error(`Agent stream failed with ${res.status}`);
    }

    const decoder = new TextDecoder();
    let buffer = "";

    for await (const chunk of res.body as unknown as AsyncIterable<Uint8Array>) {
        buffer += decoder.decode(chunk, { stream: true });

        const frames = buffer.split("\n\n");
        buffer = frames.pop() ?? "";

        for (const frame of frames) {
            for (const line of frame.split("\n")) {
                if (!line.startsWith("data:")) continue;
                const payload = line.slice(5).trim();
                if (!payload) continue;
                try {
                    onEvent(JSON.parse(payload) as TraceEvent);
                } catch {
                    console.warn("Skipping malformed trace event:", payload);
                }
            }
        }
    }
}
