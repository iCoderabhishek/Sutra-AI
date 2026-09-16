from abc import ABC, abstractmethod
from typing import Type, Any
from pydantic import BaseModel

class BaseTool(ABC):
    name: str
    desciption: str

    args_schema: Type[BaseModel]

    @abstractmethod
    async def execute(self, **kwargs) -> Any:
        """
        The actual logic of the tool goes here.
        Must be implemented by subclasses.
        """
        pass

    def get_schema(self) -> dict:
        """
        Automatically generates the strict JSON schema that OpenAI or other LLM expects.
        """
        return {
            "type": "function",
            "function": {
                "name": self.name,
                "description": self.description,
                "parameters": self.args_schema.model_json_schema()
            }
        }