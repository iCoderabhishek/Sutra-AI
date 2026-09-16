import { Router } from "express";
import { createAgent, getAgent, listAgents, triggerAgent } from "./services";
import { authMiddleware } from "../auth/middleware";


const router = Router();

router.post("/", authMiddleware, createAgent);

router.get("/", authMiddleware, listAgents);

router.get("/:agentId", authMiddleware, getAgent);

router.post("/:agentId/run", authMiddleware, triggerAgent)

export default router