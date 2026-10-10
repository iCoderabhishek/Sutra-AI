import type { Request, Response, NextFunction } from "express";

let cache: any[] | null = null;
let lastFetch = 0;

export const getModels = async (req: Request, res: Response, next: NextFunction) => {
    try {
        const now = Date.now();
        if (cache && now - lastFetch < 3600 * 1000) {
            return res.json(cache);
        }

        const response = await fetch("https://openrouter.ai/api/v1/models");
        const data = await response.json() as any;
        
        const modelsList = data.data
            .filter((m: any) => m.pricing?.prompt === "0" && m.pricing?.completion === "0")
            .map((m: any) => ({
                id: m.id,
                name: m.name,
                context_length: m.context_length
            }));
            
        // Include auto as first option
        const result = [
            { id: "auto", name: "Auto (Best Available)", context_length: 0 },
            ...modelsList
        ];
        
        cache = result;
        lastFetch = now;
        
        res.json(result);
    } catch (error) {
        if (cache) return res.json(cache);
        next(error);
    }
};
