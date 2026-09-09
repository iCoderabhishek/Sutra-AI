import { Router } from "express";
import { getUser } from "./services";


const router = Router()



router.get("/me", getUser)




export default router