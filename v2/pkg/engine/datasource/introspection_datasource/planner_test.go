package introspection_datasource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wundergraph/astjson"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/asttransform"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/datasourcetesting"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/plan"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/internal/unsafeparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"
)

const (
	schema = `
		type Query {
			friend: String
		}
	`

	schemaWithCustomRootOperationTypes = `
		schema {
			query: CustomQuery
			mutation: CustomMutation
			subscription: CustomSubscription
		}

		type CustomQuery {
			friend: String
		}

		type CustomMutation {
			addFriend: Boolean
		}

		type CustomSubscription {
			lastAddedFriend: String
		}
	`

	typeIntrospection = `
		query typeIntrospection {
			__type(name: "Query") {
				name
				kind
			}
		}
	`

	schemaIntrospection = `
		query typeIntrospection {
			__schema {
				queryType {
					name
				}
			}
		}
	`

	schemaIntrospectionForAllRootOperationTypeNames = `
		query typeIntrospection {
			__schema {
				queryType {
					name
				}
				mutationType {
					name
				}
				subscriptionType {
					name
				}
			}
		}
	`

	typeIntrospectionWithArgs = `
		query typeIntrospection {
			__type(name: "Query") {
				fields(includeDeprecated: true) {
					name
				}
				enumValues(includeDeprecated: true) {
					name
				}
			}
		}
	`
)

