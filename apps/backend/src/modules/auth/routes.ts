import { Router } from "express";
import { getUser, getGoogleCallback, redirectToGoogle } from "./services";


const router = Router()



router.get("/me", getUser)
router.get("/google", redirectToGoogle)
router.get("/callback", getGoogleCallback)



export default router