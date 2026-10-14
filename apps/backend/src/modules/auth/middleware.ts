import type { NextFunction, Request, Response } from "express";
import { readAuthSession } from "../../libs/session";
import jwt from "jsonwebtoken";
import { SESSION_SECRET } from "../../libs/env";

export const authMiddleware = async (req: Request, res: Response, next: NextFunction) => {
    // 1. Check for Bearer token first (CLI flow)
    const authHeader = req.headers.authorization;
    if (authHeader && authHeader.startsWith('Bearer ')) {
        const token = authHeader.split(' ')[1];
        if (token) {
            try {
                const decoded = jwt.verify(token, SESSION_SECRET) as unknown as { userId: string };
                req.user = {
                    access_token: token,
                    userId: decoded.userId
                };
                return next();
            } catch (error) {
                return res.status(401).json({ message: "Unauthorized: Invalid or expired token" });
            }
        }
    }

    // 2. Fall back to cookie session (Web flow)
    const session = readAuthSession(req);
    if (!session) {
        return res.status(401).json({ message: "Unauthorized: No valid session found" });
    }

    req.user = {
        access_token: session.access_token,
        refresh_token: session.refresh_token,
        userId: session.userId
    };

    next();
}