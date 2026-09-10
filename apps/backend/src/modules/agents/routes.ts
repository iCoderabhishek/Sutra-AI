import { Router } from "express";
import { createAgent, getAgent, listAgents } from "./services";
import { authMiddleware } from "../auth/middleware";


const router = Router();

router.post("/", authMiddleware, createAgent);

router.get("/", authMiddleware, listAgents);

router.get("/:agentId", authMiddleware, getAgent);
