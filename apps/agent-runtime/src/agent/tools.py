from tools.scrapper import ScrapperTool
from tools.search import SearchTool
from tools.emailer import EmailTool

tools = [
    ScrapperTool(), 
    SearchTool(), 
    EmailTool()
]

# name  instance lookup dict

tool_map = {t.name: t for t in tools}
