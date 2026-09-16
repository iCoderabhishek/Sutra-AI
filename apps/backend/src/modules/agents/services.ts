import { prisma } from "@sutra/db";
import type { Request, Response, NextFunction } from "express";
import AgentSchema, { PaginationSchema } from "./schema";
import { scheduleAgent } from "./scheduler"
import { runAgent } from "./runner";
import { triggerAgentRun } from "../../libs/agent-proxy";
import { deductCredits, hasEnoughCredits } from "../../libs/credits";

export const createAgent = async (req: Request, res: Response, next: NextFunction) => {

    const result = AgentSchema.safeParse(req.body);

    if (!result.success) {
        return res.status(400).json({ message: "Invalid request body", error: result.error });
    }

    const { name, desc, prompt, template, instruction, tools, schedule, status } = result.data;

    if (!name || !prompt || !tools || !schedule) {
        return res.status(400).json({ message: "All fields are required" });
    }

    if (status) {
        if (!['ACTIVE', 'INACTIVE', 'PAUSED'].includes(status)) {
            return res.status(400).json({ message: "Invalid status" });
        }
    }

    try {
        const agent = await prisma.agent.create({
            data: {
                name,
                desc,
                prompt,
                template,
                instruction,
                tools,
                schedule,
                status,
                userId: req.user.userId,
            },
        });

        if (agent.status === 'ACTIVE' && agent.schedule) {
            await scheduleAgent(agent);
        }

        return res.status(201).json(agent);
    } catch (error) {
        return res.status(500).json({ message: "Failed to create agent", error });
    }

};


export const listAgents = async (req: Request, res: Response, next: NextFunction) => {

    const result = PaginationSchema.safeParse(req.query);

    if (!result.success) {
        return res.status(400).json({ message: "Invalid pagination parameters", error: result.error });
    }

    const { page, limit } = result.data;

    try {
        const [agents, total] = await Promise.all([
            prisma.agent.findMany({
                where: { userId: req.user.userId },
                skip: (page - 1) * limit,
                take: limit,
                orderBy: { createdAt: 'desc' }
            }),
            prisma.agent.count({
                where: { userId: req.user.userId }
            })
        ]);

        return res.status(200).json({
            data: agents,
            meta: {
                total,
                page,
                limit,
                totalPages: Math.ceil(total / limit)
            }
        });

    } catch (error) {
        console.error("Failed to list agents:", error);
        return res.status(500).json({ message: "Failed to list agents" });
    }

};


export const getAgent = async (req: Request, res: Response, next: NextFunction) => {

    const { agentId } = req.params as { agentId: string };

    try {
        const agent = await prisma.agent.findFirst({
            where: { id: agentId, userId: req.user.userId },
        });

        if (!agent) {
            return res.status(404).json({ message: "Agent not found" });
        }

        return res.status(200).json(agent);

    } catch (error) {
        console.error(`Failed to get agent with ${agentId}:`, error);
        return res.status(500).json({ message: "Failed to get agent" });
    }

};


export const triggerAgent = async (req: Request, res: Response, next: NextFunction) => {

    try {
        const { agentId } = req.params as { agentId: string };
        const agent = await prisma.agent.findFirst({
            where: { id: agentId, userId: req.user.userId },
        });

        if (!agent) {
            return res.status(404).json({ message: "Agent not found" });
        }

        if (agent.status !== 'ACTIVE') {
            return res.status(400).json({ message: "Agent is not active" });
        }

        // 
        // model JobRun {
        //   id          String     @id @default(uuid())
        //   agent       Agent      @relation(fields: [agentId], references: [id])
        //   agentId     String
        //   status      StatusEnum
        //   trace       Json
        //   totalCost   Decimal
        //   totalTokens Int
        //   startedAt   DateTime
        //   finishedAt  DateTime
        //   createdAt   DateTime   @default(now())
        //   updatedAt   DateTime   @updatedAt
        // }

        const checkCredits = await hasEnoughCredits(req.user.userId)
        if (!checkCredits.hasEnoughCredits) {
            return res.status(402).json({ message: "Insufficient credits to run this agent" });
        }

        const run = await prisma.jobRun.create({
            data: {
                agentId: agent.id,
                status: 'ACTIVE',
                trace: req.body.trace || {},
                totalCost: req.body.totalCost || 0,
                totalTokens: req.body.totalTokens || 0,
                startedAt: req.body.startedAt || new Date(),
                finishedAt: req.body.finishedAt || new Date(),

            },
        });

        const result = await triggerAgentRun({
            goal: agent.prompt as string,
            tools: agent.tools as string[]
        })

        if (result.error) {
            return res.status(500).json({ message: "Failed to trigger agent" });
        }

        const deductedCredits = await deductCredits(req.user.userId, 1)
        console.log("Credits deducted successfully", deductedCredits);


        console.log("{Agent} Agent triggered successfully", result);


        await runAgent(agent, run);

        return res.status(202).json({ message: "Agent triggered successfully", runId: run.id });


    } catch (error) {
        console.error(`Failed to trigger agent ${req.params?.agentId}:`, error);
        return res.status(500).json({ message: "Failed to trigger agent", error });
    }
}