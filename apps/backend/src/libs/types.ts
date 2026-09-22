export interface RunUsage {
    pagesScraped: number;
    emailsSent: number;
}


export interface AgentRun {
    error?: string;
    success?: boolean;
    message?: string;
    runId?: string;
    run_id?: string;
}
