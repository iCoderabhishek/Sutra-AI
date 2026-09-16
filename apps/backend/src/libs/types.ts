export interface RunUsage {
    pageScraped: number;
    promptTokens: number;
    completionTokens: number
}


export interface AgentRun {
    error?: string;
    success?: boolean;
    message?: string;
    runId?: string;
}