package generate

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type HclRequiredProvider struct {
	// Required, provider name
	Name string

	// Optional. Source for this provider
	Source string

	// Optional. Version for this provider
	Version string
}

type HclProvider struct {
	// Required, provider name
	Name string

	// Optional. Extra properties for this module.  Can supply string, bool, int, or map[string]interface{} as values
	Attributes map[string]interface{}
}

type HclModule struct {
	// Required, module name
	Name string

	// Required, source for this module
	Source string

	// Required, version
	Version string

	// Optional. Extra properties for this module.  Can supply string, bool, int, or map[string]interface{} as values
	Attributes map[string]interface{}

	// Optional.  Provider details to override defaults.  These values must be supplied as strings, and raw values will be
	// accepted.  Unfortunately map[string]hcl.Traversal is not a format that is supported by hclwrite.SetAttributeValue
	// today so we must work around it (https://github.com/hashicorp/hcl/issues/347).
	ProviderDetails map[string]string
}

func convertTypeToCty(value interface{}) cty.Value {
	switch v := value.(type) {
	case string:
		return cty.StringVal(v)
	case int:
		return cty.NumberIntVal(int64(v))
	case bool:
		return cty.BoolVal(v)
	default:
		panic("Unknown attribute value type")
	}
}

// Helper to create various types of new hclwrite.Block using generic inputs
func HclCreateGenericBlock(hcltype string, labels []string, attr map[string]interface{}) *hclwrite.Block {
	block := hclwrite.NewBlock(hcltype, labels)

	// Source and version require some special handling, should go at the top of a block declaration
	sourceFound := false
	versionFound := false

	// We need/want to guarentee the ordering of the attributes, do that here
	var keys []string
	for k := range attr {
		switch k {
		case "source":
			sourceFound = true
		case "version":
			versionFound = true
		default:
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	if sourceFound || versionFound {
		var newKeys []string
		if sourceFound {
			newKeys = append(newKeys, "source")
		}
		if versionFound {
			newKeys = append(newKeys, "version")
		}
		keys = append(newKeys, keys...)
	}

	// Write block data
	// TODO lists
	for _, key := range keys {
		val := attr[key]
		switch v := val.(type) {
		case string, int, bool:
			block.Body().SetAttributeValue(key, convertTypeToCty(v))
		case hcl.Traversal:
			block.Body().SetAttributeTraversal(key, v)
		case map[string]interface{}:
			data := map[string]cty.Value{}
			for attrKey, attrVal := range v {
				data[attrKey] = convertTypeToCty(attrVal)
			}
			block.Body().SetAttributeValue(key, cty.ObjectVal(data))
		default:
			panic("Unknown type")
		}
	}

	return block
}

// Create tokens for map of traversals.  Used as a workaround for writing complex types where the built-in
// SetAttributeValue won't work
func createMapTraversalTokens(input map[string]string) hclwrite.Tokens {
	// Sort input
	keys := []string{}
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tokens := hclwrite.Tokens{
		{Type: hclsyntax.TokenOBrace, Bytes: []byte("{"), SpacesBefore: 1},
		{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
	}

	for _, k := range keys {
		tokens = append(tokens, []*hclwrite.Token{
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(k)},
			{Type: hclsyntax.TokenEqual, Bytes: []byte("=")},
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(" " + input[k]), SpacesBefore: 1},
			{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
		}...)
	}

	tokens = append(tokens, []*hclwrite.Token{
		{Type: hclsyntax.TokenNewline},
		{Type: hclsyntax.TokenCBrace, Bytes: []byte("}")},
	}...)

	return tokens
}

// Create a module statement in the HCL output
func CreateModule(module *HclModule) *hclwrite.Block {
	if module.Attributes == nil {
		module.Attributes = make(map[string]interface{})
	}
	if module.Source != "" {
		module.Attributes["source"] = module.Source

	}
	if module.Version != "" {
		module.Attributes["version"] = module.Version
	}
	block := HclCreateGenericBlock(
		"module",
		[]string{module.Name},
		module.Attributes,
	)
	if module.ProviderDetails != nil {
		block.Body().AppendNewline()
		block.Body().SetAttributeRaw("providers", createMapTraversalTokens(module.ProviderDetails))
	}

	return block

}

// Create a provider statement in the HCL output
func CreateProvider(provider *HclProvider) *hclwrite.Block {
	return HclCreateGenericBlock("provider", []string{provider.Name}, provider.Attributes)
}

// Convert blocks to a string
func CreateHclStringOutput(blocks []*hclwrite.Block) string {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	for _, b := range blocks {
		body.AppendBlock(b)
		body.AppendNewline()
	}
	return fmt.Sprintf("%s", file.Bytes())
}

// Create required providers block
func CreateRequiredProviders(providers []*HclRequiredProvider) *hclwrite.Block {
	block := HclCreateGenericBlock("terraform", nil, nil)
	providerDetails := map[string]interface{}{}
	for _, provider := range providers {
		details := map[string]interface{}{}
		if provider.Source != "" {
			details["source"] = provider.Source
		}
		if provider.Version != "" {
			details["version"] = provider.Version
		}
		providerDetails[provider.Name] = details
	}
	block.Body().AppendBlock(
		HclCreateGenericBlock(
			"required_providers",
			nil,
			providerDetails,
		),
	)

	return block
}

// helper to create a hcl.Traversal in the order of supplied []string
//
// e.g. []string{"a", "b", "c"} as input results in traversal having value a.b.c
func CreateSimpleTraversal(input []string) hcl.Traversal {
	traversers := []hcl.Traverser{}

	for i, val := range input {
		if i == 0 {
			traversers = append(traversers, hcl.TraverseRoot{Name: val})
		} else {
			traversers = append(traversers, hcl.TraverseAttr{Name: val})
		}
	}
	return traversers
}

// Simple helper to combine multiple blocks (or slices of blocks) into a single slice to be rendered to string
func CombineHclBlocks(results ...interface{}) []*hclwrite.Block {
	blocks := []*hclwrite.Block{}
	// Combine all blocks into single flat slice
	for _, result := range results {
		switch v := result.(type) {
		case *hclwrite.Block:
			if v != nil {
				blocks = append(blocks, v)
			}
		case []*hclwrite.Block:
			if len(v) > 0 {
				blocks = append(blocks, v...)
			}
		default:
			panic("Unknown type supplied!")
		}
	}

	return blocks
}
