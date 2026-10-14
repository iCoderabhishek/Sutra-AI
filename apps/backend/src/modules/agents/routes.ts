import { Router } from "express";
import { createAgent, deleteAgent, getAgent, listAgents, triggerAgent, updateAgent } from "./services";
import { authMiddleware } from "../auth/middleware";


const router = Router();

router.post("/", authMiddleware, createAgent);

router.get("/", authMiddleware, listAgents);

router.get("/:agentId", authMiddleware, getAgent);

router.patch("/:agentId", authMiddleware, updateAgent);

router.delete("/:agentId", authMiddleware, deleteAgent);

router.post("/:agentId/run", authMiddleware, triggerAgent)

export default router