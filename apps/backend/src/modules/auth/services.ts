import type { Request, Response, NextFunction } from "express";
import axios from "axios";
import { GOOGLE_CLIENT_ID, GOOGLE_CALLBACK_URL, GOOGLE_CLIENT_SECRET } from "../../libs/env";
import { prisma } from "@sutra/db";
import { readAuthSession, setAuthSession } from "../../libs/session";
import { ensureCredits } from "../../libs/credits";
import jwt from "jsonwebtoken";
import { SESSION_SECRET } from "../../libs/env";

export const getMe = async (req: Request, res: Response, next: NextFunction) => {
    const user = await prisma.user.findUnique({
        where: { id: req.user.userId }
    })

    if (!user) {
        res.status(401).json({ message: "Unauthorized" })
        return
    }

    res.json({
        id: user.id,
        email: user.email,
        name: user.name,
        avatar: user.avatar,
        githubInstallationId: user.googleId,
        accessToken: req.user.access_token,
        userId: req.user.userId,
        refreshToken: req.user.refresh_token,
    })
}

export const redirectToGoogle = (req: Request, res: Response) => {
    const { cli_redirect } = req.query;
    // Encode cli_redirect into the state parameter to safely pass it through Google
    const state = cli_redirect ? Buffer.from(cli_redirect as string).toString('base64') : '';
    
    const params = new URLSearchParams({
        client_id: GOOGLE_CLIENT_ID,
        redirect_uri: GOOGLE_CALLBACK_URL,
        response_type: 'code',
        scope: 'profile email',
        state: state
    });

    const url = `https://accounts.google.com/o/oauth2/v2/auth?${params.toString()}`;
    res.redirect(url);
}

export const getGoogleCallback = async (req: Request, res: Response, next: NextFunction) => {

    const { code, state } = req.query

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

        await ensureCredits(user.id);

        setAuthSession(req, access_token, user.id, data.refresh_token);

        // Decode cli_redirect from the state parameter
        let cliRedirect: string | null = null;
        if (state && typeof state === 'string') {
            try {
                cliRedirect = Buffer.from(state, 'base64').toString('utf-8');
            } catch (e) {
                console.warn("Failed to decode OAuth state");
            }
        }

        if (cliRedirect) {
            const token = jwt.sign({ userId: user.id }, SESSION_SECRET, { expiresIn: '7d' });
            const redirectUrl = new URL(cliRedirect);
            redirectUrl.searchParams.append("token", token);
            return res.redirect(redirectUrl.toString());
        }

        return res.json({
            message: "Successfully authenticated with Google",
            user: user
        });

    } catch (error) {
        console.log("Error fetching Google user:", error);
        return res.status(500).json({ error: "Authentication failed" });
    }
}