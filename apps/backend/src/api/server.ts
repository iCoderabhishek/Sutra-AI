import express from 'express'
import dotenv from "dotenv"
import authRoutes from '../modules/auth/routes'
import cookieSession from "cookie-session"
import { SESSION_SECRET } from '../libs/env'
import "../libs/queue";
import agentRoutes from '../modules/agents/routes'



dotenv.config()
const app = express()

const PORT = process.env.PORT || 8000

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


app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`)
})
