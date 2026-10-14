
from libs.tavily_client import tavily_client
from tools.base import BaseTool
from typing import Type, Any
from pydantic import BaseModel

class WebSearchArgs(BaseModel):
    """Arguments for the web_search tool"""
    query: str
    """The search query"""
    


class WebSearchTool(BaseTool):
    name: str = "web_search"
    description: str = "Search the web for information"
    args_schema: Type[BaseModel] = WebSearchArgs
    async def execute(self, **kwargs) -> Any:
        query = kwargs.get("query")
        if not query:
            return "Error: No query provided."
            
        print(f"[WebSearchTool] Searching for: {query}")
        
        try:
            response = await tavily_client.search(query=query, search_depth="basic")
            results = response.get("results", [])
            if not results:
                return "No results found."
                
            formatted_results = "\n\n".join([f"Source: {res['url']}\nContent: {res['content']}" for res in results])
            return formatted_results
            
        except Exception as e:
            return f"Error executing search: {str(e)}"