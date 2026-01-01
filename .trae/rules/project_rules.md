English project; Persian user chat allowed. Repo is English-only: code/comments/docs/commits/logs/UI/DB schemas; no non-English chars. Persian must never be hardcoded or stored in the repo.

No hardcoding: no magic numbers/strings. Use constants (logic), tokens.json (UI/theme), env vars (secrets/credentials/endpoints).

SOLID: especially Single Responsibility.

Frontend: semantic HTML; no inline styles/logic; token-driven external CSS; keep business logic out of UI; split presentational vs container components.

Backend/API: correct HTTP methods; envelope {data, meta, error}; handle errors (no crashes), proper status codes; log full stacks internally; return safe generic English errors; validate input schemas; ORM/parameterized queries only.

Data/Config/Logs: migrations only; consistent English naming (snake_case or camelCase); config via env; structured JSON logs with DEBUG/INFO/WARN/ERROR; no print()/raw console.log().