package main

import "runtime"

// runtimeSetFinalizerForBuildTime 用 SetFinalizer 强制持有 BuildTime 指针，
// 防止编译器/链接器对 -X main.BuildTime 注入的符号进行 dead-code 优化。
//
// SetFinalizer 会让编译器无法证明该值无副作用，从而保留符号。
func runtimeSetFinalizerForBuildTime() {
	v := BuildTime
	runtime.SetFinalizer(&v, func(p *string) {})
}