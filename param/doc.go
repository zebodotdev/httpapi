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
// Use Object to define inline nested objects. Use Array for arrays whose items
// can be decoded directly into a Go type. Use ArrayOf when each array item
// should be parsed by another Shape, such as an inline Object with its own
// parameters and rules.
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
//	var initiateOrder = param.JSON[createOrderParams]().
//		Param(param.Required("line_items", param.Array[lineItemParams]()).
//			Null(param.NullRejected).
//			MinItems(1).
//			Parse(parseLineItems)).
//		Param(param.Optional("customer_id", param.String()).
//			Parse(parseCustomerID)).
//		Param(param.Optional("customer_data",
//			param.Object[customerDataParams]().
//				Param(param.Required("name", param.String())).
//				Param(param.Optional("email_address", param.String())).
//				Param(param.Optional("phone_number", param.String())).
//				AtLeastOne("email_address", "phone_number").
//				Parse(parseCustomerData),
//		)).
//		Param(param.Optional("created_from",
//			param.Object[createdFromParams]().
//				Param(param.Optional("source", param.String())).
//				Param(param.Optional("resource_type", param.String())).
//				Param(param.Optional("resource_id", param.String())).
//				Parse(parseCreatedFrom),
//		).AvailableTo(Worker, Dashboard, Admin)).
//		ExactlyOne("customer_id", "customer_data").
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
