import type { Request, Response, NextFunction } from "express";
import axios from "axios";
import { GOOGLE_CLIENT_ID, GOOGLE_CALLBACK_URL, GOOGLE_CLIENT_SECRET } from "../../lib/env";
import { prisma } from "@sutra/db";

export const getUser = (req: Request, res: Response, next: NextFunction) => {


    return res.json({
        user: req.user
    })

}

export const redirectToGoogle = (req: Request, res: Response) => {
    const url = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${GOOGLE_CLIENT_ID}&redirect_uri=${GOOGLE_CALLBACK_URL}&response_type=code&scope=profile email`;
    res.redirect(url);
}

export const getGoogleCallback = async (req: Request, res: Response, next: NextFunction) => {

    const { code } = req.query

    if (!code) {
        return res.status(400).json({ error: "Authorization code missing" });
    }

    try {
        const { data } = await axios.post('https://oauth2.googleapis.com/token', {
            client_id: GOOGLE_CLIENT_ID,
            client_secret: GOOGLE_CLIENT_SECRET,
            code,
            redirect_uri: GOOGLE_CALLBACK_URL,
            grant_type: 'authorization_code',
        });

        const { access_token } = data

        const { data: googleUser } = await axios.get("https://www.googleapis.com/oauth2/v2/userinfo", {
            headers: {
                Authorization: `Bearer ${access_token}`
            }
        })


        // Upsert user to the database
        const user = await prisma.user.upsert({
            where: { googleId: googleUser.id },
            update: {
                name: googleUser.name,
                email: googleUser.email,
            },
            create: {
                googleId: googleUser.id,
                name: googleUser.name,
                email: googleUser.email,
            }
        });

        return res.json({
            message: "Successfully authenticated with Google",
            user: user
        });

    } catch (error) {
        console.log("Error fetching Google user:", error);
        return res.status(500).json({ error: "Authentication failed" });
    }
}