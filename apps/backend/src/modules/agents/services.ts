import { prisma } from "@sutra/db";
import type { Request, Response, NextFunction } from "express";
import AgentSchema from "./schema";
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

export const listAgents = (req: Request, res: Response, next: NextFunction) => { };

export const getAgent = (req: Request, res: Response, next: NextFunction) => { };