# Gervigreindarfulltrúar í skólastarfi

Glærur (Typst) fyrir kennara sem hafa ekki notað gervigreindarfulltrúa. Um 30–40 mínútur.

- `glaerur.typ` — frumskrá
- `glaerur.pdf` — glærur til sýningar
- `glaerur-glosur.pdf` — sama, með glósum og tímaáætlun fyrir kynni

```bash
typst compile glaerur.typ
typst compile glaerur.typ --input notes=true glaerur-glosur.pdf
```

Letur: Inter (annars Noto Sans / Liberation Sans / DejaVu Sans).
