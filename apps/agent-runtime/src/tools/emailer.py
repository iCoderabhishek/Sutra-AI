from typing import Type, Any
from pydantic import BaseModel
from tools.base import BaseTool
from libs.smtp_mail import send_email

class EmailArgs(BaseModel):
    """Arguments for the emailer tool"""
    to_email: str
    """The recipient's email address"""
    subject: str
    """The subject line of the email"""
    body: str
    """The main text content of the email"""

class EmailTool(BaseTool):
    name: str = "send_email"
    description: str = "Send an email to a user with the specified subject and body."
    args_schema: Type[BaseModel] = EmailArgs

    async def execute(self, **kwargs) -> Any:
        to_email = kwargs.get("to_email")
        subject = kwargs.get("subject")
        body = kwargs.get("body")
        
        if not to_email or not subject or not body:
            return "Error: Missing required email fields (to_email, subject, body)."
            
        print(f"[EmailTool] Sending email to: {to_email}")
        
        try:
            # We use the SMTP mailer we just created!
            success = await send_email(to_email=to_email, subject=subject, body=body)
            
            if success:
                return f"Successfully sent email to {to_email}."
            else:
                return f"Failed to send email to {to_email}."
                
        except Exception as e:
            return f"Error executing emailer: {str(e)}"
