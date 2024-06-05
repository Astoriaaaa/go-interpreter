package evaluator

import (
	"fmt"
	"interpreter/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.Array:
				length := len(arg.Elements)
				return &object.Integer{Value: int64(length)}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}

		},
	},
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.Array:
				length := len(arg.Elements)
				if length < 1 {
					return NULL
				}
				return arg.Elements[0]
			default:
				return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
			}

		},
	},
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.Array:
				length := len(arg.Elements)
				if length < 1 {
					return NULL
				}
				return arg.Elements[length-1]
			default:
				return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
			}

		},
	},
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.Array:
				length := len(arg.Elements)
				if length < 1 {
					return NULL
				}
				return &object.Array{Elements: arg.Elements[1:]}
			default:
				return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
			}

		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Array{Elements: append(arg.Elements, args[1])}
			default:
				return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
			}

		},
	},
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
}
