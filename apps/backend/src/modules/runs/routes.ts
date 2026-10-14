import { Router } from "express";
import { authMiddleware } from "../auth/middleware";
import { getRun, listRuns, streamRun } from "./services";

const router = Router();

router.get("/", authMiddleware, listRuns);
router.get("/:runId", authMiddleware, getRun);
router.get("/:runId/stream", authMiddleware, streamRun);

export default router;
