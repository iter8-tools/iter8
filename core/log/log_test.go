package log

func ExampleStackTrace() {
	Logger.WithStackTrace("\nmy \nnew \nmulti \nline \ntrace").Error("experiment did not run")
}
