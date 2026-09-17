from libs import firecrawl_client
from typing import Type, Any
from tools.base import BaseTool
from pydantic import BaseModel

class ScrapperArgs(BaseModel):
    """Arguments for the scrapper tool"""
    url: str
    """The url to scrapper"""
    

class ScrapperTool(BaseTool):
    name: str = "scrapper"
    description: str = "Scrapper a url for information"
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
            response = await firecrawl_client.scrape(url, params={'formats': ['markdown']})
            
            # The response is a dictionary. We want to extract just the markdown content 
            # so the LLM doesn't get confused by all the extra metadata.
            markdown_content = response.get('markdown', '')
            
            if not markdown_content:
                return "No content could be extracted from this URL."
                
            return markdown_content
            
        except Exception as e:
            return f"Error executing scrapper: {str(e)}"