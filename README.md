# vt100

The source of the vt100 site: calm architecture, and the small tools I use to run my hobby projects calmly.

It's one Go binary. The pages, the stylesheet, the fonts and the images are embedded in it, and it loads nothing from anywhere else.

```bash
mise run dev    # http://localhost:8080
mise run ci     # what Punchcard runs: govulncheck, tests, build, vet
```

`PORT` changes the port. `mise run build` makes a static Linux binary in `bin/vt100`.

## What's where

- `tools.go`: the tools and the text on their pages
- `templates/`: the layout, the front page, a tool page and the 404 page
- `static/site.css`: the [Nord](https://www.nordtheme.com) palette and the layout
- `static/fonts/`: [Inter](https://rsms.me/inter/) and [JetBrains Mono](https://www.jetbrains.com/lp/mono/), both under the SIL Open Font License (licenses alongside)

## License

MIT for the code. The fonts keep their own license.
