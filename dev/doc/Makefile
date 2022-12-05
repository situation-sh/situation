

mdoc: $(shell find -name "*.go")
	go build -o mdoc $^

run: mdoc
	./$^ -d ../../modules
