---
trigger: always_on
---

- Don't install Go web frameworks, always use the native Go http package
- Don't use ORM (Gorm)
- Do not compile directly; the code is tested using Docker Compose
- Always log important activity using the Logging Service