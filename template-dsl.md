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

| Command | Expected Input (if any) | Output |
| :---- | :---------------------- | :----- |
| `%m` | N/A | current month - ex: `06` |
| `%d` | N/A | current day - ex: `14` |
| `%y` | N/A | current year - ex: `2024` |
| `%uli:<X>` | `<X>` is the title of an unordered list of items | `<X>` followed by an empty unordered list |
| `%nli:<X>` | `<X>` is the title of numbered list of items | `<X>` followed by an empty numbered list |
| `%br` | N/A | horizontal line break (`-----`) |
