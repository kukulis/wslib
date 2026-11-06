# Git policy

Never run `git commit` or `git push` (or any equivalent action that creates a commit or pushes to a remote), even if explicitly asked. The user handles all commits and pushes themselves.

# Running PHP tests

Run the PHP tests in the `wshub_user` container, as the host user, from the repo root:

```
docker compose exec -u 1000:1000 wshub_user bash -c "cd wshub_user && vendor/bin/phpunit tests"
```

To run one test file, replace `tests` with its path, for example `tests/Handler/FileUploadHandlerTest.php`. Do not run phpunit directly on the host: the PHP version differs and `vendor/` is not writable there.

# Writing tests

- Readability comes first, not fewer lines. Do not extract a helper only to shorten code.
- A helper does one job on one abstraction layer. A factory only creates the object. It does not also start, configure, or run it. The test body shows the steps itself: arrange (create, start, set up mocks), act, assert.
- Do not use void helpers that change mock behaviour behind the scenes. Separate building a thing (for example a closure) from assigning it: the helper returns the closure and the test assigns it to the mock.
- A test must make the state of the thing under test visible. After every call that advances a process (such as `uploadFileCycle()`), assert whether it is finished, and assert its outcome (success, failure or cancelled) right away. Check intermediate states too, for example "not finished yet, so not successful".
- Give such assertions a message that names the state, for example "the cycle must be finished" or "the upload must have failed", so the reader does not have to infer it from `true` or `false`.
- Every test of the same flow asserts the outcome (such as `isUploadSuccess()`), including tests that mainly check something else, like a file path.
- Mocks in `tests/Mock` follow the `ClientMock` pattern: a nullable closure per method, a fluent setter per closure, and a safe default when no closure is set.
