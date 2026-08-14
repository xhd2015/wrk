# Scenario

**Feature**: bare `--bring` / `--bring` followed by a flag requires a value

```
# Varargs: ≥1 non-flag token required; tokens starting with - are not values
wrk --bring            -> non-zero; requires a value
wrk --bring --no-dep   -> non-zero; requires a value (flag is not a value)
```

## Steps

- Leaves set `req.InProcess = true` and `req.Args` to the bare / next-is-flag forms.
- No dep fixtures (parse fails before bring apply).
