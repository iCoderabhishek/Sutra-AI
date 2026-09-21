import type { Request, Response, NextFunction } from "express";
import { prisma } from "@sutra/db";

export const getCredits = async (req: Request, res: Response, next: NextFunction) => {
    try {
        const userCredits = await prisma.credits.findFirst({
            where: { userId: req.user.userId }
        })

        return res.json({
            availableCredits: userCredits ? Number(userCredits.balance) : 0,
            plan: userCredits?.plan ?? null,
        })
    } catch (error) {
        console.error("Failed to fetch credits:", error);
        return res.status(500).json({ message: "Failed to fetch credits" });
    }
}