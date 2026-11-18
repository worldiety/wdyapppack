# wdyapppack

Packages prebuilt programs — Go, Rust, Zig, C — as macOS `.app` bundles, and
signs, notarizes and wraps them for download.

```
go install github.com/worldiety/wdyapppack/cmd/wdyapppack@latest
wdyapppack
```

The program explains itself. Run it without arguments for the overview, and
`wdyapppack help <command|topic>` for everything else; `wdyapppack help agents`
is written for AI agents and scripts, and `wdyapppack help --format json`
describes the whole interface as data.

```
wdyapppack build --binary dist/mytool --name "My Tool" \
    --identifier com.example.mytool --version 1.0.0 --icon icon.png
wdyapppack sign "dist/My Tool.app" --identity "Developer ID Application: …"
wdyapppack notarize "dist/My Tool.app" --keychain-profile <profile>
wdyapppack dmg "dist/My Tool.app"
```

Building bundles, icons, universal programs and zip archives works on every
operating system; signing, notarization and disk images need macOS.

wdyapppack uses nothing but the Go standard library. Its requirements live in
[`requirements/`](requirements), traced to the code and the tests with
[speclink](https://github.com/worldiety/speclink); [`SPECIFICATION.md`](SPECIFICATION.md)
is derived from them. Working on the code: [`AGENTS.md`](AGENTS.md).

## License

BSD-2-Clause, see [LICENSE](LICENSE).