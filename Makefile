.PHONY: generate-7zip

generate-7zip:
	go generate ./szextractor/formulas
	sed -i -e 's/{/{\n/g' -e 's/}/,\n}/g' szextractor/formulas/formulas.go
	gofmt -w szextractor/formulas/formulas.go
