import type { Request, Response, NextFunction } from "express";

export const getUser = (req: Request, res: Response, next: NextFunction) => {


    return res.json({
        "user": "user"
    })

}