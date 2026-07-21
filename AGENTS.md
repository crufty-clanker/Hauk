# AGENTS.md

## Hauk — Self-hosted location sharing service

**Stack:** PHP 8.2+/8.3 backend, vanilla JavaScript frontend (no framework), Android app (Kotlin/Gradle).

**Directory layout:**
- `backend-php/` — PHP backend. API endpoints, config (`include/`), session store (Memcached/Redis).
- `backend-go/` — Go backend.
- `frontend/` — Single-page frontend: `index.html`, `main.js`, `style.css`, `lib/`, `assets/`.
- `android/` — Android companion app: Gradle build, Kotlin source under `app/`.
- `docker/` — Dockerfile templates for local podman testing only (not used in CI).
- `.github/workflows/` — CI pipelines for PHP, frontend, Android, and Go.

**Key files:**
- `composer.json` — PHP dependencies (PHPStan, PHP-CS-Fixer, PHPUnit, etc.).
- `package.json` — Frontend tooling (ESLint, Prettier, html-validate).
- `phpstan.neon` — PHP static analysis config.
- `phpunit.xml` — PHP test runner config.
- `install.sh` — One-command installer for self-hosted deployments.
- `.github/dependabot.yml` — Automated dependency updates. Configured with `delete-dependent-on-merge: true` and grouped by ecosystem (`github-actions`, `php-backend`, `frontend`, `android-app`, `docker`) so each ecosystem gets a single PR.

**CI/CD:**
- **PHP CI** — Lint (PHP-CS-Fixer) + static analysis (PHPStan) + PHPUnit on PHP 8.2 & 8.3.
- **Frontend CI** — ESLint + Prettier formatting check + html-validate.
- **Android CI** — Build debug APK + unit tests with JDK 17 / Gradle.
- **Docker** — Removed from CI. Use podman locally with the Dockerfiles in `docker/`.

**Pre-commit checks (run locally before committing):**
- Modified **PHP** files → `composer run lint` + `composer run analyse` + `composer run test`
- Modified **frontend** files → `npm run lint` (ESLint + Prettier + html-validate)
- Modified **Android** files → `cd android && ./gradlew testDebugUnitTest --no-daemon`
- Modified **Go** files → `cd backend-go && go vet ./... && go test -race ./...`

**Run GitHub Actions locally (before pushing — required):**

Use [`act`](https://github.com/nektos/act) to run the relevant workflows locally before pushing. This catches CI failures early and avoids wasting CI resources on broken commits.

- **Prerequisites:** Docker (or podman with `--compat` flag) must be running.
- **Installation:** `brew install act` (macOS) or follow [act's install instructions](https://github.com/nektos/act?tab=readme-ov-file#installation) for other platforms.
- A `.actrc` file is included in the repo root — it configures defaults like the container OS and working directory.

Run **all** workflows:
```bash
act push
```

Run only the workflow relevant to your changes:
```bash
act push -W .github/workflows/php-ci.yml       # PHP changes
act push -W .github/workflows/frontend-ci.yml   # Frontend changes
act push -W .github/workflows/android-ci.yml    # Android changes
act push -W .github/workflows/go-ci.yml         # Go changes
```

**Tip:** `act` defaults to the `push` event, so you usually only need `act push`. Use `-l` to list local environment variables or `-w` to target a single workflow file.


**Important:**
- No official Docker image is built in CI — use `docker/` Dockerfiles with podman for local testing only.
- `git config pull.rebase true` is set locally — branches rebase on pull.
- Always run `act push` (or the relevant workflow) before pushing — broken CI wastes reviewer time and blocks merges.

**Requirements:** PHP web server + Memcached or Redis (optional: LDAP). Android 6+ for the app.

**Install:** `sudo ./install.sh -c <web_root>` or copy `backend-php/` + `frontend/` to your web root and edit `backend-php/include/config.php`. Docker users mount `./config/hauk:/etc/hauk`.
