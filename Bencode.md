# Bencode Format

Bencoding is used to encode data types in a compact way. It supports four main types: strings, integers, lists, and dictionaries.

## String

Format: `<length>:<string>`

**Example:** The string `jyotishmoy` encodes to:

```
10:jyotishmoy
```

## Integer

Format: `i<integer>e`

**Example:** The integer `10` encodes to:

```
i10e
```

## List

Format: `l<bencoded value>...e`

**Example:** The list `["a", "b", 1]` encodes to:

```
l1:a1:bi1ee
```

## Dictionary

Format: `d<bencoded string><bencoded value>...e`

**Example:** The dictionary `{"a": 1, "b": 2}` encodes to:

```
d1:ai1e1:bi2ee
```
