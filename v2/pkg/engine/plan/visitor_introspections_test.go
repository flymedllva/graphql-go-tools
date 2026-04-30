package plan

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wundergraph/astjson"

	"github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/internal/unsafeparser"
)

func TestVisitorResolveSkipArrayItemForIncludeDeprecated(t *testing.T) {
	deprecatedItem := astjson.MustParseBytes([]byte(`{"isDeprecated":true}`))
	nonDeprecatedItem := astjson.MustParseBytes([]byte(`{"isDeprecated":false}`))
	ctxWithVars := func(raw string) *resolve.Context {
		if raw == "" {
			raw = `{}`
		}
		return &resolve.Context{Variables: astjson.MustParseBytes([]byte(raw))}
	}

	type testCase struct {
		name                       string
		query                      string
		fieldName                  string
		enclosingTypeName          string
		variables                  string
		expectDeprecatedSkipped    bool
		expectNonDeprecatedSkipped bool
	}

	testCases := []testCase{
		{
			name: "inputFields literal true keeps deprecated",
			query: `
				query {
					__type(name: "Query") {
						inputFields(includeDeprecated: true) { name }
					}
				}
			`,
			fieldName: "inputFields", enclosingTypeName: "__Type",
			expectDeprecatedSkipped: false, expectNonDeprecatedSkipped: false,
		},
		{
			name: "inputFields literal false skips deprecated",
			query: `
				query {
					__type(name: "Query") {
						inputFields(includeDeprecated: false) { name }
					}
				}
			`,
			fieldName: "inputFields", enclosingTypeName: "__Type",
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "inputFields literal null uses default include",
			query: `
				query {
					__type(name: "Query") {
						inputFields(includeDeprecated: null) { name }
					}
				}
			`,
			fieldName: "inputFields", enclosingTypeName: "__Type",
			expectDeprecatedSkipped: false, expectNonDeprecatedSkipped: false,
		},
		{
			name: "inputFields variable true keeps deprecated",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						inputFields(includeDeprecated: $includeDeprecated) { name }
					}
				}
			`,
			fieldName: "inputFields", enclosingTypeName: "__Type", variables: `{"includeDeprecated":true}`,
			expectDeprecatedSkipped: false, expectNonDeprecatedSkipped: false,
		},
		{
			name: "inputFields variable false skips deprecated",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						inputFields(includeDeprecated: $includeDeprecated) { name }
					}
				}
			`,
			fieldName: "inputFields", enclosingTypeName: "__Type", variables: `{"includeDeprecated":false}`,
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "inputFields variable null behaves as false",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						inputFields(includeDeprecated: $includeDeprecated) { name }
					}
				}
			`,
			fieldName: "inputFields", enclosingTypeName: "__Type", variables: `{"includeDeprecated":null}`,
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "inputFields variable omitted behaves as false",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						inputFields(includeDeprecated: $includeDeprecated) { name }
					}
				}
			`,
			fieldName: "inputFields", enclosingTypeName: "__Type", variables: `{}`,
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "inputFields without argument includes deprecated by default",
			query: `
				query {
					__type(name: "Query") {
						inputFields { name }
					}
				}
			`,
			fieldName: "inputFields", enclosingTypeName: "__Type",
			expectDeprecatedSkipped: false, expectNonDeprecatedSkipped: false,
		},
		{
			name: "fields literal true keeps deprecated",
			query: `
				query {
					__type(name: "Query") {
						fields(includeDeprecated: true) { name }
					}
				}
			`,
			fieldName: "fields", enclosingTypeName: "__Type",
			expectDeprecatedSkipped: false, expectNonDeprecatedSkipped: false,
		},
		{
			name: "fields literal false skips deprecated",
			query: `
				query {
					__type(name: "Query") {
						fields(includeDeprecated: false) { name }
					}
				}
			`,
			fieldName: "fields", enclosingTypeName: "__Type",
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "fields literal null uses default skip",
			query: `
				query {
					__type(name: "Query") {
						fields(includeDeprecated: null) { name }
					}
				}
			`,
			fieldName: "fields", enclosingTypeName: "__Type",
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "fields variable true keeps deprecated",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						fields(includeDeprecated: $includeDeprecated) { name }
					}
				}
			`,
			fieldName: "fields", enclosingTypeName: "__Type", variables: `{"includeDeprecated":true}`,
			expectDeprecatedSkipped: false, expectNonDeprecatedSkipped: false,
		},
		{
			name: "fields variable false skips deprecated",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						fields(includeDeprecated: $includeDeprecated) { name }
					}
				}
			`,
			fieldName: "fields", enclosingTypeName: "__Type", variables: `{"includeDeprecated":false}`,
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "fields variable null behaves as false",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						fields(includeDeprecated: $includeDeprecated) { name }
					}
				}
			`,
			fieldName: "fields", enclosingTypeName: "__Type", variables: `{"includeDeprecated":null}`,
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "fields variable omitted behaves as false",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						fields(includeDeprecated: $includeDeprecated) { name }
					}
				}
			`,
			fieldName: "fields", enclosingTypeName: "__Type", variables: `{}`,
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "fields without argument skips deprecated by default",
			query: `
				query {
					__type(name: "Query") {
						fields { name }
					}
				}
			`,
			fieldName: "fields", enclosingTypeName: "__Type",
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "args without argument includes deprecated by default",
			query: `
				query {
					__type(name: "Query") {
						fields(includeDeprecated: true) {
							args { name }
						}
					}
				}
			`,
			fieldName: "args", enclosingTypeName: "__Field",
			expectDeprecatedSkipped: false, expectNonDeprecatedSkipped: false,
		},
		{
			name: "args literal false skips deprecated",
			query: `
				query {
					__type(name: "Query") {
						fields(includeDeprecated: true) {
							args(includeDeprecated: false) { name }
						}
					}
				}
			`,
			fieldName: "args", enclosingTypeName: "__Field",
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
		{
			name: "args variable true keeps deprecated",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						fields(includeDeprecated: true) {
							args(includeDeprecated: $includeDeprecated) { name }
						}
					}
				}
			`,
			fieldName: "args", enclosingTypeName: "__Field", variables: `{"includeDeprecated":true}`,
			expectDeprecatedSkipped: false, expectNonDeprecatedSkipped: false,
		},
		{
			name: "args variable null behaves as false",
			query: `
				query($includeDeprecated: Boolean) {
					__type(name: "Query") {
						fields(includeDeprecated: true) {
							args(includeDeprecated: $includeDeprecated) { name }
						}
					}
				}
			`,
			fieldName: "args", enclosingTypeName: "__Field", variables: `{"includeDeprecated":null}`,
			expectDeprecatedSkipped: true, expectNonDeprecatedSkipped: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			operation := unsafeparser.ParseGraphqlDocumentString(tc.query)

			v := &Visitor{Operation: &operation}
			skip := v.resolveSkipArrayItem(findFieldRefByName(t, &operation, tc.fieldName), tc.fieldName, tc.enclosingTypeName)
			require.NotNil(t, skip)

			ctx := ctxWithVars(tc.variables)
			require.Equal(t, tc.expectDeprecatedSkipped, skip(ctx, deprecatedItem))
			require.Equal(t, tc.expectNonDeprecatedSkipped, skip(ctx, nonDeprecatedItem))
		})
	}
}

func findFieldRefByName(t *testing.T, operation *ast.Document, fieldName string) int {
	t.Helper()

	for ref := range operation.Fields {
		if operation.FieldNameString(ref) == fieldName {
			return ref
		}
	}

	t.Fatalf("field %q not found", fieldName)
	return -1
}
