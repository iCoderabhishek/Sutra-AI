import z from "zod";

const AgentSchema = z.object({
    name: z.string(),
    desc: z.string().optional(),
    prompt: z.object().optional(),
    template: z.string().optional(),
    instruction: z.object().optional(),
    tools: z.array(z.string()).optional(),
    schedule: z.object().optional(),
    status: z.enum(["ACTIVE", "INACTIVE", "PAUSED"]),
});

export const PaginationSchema = z.object({
    page: z.coerce.number().min(1).default(1),
    limit: z.coerce.number().min(1).max(100).default(10),
});


export default AgentSchema;