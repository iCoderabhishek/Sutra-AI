import { Router } from "express";
import { getUser, getGoogleCallback, redirectToGoogle } from "./services";
import { authMiddleware } from "./middleware";


const router = Router()



router.get("/me", authMiddleware, getUser)
router.get("/google", redirectToGoogle)
router.get("/callback", getGoogleCallback)



export default router