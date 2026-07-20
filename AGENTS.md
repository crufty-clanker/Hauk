# AGENTS.md

## Hauk — Self-hosted location sharing service

**Stack:** PHP 8.2+/8.3 backend, vanilla JavaScript frontend (no framework), Android app (Kotlin/Gradle).

**Directory layout:**
- `backend-php/` — PHP backend. API endpoints, config (`include/`), session store (Memcached/Redis).
- `frontend/` — Single-page frontend: `index.html`, `main.js`, `style.css`, `lib/`, `assets/`.
- `android/` — Android companion app: Gradle build, Kotlin source under `app/`.
- `docker/` — Dockerfile templates for local podman testing only (not used in CI).
- `.github/workflows/` — CI pipelines for PHP, frontend, and Android.

**Key files:**
- `composer.json` — PHP dependencies (PHPStan, PHP-CS-Fixer, PHPUnit, etc.).
- `package.json` — Frontend tooling (ESLint, Prettier, html-validate).
- `phpstan.neon` — PHP static analysis config.
- `phpunit.xml` — PHP test runner config.
- `install.sh` — One-command installer for self-hosted deployments.
- `.github/dependabot.yml` — Automated dependency updates. Configured with `delete-dependent-on-merge: true` and grouped by ecosystem (`github-actions`, `php-backend`, `frontend`, `android-app`, `docker`) so each ecosystem gets a single PR.

**CI/CD:**
- **PHP CI** — Lint (PHP-CS-Fixer) + static analysis (PHPStan) on PHP 8.2 & 8.3.
- **Frontend CI** — ESLint + Prettier formatting check + html-validate.
- **Android CI** — Build debug APK + unit tests with JDK 17 / Gradle.


**Important:**
- No official Docker image is built in CI — use `docker/` Dockerfiles with podman for local testing only.
- `git config pull.rebase true` is set locally — branches rebase on pull.

**Requirements:** PHP web server + Memcached or Redis (optional: LDAP). Android 6+ for the app.

**Install:** `sudo ./install.sh -c <web_root>` or copy `backend-php/` + `frontend/` to your web root and edit `backend-php/include/config.php`. Docker users mount `./config/hauk:/etc/hauk`.
