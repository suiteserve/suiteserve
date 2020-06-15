# Hello
**Request**
```
{
  "cmd" string: "hello"
  "seq" int,
  "version" string: "1"
}
```

# Entry
**Request**
```
{
  "cmd" string: "create_suite",
  "seq" int,
  "name" string,
  "failure_types" []{
    "name"        string,
    "description" string,
  },
  "tags" []string,
  "env_vars" []{
    "key"   string,
    "value" string,
  },
  "planned_cases" int,
  "started_at" int,
}
```
