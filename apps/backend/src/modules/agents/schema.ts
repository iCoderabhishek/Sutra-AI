import z from "zod";

// z.object() with no shape strips every key, which silently emptied prompt,
// instruction and schedule. These columns are free-form Json.
// z.any() rather than z.unknown() so the inferred type stays assignable to
// Prisma's InputJsonValue.
const JsonObject = z.record(z.string(), z.any());

const AgentSchema = z.object({
    name: z.string(),
    desc: z.string().optional(),
    prompt: JsonObject.optional(),
    template: z.string().optional(),
    instruction: JsonObject.optional(),
    tools: z.array(z.string()).optional(),
    schedule: JsonObject.optional(),
    status: z.enum(["ACTIVE", "INACTIVE", "PAUSED"]),
});

export const AgentUpdateSchema = z.object({
    name: z.string().min(1).optional(),
    desc: z.string().optional(),
    prompt: JsonObject.optional(),
    template: z.string().optional(),
    instruction: JsonObject.optional(),
    tools: z.array(z.string()).optional(),
    schedule: JsonObject.nullable().optional(),
    status: z.enum(["ACTIVE", "INACTIVE", "PAUSED"]).optional(),
}).refine((data) => Object.keys(data).length > 0, {
    message: "No fields to update",
});

export const PaginationSchema = z.object({
    page: z.coerce.number().min(1).default(1),
    limit: z.coerce.number().min(1).max(100).default(10),
});


export default AgentSchema;
