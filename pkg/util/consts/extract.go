// Package consts provides utilities for extracting constant values from Go source code.
package consts

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strconv"
	"strings"
)

// ExtractInt parses the given Go source file and returns all integer constants (including iota-based ones).
// It evaluates per-line iota (reset to 0 in each const block), supports omitted RHS (reuse previous),
// and handles common integer constant expressions: literals, iota, unary +/- and binary + - * / % | & ^ << >>, and parentheses.
func ExtractInt(filePath string) (map[string]int, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	out := make(map[string]int)

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}

		// iota resets per const block
		lineIota := 0
		// previous RHS(Right Hand Side) list for implicit repetition
		var prevRHS []ast.Expr

		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// if the ValueSpec's RHS is empty, reuse the previous RHS
			rhs := vs.Values
			if len(rhs) == 0 {
				rhs = prevRHS
			} else {
				prevRHS = rhs
			}

			// Expand RHS to match names count (spec rule: missing expressions are the last expression)
			expanded := expandRHS(rhs, len(vs.Names))

			for i, name := range vs.Names {
				expr := expanded[i]
				if expr == nil {
					// Nothing to evaluate
					continue
				}
				if v, ok := evalIntConst(expr, lineIota); ok {
					out[name.Name] = v
				}
			}

			lineIota++
		}
	}

	return out, nil
}

// ExtractString parses the given Go source file and returns all string constants.
// It supports literal strings, concatenation with '+', parentheses, and omitted RHS reuse.
func ExtractString(filePath string) (map[string]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}

		lineIota := 0 // unused for strings but kept for symmetry
		var prevRHS []ast.Expr

		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			rhs := vs.Values
			if len(rhs) == 0 {
				rhs = prevRHS
			} else {
				prevRHS = rhs
			}

			expanded := expandRHS(rhs, len(vs.Names))

			for i, name := range vs.Names {
				expr := expanded[i]
				if expr == nil {
					continue
				}
				if v, ok := evalStringConst(expr, lineIota); ok {
					out[name.Name] = v
				}
			}

			lineIota++
		}
	}

	return out, nil
}

// expandRHS expands RHS expressions to match the number of names, repeating the last expression as needed.
func expandRHS(rhs []ast.Expr, n int) []ast.Expr {
	if n == 0 {
		return nil
	}
	if len(rhs) == 0 {
		// no previous RHS to reuse
		return make([]ast.Expr, n)
	}
	out := make([]ast.Expr, n)
	last := rhs[len(rhs)-1]
	for i := range n {
		if i < len(rhs) {
			out[i] = rhs[i]
		} else {
			out[i] = last
		}
	}
	return out
}

// evalIntConst evaluates an integer constant expression with the given iota value.
func evalIntConst(e ast.Expr, iotaVal int) (int, bool) {
	switch ex := e.(type) {
	case *ast.ParenExpr:
		return evalIntConst(ex.X, iotaVal)
	case *ast.Ident:
		if ex.Name == "iota" {
			return iotaVal, true
		}
		return 0, false
	case *ast.BasicLit:
		if ex.Kind == token.INT {
			if v, err := strconv.ParseInt(ex.Value, 0, 64); err == nil {
				return int(v), true
			}
		}
		return 0, false
	case *ast.UnaryExpr:
		if ex.Op == token.ADD || ex.Op == token.SUB {
			if v, ok := evalIntConst(ex.X, iotaVal); ok {
				if ex.Op == token.SUB {
					return -v, true
				}
				return v, true
			}
		}
		return 0, false
	case *ast.BinaryExpr:
		lv, lok := evalIntConst(ex.X, iotaVal)
		rv, rok := evalIntConst(ex.Y, iotaVal)
		if !lok || !rok {
			return 0, false
		}
		switch ex.Op {
		case token.ADD:
			return lv + rv, true
		case token.SUB:
			return lv - rv, true
		case token.MUL:
			return lv * rv, true
		case token.QUO:
			if rv == 0 {
				return 0, false
			}
			return lv / rv, true
		case token.REM:
			if rv == 0 {
				return 0, false
			}
			return lv % rv, true
		case token.OR:
			return lv | rv, true
		case token.AND:
			return lv & rv, true
		case token.XOR:
			return lv ^ rv, true
		case token.SHL:
			if rv < 0 {
				return 0, false
			}
			return int(uint64(lv) << uint(rv)), true
		case token.SHR:
			if rv < 0 {
				return 0, false
			}
			return int(uint64(lv) >> uint(rv)), true
		default:
			return 0, false
		}
	default:
		return 0, false
	}
}

// evalStringConst evaluates a string constant expression, supporting literals, concatenation with '+', and parentheses.
func evalStringConst(e ast.Expr, _ int) (string, bool) {
	switch ex := e.(type) {
	case *ast.ParenExpr:
		return evalStringConst(ex.X, 0)
	case *ast.BasicLit:
		if ex.Kind == token.STRING || ex.Kind == token.CHAR {
			if v, err := strconv.Unquote(ex.Value); err == nil {
				return v, true
			}
			// if unquote fails, return raw without quotes best-effort
			return ex.Value, true
		}
		return "", false
	case *ast.BinaryExpr:
		if ex.Op == token.ADD {
			ls, lok := evalStringConst(ex.X, 0)
			rs, rok := evalStringConst(ex.Y, 0)
			if lok && rok {
				return ls + rs, true
			}
		}
		return "", false
	case *ast.Ident:
		// strings rarely use iota; not supported
		return "", false
	default:
		return "", false
	}
}

// ExtractComment extracts comments of constant declarations from a Go source file.
// It needs a file path and a custom parse function to process the comments.
// parse function will handle the extracted comments and write the results to the values of output map.
func ExtractComment(filePath string, parse func(string) string) (map[string]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}
	cmap := ast.NewCommentMap(fset, file, file.Comments)

	out := make(map[string]string)

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}

		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, cg := range cmap[vs] {
				comment := parse(cg.Text())
				for _, name := range vs.Names {
					out[name.Name] = comment
				}
			}
		}
	}

	return out, nil
}

// ParseErrExternal trims the comment to its relevant part.
// e.g.
//
//	// ErrUserFound - 404: User not found.
//	return User not found
func ParseErrExternal(com string) string {
	if !strings.HasPrefix(com, "Err") {
		log.Println("Warn: comment does not start with 'Err', ", com)
		return ""
	}

	return strings.Split(strings.Split(com, ".")[0], ": ")[1]
}

// ParseErrHTTPStatus trims the comment to extract the HTTP status code.
// e.g.
//
//	// ErrUserFound - 404: User not found.
//	return "404"
func ParseErrHTTPStatus(com string) string {
	if !strings.HasPrefix(com, "Err") {
		log.Println("Warn: comment does not start with 'Err', ", com)
		return ""
	}

	return strings.Split(strings.Split(com, ":")[0], " ")[2]
}
