import { Router } from "express";
import { authMiddleware } from "../auth/middleware";
import { getModels } from "./services";

const router = Router();

router.get("/", authMiddleware, getModels);

export default router;
