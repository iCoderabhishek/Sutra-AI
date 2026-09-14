// Credit System & Cost Calculation
// Raw Costs Per Agent Run (Morning Briefing)
// Step	Service	Raw Cost
// Scrape 5 URLs	Firecrawl (1 credit/page)	~$0.004 (on Hobby plan: $16/mo for 3000 credits)
// LLM Call 1: Analyze scraped text (~4000 input tokens, ~500 output tokens)	GPT-4o-mini	~$0.0009
// LLM Call 2: Generate final briefing (~2000 input, ~800 output)	GPT-4o-mini	~$0.0008
// Send email	Resend (free tier: 3000/mo)	$0.00
// Total raw cost per run		~$0.006

// HERE 1 credit = $0.01 internally.