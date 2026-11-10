using Microsoft.AspNetCore.Mvc;

namespace SampleApp.Controllers;

[ApiController]
[Route("sample-items")]
public class SampleItemController : ControllerBase
{
    [HttpGet]
    public IActionResult List() => Ok(Array.Empty<object>());

    [HttpGet("{id}")]
    public IActionResult Get(string id) => Ok(new { id });
}
