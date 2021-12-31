go_library(
    name = "vso_hash",
    srcs = ["vso_hash.go"],
)

go_test(
    name = "vso_hash_test",
    srcs = ["vso_hash_test.go"],
    deps = [
        ":testify",
        ":vso_hash",
    ],
)

go_benchmark(
    name = "vso_hash_benchmark",
    srcs = ["vso_hash_benchmark_test.go"],
    deps = [
        ":vso_hash",
    ],
)

go_module(
    name = "testify",
    install = [
        "assert",
        "require",
    ],
    licences = ["MIT"],
    module = "github.com/stretchr/testify",
    version = "v1.7.0",
    deps = [
        ":difflib",
        ":spew",
        ":yaml",
    ],
)

go_module(
    name = "difflib",
    install = ["difflib"],
    licences = ["BSD-3-Clause"],
    module = "github.com/pmezard/go-difflib",
    version = "v1.0.0",
)

go_module(
    name = "spew",
    install = ["spew"],
    licences = ["ISC"],
    module = "github.com/davecgh/go-spew",
    version = "v1.1.1",
)

go_module(
    name = "yaml",
    licences = ["MIT"],
    module = "gopkg.in/yaml.v3",
    version = "v3.0.0-20210107192922-496545a6307b",
)
