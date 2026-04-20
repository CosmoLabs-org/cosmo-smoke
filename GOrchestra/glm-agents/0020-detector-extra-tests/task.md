# Task

Add tests to internal/detector/ package. Create file internal/detector/detector_extra_test.go in package detector. Test cases: detect Go project from go.mod presence, detect Node project from package.json, detect Docker project from Dockerfile, detect Python project from requirements.txt or pyproject.toml, detect Rust project from Cargo.toml, unknown project type returns default template. Verify: go test ./internal/detector/ -v passes.
