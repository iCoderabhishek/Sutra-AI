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

export default AgentSchema;