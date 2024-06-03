package evaluator

import (
	"fmt"
	"interpreter/ast"
	"interpreter/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Enviroment) object.Object {
	switch node := node.(type) {

	case *ast.Program:
		return evalProgram(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpressions(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, right, left)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.BlockStatements:
		return evalStatements(node.Statements, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		exp := Eval(node.Value, env)
		if isError(exp) {
			return exp
		}
		env.Set(node.Name.Value, exp)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters, Body: node.Body, Env: env}

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		params := evalExpressions(node.Arguments, env)
		if len(params) == 1 && isError(params[0]) {
			return params[0]
		}

		return applyFunction(function, params)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.Array:
		ele := evalExpressions(node.Items, env)
		if len(ele) == 1 && isError(ele[0]) {
			return ele[0]
		}
		return &object.Array{Elements: ele}
	case *ast.IndexExpression:
		leftexp := Eval(node.LeftExpression, env)
		index := Eval(node.Index, env)
		if isError(leftexp) {
			return leftexp
		}
		if isError(index) {
			return index
		}
		return evalIndexExpression(leftexp, index)

	}

	return nil
}

func nativeBoolObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

func evalIndexExpression(left object.Object, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array object.Object, index object.Object) object.Object {
	arrayObj := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObj.Elements[idx]
}

func evalProgram(program *ast.Program, env *object.Enviroment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalPrefixExpressions(op string, val object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExpression(val)
	case "-":
		return evalMinusPrefixOperator(val)
	default:
		return newError("unknown operator: %s%s", op, val.Type())
	}
}

func evalBangOperatorExpression(val object.Object) object.Object {
	if val == TRUE {
		return FALSE
	} else if val == FALSE {
		return TRUE
	} else if val == NULL {
		return TRUE
	}
	return FALSE
}

func evalMinusPrefixOperator(val object.Object) object.Object {
	if val.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", val.Type())
	}

	value := val.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(op string, right object.Object, left object.Object) object.Object {
	switch {
	case right.Type() == object.INTEGER_OBJ && left.Type() == object.INTEGER_OBJ:
		return evalInfixIntegerExpression(op, right, left)
	case right.Type() == object.STRING_OBJ && left.Type() == object.STRING_OBJ:
		return evalInfixStringExpression(op, right, left)
	case op == "==":
		return nativeBoolObject(left == right)
	case op == "!=":
		return nativeBoolObject(right != left)
	case right.Type() != left.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}

}

func evalInfixStringExpression(op string, right object.Object, left object.Object) object.Object {
	if op != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), op, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	return &object.String{Value: leftVal + rightVal}
}

func evalInfixIntegerExpression(op string, right object.Object, left object.Object) object.Object {
	right_val := right.(*object.Integer).Value
	left_val := left.(*object.Integer).Value

	switch op {
	case "+":
		return &object.Integer{Value: right_val + left_val}
	case "-":
		return &object.Integer{Value: left_val - right_val}
	case "*":
		return &object.Integer{Value: right_val * left_val}
	case "/":
		return &object.Integer{Value: left_val / right_val}
	case ">":
		return nativeBoolObject(left_val > right_val)
	case "<":
		return nativeBoolObject(left_val < right_val)
	case "==":
		return nativeBoolObject(left_val == right_val)
	case "!=":
		return nativeBoolObject(left_val != right_val)
	}

	return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
}

func evalIfExpression(ie *ast.IfExpression, env *object.Enviroment) object.Object {

	res := Eval(ie.Condition, env)
	if isError(res) {
		return res
	}
	if res.Type() == object.BOOLEAN_OBJ && (res == NULL || res == FALSE) {
		if ie.Alternatives == nil {
			return NULL
		}
		return Eval(ie.Alternatives, env)
	}

	return Eval(ie.Consequence, env)

}

func evalStatements(stmts []ast.Statement, env *object.Enviroment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)
		if result != nil {
			if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Enviroment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	} else if val, ok := builtins[node.Value]; ok {
		return val
	}
	return newError("identifier not found: %s", node.Value)
}

func evalExpressions(exps []ast.Expression, env *object.Enviroment) []object.Object {
	res := []object.Object{}
	for _, exp := range exps {
		evaluated := Eval(exp, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		res = append(res, evaluated)
	}
	return res
}

func applyFunction(fn object.Object, params []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		new_env := object.NewEnclosedEnviroment(fn.Env)
		for paramID, p := range fn.Parameters {
			new_env.Set(p.Value, params[paramID])
		}
		evaluated := Eval(fn.Body, new_env)
		if evaluated, ok := evaluated.(*object.ReturnValue); ok {
			return evaluated.Value
		}
		return evaluated

	case *object.Builtin:
		if len(params) != 1 {
			return newError("wrong number of arguments. got=%d, want=1", len(params))
		}
		return fn.Fn(params...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func newError(format string, a ...interface{}) object.Object {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