func TestIntrospectionDataSourcePlanning(t *testing.T) {
	runTest := func(schema string, introspectionQuery string, expectedPlan plan.Plan) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			def := unsafeparser.ParseGraphqlDocumentString(schema)
			err := asttransform.MergeDefinitionWithBaseSchema(&def)
			require.NoError(t, err)

			var (
				introspectionData introspection.Data
				report            operationreport.Report
			)

			gen := introspection.NewGenerator()
			gen.Generate(&def, &report, &introspectionData)
			require.False(t, report.HasErrors())

			cfgFactory := IntrospectionConfigFactory{introspectionData: &introspectionData}
			introspectionDataSources := cfgFactory.BuildDataSourceConfigurations()

			planConfiguration := plan.Configuration{
				DataSources: introspectionDataSources,
				Fields:      cfgFactory.BuildFieldConfigurations(),
			}

			datasourcetesting.RunTest(schema, introspectionQuery, "", expectedPlan, planConfiguration)(t)
		}
	}

	dataSourceIdentifier := []byte("introspection_datasource.Source")

	t.Run("type introspection request", runTest(schema, typeIntrospection,
		&plan.SynchronousResponsePlan{
			Response: &resolve.GraphQLResponse{
				RawFetches: []*resolve.FetchItem{
					{
						Fetch: &resolve.SingleFetch{
							DataSourceIdentifier: dataSourceIdentifier,
							FetchConfiguration: resolve.FetchConfiguration{
								Input:      `{"request_type":2,"type_name":"$$0$$"}`,
								DataSource: &Source{},
								Variables: resolve.NewVariables(
									&resolve.ContextVariable{
										Path:     []string{"a"},
										Renderer: resolve.NewPlainVariableRenderer(),
									},
								),
								PostProcessing: resolve.PostProcessingConfiguration{
									MergePath: []string{"__type"},
								},
							},
						},
					},
				},
				Data: &resolve.Object{
					Fields: []*resolve.Field{
						{
							Name: []byte("__type"),
							Position: resolve.Position{
								Line:   3,
								Column: 4,
							},
							Value: &resolve.Object{
								Path:     []string{"__type"},
								Nullable: true,
								PossibleTypes: map[string]struct{}{
									"__Type": {},
								},
								TypeName: "__Type",
								Fields: []*resolve.Field{
									{
										Name: []byte("name"),
										Value: &resolve.String{
											Path:     []string{"name"},
											Nullable: true,
										},
										Position: resolve.Position{
											Line:   4,
											Column: 5,
										},
									},
									{
										Name: []byte("kind"),
										Value: &resolve.Enum{
											TypeName: "__TypeKind",
											Path:     []string{"kind"},
											Values: []string{
												"SCALAR",
												"OBJECT",
												"INTERFACE",
												"UNION",
												"ENUM",
												"INPUT_OBJECT",
												"LIST",
												"NON_NULL",
											},
											InaccessibleValues: []string{},
										},
										Position: resolve.Position{
											Line:   5,
											Column: 5,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	))

	t.Run("schema introspection request", runTest(schema, schemaIntrospection,
		&plan.SynchronousResponsePlan{
			Response: &resolve.GraphQLResponse{
				RawFetches: []*resolve.FetchItem{
					{
						Fetch: &resolve.SingleFetch{
							DataSourceIdentifier: dataSourceIdentifier,
							FetchConfiguration: resolve.FetchConfiguration{
								Input:      `{"request_type":1}`,
								DataSource: &Source{},
								PostProcessing: resolve.PostProcessingConfiguration{
									MergePath: []string{"__schema"},
								},
							},
						},
					},
				},
				Data: &resolve.Object{
					Fields: []*resolve.Field{
						{
							Name: []byte("__schema"),
							Position: resolve.Position{
								Line:   3,
								Column: 4,
							},
							Value: &resolve.Object{
								Path: []string{"__schema"},
								PossibleTypes: map[string]struct{}{
									"__Schema": {},
								},
								TypeName: "__Schema",
								Fields: []*resolve.Field{
									{
										Name: []byte("queryType"),
										Value: &resolve.Object{
											PossibleTypes: map[string]struct{}{
												"__Type": {},
											},
											TypeName: "__Type",
											Path:     []string{"queryType"},
											Fields: []*resolve.Field{
												{
													Name: []byte("name"),
													Value: &resolve.String{
														Path:     []string{"name"},
														Nullable: true,
													},
													Position: resolve.Position{
														Line:   5,
														Column: 6,
													},
												},
											},
										},
										Position: resolve.Position{
											Line:   4,
											Column: 5,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	))

	t.Run("schema introspection request with custom root operation types", runTest(schemaWithCustomRootOperationTypes, schemaIntrospectionForAllRootOperationTypeNames,
		&plan.SynchronousResponsePlan{
			Response: &resolve.GraphQLResponse{
				RawFetches: []*resolve.FetchItem{
					{
						Fetch: &resolve.SingleFetch{
							DataSourceIdentifier: dataSourceIdentifier,
							FetchConfiguration: resolve.FetchConfiguration{
								Input:      `{"request_type":1}`,
								DataSource: &Source{},
								PostProcessing: resolve.PostProcessingConfiguration{
									MergePath: []string{"__schema"},
								},
							},
						},
					},
				},
				Data: &resolve.Object{
					Fields: []*resolve.Field{
						{
							Name: []byte("__schema"),
							Position: resolve.Position{
								Line:   3,
								Column: 4,
							},
							Value: &resolve.Object{
								Path: []string{"__schema"},
								PossibleTypes: map[string]struct{}{
									"__Schema": {},
								},
								TypeName: "__Schema",
								Fields: []*resolve.Field{
									{
										Name: []byte("queryType"),
										Value: &resolve.Object{
											Path: []string{"queryType"},
											PossibleTypes: map[string]struct{}{
												"__Type": {},
											},
											TypeName: "__Type",
											Fields: []*resolve.Field{
												{
													Name: []byte("name"),
													Value: &resolve.String{
														Path:     []string{"name"},
														Nullable: true,
													},
													Position: resolve.Position{
														Line:   5,
														Column: 6,
													},
												},
											},
										},
										Position: resolve.Position{
											Line:   4,
											Column: 5,
										},
									},
									{
										Name: []byte("mutationType"),
										Value: &resolve.Object{
											Path:     []string{"mutationType"},
											Nullable: true,
											PossibleTypes: map[string]struct{}{
												"__Type": {},
											},
											TypeName: "__Type",
											Fields: []*resolve.Field{
												{
													Name: []byte("name"),
													Value: &resolve.String{
														Path:     []string{"name"},
														Nullable: true,
													},
													Position: resolve.Position{
														Line:   8,
														Column: 6,
													},
												},
											},
										},
										Position: resolve.Position{
											Line:   7,
											Column: 5,
										},
									},
									{
										Name: []byte("subscriptionType"),
										Value: &resolve.Object{
											Path:     []string{"subscriptionType"},
											Nullable: true,
											PossibleTypes: map[string]struct{}{
												"__Type": {},
											},
											TypeName: "__Type",
											Fields: []*resolve.Field{
												{
													Name: []byte("name"),
													Value: &resolve.String{
														Path:     []string{"name"},
														Nullable: true,
													},
													Position: resolve.Position{
														Line:   11,
														Column: 6,
													},
												},
											},
										},
										Position: resolve.Position{
											Line:   10,
											Column: 5,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	))

	t.Run("type introspection request with fields args", runTest(schema, typeIntrospectionWithArgs,
		&plan.SynchronousResponsePlan{
			Response: &resolve.GraphQLResponse{
				RawFetches: []*resolve.FetchItem{
					{
						Fetch: &resolve.SingleFetch{
							DataSourceIdentifier: dataSourceIdentifier,
							FetchConfiguration: resolve.FetchConfiguration{
								Input:      `{"request_type":2,"type_name":"$$0$$"}`,
								DataSource: &Source{},
								Variables: resolve.NewVariables(
									&resolve.ContextVariable{
										Path:     []string{"a"},
										Renderer: resolve.NewPlainVariableRenderer(),
									},
								),
								PostProcessing: resolve.PostProcessingConfiguration{
									MergePath: []string{"__type"},
								},
							},
						},
					},
				},
				Data: &resolve.Object{
					Fields: []*resolve.Field{
						{
							Name: []byte("__type"),
							Position: resolve.Position{
								Line:   3,
								Column: 4,
							},
							Value: &resolve.Object{
								Path:     []string{"__type"},
								Nullable: true,
								PossibleTypes: map[string]struct{}{
									"__Type": {},
								},
								TypeName: "__Type",
								Fields: []*resolve.Field{
									{
										Name: []byte("fields"),
										Value: &resolve.Array{
											Path:     []string{"fields"},
											Nullable: true,
											Item: &resolve.Object{
												PossibleTypes: map[string]struct{}{
													"__Field": {},
												},
												TypeName: "__Field",
												Fields: []*resolve.Field{
													{
														Name: []byte("name"),
														Value: &resolve.String{
															Path: []string{"name"},
														},
														Position: resolve.Position{
															Line:   5,
															Column: 6,
														},
													},
												},
											},
											SkipItem: func(ctx *resolve.Context, arrayItem *astjson.Value) bool {
												return false
											},
										}, Position: resolve.Position{
											Line:   4,
											Column: 5,
										},
									},
									{
										Name: []byte("enumValues"),
										Value: &resolve.Array{
											Path:     []string{"enumValues"},
											Nullable: true,
											Item: &resolve.Object{
												PossibleTypes: map[string]struct{}{
													"__EnumValue": {},
												},
												TypeName: "__EnumValue",
												Fields: []*resolve.Field{
													{
														Name: []byte("name"),
														Value: &resolve.String{
															Path: []string{"name"},
														},
														Position: resolve.Position{
															Line:   8,
															Column: 6,
														},
													},
												},
											},
											SkipItem: func(ctx *resolve.Context, arrayItem *astjson.Value) bool {
												return false
											},
										}, Position: resolve.Position{
											Line:   7,
											Column: 5,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	))
}

func TestIntrospectionIncludeDeprecatedExtractedLiteralTrue(t *testing.T) {
	const testSchema = `
		type Query {
			active: String
			legacy: String @deprecated(reason: "old")
		}
	`

	const introspectionQuery = `
		query IntrospectionQuery {
			__type(name: "Query") {
				fields(includeDeprecated: true) {
					name
				}
			}
		}
	`

	planned := buildIntrospectionPlan(t, testSchema, introspectionQuery, true)
	skipItem := skipItemFromTypeField(t, planned, "fields")

	deprecatedField := astjson.MustParseBytes([]byte(`{"name":"legacy","isDeprecated":true}`))

	ctx := resolve.NewContext(context.Background())
	ctx.Variables = astjson.MustParseBytes([]byte(`{}`))
	// With ExtractVariables enabled, literal true is turned into an operation variable.
	// We still must include deprecated fields even when runtime ctx.Variables is empty.
	require.False(t, skipItem(ctx, deprecatedField))
}

func TestIntrospectionInputFieldsAreNotFilteredByIncludeDeprecated(t *testing.T) {
	const testSchema = `
		type Query {
			friend(filter: TestAllFieldsDeprecatedFilterInput): String
		}

		input TestAllFieldsDeprecatedFilterInput {
			foo: String @deprecated(reason: "old")
		}
	`

	const introspectionQuery = `
		query IntrospectionQuery {
			__type(name: "TestAllFieldsDeprecatedFilterInput") {
				inputFields {
					name
				}
			}
		}
	`

	planned := buildIntrospectionPlan(t, testSchema, introspectionQuery, false)
	skipItem := skipItemFromTypeField(t, planned, "inputFields")
	require.Nil(t, skipItem)
}

func TestIntrospectionInputFieldsMixedDeprecatedAreNotFilteredByIncludeDeprecated(t *testing.T) {
	const testSchema = `
		type Query {
			friend(filter: TestAllFieldsDeprecatedFilterInput): String
		}

		input TestAllFieldsDeprecatedFilterInput {
			foo: String @deprecated(reason: "old")
			bar: String
		}
	`

	const introspectionQuery = `
		query IntrospectionQuery {
			__type(name: "TestAllFieldsDeprecatedFilterInput") {
				inputFields(includeDeprecated: false) {
					name
				}
			}
		}
	`

	planned := buildIntrospectionPlan(t, testSchema, introspectionQuery, false)
	skipItem := skipItemFromTypeField(t, planned, "inputFields")
	require.Nil(t, skipItem)
}

func buildIntrospectionPlan(t *testing.T, schema, introspectionQuery string, withExtractVariables bool) *plan.SynchronousResponsePlan {
	t.Helper()

	def := unsafeparser.ParseGraphqlDocumentString(schema)
	err := asttransform.MergeDefinitionWithBaseSchema(&def)
	require.NoError(t, err)

	var (
		introspectionData introspection.Data
		report            operationreport.Report
	)

	gen := introspection.NewGenerator()
	gen.Generate(&def, &report, &introspectionData)
	require.False(t, report.HasErrors())

	cfgFactory := IntrospectionConfigFactory{introspectionData: &introspectionData}
	planConfiguration := plan.Configuration{
		DataSources: introspectionDataSources(cfgFactory),
		Fields:      cfgFactory.BuildFieldConfigurations(),
	}

	op := unsafeparser.ParseGraphqlDocumentString(introspectionQuery)
	normalizationOptions := []astnormalization.Option{
		astnormalization.WithInlineFragmentSpreads(),
		astnormalization.WithRemoveFragmentDefinitions(),
		astnormalization.WithRemoveUnusedVariables(),
	}
	if withExtractVariables {
		normalizationOptions = append(normalizationOptions, astnormalization.WithExtractVariables())
	}

	norm := astnormalization.NewWithOpts(normalizationOptions...)
	report = operationreport.Report{}
	norm.NormalizeOperation(&op, &def, &report)
	require.False(t, report.HasErrors(), report.Error())

	planner, err := plan.NewPlanner(planConfiguration)
	require.NoError(t, err)

	actualPlan := planner.Plan(&op, &def, "", &report)
	require.False(t, report.HasErrors(), report.Error())

	syncPlan, ok := actualPlan.(*plan.SynchronousResponsePlan)
	require.True(t, ok)
	return syncPlan
}

func introspectionDataSources(cfgFactory IntrospectionConfigFactory) []plan.DataSource {
	return cfgFactory.BuildDataSourceConfigurations()
}

func skipItemFromTypeField(t *testing.T, syncPlan *plan.SynchronousResponsePlan, fieldName string) resolve.SkipArrayItem {
	t.Helper()

	rootTypeField := findFieldByName(t, syncPlan.Response.Data, "__type")
	typeObject, ok := rootTypeField.Value.(*resolve.Object)
	require.True(t, ok)

	introspectionField := findFieldByName(t, typeObject, fieldName)
	arrayValue, ok := introspectionField.Value.(*resolve.Array)
	require.True(t, ok)

	return arrayValue.SkipItem
}

func findFieldByName(t *testing.T, object *resolve.Object, name string) *resolve.Field {
	t.Helper()

	for _, field := range object.Fields {
		if string(field.Name) == name {
			return field
		}
	}

	require.FailNowf(t, "field not found", "field %q not found", name)
	return nil
}
