---
trigger: always_on
---

Technologies:
- About Go web frameworks, always only use `net/http` + `chi`
- Don't use ORM (Gorm)
- The GitHub repository for this backend service is "github.com/xeodocs/xeodocs-backend"

Implementation:
- If an endpoint returns a list of elements, in the case that no elements exist yet, always return an empty array `[]`, never return `null`

Code:
- Always write the code in English, as well as the comments, even when the prompt is not in English.

Architecture:
- This backend uses a Modular Monolith architecture strategy in Go.