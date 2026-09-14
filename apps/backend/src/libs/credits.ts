// Credit System & Cost Calculation
// Raw Costs Per Agent Run (Morning Briefing)
// Step	Service	Raw Cost
// Scrape 5 URLs	Firecrawl (1 credit/page)	~$0.004 (on Hobby plan: $16/mo for 3000 credits)
// LLM Call 1: Analyze scraped text (~4000 input tokens, ~500 output tokens)	GPT-4o-mini	~$0.0009
// LLM Call 2: Generate final briefing (~2000 input, ~800 output)	GPT-4o-mini	~$0.0008
// Send email	Resend (free tier: 3000/mo)	$0.00
// Total raw cost per run		~$0.006

import type { RunUsage } from "./types";

// HERE 1 credit = $0.01 internally.

export const CREDIT_USD_VALUE = 0.01;

export const PRICING = {
    FIRECRAWL_PER_PAGE: 0.0008,
    GPT_4O_MINI_INPUT_1K: 0.000015,
    GPT_4O_MINI_OUTPUT_1K: 0.0006,
    RESEND_PER_EMAIL: 0.0,
}

export const calculateRunCost = (usage: RunUsage) => {

    const fircrawlCost = usage.pageScraped * PRICING.FIRECRAWL_PER_PAGE

    const inputCost = (usage.promptTokens / 1000) * PRICING.GPT_4O_MINI_INPUT_1K
    const outputCost = (usage.completionTokens / 1000) * PRICING.GPT_4O_MINI_OUTPUT_1K

    const llmCost = inputCost + outputCost

    const totalCost = fircrawlCost + llmCost + PRICING.RESEND_PER_EMAIL

    const credits = totalCost / CREDIT_USD_VALUE

    return {
        totalCost,
        credits
    }
}
