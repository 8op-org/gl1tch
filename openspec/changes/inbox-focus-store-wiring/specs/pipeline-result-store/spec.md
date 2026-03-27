## ADDED Requirements

### Requirement: pipeline CLI records runs
`cmd/pipeline.go` `pipelineRunCmd` MUST open the store and pass it to `pipeline.Run()`.

#### Scenario: pipeline run succeeds
- **WHEN** `orcai pipeline run <yaml>` completes with exit 0
- **THEN** a row exists in `runs` table with `exit_status = 0` and `stdout` populated

#### Scenario: pipeline run fails
- **WHEN** `orcai pipeline run <yaml>` exits non-zero
- **THEN** a row exists in `runs` table with `exit_status != 0`

#### Scenario: store open fails
- **WHEN** store cannot be opened (permissions, disk full)
- **THEN** pipeline still runs; store error is logged to stderr but does not abort the pipeline
