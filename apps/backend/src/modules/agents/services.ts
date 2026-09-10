import { prisma } from "@sutra/db";
import type { Request, Response, NextFunction } from "express";
import AgentSchema, { PaginationSchema } from "./schema";
import { scheduleAgent } from "./scheduler"

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
                userId: req.user.id,
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
                where: { userId: req.user.id },
                skip: (page - 1) * limit,
                take: limit,
                orderBy: { createdAt: 'desc' }
            }),
            prisma.agent.count({
                where: { userId: req.user.id }
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
            where: { id: agentId, userId: req.user.id },
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