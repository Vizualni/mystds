// (Finite|Function)StateMachine is a generic interface for a state machine
// using functions to define states.
package myfsm

import "context"

type errstate struct {
	error
}

func (e errstate) Error() string {
	return e.error.Error()
}

func (e errstate) Transition() Transitioner {
	return nil
}

type retval struct {
	Value any
}

func Return(value any) retval {
	return retval{Value: value}
}

func Error(err error) errstate {
	return errstate{error: err}
}

func (v retval) Transition() Transitioner {
	return nil
}

type Func func() Transitioner

func (f Func) Transition() Transitioner {
	return f()
}

type Transitioner interface {
	Transition() Transitioner
}

func Start(ctx context.Context, initial Transitioner) (ret any, err error) {
	var reterror error
	var retValue any
	for f := initial; f != nil && reterror == nil; {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var next Transitioner
		func() {
			defer func() {
				if r := recover(); r != nil {
					reterror = r.(error)
				}
			}()
			next = f.Transition()
		}()

		errstate, ok := next.(errstate)
		if ok {
			reterror = errstate
			break
		}

		if valstate, ok := next.(retval); ok {
			retValue = valstate.Value
			break
		}

		if next == nil {
			retValue = nil
			break
		}

		f = next
	}
	return retValue, reterror
}
