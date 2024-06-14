# `touchlog` Template DSL

The `touchlog` template DSL is simple direct language that leads to replicatable templated outputs.

## Example File

Standard Template:

```plaintext
> month: %m
> day: %d
> year: %y

%li:|> events

%li:|> emotions

%li:|> things to remember

```

You can find this file stored in `$TOUCHLOGHOME/templates/standard.logplate` where `$TOUCHLOGHOME` is where `touchlog` is installed.

Example Output:

```plaintext
> month: 06
> day: 14
> year: 2024

|> events
- 

|> emotions
- 

| things to remember
- 

```

## Language Specification

TODO Define language specification
