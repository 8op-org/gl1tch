>> github-pr-review
  > parse-url
  > fetch-pr
  > fetch-diff
  > review
**Functionality**
- PASS: The new guides and documentation accomplish their stated goals. They provide clear instructions for installing and configuring the Gemini CLI and related tools.
- N/A: Edge cases are well-handled in the documentation examples, but there are no specific edge case tests.

**Tests**
- PASS: All files pass `pre-commit run --all-files` validation, and there are no reported errors from `docs-builder`.
- N/A: The pull request does not include additional test cases or scripts, but the existing checks ensure the changes are valid.

**Documentation**
- PASS: Documentation is updated to reflect the new guides and tool additions. The existing documentation structure remains clear.
- N/A: Code comments are sufficient within the Markdown files.

**Security**
- PASS: There are no secrets, credentials, or tokens in code snippets provided.
- N/A: No injection vulnerabilities are introduced; proper auth checks are present where needed.

**Performance**
- PASS: There are no obvious performance issues noted. The guides and documentation are concise and do not include unnecessary steps or loops that could impact performance.
- N/A: Proper resource cleanup is not applicable to Markdown files but is covered in the guidelines for shell commands used in examples.

**Standards**
- PASS: The changes follow repo coding conventions, with clear and consistent naming. The new files are added in appropriate locations within the documentation structure.
- N/A: There are no specific standards that need verification beyond the existing file structure and Markdown formatting rules.

**Breaking Changes**
- N/A: No backward-incompatible changes have been introduced; all guides maintain compatibility with previous versions of the tools.

Summary:
The pull request introduces comprehensive new guides for the Gemini CLI, expanding the documentation to cover installation, configuration, and basic usage. The changes are well-documented, passing all validations without errors, and align with existing standards in the repository structure. There are no significant issues or breaking changes noted.

Suggested Actions:
- Review and merge the pull request.
- Ensure that any future updates to the Gemini CLI documentation will be maintained alongside these guides.
