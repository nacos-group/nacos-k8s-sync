FINDFILES=find . \( -path ./common-protos -o -path ./.git -o -path ./out -o -path ./.github -o -path ./licenses -o -path ./vendor \) -prune -o -type f

tidy-go:
	@go mod tidy

lint-go:
	@${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' \) \) -print0 | ${XARGS} common/scripts/lint_go.sh

format-go: tidy-go
	@${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' \) \) -print0 | ${XARGS} common/scripts/format_go.sh

precommit: format-go lint-go