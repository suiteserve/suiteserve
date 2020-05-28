package seed

import (
	"github.com/tmazeika/testpass/repo"
)

var logEntries = []repo.UnsavedLogEntry{
	{
		Level:   "info",
	},
	{
		Level:   "info",
		Message: "Hello, world!",
	},
	{
		Level: "info",
		Message: `Cras a lorem nec erat ullamcorper vulputate vestibulum feugiat
 leo. Sed non enim dictum, feugiat sapien nec, lacinia risus. Vivamus eget
 libero elementum, sagittis quam id, volutpat purus.

Mauris non turpis porttitor, faucibus augue et, tristique nisi.`,
	},
	{
		Level: "info",
		Message: `Sed tempus, nisi sed luctus porta, dui nunc imperdiet risus,
 sed commodo elit ex in ex. Donec pellentesque tellus leo, aliquet ullamcorper
 sem ullamcorper id. Aliquam vel velit sit amet sem ultrices mattis vitae non
 erat.`,
	},
	{
		Level:   "debug",
		Message: "Aenean euismod ultrices placerat.",
	},
	{
		Level:   "debug",
		Message: "Sed commodo elit ex in ex",
	},
	{
		Level:   "warn",
		Message: "Mauris non turpis porttitor!",
	},
	{
		Level:   "warn",
		Message: "PRAESENT TINCIDUNT ORCI QUIS GRAVIDA FAUCIBUS!",
	},
	{
		Level: "error",
		Trace: `Exception in thread "main" java.lang.RuntimeException: A test exception
  at com.example.stacktrace.StackTraceExample.methodB(StackTraceExample.java:13)
  at com.example.stacktrace.StackTraceExample.methodA(StackTraceExample.java:9)
  at com.example.stacktrace.StackTraceExample.main(StackTraceExample.java:5)`,
	},
	{
		Level: "error",
		Trace: `Exception in thread "main" java.lang.RuntimeException:
  at com.example.stacktrace.StackTraceExample.methodB(StackTraceExample.java:13)
  at com.example.stacktrace.StackTraceExample.methodA(StackTraceExample.java:9)
  at com.example.stacktrace.StackTraceExample.main(StackTraceExample.java:5)`,
		Message: "A test exception",
	},
}
