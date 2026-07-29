// Package param defines reusable request payload parsers.
//
// A parser describes the JSON parameters an endpoint accepts, how null values
// behave, which caller labels may supply restricted parameters, and how the
// accepted values become the endpoint's final domain parameters.
//
// The package treats parsing as the request acceptance boundary. Parsing either
// returns a complete, domain-ready value or an Error explaining the first
// unacceptable parameter. There is no separate validation pass for endpoint
// authors to remember.
//
// Use Required and Optional to describe accepted JSON parameters. Each parameter
// has a wire shape, optional size or item bounds, a null policy, optional caller
// availability, and an optional parser. The parser is allowed to change the
// value type: for example, a string wire value can become a domain ID, and an
// array of wire structs can become domain line items.
//
// Use MutuallyExclusive or AtLeastOneOf when parameters should be declared
// together with their presence rule. The group can be passed to Param anywhere a
// single parameter can be passed. MutuallyExclusive is optional by default; call
// Required when exactly one of the grouped parameters must be present.
//
// Use Enum for string parameters with fixed allowed values. The enum check runs
// before custom parsers so endpoint code does not need to re-check membership,
// and Describe exposes the allowed values for transcription.
//
// Use Object to define inline nested objects. Use Array for arrays whose items
// can be decoded directly into a Go type. Use ArrayOf when each array item
// should be parsed by another Shape, such as an inline Object with its own
// parameters and rules.
//
// Use DiscriminatedObject for JSON objects where one string parameter selects
// the accepted object shape. Variant shapes should omit the discriminator
// parameter; it is parsed, validated, and stripped before the selected variant
// object parser runs.
//
// Common cleanup parsers such as TrimmedString, TrimmedStringList,
// NonEmptyTrimmedString, RFC3339Timestamp, OptionalRFC3339Timestamp, and
// OptionalRFC3339TimestampPointer are provided for the small bits of request
// cruft endpoints commonly want to normalize before calling domain code. Custom
// parsers can accept either the raw value or the parameter path plus raw value.
//
// Availability is intentionally generic: callers are values from the caller
// package, not service-specific concepts. A restricted parameter sent by an
// unavailable caller is reported as an unexpected parameter so the API does not
// reveal hidden parameter names or the caller labels allowed to use them.
//
// A typical endpoint parser is defined once at package initialization:
//
//	var (
//		Worker    = caller.Define("worker")
//		Dashboard = caller.Define("dashboard")
//		Admin     = caller.Define("admin")
//	)
//
//	var lineItem = param.DiscriminatedObject[lineItemParams]("type").
//		Variant("product",
//			param.Object[lineItemParams]().
//				Param(param.Required("product_id", param.String())).
//				Parse(parseProductLineItem),
//		).
//		Variant("fee",
//			param.Object[lineItemParams]().
//				Param(param.Required("amount", param.Int())).
//				Parse(parseFeeLineItem),
//		)
//
//	var initiateOrder = param.JSON[createOrderParams]().
//		Param(param.Required("line_items", param.ArrayOf(lineItem)).
//			Null(param.NullRejected).
//			MinItems(1)).
//		Param(param.Required("mode", param.Enum("payment", "subscription"))).
//		Param(param.MutuallyExclusive(
//			param.Optional("customer_id", param.String()).
//				Parse(parseCustomerID),
//			param.Optional("customer_data",
//				param.Object[customerDataParams]().
//					Param(param.Required("name", param.String())).
//					Param(param.AtLeastOneOf(
//						param.Optional("email_address", param.String()),
//						param.Optional("phone_number", param.String()),
//					)).
//					Parse(parseCustomerData),
//			),
//		).Required()).
//		Param(param.Optional("created_from",
//			param.Object[createdFromParams]().
//				Param(param.Optional("source", param.String())).
//				Param(param.Optional("resource_type", param.String())).
//				Param(param.Optional("resource_id", param.String())).
//				Parse(parseCreatedFrom),
//		).AvailableTo(Worker, Dashboard, Admin)).
//		Parse(parseInitiateOrder)
//
// Runtime usage stays direct:
//
//	params, err := initiateOrder.Parse(r)
//	if err != nil {
//		// Convert *param.Error into your service's public error response.
//		return
//	}
package param
