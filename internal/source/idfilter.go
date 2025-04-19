package source

import (
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type IDFilter struct {
	expression string
	program    *vm.Program
	now        time.Time
}

type env struct {
	ID        uint64         `expr:"id"`
	Sid       string         `expr:"sid"`
	CreatedAt time.Time      `expr:"createdAt"`
	Age       *time.Duration `expr:"age"`
	AgeDays   *int64         `expr:"ageDays"`
}

func NewIDFilter(expression string, now time.Time) (IDFilter, error) {
	envSample := env{} //nolint:exhaustruct

	program, err := expr.Compile(
		expression,
		expr.Env(envSample),
		expr.AsBool(),
		expr.Function("now", func(_ ...any) (any, error) {
			return now, nil
		}),
	)

	if err != nil {
		return IDFilter{}, &CompileFilterError{Expression: expression, Err: err}
	}

	return IDFilter{expression: expression, program: program, now: now}, nil
}

func (filter *IDFilter) Test(id ID) (bool, error) {
	env := filter.createEnv(id)

	result, err := expr.Run(filter.program, &env)

	if err != nil {
		return false, &RunFilterError{Expression: filter.expression, ID: id, Err: err}
	}

	return result.(bool), nil //nolint:forcetypeassert
}

func (filter *IDFilter) createEnv(id ID) env {
	createdAt, err := id.Time()

	if err != nil {
		panic(err)
	}

	result := env{
		ID:        id.Uint64(),
		Sid:       id.String(),
		CreatedAt: createdAt,
		Age:       nil,
		AgeDays:   nil,
	}

	if now := filter.now; createdAt.Before(now) {
		const hoursInDay = 24

		age := now.Sub(createdAt)
		ageDays := int64(age.Hours() / hoursInDay)

		result.Age = &age
		result.AgeDays = &ageDays
	}

	return result
}
