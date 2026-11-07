using Microsoft.AspNetCore.Mvc;

namespace {{.Name}}.Controllers;

[ApiController]
[Route("{{.EntityPluralKebab}}")]
public class {{.Entity}}Controller : ControllerBase
{
    [HttpGet]
    public IActionResult List() => Ok(Array.Empty<object>());

    [HttpGet("{id}")]
    public IActionResult Get(string id) => Ok(new { id });
}
