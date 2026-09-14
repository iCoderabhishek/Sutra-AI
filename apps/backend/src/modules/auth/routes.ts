import { Router } from "express";
import { getMe, getGoogleCallback, redirectToGoogle } from "./services";
import { authMiddleware } from "./middleware";


const router = Router()



router.get("/me", authMiddleware, getMe)
router.get("/google", redirectToGoogle)
router.get("/callback", getGoogleCallback)



export default router