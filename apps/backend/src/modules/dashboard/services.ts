import { prisma } from "@sutra/db";
import type { Request, Response, NextFunction } from "express";

export const getDashboard = async (req: Request, res: Response, next: NextFunction) => {
    const userId = req.user.userId;

    try {
        const [agentsByStatus, runsByStatus, totals, recentRuns, credits] = await Promise.all([
            prisma.agent.groupBy({
                by: ["status"],
                where: { userId },
                _count: { _all: true },
            }),
            prisma.jobRun.groupBy({
                by: ["status"],
                where: { agent: { userId } },
                _count: { _all: true },
            }),
            prisma.jobRun.aggregate({
                where: { agent: { userId } },
                _sum: { totalCost: true, totalTokens: true },
                _count: { _all: true },
            }),
            prisma.jobRun.findMany({
                where: { agent: { userId } },
                orderBy: { createdAt: "desc" },
                take: 10,
                select: {
                    id: true,
                    status: true,
                    totalCost: true,
                    totalTokens: true,
                    startedAt: true,
                    finishedAt: true,
                    createdAt: true,
                    agent: { select: { id: true, name: true, template: true } },
                },
            }),
            prisma.credits.findFirst({ where: { userId } }),
        ]);

        const countOf = (rows: { status: string; _count: { _all: number } }[]) =>
            rows.reduce<Record<string, number>>((acc, row) => {
                acc[row.status] = row._count._all;
                return acc;
            }, {});

        return res.status(200).json({
            agents: {
                total: agentsByStatus.reduce((sum, row) => sum + row._count._all, 0),
                byStatus: countOf(agentsByStatus as never),
            },
            runs: {
                total: totals._count._all,
                byStatus: countOf(runsByStatus as never),
            },
            usage: {
                totalCostUsd: Number(totals._sum.totalCost ?? 0),
                totalTokens: totals._sum.totalTokens ?? 0,
            },
            credits: {
                balance: credits ? Number(credits.balance) : 0,
                plan: credits?.plan ?? null,
            },
            recentRuns: recentRuns.map((run) => ({ ...run, totalCost: Number(run.totalCost) })),
        });
    } catch (error) {
        return next(error);
    }
};
