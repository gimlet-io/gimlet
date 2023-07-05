package manifest

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

var (
	rootSchema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
		},
	}

	variableSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "type",
			},
			{
				Name: "description",
			},
			{
				Name: "default",
			},
			{
				Name: "sensitive",
			},
		},
	}
)

// ParseVariables parses the given data from a terraform Module of input variables into a tfconfig.Variable slice.
func ParseVariables(data []byte, out *[]*tfconfig.Variable) error {
	file, diags := hclsyntax.ParseConfig(data, "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("cannot parse config: %s", diags)
	}

	bodyCont, diags := file.Body.Content(rootSchema)
	if diags.HasErrors() {
		return fmt.Errorf("cannot get body content: %s", diags)
	}

	for _, block := range bodyCont.Blocks {
		content, _, contentDiags := block.Body.PartialContent(variableSchema)
		diags = append(diags, contentDiags...)

		name := block.Labels[0]
		v := &tfconfig.Variable{
			Name: name,
			Pos:  sourcePosHCL(block.DefRange),
		}

		if attr, defined := content.Attributes["type"]; defined {
			// We handle this particular attribute in a somewhat-tricky way:
			// since Terraform may evolve its type expression syntax in
			// future versions, we don't want to be overly-strict in how
			// we handle it here, and so we'll instead just take the raw
			// source provided by the user, using the source location
			// information in the expression object.
			//
			// However, older versions of Terraform expected the type
			// to be a string containing a keyword, so we'll need to
			// handle that as a special case first for backward compatibility.

			var typeExpr string

			var typeExprAsStr string
			valDiags := gohcl.DecodeExpression(attr.Expr, nil, &typeExprAsStr)
			if !valDiags.HasErrors() {
				typeExpr = typeExprAsStr
			} else {
				rng := attr.Expr.Range()
				typeExpr = string(rng.SliceBytes(file.Bytes))
			}

			v.Type = typeExpr
		}

		if attr, defined := content.Attributes["description"]; defined {
			var description string
			valDiags := gohcl.DecodeExpression(attr.Expr, nil, &description)
			diags = append(diags, valDiags...)
			v.Description = description
		}

		if attr, defined := content.Attributes["default"]; defined {
			// To avoid the caller needing to deal with cty here, we'll
			// use its JSON encoding to convert into an
			// approximately-equivalent plain Go interface{} value
			// to return.
			val, valDiags := attr.Expr.Value(nil)
			diags = append(diags, valDiags...)
			if val.IsWhollyKnown() { // should only be false if there are errors in the input
				valJSON, err := ctyjson.Marshal(val, val.Type())
				if err != nil {
					// Should never happen, since all possible known
					// values have a JSON mapping.
					return fmt.Errorf("failed to serialize default value as JSON: %s", err)
				}
				var def interface{}
				err = json.Unmarshal(valJSON, &def)
				if err != nil {
					// Again should never happen, because valJSON is
					// guaranteed valid by ctyjson.Marshal.
					return fmt.Errorf("failed to re-parse default value from JSON: %s", err)
				}
				v.Default = def
			}
		} else {
			v.Required = true
		}

		if attr, defined := content.Attributes["sensitive"]; defined {
			var sensitive bool
			valDiags := gohcl.DecodeExpression(attr.Expr, nil, &sensitive)
			diags = append(diags, valDiags...)
			v.Sensitive = sensitive
		}

		*out = append(*out, v)
	}
	return nil
}

func sourcePosHCL(rng hcl.Range) tfconfig.SourcePos {
	// We intentionally throw away the column information here because
	// current and legacy HCL both disagree on the definition of a column
	// and so a line-only reference is the best granularity we can do
	// such that the result is consistent between both parsers.
	return tfconfig.SourcePos{
		Filename: rng.Filename,
		Line:     rng.Start.Line,
	}
}
