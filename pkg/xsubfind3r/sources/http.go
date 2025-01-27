package sources

import (
	hqgohttp "go.source.hueristiq.com/http"
)

// init is a special Go function that runs automatically when the package is imported.
// It is used here to configure default HTTP headers for the `hqgohttp.DefaultClient`.
//
// Default Client Configuration:
// The `hqgohttp.DefaultClient` is a pre-configured HTTP client that is customized
// with default headers to standardize requests made by the `sources` package.
//
// Headers Set:
//
//   - "Accept": "*/*"
//     Specifies that the client accepts any type of content in the response.
//
//   - "Accept-Language": "en"
//     Indicates that the preferred language for the response is English.
//
//   - "Connection": "close"
//     Signals that the connection should not be kept alive after the response is received.
//     This is useful in reducing resource usage on both client and server sides.
//
// Example Usage:
// The configured `hqgohttp.DefaultClient` can be used for making HTTP requests
// without needing to manually set these headers each time:
//
//	response, err := hqgohttp.DefaultClient.Get("https://example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Response:", response)
//
// Implementation Notes:
//   - The `init` function is executed exactly once when the package is imported.
//   - This setup ensures all HTTP requests made using `hqgohttp.DefaultClient` include
//     consistent headers, improving usability and compliance with expected defaults.
func init() {
	hqgohttp.DefaultClient.Headers["Accept"] = "*/*"
	hqgohttp.DefaultClient.Headers["Accept-Language"] = "en"
	hqgohttp.DefaultClient.Headers["Connection"] = "close"
}
