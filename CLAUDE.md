# Git policy

Never run `git commit` or `git push` (or any equivalent action that creates a commit or pushes to a remote), even if explicitly asked. The user handles all commits and pushes themselves.

# Running PHP tests

Run the PHP tests in the `wshub_user` container, as the host user, from the repo root:

```
docker compose exec -u 1000:1000 wshub_user bash -c "cd wshub_user && vendor/bin/phpunit tests"
```

To run one test file, replace `tests` with its path, for example `tests/Handler/FileUploadHandlerTest.php`. Do not run phpunit directly on the host: the PHP version differs and `vendor/` is not writable there.
