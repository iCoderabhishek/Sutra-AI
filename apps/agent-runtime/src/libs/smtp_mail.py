import smtplib
import asyncio
from email.message import EmailMessage
from libs.env import settings

async def send_email(to_email: str, subject: str, body: str) -> bool:
    """
    Sends an email using standard SMTP.
    Runs in a background thread to prevent blocking the FastAPI event loop.
    """
    def _send_sync():
        if not settings.SMTP_HOST:
            raise ValueError("SMTP_HOST is not configured in .env")

        msg = EmailMessage()
        msg.set_content(body)
        msg['Subject'] = subject
        msg['From'] = settings.SMTP_USER
        msg['To'] = to_email

        # connect to SMTP server (using STARTTLS for port 587)
        # using gmail smtp
        with smtplib.SMTP(settings.SMTP_HOST, settings.SMTP_PORT) as server:
            server.starttls()
            server.login(settings.SMTP_USER, settings.SMTP_PASSWORD)
            server.send_message(msg)

    try:
        # Offload the blocking SMTP network call to a separate thread
        await asyncio.to_thread(_send_sync)
        print(f"[SMTP] Successfully sent email to {to_email}")
        return True
    except Exception as e:
        print(f"[SMTP] Error sending email: {e}")
        return False
