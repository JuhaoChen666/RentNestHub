# Git Workflow

## Commit Messages

Use Conventional Commits:

```text
<type>(<scope>): <summary>
```

Examples:

```text
feat(house-publish): validate listing form payload
feat(search): add rent range filters
feat(ai-recommend): rank houses by tenant needs
fix(api): return 404 for missing houses
docs(pr): add develop initialization message
chore(deps): update frontend lockfile
```

Recommended types:

- `feat`: user-facing feature
- `fix`: bug fix
- `docs`: documentation
- `test`: tests only
- `refactor`: behavior-preserving code change
- `chore`: tooling, dependencies, build, or repository maintenance

## Branch Names

Use slash-separated feature branches:

```text
feature/house-publish
feature/search
feature/ai-recommend
```

Branch from `develop` for feature work and merge back into `develop` after review.

## PR Size

Keep PRs atomic. A PR should normally stay under 300 changed lines.

Accepted exceptions:

- conflict-resolution PRs
- generated lockfiles or checksum files
- one-time initialization work that cannot be meaningfully split

If a PR exceeds the limit, explain the reason in the PR body before review.

## PR Message

Every PR should include:

- what changed
- why it changed
- validation commands or screenshots
- risk and rollback notes

