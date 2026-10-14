export { };

declare global {
    namespace Express {
        interface AuthUser {
            userId: string;
            access_token: string;
            refresh_token?: string;
        }
        interface Request {
            user: AuthUser;
        }
    }
}
