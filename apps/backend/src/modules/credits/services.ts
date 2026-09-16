import type { Request, Response, NextFunction } from "express";
import { hasEnoughCredits, calculateRunCost } from "../../libs/credits";
import { prisma } from "@sutra/db";

export const getCredits = async (req: Request, res: Response, next: NextFunction) => {

    try {
        const creditsUser = await prisma.credits.findFirst({
            where: { userId: req.user.userId }
        })

        // Convert Prisma Decimal to number, or default to 0
        const availableCredits = creditsUser?.balance?.toNumber() ?? 0;

        res.json({
            availableCredits
        })


    } catch (error) {
        console.error("(CREDITs) Failed to grab credits", error);
        res.status(500).json({ error: "Failed to grab credits" });
    }

}