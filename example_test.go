package html2text_test

import (
	"fmt"

	. "github.com/cention-sany/html2text"
)

func ExampleFromString() {
	inputHtml := `
	  <html>
		<head>
		  <title>My Mega Service</title>
		  <link rel=\"stylesheet\" href=\"main.css\">
		  <style type=\"text/css\">body { color: #fff; }</style>
		</head>

		<body>
		  <div class="logo">
			<a href="http://mymegaservice.com/"><img src="/logo-image.jpg" alt="Mega Service"/></a>
		  </div>

		  <h1>Welcome to your new account on my service!</h1>

		  <p>
			  Here is some more information:

			  <ul>
				  <li>Link 1: <a href="https://example.com">Example.com</a></li>
				  <li>Link 2: <a href="https://example2.com">Example2.com</a></li>
				  <li>Something else</li>
			  </ul>
		  </p>
		</body>
	  </html>
	`

	text, err := FromString(inputHtml)
	if err != nil {
		panic(err)
	}
	fmt.Println(text)

	// Output:
	// Mega Service ( http://mymegaservice.com/ )
	//
	// ******************************************
	// Welcome to your new account on my service!
	// ******************************************
	//
	// Here is some more information:
	//
	// * Link 1: Example.com ( https://example.com )
	// * Link 2: Example2.com ( https://example2.com )
	// * Something else
}
