# ttgenlib


A framework to simplify building table test generators for functions and methods, with:

* Customizable `context.Context` treatment.
* Fields and function/method parameters are getting nice helpers to use mock objects for them.

You only need to define your messages renderer (result err processing) and provide mock lookup.
The standard lookup function will probably be sufficient for your needs at that.

See [example](internal/cmd/example/example.go) for implementation details. 

