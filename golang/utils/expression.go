package utils

import (
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"math"

	"github.com/PaesslerAG/gval"
)

func factorialRecursive(n float64) float64 {
	if n <= 1 {
		return 1
	}
	return n * factorialRecursive(n-1)
}

var languages = []gval.Language{gval.Function("sqrt", math.Sqrt), gval.Function("max", math.Max), gval.Function("min", math.Min),
	gval.Function("abs", math.Abs), gval.Function("log", math.Log), gval.Function("pow", math.Pow), gval.Function("acos", math.Acos),
	gval.Function("cos", math.Cos), gval.Function("asin", math.Asin), gval.Function("j0", math.J0), gval.Function("j1", math.J1),
	gval.Function("atan", math.Atan), gval.Function("tan", math.Tan), gval.Function("sin", math.Sin), gval.Function("jc", factorialRecursive)}

func EvaluableExpression(expr string) string {
	env := map[string]interface{}{"x": 0, "y": 1}
	// 计算复杂表达式
	value, err := gval.Evaluate(expr, env, languages...)
	if err != nil {
		return "err"
	}
	return core.ToString(value)
}
