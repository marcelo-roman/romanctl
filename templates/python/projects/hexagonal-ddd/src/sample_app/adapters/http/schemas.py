from pydantic import BaseModel, ConfigDict
from pydantic.alias_generators import to_camel


class CamelModel(BaseModel):
    model_config = ConfigDict(alias_generator=to_camel, populate_by_name=True)


class CreatedResponse(CamelModel):
    id: str


class Problem(CamelModel):
    status: int
    title: str
    detail: str = ""
