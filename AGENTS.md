# AGENTS.md

## Hauk — Self-hosted location sharing service

**Stack:** PHP 8.2+/8.3 backend, vanilla JavaScript frontend (no framework), Android app (Kotlin/Gradle).

**Directory layout:**
- `backend-php/` — PHP backend. API endpoints, config (`include/`), session store (Memcached/Redis).
- `frontend/` — Single-page frontend: `index.html`, `main.js`, `style.css`, `lib/`, `assets/`.
- `android/` — Android companion app: Gradle build, Kotlin source under `app/`.
- `docker/` — Dockerfile + compose assets for the official `bilde2910/hauk` image (multi-arch: amd64/armv7/arm64).
- `.github/workflows/` — CI pipelines for PHP, frontend, and Android.

**Key files:**
- `composer.json` — PHP dependencies (PHPStan, PHP-CS-Fixer, PHPUnit, etc.).
- `package.json` — Frontend tooling (ESLint, Prettier, html-validate).
- `phpstan.neon` — PHP static analysis config.
- `phpunit.xml` — PHP test runner config.
- `install.sh` — One-command installer for self-hosted deployments.
- `.github/dependabot.yml` — Automated dependency updates.

**CI/CD:**
- **PHP CI** — Lint (PHP-CS-Fixer) + static analysis (PHPStan) on PHP 8.2 & 8.3.
- **Frontend CI** — ESLint + Prettier formatting check + html-validate.
- **Android CI** — Build debug APK + unit tests with JDK 17 / Gradle.


**Requirements:** PHP web server + Memcached or Redis (optional: LDAP). Android 6+ for the app.

**Install:** `sudo ./install.sh -c <web_root>` or copy `backend-php/` + `frontend/` to your web root and edit `backend-php/include/config.php`. Docker users mount `./config/hauk:/etc/hauk`.
