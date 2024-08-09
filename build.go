//go:generate sh -c "echo Generating for $TARGET_GOARCH"
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target $TARGET_GOARCH -output-dir ./cmd -cc clang -no-strip KProbePWRU ./bpf/kprobe_pwru.c -- -I./bpf/headers -Wno-address-of-packed-member

package main
