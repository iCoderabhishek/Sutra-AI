import "dotenv/config";

import express from 'express'
import type { Request, Response, NextFunction } from 'express'
import authRoutes from '../modules/auth/routes'
import cookieSession from "cookie-session"
import { SESSION_SECRET } from '../libs/env'
import "../libs/queue";
import agentRoutes from '../modules/agents/routes'
import creditRoutes from '../modules/credits/routes'
import runRoutes from '../modules/runs/routes'
import dashboardRoutes from '../modules/dashboard/routes'
import { reconcileSchedules } from '../modules/agents/scheduler'

const app = express()

app.use(express.json())
const PORT = process.env.PORT || 4000

app.get("/", (req, res) => {
    res.send({
        "śtatuś": "OK",
        "message": "I am alive"
    })
})

app.use(cookieSession({
    name: 'session',
    keys: [SESSION_SECRET],
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    maxAge: 24 * 60 * 60 * 1000 // 24 hours
}))
app.use("/api/v1/auth", authRoutes)
app.use("/api/v1/agents", agentRoutes)
app.use("/api/v1/credits", creditRoutes)
app.use("/api/v1/runs", runRoutes)
app.use("/api/v1/dashboard", dashboardRoutes)

app.use((req, res) => {
    res.status(404).json({ message: "Not found" })
})

app.use((err: unknown, req: Request, res: Response, _next: NextFunction) => {
    console.error(`{error} ${req.method} ${req.originalUrl}:`, err)

    if (res.headersSent) return

    res.status(500).json({ message: "Internal server error" })
})

app.listen(PORT, async () => {
    console.log(`Server is running on port ${PORT}`)
    await reconcileSchedules()
})
