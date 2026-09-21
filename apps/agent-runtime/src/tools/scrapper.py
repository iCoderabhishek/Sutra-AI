from typing import Type, Any

from pydantic import BaseModel

from libs.firecrawl_client import firecrawl_client
from tools.base import BaseTool


class ScrapperArgs(BaseModel):
    """Arguments for the scrapper tool"""
    url: str
    """The url to scrape"""


class ScrapperTool(BaseTool):
    name: str = "scrapper"
    description: str = "Scrape a url and return its main content as markdown"
    args_schema: Type[BaseModel] = ScrapperArgs

    async def execute(self, **kwargs) -> Any:
        """
        Execute the scrapper tool
        """
        url = kwargs.get("url")
        if not url:
            return "Error: No url provided."

        print(f"[ScrapperTool] Scrapping url: {url}")

        try:
            document = await firecrawl_client.scrape(
                url,
                formats=["markdown"],
                only_main_content=True,
            )

            markdown_content = getattr(document, "markdown", None)

            if not markdown_content:
                return "No content could be extracted from this URL."

            return markdown_content

        except Exception as e:
            return f"Error executing scrapper: {str(e)}"
