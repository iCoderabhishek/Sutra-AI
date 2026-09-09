import { Router } from "express";
import { getUser, getGoogleCallback } from "./services";


const router = Router()



router.get("/me", getUser)
router.post("/google", getGoogleCallback)



export default router