package generate_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/lacework/go-sdk/generate"
	"github.com/stretchr/testify/assert"
)

func TestGenericBlockCreation(t *testing.T) {
	data := generate.HclCreateGenericBlock(
		"thing",
		[]string{"a", "b"},
		map[string]interface{}{
			"a": "foo",
			"b": 1,
			"c": false,
			"d": map[string]interface{}{ // Order of map elements should be sorted when executed
				"f": 1,
				"g": "bar",
				"e": true,
			},
			"h": hcl.Traversal{
				hcl.TraverseRoot{
					Name: "module",
				},
				hcl.TraverseAttr{
					Name: "example",
				},
				hcl.TraverseAttr{
					Name: "value",
				},
			},
		},
	)

	assert.Equal(t, "thing", data.Type())
	assert.Equal(t, "a", data.Labels()[0])
	assert.Equal(t, "b", data.Labels()[1])
	assert.Equal(t, "a=\"foo\"\n", string(data.Body().GetAttribute("a").BuildTokens(nil).Bytes()))
	assert.Equal(t, "b=1\n", string(data.Body().GetAttribute("b").BuildTokens(nil).Bytes()))
	assert.Equal(t, "c=false\n", string(data.Body().GetAttribute("c").BuildTokens(nil).Bytes()))
	assert.Equal(t, "d={\n  e = true\n  f = 1\n  g = \"bar\"\n}\n", string(data.Body().GetAttribute("d").BuildTokens(nil).Bytes()))
	assert.Equal(t, "h=module.example.value\n", string(data.Body().GetAttribute("h").BuildTokens(nil).Bytes()))
}

func TestModuleBlock(t *testing.T) {
	data := generate.CreateModule(&generate.HclModule{
		Name:    "foo",
		Version: "~> 0.1",
		Attributes: map[string]interface{}{
			"bar": "foo",
		},
	})

	assert.Equal(t, "module", data.Type())
	assert.Equal(t, "foo", data.Labels()[0])
	assert.Equal(t,
		"version=\"~> 0.1\"\n",
		string(data.Body().GetAttribute("version").BuildTokens(nil).Bytes()),
	)
	assert.Equal(t,
		"bar=\"foo\"\n",
		string(data.Body().GetAttribute("bar").BuildTokens(nil).Bytes()),
	)
}
func TestModuleWithProviderBlock(t *testing.T) {
	data := generate.CreateModule(&generate.HclModule{
		Name: "foo",
		ProviderDetails: map[string]string{
			"foo.src": "test.abc",
			"foo.dst": "abc.test",
		},
	})

	assert.Equal(t, "module", data.Type())
	assert.Equal(t, "foo", data.Labels()[0])
	assert.Equal(t,
		"providers= {\nfoo.dst= abc.test\nfoo.src= test.abc\n}\n",
		string(data.Body().GetAttribute("providers").BuildTokens(nil).Bytes()))
}

func TestProviderBlock(t *testing.T) {
	data := generate.CreateProvider(&generate.HclProvider{
		Name: "foo",
		Attributes: map[string]interface{}{
			"key": "value",
		},
	})

	assert.Equal(t, "provider", data.Type())
	assert.Equal(t, "foo", data.Labels()[0])
	assert.Equal(t, "key=\"value\"\n", string(data.Body().GetAttribute("key").BuildTokens(nil).Bytes()))
}

func TestProviderBlockWithTraversal(t *testing.T) {
	data := generate.CreateProvider(&generate.HclProvider{
		Name: "foo",
		Attributes: map[string]interface{}{
			"test": hcl.Traversal{
				hcl.TraverseRoot{Name: "key"},
				hcl.TraverseAttr{Name: "value"},
			},
		},
	})

	assert.Equal(t, "provider", data.Type())
	assert.Equal(t, "foo", data.Labels()[0])
	assert.Equal(t, "test=key.value\n", string(data.Body().GetAttribute("test").BuildTokens(nil).Bytes()))
}

func TestRequiredProvidersBlock(t *testing.T) {
	provider1 := &generate.HclRequiredProvider{Name: "foo", Source: "test/test"}
	provider2 := &generate.HclRequiredProvider{Name: "bar", Version: "~> 0.1"}
	provider3 := &generate.HclRequiredProvider{Name: "lacework", Version: "~> 0.1", Source: "lacework/lacework"}
	data := generate.CreateRequiredProviders([]*generate.HclRequiredProvider{provider1, provider2, provider3})
	assert.Equal(t, testRequiredProvider, generate.CreateHclStringOutput([]*hclwrite.Block{data}))
}

var testRequiredProvider = `terraform {
  required_providers {
    bar = {
      version = "~> 0.1"
    }
    foo = {
      source = "test/test"
    }
    lacework = {
      source  = "lacework/lacework"
      version = "~> 0.1"
    }
  }
}

`
