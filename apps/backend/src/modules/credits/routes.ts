import { Router } from "express";
import { authMiddleware } from "../auth/middleware";
import { getCredits } from "./services";


const router = Router();

router.get("/credits", authMiddleware, getCredits);

export default router;