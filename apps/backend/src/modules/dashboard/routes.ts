import { Router } from "express";
import { authMiddleware } from "../auth/middleware";
import { getDashboard } from "./services";

const router = Router();

router.get("/", authMiddleware, getDashboard);

export default router;
