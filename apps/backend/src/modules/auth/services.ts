import type { Request, Response, NextFunction } from "express";
import axios from "axios";
import { GOOGLE_CLIENT_ID, GOOGLE_CALLBACK_URL, GOOGLE_CLIENT_SECRET } from "../../lib/env";

export const getUser = (req: Request, res: Response, next: NextFunction) => {


    return res.json({
        "user": "user"
    })

}


export const getGoogleCallback = async (req: Request, res: Response, next: NextFunction) => {

    const { code } = req.query
    try {
        const { data } = await axios.post('<https://oauth2.googleapis.com/token>', {
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


        // todo write to save the user to db
        // if user exists update the user
        // if user does not exists create the user
        // save user accesstoken to db
        // todo check whether the user is active or not


    } catch (error) {
        console.log("Error fetching Google user:", error);
        res.redirect("/login")

    }
}