import { REDIS_URL } from "./env";
import { Redis } from "ioredis"

const redis = new Redis(REDIS_URL, {
    maxRetriesPerRequest: 0
})

export default redis