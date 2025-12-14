---
trigger: always_on
---

Technologies:
- About Go web frameworks, always only use `net/http` + `chi`.
- Don't use ORM (Gorm).
- The GitHub repository for this backend service is "github.com/xeodocs/xeodocs-backend".
- For SQL migrations, the Goose tool is used but it is not included in the code; SQL migrations are managed by executing its command.

Implementation:
- If an endpoint returns a list of elements, in the case that no elements exist yet, always return an empty array `[]`, never return `null`.
- Whenever you need to configure the request to an API endpoint, refer to the OpenAPI description on the GitHub MCP Server.

Code:
- Always write the code, comments and documentation in English, even when the prompt is not in English.

Architecture:
- This backend uses a Modular Monolith architecture strategy in Go.