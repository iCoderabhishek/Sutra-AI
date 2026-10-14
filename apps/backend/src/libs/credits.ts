// Credit System & Cost Calculation
// Raw Costs Per Agent Run (Morning Briefing)
// Step	Service	Raw Cost
// Scrape 5 URLs	Firecrawl (1 credit/page)	~$0.004 (on Hobby plan: $16/mo for 3000 credits)
// LLM Call 1: Analyze scraped text (~4000 input tokens, ~500 output tokens)	GPT-4o-mini	~$0.0009
// LLM Call 2: Generate final briefing (~2000 input, ~800 output)	GPT-4o-mini	~$0.0008
// Send email	Resend (free tier: 3000/mo)	$0.00
// Total raw cost per run		~$0.006

import { prisma } from "@sutra/db";
import type { RunUsage } from "./types";

// HERE 1 credit = $0.01 internally.

export const CREDIT_USD_VALUE = 0.01;

// LLM cost comes from the runtime's reported cost_usd, not from here.
export const PRICING = {
    FIRECRAWL_PER_PAGE: 0.0008,
    SMTP_PER_EMAIL: 0.0,
}

export const calculateSideCarCost = (usage: RunUsage) =>
    usage.pagesScraped * PRICING.FIRECRAWL_PER_PAGE
    + usage.emailsSent * PRICING.SMTP_PER_EMAIL


export const hasEnoughCredits = async (userId: string, minCreditsRequired: number = 1) => {

    const userCredits = await prisma.credits.findUnique({
        where: { userId }
    })

    if (!userCredits || Number(userCredits.balance) < minCreditsRequired) {
        return {
            hasEnoughCredits: false,
            error: "Insufficient credits"
        }
    }

    return {
        hasEnoughCredits: true
    }
}

export const usdToCredits = (usd: number) => usd / CREDIT_USD_VALUE;

export const DEFAULT_FREE_CREDITS = 20;
export const DEFAULT_PLAN = "FREE";

// Gives a new user their starting wallet
// THEY Can use this to get started, later they can buy more credits they want
export const ensureCredits = async (userId: string) =>
    prisma.credits.upsert({
        where: { userId },
        update: {},
        create: { userId, balance: DEFAULT_FREE_CREDITS, plan: DEFAULT_PLAN },
    });

export const deductCredits = async (userId: string, creditsToDeduct: number) => {
    try {
        // updateMany so the balance check and the decrement are one statement.
        const updated = await prisma.credits.updateMany({
            where: { userId, balance: { gte: creditsToDeduct } },
            data: {
                balance: {
                    decrement: creditsToDeduct
                }
            }
        });
        return updated.count > 0;
    } catch (error) {
        console.error("Failed to deduct credits, cause: ", error);
        throw error
    }
}