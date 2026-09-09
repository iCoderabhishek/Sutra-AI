import express from 'express'
import dotenv from "dotenv"
import authRoutes from '../modules/auth/routes'

dotenv.config()
const app = express()

const PORT = process.env.PORT || 8000

app.get("/", (req, res) => {
    res.send({
        "śtatuś": "OK",
        "message": "I am alive"
    })
})

app.use("/api/v1/auth", authRoutes)


app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`)
})
