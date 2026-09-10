import type { Request } from "express"
export { };
declare global {
    namespace Express {
        interface Request {
            user?: any,
            userId?: any,
        }
    }
}
